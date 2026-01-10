$body = @{
    action = "payment"
    params = @{
        recipient = "0xEaF9A3648c1c5C7Aa194AAb84C112eFC0443964C"
        amount = "100000000000000"
    }
} | ConvertTo-Json

try {
    $response = Invoke-RestMethod -Method Post -Uri "http://localhost:8081/simulate" -Body $body -ContentType "application/json" -ErrorAction Stop
    Write-Host "✅ Response Received:" -ForegroundColor Green
    $response | Format-List
} catch {
    Write-Host "❌ Error: $_" -ForegroundColor Red
    if ($_.Exception.Response) {
        $reader = New-Object System.IO.StreamReader($_.Exception.Response.GetResponseStream())
        Write-Host "Server Response: $($reader.ReadToEnd())" -ForegroundColor Red
    }
}
