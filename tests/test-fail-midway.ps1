$body = @{
    steps = @(
        @{
            action = "payment"
            params = @{
                recipient = "0xEaF9A3648c1c5C7Aa194AAb84C112eFC0443964C"
                amount = "10000000000000" # 0.00001 TCRO (Should Succeed)
            }
        },
        @{
            action = "payment"
            params = @{
                recipient = "0xEaF9A3648c1c5C7Aa194AAb84C112eFC0443964C"
                amount = "1000000000000000000000" # 1000 TCRO (Should Fail - Insufficient Funds)
            }
        }
    )
} | ConvertTo-Json -Depth 5

Write-Host "🚀 Sending Fail-Midway Intent..." -ForegroundColor Cyan

try {
    $response = Invoke-RestMethod -Method Post -Uri "http://localhost:8081/intent" -Body $body -ContentType "application/json" -ErrorAction Stop
    Write-Host "✅ Response Received (Unexpected Success):" -ForegroundColor Green
    $response | Format-List
} catch {
    Write-Host "⚠️  Response Received (Expected Failure/Partial):" -ForegroundColor Yellow
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        $respContent = $reader.ReadToEnd()
        Write-Host "Server Response: $respContent" -ForegroundColor Yellow
        
        # Try to parse JSON to show details nicely
        try {
            $json = $respContent | ConvertFrom-Json
            Write-Host "Status: $($json.status)"
            Write-Host "Message: $($json.message)"
            Write-Host "Failed Step: $($json.failed_step_index)"
            Write-Host "Successful Hashes: $($json.tx_hashes -join ', ')"
            Write-Host "Error: $($json.error)"
        } catch {}
    } else {
        Write-Host "Error: $_" -ForegroundColor Red
    }
}
