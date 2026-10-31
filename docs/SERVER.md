# 운영 서버

> **서버에 어떻게 들어가고, 서버가 어떻게 세팅돼 있고, 무엇을 어떤 명령으로 했는지** 적는 문서다(런북).
> 왜 이 서버인지는 [decisions/0007](decisions/0007-deploy-vultr-seoul.md), 앞으로 할 일은 [PLAN.md](PLAN.md) M3, 겪은 일은 [DEVLOG.md](DEVLOG.md).
> **비밀값(비밀번호 · 키 · 서버 IP)은 이 문서에 적지 않는다.** 서버 IP는 Vultr 콘솔 → Compute → Instances에서 본다. 아래 `<서버 IP>`는 그 값으로 바꿔 친다.

## 1. 한눈에

| 항목 | 값 |
|---|---|
| 업체 · 위치 | Vultr · 서울 |
| 서버 이름 | `cbt-1` |
| 사양 | Shared CPU `vc2-1c-1gb` — 1 vCPU · 메모리 1GB · SSD 25GB · 월 $5 |
| OS | Ubuntu 24.04 LTS (보안 업데이트 2029-04까지) |
| 주소 | IPv4만 (IPv6 끔 — 이유는 DEVLOG 2026-10-30 "[설계] 정하면서 바꾼 것") |
| 접속 계정 | `sagming40` (sudo 가능). **root로는 SSH 접속 불가** |
| 접속 방법 | SSH 키만 (비밀번호 로그인 불가) |
| 열린 문 | 22(SSH) · 80 · 443 — 나머지는 방화벽(ufw)이 막는다 |
| 자동 백업 | Vultr 자동 백업 켬(+$1/월) — 2026-11-30에 유지 여부 판단 |
| 도메인 | `cbt.sagming.com` (Cloudflare, DNS만 · 회색 구름) — M3에서 연결 |

## 2. 지금 어디서 치고 있나

명령을 치기 전에 **맨 앞의 프롬프트**를 본다. 서버 명령을 PC에서, PC 명령을 서버에서 치면 엉뚱한 일이 생긴다.

| 프롬프트 | 위치 |
|---|---|
| `PS C:\Users\Windows>` | **내 PC** (PowerShell) |
| `sagming40@cbt-1:~$` | **서버** (일반 계정 — 끝이 `$`) |
| `root@cbt-1:~#` | **서버** (관리자 — 끝이 `#`. 지금은 Vultr 웹 콘솔에서만) |

## 3. 접속하기

### 3.1 기본 (내 PC에서)
```powershell
ssh -i "$HOME\.ssh\cbt_vultr" sagming40@<서버 IP>
```
- `-i`: 쓸 열쇠(개인 키) 파일. passphrase를 물으면 키를 만들 때 정한 값을 친다(화면에 안 보이는 게 정상).
- 처음 접속하는 PC면 `Are you sure you want to continue connecting (yes/no/[fingerprint])?` → `yes`. PC가 서버 지문을 기억해 두고, 나중에 다른 서버가 끼어들면 경고한다.
- 나갈 때: 서버에서 `exit`

### 3.2 짧게 `ssh cbt`로 들어가기 (PC마다 한 번, **`task deploy`에 필요**)
PC의 `~/.ssh/config`에 별명을 적어 두면 주소 · 열쇠 · 계정을 매번 안 친다. 이 파일은 PC에만 있고 저장소에는 안 들어간다. `task deploy`는 이 별명 `cbt`로 서버를 찾는다(저장소에 IP를 적지 않으려고).
```powershell
if (-not (Test-Path "$HOME\.ssh\config")) { New-Item -ItemType File "$HOME\.ssh\config" }
notepad "$HOME\.ssh\config"
```
파일을 먼저 만드는 이유: 없는 파일을 메모장으로 열어 저장하면 `config.txt`가 되고, ssh는 그 파일을 안 읽는다(`Could not resolve hostname cbt` — 노트북에서 겪음). 확인: `ssh -G cbt`에 `hostname`이 서버 IP로 나오면 별명이 읽힌 것.
메모장에 아래를 넣고 저장(`<서버 IP>`는 실제 값으로):
```text
Host cbt
    HostName <서버 IP>
    User sagming40
    IdentityFile ~/.ssh/cbt_vultr
```
그다음부터:
```powershell
ssh cbt
```
도메인(`cbt.sagming.com`)을 연결한 뒤에는 `HostName`을 도메인으로 바꿔도 된다.

### 3.3 비상구: Vultr 웹 콘솔
SSH가 막혔을 때(방화벽 · SSH 설정 실수) 쓰는 문이다.
- Vultr → Compute → Instances → `cbt-1` 이름을 눌러 상세 화면 → 오른쪽 위 모니터 모양 아이콘(**View Console**). 목록에서 그 줄 끝 `···` 메뉴에도 있다. 콘솔은 새 창(팝업)으로 열린다 — 브라우저가 막으면 주소창 오른쪽에서 허용
- 웹 화면 안에서 `root`와 **Vultr 화면에 표시된 root 비밀번호**로 로그인한다. root 금지는 SSH에만 걸려 있어서 콘솔은 된다.
- root 비밀번호는 이 비상구 말고는 쓰지 않는다. 어디에도 옮겨 적지 않는다.

### 3.4 새 PC를 추가할 때 (노트북 `laptop` 2026-10-31 웹 콘솔로 추가, 회사 PC는 아직)
PC마다 열쇠를 따로 만든다 — 한 PC를 잃으면 그 열쇠만 지우면 된다. **서버는 이미 만들어져 있어서 Vultr에 키를 등록해도 이 서버에는 안 들어간다.** 서버의 `authorized_keys`에 직접 한 줄 추가해야 한다.

1. **새 PC에서** 열쇠 만들기
   ```powershell
   New-Item -ItemType Directory -Force "$HOME\.ssh"
   ssh-keygen -t ed25519 -C "cbt-journal company-pc" -f "$HOME\.ssh\cbt_vultr"   # 노트북이면 laptop
   Get-Content "$HOME\.ssh\cbt_vultr.pub"
   ```
   마지막 줄(`ssh-ed25519 AAAA… cbt-journal company-pc`)을 복사한다. `.pub`가 아닌 파일(개인 키)은 열지도 옮기지도 않는다.
