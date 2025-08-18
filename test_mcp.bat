@echo off
echo Testing MCP Android Devices Server...
echo.

echo Building server...
go build
if %ERRORLEVEL% neq 0 (
    echo Build failed!
    pause
    exit /b 1
)
echo Build successful!
echo.

echo === Test 1: Initialize ===
echo Request: {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}
echo {"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}} | mcp_android_devices.exe
echo.

echo === Test 2: Tools List ===
echo Request: {"jsonrpc":"2.0","id":2,"method":"tools/list"}
echo {"jsonrpc":"2.0","id":2,"method":"tools/list"} | mcp_android_devices.exe
echo.

echo === Test 3: Call Android Devices Tool ===
echo Request: {"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_android_devices","arguments":{}}}
echo {"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_android_devices","arguments":{}}} | mcp_android_devices.exe
echo.

echo === Test 4: Error Case - Unknown Tool ===
echo Request: {"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}}
echo {"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"unknown_tool","arguments":{}}} | mcp_android_devices.exe
echo.

echo Testing Complete!
pause
