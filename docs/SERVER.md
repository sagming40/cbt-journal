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

### 4.7 스왑 (2026-10-31, 노트북에서)
```bash
swapon --show            # /swapfile file 2.3G — 이미 있었다
ls -l /swapfile          # -rw------- root root
grep swap /etc/fstab     # /swapfile swap swap defaults 0 0 → 재부팅 뒤에도 켜짐
cat /proc/sys/vm/swappiness                                       # 60 (기본값)
echo 'vm.swappiness=10' | sudo tee /etc/sysctl.d/99-swap.conf
sudo sysctl --system | grep swappiness                            # vm.swappiness = 10
```
- **Vultr가 서버를 만들 때 `/swapfile` 2.3GB를 넣어 둔다.** 접속 화면의 `Swap usage: 0%`는 "없다"가 아니라 "안 쓴다"다. 새로 만들려던 `fallocate`는 `Text file busy`로 실패했다(쓰는 중인 파일) — 바뀐 것은 없다.
- 2.3GB를 그대로 쓴다(디스크 여유 16G). swappiness 10: 메모리가 정말 모자랄 때만 디스크로 내려 DB가 느려지지 않게.

### 4.8 앱 전용 계정 `cbt` · 폴더 (2026-10-31, 노트북에서)
```bash
sudo useradd --system --user-group --no-create-home --home-dir /nonexistent --shell /usr/sbin/nologin cbt
sudo install -d -m 755 /opt/cbt    # 실행 파일(releases/ · current) — 하위 폴더는 release.sh가 만든다
sudo install -d -m 755 /etc/cbt    # cbt.env(640 root:cbt) · backup.pub — 권한은 파일마다
sudo install -d -m 700 /srv/cbt    # 운영 compose · DB 비밀번호 .env — 폴더째 root만
```
- 결과: `uid=999(cbt) gid=987(cbt)`, `getent passwd cbt` → `…:/nonexistent:/usr/sbin/nologin`, `sudo su - cbt` → `This account is currently not available.`(로그인 거부가 정상, 앞의 `cannot change directory to /nonexistent` 경고도 정상).
- 폴더 주인은 전부 root — 앱(`cbt`)은 읽기만 한다. 앱은 파일을 쓰지 않는다(로그는 journald).

### 4.9 비밀값 · 설정 파일 (2026-10-31, 노트북에서)
PC에서 원본을 보낸다:
```powershell
scp deploy/compose.prod.yaml deploy/compose.env.example deploy/cbt.env.example cbt:~/
```
서버에서 제자리에 놓고, 비밀번호를 만들어 **화면에 띄우지 않고** 바로 채운다(같은 터미널에서 이어서):
```bash
sudo install -m 644 -o root -g root ~/compose.prod.yaml   /srv/cbt/compose.yaml
sudo install -m 600 -o root -g root ~/compose.env.example /srv/cbt/.env
sudo install -m 640 -o root -g cbt  ~/cbt.env.example     /etc/cbt/cbt.env
PG=$(openssl rand -hex 32); VK=$(openssl rand -hex 32)
sudo sed -i "s/^POSTGRES_PASSWORD=$/POSTGRES_PASSWORD=$PG/; s/^VALKEY_PASSWORD=$/VALKEY_PASSWORD=$VK/" /srv/cbt/.env
sudo sed -i "s/<POSTGRES_PASSWORD>/$PG/; s/<VALKEY_PASSWORD>/$VK/" /etc/cbt/cbt.env
unset PG VK
rm ~/compose.prod.yaml ~/compose.env.example ~/cbt.env.example
```
확인 — 내용은 보지 않고 길이 · 일치만(`cat`으로 열지 않는다):
```bash
sudo awk -F= '/PASSWORD/ {print $1, length($2)}' /srv/cbt/.env      # 둘 다 64
sudo grep -n '<' /etc/cbt/cbt.env                                    # 2번 줄 주석 하나만(자리표 남지 않음)
sudo bash -c 'source /srv/cbt/.env; grep -q ":$POSTGRES_PASSWORD@" /etc/cbt/cbt.env && grep -q ":$VALKEY_PASSWORD@" /etc/cbt/cbt.env && echo 두 파일 비밀번호 일치'
sudo ls -la /srv/cbt /etc/cbt     # .env -rw------- root root / compose.yaml -rw-r--r-- / cbt.env -rw-r----- root cbt
sudo docker compose -f /srv/cbt/compose.yaml config --quiet && echo compose 설정 OK
```
- 비밀번호는 서버 두 파일에만 있다. 잃어도 서버에서 다시 만들면 된다(DB 비밀번호를 바꾸려면 Postgres 안에서도 바꿔야 한다 — 볼륨을 처음 만들 때 한 번만 정해지므로).
- `ls -l`은 점으로 시작하는 `.env`를 숨긴다 → `-a`.

