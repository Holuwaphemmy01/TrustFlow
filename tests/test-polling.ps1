$body = @{
    steps = @(
        @{
            action = "payment"
            params = @{
                recipient = "0xEaF9A3648c1c5C7Aa194AAb84C112eFC0443964C"
                amount = "100"
            }
        }
    )
} | ConvertTo-Json -Depth 5

Write-Host "Submitting Intent..." -ForegroundColor Cyan
$response = Invoke-RestMethod -Method Post -Uri "http://localhost:8081/intent" -Body $body -ContentType "application/json"
$id = $response.intent_id
Write-Host "Intent Submitted! ID: $id" -ForegroundColor Green

Write-Host "Checking Status via Polling API..." -ForegroundColor Cyan
$status = Invoke-RestMethod -Method Get -Uri "http://localhost:8081/status/$id"
$status | Format-List

Write-Host "Step Details:" -ForegroundColor Cyan
$status.steps | Format-Table -AutoSize
