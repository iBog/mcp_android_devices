#!/usr/bin/env powershell

# Test script for MCP Android Devices Server
Write-Host "Testing MCP Android Devices Server..." -ForegroundColor Green

# Build the server
Write-Host "`nBuilding server..." -ForegroundColor Yellow
go build
if ($LASTEXITCODE -ne 0) {
    Write-Host "Build failed!" -ForegroundColor Red
    exit 1
}

Write-Host "Build successful!" -ForegroundColor Green

# Test 1: Initialize
Write-Host "`n=== Test 1: Initialize ===" -ForegroundColor Cyan
$initRequest = '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}'
$initResponse = $initRequest | .\mcp_android_devices.exe
Write-Host "Request: $initRequest" -ForegroundColor Gray
Write-Host "Response: $initResponse" -ForegroundColor White

# Test 2: Tools List
Write-Host "`n=== Test 2: Tools List ===" -ForegroundColor Cyan
$listRequest = '{"jsonrpc":"2.0","id":2,"method":"tools/list"}'
$listResponse = $listRequest | .\mcp_android_devices.exe
Write-Host "Request: $listRequest" -ForegroundColor Gray
Write-Host "Response: $listResponse" -ForegroundColor White

# Test 3: Call Tool
Write-Host "`n=== Test 3: Call Android Devices Tool ===" -ForegroundColor Cyan
$callRequest = '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_android_devices","arguments":{}}}'
$callResponse = $callRequest | .\mcp_android_devices.exe
Write-Host "Request: $callRequest" -ForegroundColor Gray
Write-Host "Response: $callResponse" -ForegroundColor White

# Test 4: Error case - Unknown tool
Write-Host "`n=== Test 4: Error Case - Unknown Tool ===" -ForegroundColor Cyan
$errorRequest = '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}'
$errorResponse = $errorRequest | .\mcp_android_devices.exe
Write-Host "Request: $errorRequest" -ForegroundColor Gray
Write-Host "Response: $errorResponse" -ForegroundColor White

# Test 5: Error case - Unknown method
Write-Host "`n=== Test 5: Error Case - Unknown Method ===" -ForegroundColor Cyan
$unknownRequest = '{"jsonrpc":"2.0","id":5,"method":"unknown/method"}'
$unknownResponse = $unknownRequest | .\mcp_android_devices.exe
Write-Host "Request: $unknownRequest" -ForegroundColor Gray
Write-Host "Response: $unknownResponse" -ForegroundColor White

Write-Host "`n=== Testing Complete ===" -ForegroundColor Green
Write-Host "All tests executed. Check responses above for correctness." -ForegroundColor Yellow