### 4.10 운영 Postgres · Valkey (2026-10-31, 노트북에서)
```bash
sudo docker compose -f /srv/cbt/compose.yaml up -d
sudo docker compose -f /srv/cbt/compose.yaml ps                    # 둘 다 Up (healthy), 127.0.0.1:5432 · 127.0.0.1:6379
sudo docker compose -f /srv/cbt/compose.yaml exec valkey valkey-cli ping   # PONG (VALKEYCLI_AUTH로 비밀번호)
sudo docker inspect --format '{{.Name}} {{.HostConfig.LogConfig.Config}}' $(sudo docker compose -f /srv/cbt/compose.yaml ps -q)   # max-file:3 max-size:10m
sudo ss -tlnp | grep -E ':5432|:6379'                               # 127.0.0.1만 LISTEN
sudo docker compose -f /srv/cbt/compose.yaml exec -T postgres psql -U cbt -d cbt_journal -tAc 'show server_version;'   # 18.6
```
- `psql`은 `exec -T`(TTY 없이)로 부른다. TTY를 붙인 `exec … -c 'select version();'`은 아무것도 출력하지 않았다.
- 이미지 `postgres:18` · `valkey/valkey:9`, 볼륨 `cbt_pgdata` · `cbt_valkeydata`.
- 바깥에서 막혔는지 (PC에서, IP는 ssh 별명에서 읽어 저장소 · 화면에 적지 않는다):
  ```powershell
  $ip = ((ssh -G cbt | Select-String '^hostname ') -split ' ')[1]
  Test-NetConnection $ip -Port 5432 | Select-Object RemotePort, TcpTestSucceeded   # False (6379도 False, 22는 True)
  ```

### 4.11 DNS (2026-10-31, Cloudflare 화면에서)
- `sagming.com` → DNS → 레코드 추가: 형식 `A` · 이름 `cbt` · IPv4 서버 IP · 프록시 **끔(회색 구름, DNS 전용)** · TTL 자동.
- 확인 (PC): `nslookup cbt.sagming.com 1.1.1.1`과 `nslookup cbt.sagming.com`(지금 네트워크의 DNS) 둘 다 서버 IP.
- 프록시를 끄는 이유: 켜면 Cloudflare가 중간에 끼어 Caddy 인증서 발급 · 실제 접속 IP 규칙(ARCHITECTURE 3.3절)이 꼬인다. Cloudflare 화면의 "프록싱 필요" 안내 · `www` · 루트 도메인 권장 사항은 따르지 않는다(쓰는 이름은 `cbt` 하나).

