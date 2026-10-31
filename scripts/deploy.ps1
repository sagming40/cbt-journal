# deploy.ps1 ─ PC 쪽 배포: 확인 → 빌드 → 보낼 폴더 만들기 → scp 한 번 → ssh(sudo) 한 번 (ARCHITECTURE 7.3절)
# task deploy로 부른다. 서버 쪽 일은 deploy/release.sh가 한다.
# 비유: 택배 보내기 ─ 물건이 맞는지 확인하고(깨끗한 main), 포장해서(빌드 + 폴더), 한 번에 부치고(scp),
#       받는 쪽에 "열어서 설치해 줘"라고 한 번 전화한다(ssh). 설치 · 확인 · 되돌리기는 받는 쪽이 한다.
#
# 쓰는 법:
#   task deploy                  배포
#   task deploy -- -DryRun       확인 · 빌드 · 폴더 만들기까지만 하고 서버에는 안 보낸다
#   task deploy -- -SshHost 별명  ~/.ssh/config의 다른 별명으로 (기본 cbt, SERVER.md 3.2절)
#
# 이 파일은 UTF-8(BOM 있음) + CRLF여야 한다. Windows PowerShell 5.1은 BOM이 없으면 한글을 CP949로 읽어 깨진다.

param(
	[string]$SshHost = 'cbt',
	[switch]$DryRun
)

$ErrorActionPreference = 'Stop'
# 저장소 맨 위에서 돈다 (task가 거기서 부르지만, 직접 실행해도 같게).
Set-Location (Split-Path -Parent $PSScriptRoot)

function Fail([string]$msg) {
	Write-Host "[deploy] 실패: $msg" -ForegroundColor Yellow
	exit 1
}
function Log([string]$msg) { Write-Host "[deploy] $msg" }

# 바깥 프로그램(git · task · scp · ssh)은 실패해도 PowerShell이 멈추지 않는다 → 종료 코드를 직접 본다.
function Check([string]$what) {
	if ($LASTEXITCODE -ne 0) { Fail "$what (종료 코드 $LASTEXITCODE)" }
}

# ── 1. 확인: 커밋 안 한 변경이 없고, 지금 커밋이 origin/main과 같아야 한다 ──
# 두 PC가 push 안 한 서로 다른 코드를 번갈아 올리는 일을 막는다.
git fetch --quiet origin main; Check 'git fetch'
$dirty = git status --porcelain; Check 'git status'
if ($dirty) {
	Fail "커밋 안 한 변경이 있다 (새 파일 포함) ─ 커밋 · push 뒤 다시`n$($dirty -join "`n")"
}
$head = git rev-parse HEAD; Check 'git rev-parse HEAD'
$origin = git rev-parse origin/main; Check 'git rev-parse origin/main'
if ($head -ne $origin) {
	Fail "지금 커밋($($head.Substring(0, 12)))이 origin/main($($origin.Substring(0, 12)))과 다르다 ─ push 또는 pull 먼저"
}
# build:linux와 같은 방법으로 커밋 번호를 얻는다(보통 12자리, 겹치면 더 길어진다). release.sh가 실행 파일 버전과 맞춰 본다.
$commit = git describe --always --abbrev=12 --exclude='*'; Check 'git describe'
Log "배포할 커밋: $commit"

# release.sh가 CRLF면 서버 bash가 "$'\r': command not found"로 죽는다 (.gitattributes가 LF로 지키지만 한 번 더).
$releaseSh = 'deploy/release.sh'
if ([IO.File]::ReadAllText((Resolve-Path $releaseSh)).Contains("`r")) {
	Fail "$releaseSh 에 CRLF가 섞여 있다"
}

# ── 2. 빌드 ──
task build:linux; Check 'task build:linux'
# 빌드가 추적 중인 파일을 바꿨으면 실행 파일이 커밋과 다를 수 있다 → 다시 본다.
$dirty = git status --porcelain; Check 'git status'
if ($dirty) { Fail "빌드 뒤 작업 트리가 바뀌었다`n$($dirty -join "`n")" }

# ── 3. 보낼 폴더: bin/cbt-deploy-<커밋>/ (bin은 .gitignore) ──
$stageName = "cbt-deploy-$commit"
$stage = "bin/$stageName"
if (Test-Path $stage) { Remove-Item -Recurse -Force $stage }
New-Item -ItemType Directory $stage | Out-Null
Copy-Item 'bin/app-linux-amd64' "$stage/"
Copy-Item $releaseSh "$stage/"
Log "보낼 폴더: $stage"

# 서버에서 한 번에 돌 명령. 성공 · 실패와 상관없이 임시 폴더를 지우고, release.sh의 종료 코드를 그대로 돌려준다.
$remote = "sudo bash ~/$stageName/release.sh $commit; rc=`$?; rm -rf ~/$stageName; exit `$rc"

if ($DryRun) {
	Log "DryRun ─ 여기서 멈춘다. 실제로는:"
	Log "  scp -r $stage ${SshHost}:"
	Log "  ssh -t $SshHost `"$remote`""
	exit 0
}

# ── 4. 복사 1번 · 실행 1번 ──
Log "복사 → ${SshHost}:~/$stageName (열쇠 passphrase를 물으면 입력)"
scp -r $stage "${SshHost}:"; Check 'scp'
# -t: 서버의 sudo가 비밀번호를 물을 수 있게 터미널을 연결한다.
Log "서버에서 release.sh 실행 (passphrase · sudo 비밀번호를 물으면 입력)"
ssh -t $SshHost $remote; Check '서버 쪽 배포 ─ 위 [release] 줄을 본다'

Remove-Item -Recurse -Force $stage
Log "끝: $commit ─ https://cbt.sagming.com"