2. **이미 접속되는 PC로 서버에 들어가서** 그 한 줄을 추가
   ```bash
   echo 'ssh-ed25519 AAAA…(복사한 한 줄 전체)' >> ~/.ssh/authorized_keys
   ```
   `>>`(두 개)여야 한다. `>`(하나)면 기존 열쇠를 지워 버려 다른 PC가 못 들어온다.
   - **접속되는 PC가 없어 웹 콘솔(3.3절)로 넣을 때**: 콘솔은 `root`로 들어가서 `~`가 `/root`다. 경로를 `/home/sagming40/.ssh/authorized_keys`로 바꿔 친다(기존 파일에 덧붙이는 거라 주인은 그대로 `sagming40`). 확인: `tail -n 2 /home/sagming40/.ssh/authorized_keys`
   - 웹 콘솔에 붙여넣기가 안 되면 콘솔 화면의 클립보드 · 텍스트 보내기 버튼을 쓰거나, 집 PC로 넣는다. 열쇠 한 줄을 손으로 옮겨 치지 않는다(한 글자만 틀려도 안 맞는다).
3. 새 PC에서 3.1 명령으로 접속 확인.
4. 지금 등록된 열쇠 보기 (서버에서): `cat ~/.ssh/authorized_keys` — 줄 끝 메모(`home-pc` · `company-pc`)로 구분한다. PC를 잃으면 그 줄만 지운다.

### 3.5 실제로 친 순서: 노트북 추가 (2026-10-31)
**상황**: 노트북에서 처음 서버 작업을 하려는데 서버에 들어갈 수 있는 기기가 집 PC뿐이었다. 집 PC 없이 Vultr 웹 콘솔로 열쇠를 넣었다.
```powershell
# ── 노트북 PowerShell ──
New-Item -ItemType Directory -Force "$HOME\.ssh"     # 열쇠 보관 폴더 만들기 (이미 있으면 그대로)
ssh-keygen -t ed25519 -C "cbt-journal laptop" -f "$HOME\.ssh\cbt_vultr"
                                                     # 열쇠 한 쌍 만들기: cbt_vultr(개인 키, 노트북에만) + cbt_vultr.pub(공개 키, 서버에 줄 것)
                                                     # -t ed25519 짧고 안전한 종류 / -C 줄 끝 메모(어느 기기 열쇠인지) / passphrase = 열쇠를 쓸 때마다 묻는 암호
Get-Content "$HOME\.ssh\cbt_vultr.pub"               # 공개 키 한 줄을 화면에 띄워 복사
```
```bash
# ── Vultr 웹 콘솔 (root로 로그인, 3.3절) ──
echo 'ssh-ed25519 AAAA… cbt-journal laptop' >> /home/sagming40/.ssh/authorized_keys
                                                     # sagming40의 "들어와도 되는 열쇠 목록"에 한 줄 덧붙이기. root라서 ~가 아니라 전체 경로
tail -n 2 /home/sagming40/.ssh/authorized_keys       # 마지막 두 줄 보기 → home-pc 줄 + laptop 줄
exit                                                 # 콘솔에서 나가기
```
```powershell
# ── 노트북 PowerShell ──
notepad "$HOME\.ssh\config"                          # ssh 별명 파일 쓰기(3.2절) → 메모장이 config.txt로 저장해 버림
ssh cbt                                              # → Could not resolve hostname cbt (별명을 못 읽음)
Rename-Item "$HOME\.ssh\config.txt" config           # (Claude가 침) 이름 바로잡기
ssh -G cbt                                           # 접속하지 않고 "별명을 어떻게 읽었나"만 출력 → hostname · user · identityfile 확인
ssh cbt                                              # 첫 접속: 서버 지문 질문에 yes → passphrase → sagming40@cbt-1:~$
winget install --id FiloSottile.age -e               # 백업을 열 때 쓰는 age 설치 (4.13절)
age --version                                        # 새 터미널에서 설치 확인 → v1.3.2
```
- 나중에 `WARNING: REMOTE HOST IDENTIFICATION HAS CHANGED!`가 뜨면 접속을 이어 가지 않는다 — 처음 `yes`로 저장한 서버 지문과 다르다는 뜻이다.
- 어느 와이파이에서든 접속된다(22번은 어디서 오든 열려 있고 열쇠로만 들어온다). 22번을 막는 와이파이(회사 · 학교 등)면 `Connection timed out` → 휴대폰 핫스팟이나 웹 콘솔.

## 4. 지금까지 한 설정 (순서대로)

> 각 절은 **상황**(왜 · 어디서 쳤나) → 친 명령(줄마다 뜻 · 옵션을 어떻게 줬나 · 나온 결과) → 메모 순서다. 헛디딘 명령도 남긴다 — 같은 실수를 안 하려고.
> 4.1 ~ 4.6은 2026-10-30 집 PC에서, 4.7 ~ 4.15는 2026-10-31 노트북에서 쳤다. 노트북을 서버에 연결한 과정은 3.5절.

| 절 | 한 일 | 날짜 · 기기 |
|---|---|---|
| 4.1 ~ 4.6 | 업데이트 · 방화벽 · 일반 계정 · SSH 잠금 · 자동 업데이트 · Docker | 2026-10-30 집 PC |
| 4.7 | 스왑 확인 · swappiness | 2026-10-31 노트북 |
| 4.8 | 앱 전용 계정 `cbt` · 폴더 | 〃 |
| 4.9 | 비밀값 · 설정 파일 | 〃 |
| 4.10 | 운영 Postgres · Valkey | 〃 |
| 4.11 | DNS | 〃 (Cloudflare 화면) |
| 4.12 | Caddy · HTTPS 인증서 | 〃 |
| 4.13 | 백업 열쇠 `age` | 〃 |
| 4.14 | 운영 파일 · 매일 백업 | 〃 |
| 4.15 | 첫 배포 · 내 계정 | 〃 |

### 4.1 첫 접속 · 업데이트 · 재부팅 (root로, 이때만 root 접속 가능했음)
**상황**: 서버를 막 만든 직후. Vultr가 준 root로 처음 들어가 밀린 업데이트부터 했다.
```bash
cloud-init status --wait      # cloud-init = 서버가 처음 켜질 때 Vultr 설정을 깔아 주는 프로그램. --wait: 끝날 때까지 기다렸다가 상태 출력 → status: done
apt update && apt upgrade -y  # apt update: 설치 가능한 패키지 목록 새로고침 / &&: 앞이 성공해야 뒤를 실행 / apt upgrade -y: 전부 업데이트, -y는 "계속할까?"에 자동 yes
reboot                        # 새 커널 적용 (접속이 끊기는 게 정상, 1분 뒤 다시 접속)
uname -r                      # 지금 돌고 있는 커널 버전 → 6.8.0-147-generic (재부팅으로 새 커널이 올라왔는지 확인)
```
업데이트 중 질문 화면(보라색):
- `sshd_config`를 바꿀까 → **`keep the local version currently installed`**
- PAM 설정을 덮어쓸까(`Override local changes to /etc/pam.d/common-*?`) → **`<No>`**
- "어떤 서비스를 재시작할까" → Enter
- 원칙: **"이미 수정된 설정 파일을 덮어쓸까?"는 지금 것 유지.** Vultr가 맞춰 둔 상태가 정상이다.