### 4.12 Caddy (2026-10-31, 노트북에서)
설치 — [공식 문서](https://caddyserver.com/docs/install)의 apt 저장소 그대로:
```bash
sudo apt install --yes debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo chmod o+r /usr/share/keyrings/caddy-stable-archive-keyring.gpg
sudo chmod o+r /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy           # 2.11.7, 설치하자마자 systemd 서비스 caddy로 켜진다
```
설정 — PC에서 `scp deploy/Caddyfile cbt:~/` 한 뒤 서버에서:
```bash
sudo install -m 644 -o root -g root ~/Caddyfile /etc/caddy/Caddyfile && rm ~/Caddyfile
sudo caddy validate --config /etc/caddy/Caddyfile --adapter caddyfile   # Valid configuration
sudo systemctl reload caddy
sudo journalctl -u caddy --since "5 min ago" --no-pager | grep -i 'certificate obtained'
```
확인 (PC):
```powershell
curl.exe -sI https://cbt.sagming.com    # 앱 배포 전엔 502 Bad Gateway (인증서 오류 없이 응답 = HTTPS 됨)
curl.exe -sI http://cbt.sagming.com     # 308 → Location: https://cbt.sagming.com/
```
- 인증서: `CN=cbt.sagming.com`, 발급자 Let's Encrypt(`YE2`), TLS 1.3, 만료 2027-01-29 — 갱신은 Caddy가 알아서 한다.
- IP로: `https://IP`는 TLS 연결 실패, `http://IP`는 `https://IP`로 308 후 실패 → 앱은 안 보인다.
- **인증서를 받자마자 20초 안에 스캐너가 들어왔다**(`/.env` · `/phpinfo.php` · `/wp-admin/…` · `/database.sql`). 새 인증서는 공개 목록(Certificate Transparency)에 올라가고 봇이 그 목록을 지켜본다. 앱이 없어 전부 502였다. 정상적인 인터넷 소음이다 — 그래서 기본 잠금 · 요청 제한이 먼저 있어야 했다.
- Caddy는 접근 로그를 안 남기지만, **오류(502 등) 때는 `journalctl -u caddy`에 접속 IP · 경로 · 헤더를 남긴다**. 쿠키 · `Authorization`은 Caddy가 기본으로 가린다. 경로에는 id만 있고 본문은 안 남아서 REQUIREMENTS 5.1절과 맞는다.

### 4.13 백업 열쇠 `age` (2026-10-31, 노트북에서)
PC에서 열쇠 한 쌍을 만든다(저장소 밖):
```powershell
age-keygen -o "$HOME\cbt-backup.key"     # 화면의 Public key: age1… = 공개키(비밀 아님)
age-keygen -y "$HOME\cbt-backup.key"     # 공개키를 다시 보고 싶을 때
```
서버에는 공개키만:
```bash
sudo apt install --yes age               # 1.1.1 (PC 1.3.2와 호환)
echo 'age1…' | sudo tee /etc/cbt/backup.pub
sudo chmod 644 /etc/cbt/backup.pub
```
짝 시험 — 서버에서 잠그고 PC에서 연다:
```bash
echo '열쇠 짝 시험' | age -R /etc/cbt/backup.pub -o ~/keytest.age    # 서버
```
```powershell
scp cbt:~/keytest.age $env:TEMP\
age -d -i "$HOME\cbt-backup.key" "$env:TEMP\keytest.age"     # 열쇠 짝 시험
Remove-Item "$env:TEMP\keytest.age"
```
- **개인키(`AGE-SECRET-KEY-1…`)는 서버 · 저장소 · 대화 · 클라우드에 두지 않는다.** 보관: 노트북 `$HOME\cbt-backup.key` + 종이. 백업을 열 PC마다 같은 파일을 `$HOME\cbt-backup.key`에 둔다(USB로 옮기고 USB에서는 지운다).
- 잃으면 모든 백업을 못 연다 → 새 열쇠를 만들어 `backup.pub`를 바꾸고, 그 전 백업은 버린다.
- `age-keygen`의 `writing secret key to a world-readable file` 경고는 Unix 권한 기준이다. Windows 홈 폴더는 기본으로 본인 · 관리자만 읽는다.

### 4.14 운영 파일 · 백업 타이머 (2026-10-31, 노트북에서)
PC에서 `scp deploy/cbtctl deploy/cbt-backup deploy/cbt.service deploy/cbt-backup.service deploy/cbt-backup.timer cbt:~/` 한 뒤 서버에서:
```bash
sudo install -m 755 -o root -g root ~/cbtctl ~/cbt-backup /usr/local/bin/
sudo install -m 644 -o root -g root ~/cbt.service ~/cbt-backup.service ~/cbt-backup.timer /etc/systemd/system/
rm ~/cbtctl ~/cbt-backup ~/cbt.service ~/cbt-backup.service ~/cbt-backup.timer
sudo systemctl daemon-reload
sudo systemd-analyze verify /etc/systemd/system/cbt.service /etc/systemd/system/cbt-backup.service /etc/systemd/system/cbt-backup.timer
sudo systemctl start cbt-backup                     # 손으로 한 번
sudo journalctl -u cbt-backup -n 10 --no-pager      # [backup] 만듦: /var/backups/cbt/cbt-…Z.dump.age
sudo systemctl enable --now cbt-backup.timer
systemctl list-timers cbt-backup.timer --no-pager   # NEXT 19:0x UTC = 04:0x KST
```
- 첫 배포 전에는 `verify`가 `cbt.service: Command /opt/cbt/current/app is not executable`를 낸다 — 실행 파일은 첫 `task deploy`가 넣는다. 나머지 두 파일은 출력 없음(정상).
- `cbt.service`는 여기서 켜지 않는다. 첫 배포의 `release.sh`가 `enable` · `restart` 한다.
- 결과: 빈 DB 백업 1096바이트(`-rw-r--r-- root`), PC에서 `task backup:pull` → `~/cbt-backups/`, 개인키로 풀어 896바이트 · 머리글 `PGDMP`(pg_dump 형식).

### 4.15 첫 배포 · 내 계정 (2026-10-31, 노트북에서)
```powershell
task deploy     # PC에서. passphrase 2번(복사 · 실행) + sudo 비밀번호 1번
```
```bash
sudo cbtctl user create --email 내이메일 --nickname 내닉네임   # 서버에서. 비밀번호는 실행 중에 묻는다
systemctl is-active cbt                                        # active
sudo systemd-analyze verify /etc/systemd/system/cbt.service    # 출력 없음 (4.14절의 "실행 파일 없음"이 사라짐)
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
