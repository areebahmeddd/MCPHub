# Test Python MCP Hub with different project type
Write-Host "Testing Python MCP project..." -ForegroundColor Green

# Test health endpoint first
$healthResult = Invoke-RestMethod -Uri "http://localhost:8080/health" -Method Get
Write-Host "Health check: $($healthResult.status)" -ForegroundColor Green

# Create temp directory with Python test files
$tempDir = "temp-python-test"
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

# Copy Python MCP config
Copy-Item "examples\python-mcp.json" "$tempDir\mcp.json"

# Create sample Python files
@"
from flask import Flask, jsonify
import os

app = Flask(__name__)

@app.route('/health')
def health():
    return jsonify({'status': 'healthy'})

@app.route('/')
def home():
    return jsonify({'message': 'Python Analytics Service'})

if __name__ == '__main__':
    port = int(os.environ.get('PORT', 8000))
    app.run(host='0.0.0.0', port=port)
"@ | Out-File -FilePath "$tempDir\main.py" -Encoding UTF8

@"
flask==2.3.3
pandas==2.0.3
numpy==1.24.3
requests==2.31.0
"@ | Out-File -FilePath "$tempDir\requirements.txt" -Encoding UTF8

# Create ZIP
$pythonZipPath = "python-test.zip"
if (Test-Path $pythonZipPath) { Remove-Item $pythonZipPath }
Compress-Archive -Path "$tempDir\*" -DestinationPath $pythonZipPath -Force

# Test API
try {
    $httpClient = New-Object System.Net.Http.HttpClient
    $form = New-Object System.Net.Http.MultipartFormDataContent
    
    $fileBytes = [System.IO.File]::ReadAllBytes((Resolve-Path $pythonZipPath))
    $fileContent = New-Object System.Net.Http.ByteArrayContent -ArgumentList @(,$fileBytes)
    $fileContent.Headers.ContentType = [System.Net.Http.Headers.MediaTypeHeaderValue]::Parse("application/zip")
    
    $form.Add($fileContent, "zip", "python-test.zip")
    
    $response = $httpClient.PostAsync("http://localhost:8080/api/v1/generate-dockerfile", $form).Result
    $responseContent = $response.Content.ReadAsStringAsync().Result
    
    if ($response.IsSuccessStatusCode) {
        $jsonResponse = $responseContent | ConvertFrom-Json
        
        Write-Host "`nPython Dockerfile Generated:" -ForegroundColor Cyan
        Write-Host $jsonResponse.dockerfile
        
        # Save Dockerfile
        $jsonResponse.dockerfile | Out-File -FilePath "Python-Dockerfile" -Encoding UTF8
        Write-Host "`nPython Dockerfile saved as 'Python-Dockerfile'" -ForegroundColor Green
        
    } else {
        Write-Host "API call failed: $($response.StatusCode)" -ForegroundColor Red
    }
    
} catch {
    Write-Host "Error: $($_.Exception.Message)" -ForegroundColor Red
} finally {
    if ($httpClient) { $httpClient.Dispose() }
    Remove-Item -Path $tempDir -Recurse -Force
    Remove-Item -Path $pythonZipPath -Force
}

Write-Host "`nPython test completed!" -ForegroundColor Green