### 4.2 방화벽 (ufw)
**상황**: 서버는 만들자마자 인터넷 전체에 노출된다. 쓸 문(22 · 80 · 443)만 남기고 다 닫았다. ufw = Ubuntu의 간단한 방화벽 도구.
```bash
ufw default deny incoming     # 들어오는 건 기본 차단
ufw default allow outgoing    # 나가는 건 허용 (업데이트 · 인증서 받기)
ufw allow 22/tcp              # SSH — enable보다 먼저! 안 그러면 켜는 순간 내 접속도 막힌다 (Vultr가 미리 열어 둠)
ufw allow 80/tcp              # http — Caddy가 인증서를 받을 때 · http→https 넘길 때
ufw allow 443/tcp             # https — 실제 앱 접속
ufw enable                    # 방화벽 켜기. "may disrupt existing ssh connections" → y
ufw status verbose            # 지금 규칙 보기 → 22 · 80 · 443만 ALLOW IN
```
**Docker는 ufw를 우회해 포트를 열 수 있다.** compose의 포트는 반드시 `127.0.0.1:`에 묶는다.

### 4.3 일반 계정
**상황**: root로 계속 일하면 실수 한 번이 서버 전체를 망친다. 평소엔 일반 계정으로 들어오고, 관리자 일만 `sudo`로 빌려 쓰게 했다.
```bash
adduser sagming40                                  # 계정 만들기(집 폴더 포함). 묻는 비밀번호 = sudo 쓸 때 입력하는 값 (SSH 로그인용 아님)
usermod -aG sudo sagming40                         # -aG: 기존 그룹은 두고(a) sudo 그룹에 추가(G) → 관리자 권한 빌리기 허용
rsync --archive --chown=sagming40:sagming40 ~/.ssh /home/sagming40   # root의 열쇠 목록(.ssh)을 새 계정으로 복사. --archive 권한 그대로, --chown 주인을 sagming40으로
```
확인(PC 새 탭에서 접속 → 서버에서): `sudo whoami` → `root`

### 4.4 SSH 잠그기
**상황**: 인터넷의 봇은 쉬지 않고 root 비밀번호를 찍어 본다. root 접속과 비밀번호 로그인을 아예 끄고 열쇠로만 들어오게 했다.
```bash
printf 'PermitRootLogin no\nPasswordAuthentication no\nKbdInteractiveAuthentication no\n' | sudo tee /etc/ssh/sshd_config.d/00-hardening.conf
                              # printf: \n을 줄바꿈으로 바꿔 세 줄을 만든다 → | 로 넘겨 → sudo tee: root 권한으로 파일에 쓴다
                              # (sudo echo … > 파일 은 안 된다 — > 는 sudo 밖의 내 권한으로 열리기 때문)
                              # 세 줄 = root 로그인 금지 · 비밀번호 로그인 금지 · 키보드 입력식 로그인 금지
sudo sshd -t                  # sshd(SSH 서버) 설정 문법 검사 — 아무것도 안 나오면 정상
sudo sshd -T | grep -Ei 'permitrootlogin|passwordauthentication|kbdinteractive'   # -T: 실제로 적용될 최종 설정 출력 → grep -E 여러 단어 · -i 대소문자 무시로 골라 보기 → 셋 다 no
sudo systemctl reload ssh     # 설정 다시 읽기 (열려 있는 접속은 안 끊긴다)
```
- 원래 `sshd_config`는 건드리지 않고 `sshd_config.d/00-…`로 덮어쓴다. sshd는 **먼저 읽은 값**을 쓰므로 Vultr의 `50-cloud-init.conf`보다 앞 번호를 붙인다.
- 확인 (PC에서): root 접속 → `Permission denied (publickey)` / `ssh -o PubkeyAuthentication=no sagming40@<서버 IP>` → 비밀번호를 묻지 않고 거부 / 열쇠로 sagming40 → 접속됨
- **잠그기 전에 다른 탭에 접속 하나를 열어 두고**, 새 탭에서 확인한다. 설정이 틀렸으면 열어 둔 탭에서 고친다.

### 4.5 자동 보안 업데이트 확인
**상황**: 혼자 운영하는 서버라 보안 패치를 매일 손으로 깔 수 없다. 자동으로 깔리는지 확인만 했다.
```bash
cat /etc/apt/apt.conf.d/20auto-upgrades       # cat: 파일 내용 출력 → 두 줄 모두 "1" (목록 새로고침 · 자동 설치 켜짐)
systemctl is-active unattended-upgrades       # 자동 업데이트 서비스가 돌고 있나 → active
```
Ubuntu 기본값으로 켜져 있었다. 보안 패치는 매일 저절로 설치된다(재부팅이 필요한 커널 업데이트는 5장).

### 4.6 Docker Engine (공식 저장소)
**상황**: 운영 Postgres · Valkey를 컨테이너로 띄우려고 Docker를 깔았다. Ubuntu 기본 저장소 것은 오래돼서 Docker 공식 저장소를 추가해 받았다.
```bash
sudo apt-get install -y ca-certificates curl       # 인증서 묶음(https 접속 검증) · curl(파일 받기 도구)
sudo install -m 0755 -d /etc/apt/keyrings          # -d: 폴더 만들기, -m 0755: 권한. 저장소 서명 열쇠를 둘 곳
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
                                                   # Docker 서명 열쇠 받기. -f 실패하면 오류로 끝 · -s 진행 표시 숨김 · -S 그래도 오류는 보임 · -L 주소가 옮겨졌으면 따라감 · -o 저장할 파일
sudo chmod a+r /etc/apt/keyrings/docker.asc        # a+r: 모두(a) 읽기(r) 가능 — apt가 읽어야 해서
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
                                                   # apt에 "Docker 패키지는 이 주소에서, 이 열쇠로 확인해서 받아라"를 적는다
                                                   # $(…): 명령 결과를 그 자리에 넣기 — CPU 종류(amd64) · Ubuntu 이름(noble) / > /dev/null: tee의 화면 출력은 버림
sudo apt-get update                                # 방금 추가한 저장소까지 목록 새로고침
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
                                                   # 본체(docker-ce) · 명령 도구 · 컨테이너 실행기 · 빌드 도구 · compose 다섯 개
sudo docker run --rm hello-world   # 시험용 컨테이너 한 번 실행, --rm 끝나면 지움 → Hello from Docker!
docker compose version             # compose 버전 → v5.6.1
```
- 설치 결과 Docker Engine 29.9.2 · Compose v5.6.1.
- 마지막 설치 줄에서 `docker-ce`(본체)가 빠지면 `failed to connect to the docker API`가 난다 — 설치된 패키지 수가 계획(5개)과 맞는지 본다.
- **`sagming40`을 `docker` 그룹에 넣지 않는다**(= root 권한과 같다). 서버에서는 `sudo docker …`.

