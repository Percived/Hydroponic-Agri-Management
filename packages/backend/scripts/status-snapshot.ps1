param(
    [string]$OutputPath = "docs/PROJECT_STATUS_SNAPSHOT.md"
)

$ErrorActionPreference = "Stop"

function Get-Section($title, $content) {
    "## $title`n$content`n"
}

$now = Get-Date -Format "yyyy-MM-dd HH:mm:ss zzz"

$files = rg --files 2>$null
if (-not $files) {
    $files = Get-ChildItem -Recurse -File | ForEach-Object { $_.FullName }
}

$totalFiles = ($files | Measure-Object).Count
$goFiles = ($files | Where-Object { $_ -match "\.go$" } | Measure-Object).Count
$testFiles = ($files | Where-Object { $_ -match "_test\.go$" } | Measure-Object).Count

$routes = rg -n "RegisterRoutes\(|\.GET\(|\.POST\(|\.PUT\(|\.PATCH\(|\.DELETE\(" internal 2>$null
if (-not $routes) { $routes = "(no route scan output)" }

$migrations = Get-ChildItem migrations -File -ErrorAction SilentlyContinue |
    Sort-Object Name |
    Select-Object -ExpandProperty Name
if (-not $migrations) { $migrations = "(none)" }

$todos = rg -n "TODO|FIXME" internal cmd docs migrations scripts 2>$null
if (-not $todos) { $todos = "(none)" }

$gitStatus = git status --short 2>$null
if (-not $gitStatus) { $gitStatus = "(clean or git unavailable)" }

$content = @()
$content += "# PROJECT STATUS SNAPSHOT"
$content += ""
$content += "Generated At: $now"
$content += ""
$content += Get-Section "Repository Stats" (@"
- Total files: $totalFiles
- Go files: $goFiles
- Test files: $testFiles
"@)

$content += Get-Section "Migrations" (($migrations -join "`n"))
$content += Get-Section "Route/Handler Scan (raw)" (($routes -join "`n"))
$content += Get-Section "TODO/FIXME Scan" (($todos -join "`n"))
$content += Get-Section "Git Status" (($gitStatus -join "`n"))

$content -join "`n" | Set-Content -Path $OutputPath -Encoding UTF8
Write-Output "Snapshot written to $OutputPath"
