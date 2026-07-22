# Start the Mosaic dev stack on the host, in one command.
#
#   pwsh -File scripts/dev.ps1
#   pwsh -File scripts/dev.ps1 -Stop
#
# This is the no-Docker path, and it exists because Docker is not always
# available (a broken WSL registration, no admin rights on the machine). Prefer
# docker-compose.dev.yml when you can: it also supplies ffmpeg, which this script
# can only check for.
#
# It starts the Platform and the web Shell as background jobs, waits until each
# is actually answering rather than merely launched, and prints where things are.
[CmdletBinding()]
param(
    # Stop whatever a previous run started, and nothing else.
    [switch]$Stop,
    [string]$Dsn = $env:MOSAIC_POSTGRES_DSN,
    [int]$ShellPort = 5173
)

$ErrorActionPreference = "Stop"
$root = Split-Path -Parent $PSScriptRoot          # ...\platform
$mosaic = Split-Path -Parent $root                # ...\Mosaic
$pidFile = Join-Path $root ".dev-pids"

function Stop-Stack {
    if (-not (Test-Path $pidFile)) { "nothing recorded as running"; return }
    foreach ($line in Get-Content $pidFile) {
        $parts = $line -split "="
        $p = Get-Process -Id ([int]$parts[1]) -ErrorAction SilentlyContinue
        if ($p) { Stop-Process -Id $p.Id -Force; "stopped $($parts[0]) (pid $($p.Id))" }
    }
    Remove-Item $pidFile -Force
}

if ($Stop) { Stop-Stack; return }

# --- preconditions, checked rather than assumed -----------------------------
if (-not $Dsn) {
    # Fall back to the gitignored dev env file if it has one.
    $envFile = Join-Path $root ".env.local"
    if (Test-Path $envFile) {
        $line = Select-String -Path $envFile -Pattern '^MOSAIC_POSTGRES_DSN=' -ErrorAction SilentlyContinue
        if ($line) { $Dsn = ($line.Line -split '=', 2)[1].Trim('"') }
    }
}
if (-not $Dsn) { throw "No PostgreSQL DSN. Set MOSAIC_POSTGRES_DSN or put it in platform/.env.local" }

if (-not (Get-Command ffmpeg -ErrorAction SilentlyContinue)) {
    Write-Warning "ffmpeg is not on PATH. The Platform will run in direct-play-only mode, so most real releases will not play (ADR 0050). Install it with: winget install --id Gyan.FFmpeg -e"
}

Stop-Stack | Out-Null
$pids = @()

# --- Platform ---------------------------------------------------------------
$env:MOSAIC_POSTGRES_DSN = $Dsn
if (-not $env:MOSAIC_BOOTSTRAP_ADMIN_USERNAME) { $env:MOSAIC_BOOTSTRAP_ADMIN_USERNAME = "admin" }
if (-not $env:MOSAIC_BOOTSTRAP_ADMIN_PASSWORD) { $env:MOSAIC_BOOTSTRAP_ADMIN_PASSWORD = "admin" }

"starting platform..."
$platform = Start-Process -FilePath "go" -ArgumentList "run", "./cmd/mosaic-platform" `
    -WorkingDirectory $root -PassThru -WindowStyle Hidden `
    -RedirectStandardOutput (Join-Path $root "logs\dev-platform.out.log") `
    -RedirectStandardError  (Join-Path $root "logs\dev-platform.err.log")
$pids += "platform=$($platform.Id)"

# Wait for it to answer, not merely to have been launched — `go run` compiles
# first, so the process existing says nothing about the port being open.
$ready = $false
foreach ($i in 1..90) {
    Start-Sleep -Seconds 1
    try {
        Invoke-WebRequest -Uri "http://localhost:8081/graphql" -Method POST -Body '{"query":"{__typename}"}' `
            -ContentType "application/json" -TimeoutSec 3 -UseBasicParsing | Out-Null
        $ready = $true; break
    } catch { }
}
if (-not $ready) { throw "platform did not become ready; see logs\dev-platform.err.log" }
"platform ready on :8081 (health :8080)"

# --- Web shell --------------------------------------------------------------
"starting web shell..."
$env:MOSAIC_PLATFORM_URL = "http://localhost:8081"
$shellDir = Join-Path $mosaic "web\packages\shell"
$web = Start-Process -FilePath "npm" -ArgumentList "run", "dev", "--", "--port", "$ShellPort", "--strictPort" `
    -WorkingDirectory $shellDir -PassThru -WindowStyle Hidden `
    -RedirectStandardOutput (Join-Path $root "logs\dev-web.out.log") `
    -RedirectStandardError  (Join-Path $root "logs\dev-web.err.log")
$pids += "web=$($web.Id)"

$ready = $false
foreach ($i in 1..60) {
    Start-Sleep -Seconds 1
    try {
        Invoke-WebRequest -Uri "http://localhost:$ShellPort/" -TimeoutSec 3 -UseBasicParsing | Out-Null
        $ready = $true; break
    } catch { }
}
if (-not $ready) { Write-Warning "web shell did not answer on :$ShellPort; see logs\dev-web.err.log" }

$pids | Set-Content -Path $pidFile -Encoding utf8

""
"  Shell      http://localhost:$ShellPort"
"  Platform   http://localhost:8081  (GraphQL, session, /artwork, /playback)"
"  Handoff    http://localhost:8080"
"  Logs       platform/logs/dev-*.log"
""
"  stop with: pwsh -File scripts/dev.ps1 -Stop"