### 4.7 스왑 (2026-10-31, 노트북에서)
**상황**: 메모리 1GB에 앱 · Postgres · Valkey · Caddy 넷이 돈다. 메모리가 넘치면 운영체제가 프로그램을 강제로 끈다(OOM). 스왑(책상이 차면 짐을 잠깐 넣는 서랍 = 디스크)을 1GB 만들려고 했는데, 먼저 상태를 보니 이미 있었다.
```bash
swapon --show            # 켜진 스왑 목록 → /swapfile file 2.3G (Vultr가 이미 만들어 둠)
free -h                  # 메모리 · 스왑 사용량, -h 사람이 읽기 쉬운 단위 → Mem 955Mi, Swap 2.3Gi
df -h /                  # 디스크(/) 남은 용량 → 23G 중 16G 남음
sudo fallocate -l 1G /swapfile     # ✗ 헛디딤: 1GB짜리 파일을 미리 잡으려 함 → Text file busy (지금 스왑으로 쓰는 파일이라 거부, 바뀐 것 없음). 두 번 쳐도 같음
ls -l /swapfile          # 권한 확인 → -rw------- root root (root만 읽기 — 스왑엔 메모리 내용이 그대로 내려오므로)
grep swap /etc/fstab     # fstab = 부팅 때 자동으로 붙이는 목록. 거기에 swap 줄이 있나 → /swapfile swap swap defaults 0 0 → 재부팅 뒤에도 켜짐
cat /proc/sys/vm/swappiness                                       # 지금 swappiness(서랍을 얼마나 일찍 여나, 0~100) → 60 (기본값)
echo 'vm.swappiness=10' | sudo tee /etc/sysctl.d/99-swap.conf     # 10으로 낮추는 설정 파일을 root 권한으로 쓰기(재부팅해도 유지)
sudo sysctl --system | grep swappiness                            # 설정 파일 전부 다시 읽어 바로 적용, 그중 swappiness만 보기 → vm.swappiness = 10
```
- **Vultr가 서버를 만들 때 `/swapfile` 2.3GB를 넣어 둔다.** 접속 화면의 `Swap usage: 0%`는 "없다"가 아니라 "안 쓴다"다.
- 2.3GB를 그대로 쓴다(디스크 여유 16G). swappiness 10: 메모리가 정말 모자랄 때만 디스크로 내려 DB가 느려지지 않게.
- 배운 것: 서버를 바꾸는 명령 앞에 **지금 상태를 보는 명령**을 둔다.

### 4.8 앱 전용 계정 `cbt` · 폴더 (2026-10-31, 노트북에서)
**상황**: 앱을 `sagming40`(sudo 가능)이나 root로 돌리면, 앱에 구멍이 났을 때 서버 전체가 넘어간다. 로그인도 못 하고 권한도 없는 직원 계정 `cbt`를 만들어 앱은 이 계정으로만 돌게 했다(`cbt.service`의 `User=cbt`).
```bash
id cbt                             # 먼저 상태 보기: cbt 계정이 있나 → no such user
ls -ld /opt/cbt /etc/cbt /srv/cbt  # -l 자세히 · -d 폴더 안이 아니라 폴더 자체를 → 셋 다 No such file (새로 만들어도 됨)
sudo useradd --system --user-group --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin cbt
                                   # --system 서비스용 계정(번호 1000 미만) · --user-group 같은 이름 그룹 cbt도 만듦
                                   # --no-create-home --home-dir /nonexistent 집 폴더 없음 · --shell /usr/sbin/nologin 로그인 불가
sudo install -d -m 755 /opt/cbt    # 실행 파일(releases/ · current) 둘 곳. 755 = 주인(root)만 쓰기, 나머지는 읽기 · 들어가기
sudo install -d -m 755 /etc/cbt    # cbt.env(640 root:cbt) · backup.pub 둘 곳 — 권한은 파일마다 따로 건다
sudo install -d -m 700 /srv/cbt    # 운영 compose · DB 비밀번호 .env 둘 곳. 700 = root만 (폴더째 잠금)
id cbt                             # → uid=999(cbt) gid=987(cbt) groups=987(cbt)
getent passwd cdt                  # ✗ 헛디딤(오타 cdt) → 출력 없음. 조회만 하는 명령이라 영향 없음
getent passwd cbt                  # 계정 정보 한 줄 → cbt:x:999:987::/nonexistent:/usr/sbin/nologin
sudo -u cbt -s /bin/bash -c 'echo 들어가짐' ; sudo su - cbt
                                   # 앞: root가 cbt 이름으로 명령 하나를 대신 실행(-u 누구로 · -s 셸 지정 · -c 명령) → 들어가짐 (cbtctl이 쓰는 방식)
                                   # ; 앞이 어떻든 이어서 / 뒤: cbt로 로그인 시도 → This account is currently not available. (거부가 정상)
ls -ld /opt/cbt /etc/cbt /srv/cbt  # → drwxr-xr-x root root 두 개 + drwx------ root root(/srv/cbt)
```
- `su`의 `cannot change directory to /nonexistent` 경고는 일부러 집 폴더를 안 만든 결과라 정상.
- 폴더 주인은 전부 root — 앱(`cbt`)은 읽기만 한다. 앱은 파일을 쓰지 않는다(로그는 journald).

