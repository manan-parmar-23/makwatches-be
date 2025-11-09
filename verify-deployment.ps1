param(
    [string]$Server = "139.59.71.95",
    [int]$Port = 8080
)

Write-Host ""
Write-Host "=== MakWatches Deployment Verification ==="
Write-Host ""
Write-Host "Testing: ${Server}:${Port}"
Write-Host ""

# Test 1
Write-Host "Test 1: Health Check..."
try {
    $r = Invoke-WebRequest -Uri "http://${Server}:${Port}/health" -UseBasicParsing -TimeoutSec 10
    Write-Host "Health OK: $($r.StatusCode) $($r.Content)"
}
catch {
    Write-Host "Health endpoint FAIL: $($_.Exception.Message)"
}
Write-Host ""

# Test 2
Write-Host "Test 2: Welcome Check..."
try {
    $r = Invoke-WebRequest -Uri "http://${Server}:${Port}/" -UseBasicParsing -TimeoutSec 10
    Write-Host "Welcome OK: $($r.StatusCode) $($r.Content)"
}
catch {
    Write-Host "Welcome FAIL: $($_.Exception.Message)"
}
Write-Host ""

# Test 3
Write-Host "Test 3: Products..."
try {
    $r = Invoke-WebRequest -Uri "http://${Server}:${Port}/api/v1/products?limit=1" -UseBasicParsing -TimeoutSec 10
    Write-Host "Products OK: $($r.StatusCode)"
}
catch {
    Write-Host "Products FAIL: $($_.Exception.Message)"
}
Write-Host ""

Write-Host "DONE"
