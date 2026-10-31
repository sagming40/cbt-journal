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
notepad "$HOME\.ssh\config"
```
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
- Vultr → Compute → Instances → `cbt-1` → 오른쪽 위 **View Console**
- 웹 화면 안에서 `root`와 **Vultr 화면에 표시된 root 비밀번호**로 로그인한다. root 금지는 SSH에만 걸려 있어서 콘솔은 된다.
- root 비밀번호는 이 비상구 말고는 쓰지 않는다. 어디에도 옮겨 적지 않는다.

### 3.4 새 PC를 추가할 때 (예: 회사 PC)
PC마다 열쇠를 따로 만든다 — 한 PC를 잃으면 그 열쇠만 지우면 된다. **서버는 이미 만들어져 있어서 Vultr에 키를 등록해도 이 서버에는 안 들어간다.** 서버의 `authorized_keys`에 직접 한 줄 추가해야 한다.

1. **새 PC에서** 열쇠 만들기
   ```powershell
   New-Item -ItemType Directory -Force "$HOME\.ssh"
   ssh-keygen -t ed25519 -C "cbt-journal company-pc" -f "$HOME\.ssh\cbt_vultr"
   Get-Content "$HOME\.ssh\cbt_vultr.pub"
   ```
   마지막 줄(`ssh-ed25519 AAAA… cbt-journal company-pc`)을 복사한다. `.pub`가 아닌 파일(개인 키)은 열지도 옮기지도 않는다.
2. **이미 접속되는 PC로 서버에 들어가서** 그 한 줄을 추가
   ```bash
   echo 'ssh-ed25519 AAAA…(복사한 한 줄 전체)' >> ~/.ssh/authorized_keys
   ```
   `>>`(두 개)여야 한다. `>`(하나)면 기존 열쇠를 지워 버려 다른 PC가 못 들어온다.
3. 새 PC에서 3.1 명령으로 접속 확인.
4. 지금 등록된 열쇠 보기 (서버에서): `cat ~/.ssh/authorized_keys` — 줄 끝 메모(`home-pc` · `company-pc`)로 구분한다. PC를 잃으면 그 줄만 지운다.

## 4. 지금까지 한 설정 (2026-10-30, 순서대로)

### 4.1 첫 접속 · 업데이트 · 재부팅 (root로, 이때만 root 접속 가능했음)
```bash
cloud-init status --wait      # 서버 초기 설정이 끝났는지 → status: done
apt update && apt upgrade -y  # 목록 새로고침 + 업데이트 설치
reboot                        # 새 커널 적용 (접속이 끊기는 게 정상, 1분 뒤 다시 접속)
uname -r                      # → 6.8.0-147-generic
```
업데이트 중 질문 화면(보라색):
- `sshd_config`를 바꿀까 → **`keep the local version currently installed`**
- PAM 설정을 덮어쓸까(`Override local changes to /etc/pam.d/common-*?`) → **`<No>`**
- "어떤 서비스를 재시작할까" → Enter
- 원칙: **"이미 수정된 설정 파일을 덮어쓸까?"는 지금 것 유지.** Vultr가 맞춰 둔 상태가 정상이다.

### 4.2 방화벽 (ufw)
```bash
ufw default deny incoming     # 들어오는 건 기본 차단
ufw default allow outgoing    # 나가는 건 허용 (업데이트 · 인증서 받기)
ufw allow 22/tcp              # SSH — enable보다 먼저! (Vultr가 미리 열어 둠)
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable                    # "may disrupt existing ssh connections" → y
ufw status verbose            # 22 · 80 · 443만 ALLOW IN
```
**Docker는 ufw를 우회해 포트를 열 수 있다.** compose의 포트는 반드시 `127.0.0.1:`에 묶는다.

### 4.3 일반 계정
```bash
adduser sagming40                                  # 비밀번호 = sudo 쓸 때 입력하는 값 (SSH 로그인용 아님)
usermod -aG sudo sagming40                         # 관리자 권한 빌리기 허용
rsync --archive --chown=sagming40:sagming40 ~/.ssh /home/sagming40   # root의 열쇠 목록을 복사
```
확인(PC 새 탭에서 접속 → 서버에서): `sudo whoami` → `root`

### 4.4 SSH 잠그기
```bash
printf 'PermitRootLogin no\nPasswordAuthentication no\nKbdInteractiveAuthentication no\n' | sudo tee /etc/ssh/sshd_config.d/00-hardening.conf
sudo sshd -t                  # 문법 검사 — 아무것도 안 나오면 정상
sudo sshd -T | grep -Ei 'permitrootlogin|passwordauthentication|kbdinteractive'   # 셋 다 no
sudo systemctl reload ssh     # 적용 (열려 있는 접속은 안 끊긴다)
```
- 원래 `sshd_config`는 건드리지 않고 `sshd_config.d/00-…`로 덮어쓴다. sshd는 **먼저 읽은 값**을 쓰므로 Vultr의 `50-cloud-init.conf`보다 앞 번호를 붙인다.
- 확인 (PC에서): root 접속 → `Permission denied (publickey)` / `ssh -o PubkeyAuthentication=no sagming40@<서버 IP>` → 비밀번호를 묻지 않고 거부 / 열쇠로 sagming40 → 접속됨
- **잠그기 전에 다른 탭에 접속 하나를 열어 두고**, 새 탭에서 확인한다. 설정이 틀렸으면 열어 둔 탭에서 고친다.

### 4.5 자동 보안 업데이트 확인
```bash
cat /etc/apt/apt.conf.d/20auto-upgrades       # 두 줄 모두 "1"
systemctl is-active unattended-upgrades       # active
```
Ubuntu 기본값으로 켜져 있었다. 보안 패치는 매일 저절로 설치된다(재부팅이 필요한 커널 업데이트는 5장).

### 4.6 Docker Engine (공식 저장소)
```bash
sudo apt-get install -y ca-certificates curl
sudo install -m 0755 -d /etc/apt/keyrings
sudo curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
sudo chmod a+r /etc/apt/keyrings/docker.asc
echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu $(. /etc/os-release && echo "$VERSION_CODENAME") stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
sudo apt-get update
sudo apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
sudo docker run --rm hello-world   # Hello from Docker!
docker compose version             # v5.6.1
```
- 설치 결과 Docker Engine 29.9.2 · Compose v5.6.1.
- 마지막 설치 줄에서 `docker-ce`(본체)가 빠지면 `failed to connect to the docker API`가 난다 — 설치된 패키지 수가 계획(5개)과 맞는지 본다.
- **`sagming40`을 `docker` 그룹에 넣지 않는다**(= root 권한과 같다). 서버에서는 `sudo docker …`.

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

## 6. 하지 말 것
- `do-release-upgrade` — 접속 화면의 "New release '26.04.1 LTS' available" 안내를 따르지 않는다. OS 통째 업그레이드는 새 서버를 만들어 옮기는 방식으로 한다.
- `sudo usermod -aG docker sagming40` — docker 그룹 = root 권한.
- `ufw disable` · root SSH 다시 허용 · 비밀번호 로그인 다시 허용.
- compose 포트를 `127.0.0.1:` 없이 열기 (Docker가 방화벽을 우회한다).
- Vultr에서 서버를 **Stop**해서 요금을 아끼려 하기 — 정지해도 요금이 나간다. 없애려면 **Destroy**(되돌릴 수 없음, 백업 먼저).
- 이 문서 · 저장소 · 대화에 비밀값(root · sudo 비밀번호, DB 비밀번호, 개인 키, 서버 IP) 적기.

## 7. 앞으로 채울 것 (M3)
운영 구성이 생기면 이 문서에 이어서 적는다: 스왑 · 앱 전용 계정 `cbt` · 운영용 compose(Postgres · Valkey) · Caddy · 앱 서비스(systemd) · 배포(`task deploy`) · 백업 · 복원 · 내 계정 만들기.

---

## 변경 이력
| 날짜 | 내용 |
|---|---|
| 2026-10-30 | 첫 작성: 접속 방법 · 새 PC 키 추가 · 지금까지 한 설정(업데이트 · 방화벽 · 일반 계정 · SSH 잠금 · 자동 업데이트 · Docker) · 자주 쓰는 명령 · 하지 말 것 |
| 2026-10-31 | 3.2절 ssh 별명 `cbt`가 `task deploy`에 필요하다고 표시 |