### 4.9 비밀값 · 설정 파일 (2026-10-31, 노트북에서)
**상황**: 운영 DB · Valkey 비밀번호를 정하고, 설정 파일 셋(운영 compose · 그 비밀번호 파일 · 앱 설정)을 서버에 놓았다. 원칙: 비밀번호는 **서버에서 만들고 화면에 한 번도 띄우지 않는다** → 대화 · 셸 기록 · 저장소 어디에도 안 남는다.
```powershell
# ── 노트북 PowerShell (C:\_dev\cbt-journal) ──
scp deploy/compose.prod.yaml deploy/compose.env.example deploy/cbt.env.example cbt:~/
                                   # scp = ssh로 파일 복사. 저장소의 원본 셋 → cbt(별명)의 홈(~/). passphrase 한 번
```
```bash
# ── 서버 ── 같은 터미널에서 이어서 친다 (중간에 나가면 PG · VK 변수가 사라진다)
sudo install -m 644 -o root -g root ~/compose.prod.yaml   /srv/cbt/compose.yaml   # install = 복사 + 주인(-o · -g) + 권한(-m)을 한 번에. 644 모두 읽기
sudo install -m 600 -o root -g root ~/compose.env.example /srv/cbt/.env           # 600 root만 읽기 · 쓰기 (DB 비밀번호가 들어갈 파일)
sudo install -m 640 -o root -g cbt  ~/cbt.env.example     /etc/cbt/cbt.env        # 640 root 쓰기 · cbt 그룹 읽기 · 나머지 못 봄
PG=$(openssl rand -hex 32); VK=$(openssl rand -hex 32)
                                   # openssl rand -hex 32: 무작위 32바이트를 0-9a-f 64글자로 → URL에 넣어도 안 깨진다. 결과를 변수 PG · VK에만 담음 (화면에 안 나옴)
sudo sed -i "s/^POSTGRES_PASSWORD=$/POSTGRES_PASSWORD=$PG/; s/^VALKEY_PASSWORD=$/VALKEY_PASSWORD=$VK/" /srv/cbt/.env
                                   # sed -i: 파일 안에서 바로 바꾸기. s/찾을것/바꿀것/ — 빈칸이던 PASSWORD= 뒤에 값 채우기 (^ 줄 처음 · $ 줄 끝)
sudo sed -i "s/<POSTGRES_PASSWORD>/$PG/; s/<VALKEY_PASSWORD>/$VK/" /etc/cbt/cbt.env
                                   # 앱 설정의 DB · Valkey 주소 안 자리표 <…>를 같은 값으로 바꾸기
unset PG VK                        # 변수 지우기 — 이 터미널에도 남기지 않음
```
확인 — 내용은 보지 않고 길이 · 일치만(`cat`으로 열지 않는다):
```bash
sudo awk -F= '/PASSWORD/ {print $1, length($2)}' /srv/cbt/.env      # awk: = 로 나눠(-F=) PASSWORD 줄의 이름과 값 길이만 출력 → 둘 다 64
sudo grep -c '<' /etc/cbt/cbt.env                                    # ✗ 헛디딘 확인(Claude 실수): < 가 든 줄 수 → 1 — 주석 줄의 <…> 까지 셌다
sudo bash -c 'source /srv/cbt/.env; grep -q ":$POSTGRES_PASSWORD@" /etc/cbt/cbt.env && grep -q ":$VALKEY_PASSWORD@" /etc/cbt/cbt.env && echo 두 파일 비밀번호 일치'
                                   # root 셸 하나에서 .env를 읽어(source) 그 값이 cbt.env 주소 안에 있는지 조용히(-q) 찾기 → 두 파일 비밀번호 일치
sudo ls -l /srv/cbt /etc/cbt       # ✗ 헛디딘 확인: ls는 점으로 시작하는 .env를 숨긴다 → compose.yaml · cbt.env만 보임
sudo docker compose -f /srv/cbt/compose.yaml config --quiet && echo compose 설정 OK
                                   # compose 파일을 .env와 합쳐 해석만 해 봄(-f 파일 지정 · --quiet 오류만) → compose 설정 OK
rm ~/compose.prod.yaml ~/compose.env.example ~/cbt.env.example       # 홈에 남은 원본 지우기
sudo grep -n '<' /etc/cbt/cbt.env  # 다시 보기: -n 줄 번호와 함께 → 2번 줄 주석 하나뿐 (자리표 다 바뀜)
sudo ls -la /srv/cbt               # -a 숨김 파일까지 → .env -rw------- root root / compose.yaml -rw-r--r--
```
- `cbt.env`는 `-rw-r----- root cbt`.
- 비밀번호는 서버 두 파일에만 있다. 잃어도 서버에서 다시 만들면 된다(DB 비밀번호를 바꾸려면 Postgres 안에서도 바꿔야 한다 — 볼륨을 처음 만들 때 한 번만 정해지므로).

### 4.10 운영 Postgres · Valkey (2026-10-31, 노트북에서)
**상황**: 4.9의 compose 파일로 운영 DB(Postgres) · 캐시(Valkey)를 띄웠다. 핵심 확인은 "서버 안(127.0.0.1)에서만 열리고 바깥에서는 안 들어와지는가".
```bash
sudo docker compose -f /srv/cbt/compose.yaml up -d
                                   # up: 파일대로 컨테이너 만들고 켜기 · -d 뒤에서(detached) → 이미지 postgres:18 · valkey/valkey:9 받고 볼륨 cbt_pgdata · cbt_valkeydata 생성
sudo docker compose -f /srv/cbt/compose.yaml ps
                                   # 컨테이너 상태 → 둘 다 Up (healthy), 127.0.0.1:5432 · 127.0.0.1:6379 (20초쯤 기다렸다 봄)
sudo docker compose -f /srv/cbt/compose.yaml exec valkey valkey-cli ping
                                   # exec: 돌고 있는 valkey 컨테이너 안에서 명령 실행 → PONG (compose가 VALKEYCLI_AUTH로 비밀번호를 넘겨 통과)
sudo docker compose -f /srv/cbt/compose.yaml exec postgres psql -U cbt -d cbt_journal -c 'select version();'
                                   # ✗ 헛디딤: 화면에 아무것도 안 나옴(오류도 없음) — TTY(터미널 흉내)를 붙여 실행한 탓으로 보임
sudo docker inspect --format '{{.Name}} {{.HostConfig.LogConfig.Config}}' $(sudo docker compose -f /srv/cbt/compose.yaml ps -q)
                                   # inspect: 컨테이너 설정 조회, --format 로그 설정만 / $(… ps -q): 컨테이너 id 목록을 그 자리에 → 둘 다 max-file:3 max-size:10m
sudo ss -tlnp | grep -E ':5432|:6379'
                                   # ss: 열린 포트 목록 (-t TCP · -l 듣는 중 · -n 숫자로 · -p 어느 프로그램) → 127.0.0.1:5432 · 127.0.0.1:6379만 (0.0.0.0이면 인터넷에 열린 것)
sudo docker compose -f /srv/cbt/compose.yaml exec -T postgres psql -U cbt -d cbt_journal -tAc 'show server_version;'
                                   # -T TTY 없이 · -U 사용자 · -d DB · -t 머리글 없이 · -A 표 정렬 없이 · -c 이 SQL 하나 → 18.6 (Debian 18.6-1.pgdg13+2)
```
```powershell
# ── 노트북 PowerShell ── 바깥에서 정말 막혔는지 (IP는 ssh 별명에서 읽어 저장소 · 화면에 안 적는다)
$ip = ((ssh -G cbt | Select-String '^hostname ') -split ' ')[1]
                                   # ssh -G로 별명 설정을 출력 → hostname 줄만 골라 → 공백으로 나눈 두 번째 값(IP)을 변수에
Test-NetConnection $ip -Port 22   | Select-Object RemotePort, TcpTestSucceeded   # 그 포트로 연결되나 → True (비교용)
Test-NetConnection $ip -Port 5432 | Select-Object RemotePort, TcpTestSucceeded   # → False (DB 막힘)
Test-NetConnection $ip -Port 6379 | Select-Object RemotePort, TcpTestSucceeded   # → False (Valkey 막힘)
```
- PLAN M3 완료 기준 "바깥에서 DB · Valkey 포트 접속이 실패한다"를 여기서 봤다.

