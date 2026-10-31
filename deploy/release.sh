#!/usr/bin/env bash
# release.sh ─ server 쪽 배포: 새 실행 파일을 넣고 backup · migrate · 교체 · 확인까지 한 번에 (ARCHITECTURE 7.3절)
# PC의 task deploy가 이 파일과 실행 파일을 server 임시 폴더에 copy한 뒤 부른다:
# sudo bash /tmp/cbt-deploy-…/release.sh <COMMIT>
# 비유 ─ 가게 간판 바꿔 달기: 새 간판을 창고에 들이고(release), 장부 사본을 떠 두고(backup), 진열대를 새 규격에 맞추고(migrate),
#       간판을 바꿔 단 뒤(link) 손님 응대가 되는지 확인한다 → 실패하면 옛 간판을 다시 단다

set -euo pipefail

RELEASES=/opt/cbt/releases
CURRENT=/opt/cbt/current
ENV_FILE=/etc/cbt/cbt.env
HEALTH_URL=http://127.0.0.1:8080/api/health
KEEP=5

# PC가 보낸 실행 파일은 이 script와 같은 폴더에 있다
SRC_DIR=$(dirname "$(readlink -f "$0")")
NEW_BIN=$SRC_DIR/app-linux-amd64

log() { echo "[release] $*"; }
die() {
	echo "[release] 실패: $*" >&2
	exit 1
}

# 운영 env를 읽고 cbt 계정으로 실행한다 (cbtctl과 같은 방식, 실행 파일만 고를 수 있다)
# ( … ) 안에서 돌려서 비밀값이 든 env가 이 script 쪽으로 새지 않게 한다
as_cbt() {
	(
		set -a
		# shellcheck source=/dev/null
		source "$ENV_FILE"
		set +a
		exec setpriv --reuid=cbt --regid=cbt --init-groups "$@"
	)
}

# current 링크를 바꿔 단다. 새 링크를 옆에 만들고 mv로 덮으면 "링크가 없는 순간"이 생기지 않는다.
# (ln -sfn은 지우고 → 만드는 두 동작이라 그 사이에 재시작되면 앱 파일이 없다)
switch_to() {
	ln -sfn "$1" "$CURRENT.new"
	mv -T "$CURRENT.new" "$CURRENT"
}

# /api/health가 200을 줄 때까지 1초 간격으로 30번 본다. 앱은 DB · Valkey가 안 되면 503을 준다.
healthy() {
	local try
	for try in $(seq 1 30); do
		if curl -fs -o /dev/null --max-time 2 "$HEALTH_URL"; then
			log "헬스체크 통과 (${try}번째)"
			return 0
		fi
		sleep 1
	done
	return 1
}

# ── 0. 확인 ──
[[ $EUID -eq 0 ]] || die "sudo로 실행해야 한다"
[[ $# -eq 1 ]] || die "사용법: release.sh <커밋>"
COMMIT=$1
# 커밋 번호 모양(16진수 12~40자리)만 받는다. -dirty 빌드나 ../ 같은 글자는 여기서 막힌다.
[[ $COMMIT =~ ^[0-9a-f]{12,40}$ ]] || die "커밋 번호 모양이 아니다: $COMMIT"
[[ -f $NEW_BIN ]] || die "실행 파일이 없다: $NEW_BIN"

# 두 PC가 동시에 배포하지 않게 잠근다. 9번 손잡이로 잠금 파일을 열고, 이미 잠겨 있으면 바로 실패한다.
# 비유: 탈의실 문고리 ─ 누가 쓰는 중이면 기다리지 않고 돌아선다. 스크립트가 끝나면 저절로 풀린다.
exec 9>/run/lock/cbt-deploy.lock
flock -n 9 || die "다른 배포가 진행 중이다"

# 되돌릴 곳: 지금 current가 가리키는 폴더. 첫 배포면 비어 있다.
PREV=$(readlink "$CURRENT" || true)
TARGET=$RELEASES/$COMMIT

# ── 1. releases/<커밋>/에 넣기 ──
# app.new로 넣고 버전을 확인한 뒤에야 app으로 바꾼다 → 엉뚱한 파일이 기존 release를 덮지 않는다.
install -d -m 755 "$TARGET"
install -m 755 "$NEW_BIN" "$TARGET/app.new"
GOT=$(as_cbt "$TARGET/app.new" version) || die "실행 파일이 안 돈다 (Linux용 빌드인가?)"
if [[ $GOT != "$COMMIT" ]]; then
	rm -f "$TARGET/app.new"
	die "실행 파일 버전($GOT)이 커밋($COMMIT)과 다르다"
fi
mv -f "$TARGET/app.new" "$TARGET/app"
log "넣음: $TARGET"

# ── 2. 배포 직전 백업 (7.4절) ──
if command -v cbt-backup >/dev/null; then
	log "배포 직전 백업"
	cbt-backup || die "백업 실패 ─ 배포를 멈춘다 (아직 아무것도 안 바꿨다)"
elif [[ -z $PREV ]]; then
	log "백업 스크립트가 아직 없다 ─ 첫 배포라 건너뛴다 (DB가 비어 있다)"
else
	die "cbt-backup이 없다 ─ 백업 없이 배포하지 않는다"
fi

# ── 3. 마이그레이션 (새 실행 파일로) ──
log "migrate up"
as_cbt "$TARGET/app" migrate up || die "migrate 실패 ─ 링크를 안 바꿨으니 옛 앱이 그대로 돈다"

# ── 4. 교체 · 재시작 ──
switch_to "$TARGET"
# enable: 재부팅 뒤에도 저절로 뜨게 (이미 켜져 있으면 아무것도 안 한다, 7.5절)
systemctl enable --quiet cbt
systemctl restart cbt
log "재시작: $COMMIT"

# ── 5. 확인 · 실패하면 되돌리기 ──
if ! healthy; then
	if [[ -n $PREV && $PREV != "$TARGET" ]]; then
		log "헬스체크 실패 ─ 옛 버전으로 되돌린다: $PREV"
		switch_to "$PREV"
		systemctl restart cbt
		if healthy; then
			log "옛 버전으로 돌아옴"
		else
			log "옛 버전도 응답이 없다"
		fi
	fi
	die "새 버전이 30초 안에 정상 응답을 안 했다 ─ 로그: journalctl -u cbt -n 50"
fi

# ── 6. 정리: 최근 것부터 KEEP개만 남긴다 (지금 쓰는 것은 절대 안 지운다) ──
# 폴더를 고친 시각(%T@) 순으로 줄 세우고, KEEP개 뒤의 것만 지운다.
find "$RELEASES" -mindepth 1 -maxdepth 1 -type d -printf '%T@ %p\n' |
	sort -rn | tail -n +$((KEEP + 1)) | cut -d' ' -f2- |
	while read -r old; do
		[[ $old == "$TARGET" ]] && continue
		log "정리: $old"
		rm -rf -- "$old"
	done

log "끝: $COMMIT"