### 4.11 DNS (2026-10-31, Cloudflare 화면에서)
**상황**: Caddy는 켜지자마자 `cbt.sagming.com` 인증서를 받으러 간다. 그 이름이 아직 서버를 가리키지 않으면 발급 실패가 반복되고 Let's Encrypt가 한동안 막는다 → **Caddy보다 DNS 먼저**(ARCHITECTURE 7.5절).
- `sagming.com` → DNS → 레코드 추가: 형식 `A`(이름 → IPv4 주소) · 이름 `cbt` · IPv4 서버 IP · 프록시 **끔(회색 구름, DNS 전용)** · TTL 자동.
```powershell
# ── 노트북 PowerShell ──
nslookup cbt.sagming.com 1.1.1.1   # 이 이름의 IP를 Cloudflare DNS(1.1.1.1)에 직접 묻기 → 서버 IP
nslookup cbt.sagming.com           # 지금 와이파이가 쓰는 DNS에 묻기 → 서버 IP (퍼졌는지 확인)
```
- 프록시를 끄는 이유: 켜면 Cloudflare가 중간에 끼어 Caddy 인증서 발급 · 실제 접속 IP 규칙(ARCHITECTURE 3.3절)이 꼬인다. Cloudflare 화면의 "프록싱 필요" 안내 · `www` · 루트 도메인 권장 사항은 따르지 않는다(쓰는 이름은 `cbt` 하나).

### 4.12 Caddy (2026-10-31, 노트북에서)
**상황**: 바깥 손님(443 · 80)을 받아 HTTPS 인증서를 보여 주고 안쪽 앱(127.0.0.1:8080)으로 안내하는 안내데스크를 세웠다. 설치는 [공식 문서](https://caddyserver.com/docs/install)의 apt 저장소 그대로(Ubuntu 기본 저장소 것은 오래됨).
```bash
# ── 서버 ── 설치
sudo apt install --yes debian-keyring debian-archive-keyring apt-transport-https curl
                                   # 저장소 추가에 필요한 도구들 (--yes = -y). curl은 이미 최신이었음
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
                                   # Caddy 서명 열쇠 받기 → gpg --dearmor로 apt가 읽는 형식으로 바꿔 저장. (-1 TLS 1 이상 · -s 조용히 · -L 따라가기 · -f 실패 시 오류)
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
                                   # "Caddy는 이 주소에서, 위 열쇠로 확인해서 받아라"를 apt 목록 파일로 저장 (tee라 내용이 화면에도 나옴)
sudo chmod o+r /usr/share/keyrings/caddy-stable-archive-keyring.gpg   # o+r: 다른 사용자(o)도 읽기 — apt가 읽게
sudo chmod o+r /etc/apt/sources.list.d/caddy-stable.list
sudo apt update                    # 새 저장소까지 목록 새로고침 → dl.cloudsmith.io … InRelease 받음
sudo apt install caddy             # Y → caddy 2.11.7 설치, 바로 systemd 서비스로 켜짐(이때는 기본 환영 페이지)
caddy version                      # → v2.11.7
systemctl is-active caddy          # → active
```
```powershell
# ── 노트북 PowerShell ──
scp deploy/Caddyfile cbt:~/        # 저장소의 운영 Caddyfile을 서버 홈으로
```
```bash
# ── 서버 ── 우리 설정으로 바꾸기
sudo install -m 644 -o root -g root ~/Caddyfile /etc/caddy/Caddyfile   # Caddy가 읽는 자리에 넣기
rm ~/Caddyfile
sudo caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile
                                   # 적용 전에 검사만 (--adapter caddyfile 이 형식으로 읽어라) → Valid configuration
sudo systemctl reload caddy        # 끄지 않고 설정만 다시 읽기 → 바로 인증서 발급 시작
sudo journalctl -u caddy --since "5 min ago" --no-pager | grep -Ei 'certificate obtained|error' | tail -n 5
                                   # journalctl: 시스템 로그 (-u caddy 서비스만 · --since 최근 5분 · --no-pager 한 번에 출력) → 발급 성공 · 오류 줄만, 마지막 5줄
                                   # → 발급 줄은 위로 밀렸고, 스캐너가 찔러 본 502 error 줄만 보였다 (아래 메모)
```
확인 (Claude가 노트북에서):
```powershell
curl.exe -sI https://cbt.sagming.com    # -s 조용히 · -I 머리글만 → 앱 배포 전이라 502 Bad Gateway (인증서 오류 없이 응답 = HTTPS 됨)
curl.exe -sI http://cbt.sagming.com     # → 308 Location: https://cbt.sagming.com/ (http로 와도 https로 보냄)
```
- 인증서: `CN=cbt.sagming.com`, 발급자 Let's Encrypt(`YE2`), TLS 1.3, 만료 2027-01-29 — 갱신은 Caddy가 알아서 한다.
- IP로: `https://IP`는 TLS 연결 실패, `http://IP`는 `https://IP`로 308 후 실패 → 앱은 안 보인다.
- **인증서를 받자마자 20초 안에 스캐너가 들어왔다**(`/.env` · `/phpinfo.php` · `/wp-admin/…` · `/database.sql`). 새 인증서는 공개 목록(Certificate Transparency)에 올라가고 봇이 그 목록을 지켜본다. 앱이 없어 전부 502였다. 정상적인 인터넷 소음이다 — 그래서 기본 잠금 · 요청 제한이 먼저 있어야 했다.
- Caddy는 접근 로그를 안 남기지만, **오류(502 등) 때는 `journalctl -u caddy`에 접속 IP · 경로 · 헤더를 남긴다**. 쿠키 · `Authorization`은 Caddy가 기본으로 가린다. 경로에는 id만 있고 본문은 안 남아서 REQUIREMENTS 5.1절과 맞는다.

### 4.13 백업 열쇠 `age` (2026-10-31, 노트북에서)
**상황**: 매일 DB 백업을 서버에서 잠그되, 서버는 **열지 못하게** 하려고 자물쇠(공개키)와 열쇠(개인키)를 나눴다. 자물쇠만 서버에, 열쇠는 PC · 종이에만. 서버가 털려도 백업 내용은 못 읽는다.
```powershell
# ── 노트북 PowerShell (홈 폴더) ──
age-keygen -o "$HOME\cbt-backup.key"     # 열쇠 한 쌍을 파일로(-o) → 화면에 Public key: age1… (공개키, 비밀 아님)
                                         # world-readable 경고는 Linux 권한 기준 — Windows 홈 폴더는 본인만 읽어서 괜찮음
notepad "$HOME\cbt-backup.key"           # 파일 열어 보기: # created · # public key · AGE-SECRET-KEY-1… (개인키) 세 줄 → 개인키 줄을 종이에 옮겨 적고 대조, 저장 없이 닫기
age-keygen -y "$HOME\cbt-backup.key"     # 개인키 파일에서 짝이 되는 공개키만 다시 뽑기(-y) → 서버에 붙여 넣을 age1…
```
```bash
# ── 서버 ──
age-keygen -y "$HOME\cbt-backup.key"     # ✗ 헛디딤: 노트북용 명령을 서버에서 침 → age-keygen not found (영향 없음)
sudo apt install --yes age               # 서버에 age 설치 → 1.1.1 (노트북 1.3.2와 호환)
echo 'age1…' | sudo tee /etc/cbt/backup.pub   # 공개키를 서버 자물쇠 파일로 (root 권한으로 쓰기)
sudo chmod 644 /etc/cbt/backup.pub       # 모두 읽기 가능 — 비밀이 아니고 백업 스크립트가 읽는다
echo '열쇠 짝 시험' | age -R /etc/cbt/backup.pub -o ~/keytest.age
                                         # 짝 시험: 글자를 자물쇠 파일(-R)로 잠가 ~/keytest.age로(-o)
```
```powershell
# ── 노트북 PowerShell ──
scp cbt:~/keytest.age $env:TEMP\         # 서버의 잠긴 파일을 노트북 임시 폴더로
age -d -i "$HOME\cbt-backup.key" "$env:TEMP\keytest.age"   # 풀기(-d) · 이 열쇠로(-i) → 열쇠 짝 시험 (짝이 맞음)
Remove-Item "$env:TEMP\keytest.age"      # 시험 파일 지우기 (서버 쪽은 4.14의 rm에서 지움)
```
- **개인키(`AGE-SECRET-KEY-1…`)는 서버 · 저장소 · 대화 · 클라우드에 두지 않는다.** 보관: 노트북 `$HOME\cbt-backup.key` + 종이. 백업을 열 PC마다 같은 파일을 `$HOME\cbt-backup.key`에 둔다(USB로 옮기고 USB에서는 지운다).
- 잃으면 모든 백업을 못 연다 → 새 열쇠를 만들어 `backup.pub`를 바꾸고, 그 전 백업은 버린다.
- `age-keygen`의 `writing secret key to a world-readable file` 경고는 Unix 권한 기준이다. Windows 홈 폴더는 기본으로 본인 · 관리자만 읽는다.

### 4.14 운영 파일 · 백업 타이머 (2026-10-31, 노트북에서)
**상황**: 저장소 `deploy/`에 만들어 둔 운영 파일(앱 서비스 · 운영 명령 도우미 · 백업 스크립트 · 매일 알람)을 서버 제자리에 넣고, 백업을 처음으로 실제 Postgres에서 돌려 봤다.
```powershell
# ── 노트북 PowerShell (C:\_dev\cbt-journal) ──
scp deploy/cbtctl deploy/cbt-backup deploy/cbt.service deploy/cbt-backup.service deploy/cbt-backup.timer cbt:~/   # 다섯 개를 서버 홈으로
```
```bash
# ── 서버 ──
sudo install -m 755 -o root -g root ~/cbtctl ~/cbt-backup /usr/local/bin/
                                   # 실행할 스크립트 둘 → /usr/local/bin(어디서나 이름만으로 실행되는 곳). 755 = 실행 가능
sudo install -m 644 -o root -g root ~/cbt.service ~/cbt-backup.service ~/cbt-backup.timer /etc/systemd/system/
                                   # systemd 설정 셋 → systemd가 읽는 곳. 644 = 읽기만
rm ~/cbtctl ~/cbt-backup ~/cbt.service ~/cbt-backup.service ~/cbt-backup.timer ~/keytest.age   # 홈 정리 (4.13 시험 파일 포함)
sudo systemctl daemon-reload       # systemd에 "설정 파일이 새로 생겼으니 다시 읽어라"
sudo systemd-analyze verify /etc/systemd/system/cbt.service /etc/systemd/system/cbt-backup.service /etc/systemd/system/cbt-backup.timer
                                   # 설정 검사 → cbt.service: Command /opt/cbt/current/app is not executable 한 줄만 (실행 파일은 첫 배포가 넣는다 — 정상). 두 번 쳐도 같음
sudo systemctl start cbt-backup    # 백업 서비스를 손으로 한 번 실행 (평소엔 타이머가 부른다)
sudo journalctl -u cbt-backup -n 10 --no-pager   # 그 서비스 로그 마지막 10줄(-n 10) → [backup] 만듦: /var/backups/cbt/cbt-20261031T052841Z.dump.age (1096 바이트)
ls -l /var/backups/cbt/            # → -rw-r--r-- root root 1096 … (암호화돼 있어 644 — sagming40이 scp로 가져갈 수 있게)
sudo systemctl enable --now cbt-backup.timer
                                   # enable 재부팅 뒤에도 켜지게 · --now 지금 바로도 켜기 → timers.target.wants에 링크 생성
systemctl list-timers cbt-backup.timer --no-pager   # 다음 실행 시각 → NEXT Sat 2026-10-31 19:09:36 UTC (= 04:09 KST, 0~10분 흩음)
```
```powershell
# ── 노트북 PowerShell (C:\_dev\cbt-journal) ──
task backup:pull                   # 서버 백업 전부를 ~/cbt-backups/로 (scp 한 번, passphrase 한 번)
Get-ChildItem "$HOME\cbt-backups"  # 받은 파일 목록 → cbt-20261031T052841Z.dump.age 1096
```
- 첫 배포 전에는 `verify`가 `cbt.service: Command /opt/cbt/current/app is not executable`를 낸다 — 실행 파일은 첫 `task deploy`가 넣는다. 나머지 두 파일은 출력 없음(정상).
- `cbt.service`는 여기서 켜지 않는다. 첫 배포의 `release.sh`가 `enable` · `restart` 한다.
- 결과: 빈 DB 백업 1096바이트(`-rw-r--r-- root`), PC에서 `task backup:pull` → `~/cbt-backups/`, 개인키로 풀어 896바이트 · 머리글 `PGDMP`(pg_dump 형식).

### 4.15 첫 배포 · 내 계정 (2026-10-31, 노트북에서)
**상황**: 서버 준비가 다 끝나 앱을 처음 올렸다. 가입이 꺼져 있어서 실제로 쓸 내 계정은 서버에서 명령으로 만들었다.
```powershell
# ── 노트북 PowerShell (C:\_dev\cbt-journal) ──
task deploy     # 작업 트리 깨끗 · push됐는지 확인 → 프론트 빌드 → Linux 실행 파일(커밋 번호 새김) → 서버로 복사 → 서버에서 release.sh
                # (버전 확인 → 배포 직전 백업 → migrate → current 링크 → cbt 서비스 enable · restart → /api/health 확인)
                # passphrase 2번(복사 · 실행) + sudo 비밀번호 1번. 배포한 커밋 1a04de8
```
```bash
# ── 서버 ──
sudo cbtctl user create --email 내이메일 --nickname 내닉네임
                                   # cbtctl: 운영 설정(/etc/cbt/cbt.env)을 읽고 cbt 계정으로 app을 실행하는 도우미 → app user create
                                   # 새 비밀번호 · 한 번 더 (화면에 안 보임, 12자 이상) → "계정을 만들었다" + user_id
systemctl is-active cbt            # 앱 서비스가 돌고 있나 → active
sudo systemd-analyze verify /etc/systemd/system/cbt.service    # → 출력 없음 (4.14의 "실행 파일 없음"이 사라짐)
```
```powershell
# ── 노트북 PowerShell ──
curl.exe -sI https://cbt.sagming.com   # 머리글만 → 200 OK + 보안 헤더 4개
```
- 확인 (PC): `curl.exe -sI https://cbt.sagming.com` → `200`, HSTS `max-age=86400` · `nosniff` · `DENY` · `Referrer-Policy` 있음, `Server` 없음. `/api/health` → postgres · valkey `ok`, 로그인 없이 `/api/me` → `401`.
- 로그인 · 설정 · 로그아웃: 노트북 브라우저, 휴대폰(LTE, 와이파이 끔) 둘 다 됨.
- 응답에 `Via: 1.1 Caddy`가 붙는다(Caddy가 프록시할 때 넣는 표준 헤더).


## 5. 자주 쓰는 명령 (서버에서)

| 하고 싶은 것 | 명령 |
|---|---|
| 업데이트 | `sudo apt update && sudo apt upgrade -y` |
| 재부팅이 필요한지 | `cat /var/run/reboot-required` (파일이 없으면 필요 없음) |
| 재부팅 | `sudo reboot` |
| 메모리 · 디스크 · 켜진 시간 | `free -h` · `df -h` · `uptime` |
| 방화벽 상태 | `sudo ufw status verbose` |
| 서버에서 열려 있는 포트 | `sudo ss -tlnp` (`127.0.0.1:`이면 바깥에서 못 들어옴, `0.0.0.0:`이면 바깥에 열림) |
| 최근 로그인 기록 | `last -n 10` |
| 오늘 SSH 실패 시도 | `sudo journalctl -u ssh --since today \| grep -Ei 'invalid\|failed'` |
| 서비스 상태 · 로그 | `systemctl status <이름>` · `sudo journalctl -u <이름> -n 50` |
| 컨테이너 목록 | `sudo docker ps` |
| SSH 설정 다시 확인 | `sudo sshd -T \| grep -Ei 'permitrootlogin\|passwordauthentication'` |
| 앱 상태 · 로그 | `systemctl status cbt` · `sudo journalctl -u cbt -n 50 --no-pager` |
| 운영 DB · Valkey 상태 | `sudo docker compose -f /srv/cbt/compose.yaml ps` |
| 운영 DB에 SQL 한 줄 | `sudo docker compose -f /srv/cbt/compose.yaml exec -T postgres psql -U cbt -d cbt_journal -tAc '…'` |
| 운영 명령(계정 · 마이그레이션 상태) | `sudo cbtctl user create …` · `sudo cbtctl migrate status` |
| 백업 손으로 · 결과 · 다음 시각 | `sudo systemctl start cbt-backup` · `sudo journalctl -u cbt-backup -n 10 --no-pager` · `systemctl list-timers cbt-backup.timer` |
| Caddy 설정 바꾼 뒤 | `sudo caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile` → `sudo systemctl reload caddy` |

## 6. 하지 말 것
- `do-release-upgrade` — 접속 화면의 "New release '26.04.1 LTS' available" 안내를 따르지 않는다. OS 통째 업그레이드는 새 서버를 만들어 옮기는 방식으로 한다.
- `sudo usermod -aG docker sagming40` — docker 그룹 = root 권한.
- `ufw disable` · root SSH 다시 허용 · 비밀번호 로그인 다시 허용.
- compose 포트를 `127.0.0.1:` 없이 열기 (Docker가 방화벽을 우회한다).
- Vultr에서 서버를 **Stop**해서 요금을 아끼려 하기 — 정지해도 요금이 나간다. 없애려면 **Destroy**(되돌릴 수 없음, 백업 먼저).
- 이 문서 · 저장소 · 대화에 비밀값(root · sudo 비밀번호, DB 비밀번호, 개인 키, 서버 IP) 적기.

## 7. 앞으로 채울 것 (M3)
운영 구성이 생기면 이 문서에 이어서 적는다: 복원.

---

## 변경 이력
| 날짜 | 내용 |
|---|---|
| 2026-10-30 | 첫 작성: 접속 방법 · 새 PC 키 추가 · 지금까지 한 설정(업데이트 · 방화벽 · 일반 계정 · SSH 잠금 · 자동 업데이트 · Docker) · 자주 쓰는 명령 · 하지 말 것 |
| 2026-10-31 | 3.2절 ssh 별명 `cbt`가 `task deploy`에 필요하다고 표시. 3.4절에 노트북 추가 |
| 2026-10-31 | 3.2절 `config.txt` 함정, 3.3 · 3.4절 웹 콘솔 찾는 곳 · root 경로, 노트북 열쇠 추가. 4.7절 스왑, 4.8절 `cbt` 계정 · 폴더, 4.9절 비밀값 · 설정 파일, 4.10절 운영 Postgres · Valkey, 4.11절 DNS, 4.12절 Caddy, 4.13절 백업 열쇠, 4.14절 운영 파일 · 백업 타이머, 4.15절 첫 배포 · 내 계정 |
| 2026-10-31 | 사용자 요청: 서버 작업에서 친 명령을 전부 — 4장 머리에 순서 표, 절마다 **상황** · 줄마다 뜻 · 옵션 · 결과, 헛디딘 명령(✗)까지. 3.5절 노트북 추가 실제 순서. 5장에 운영 명령 줄 추가 |
