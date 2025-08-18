# MCP Android Devices Server

A Model Context Protocol (MCP) server that provides information about connected Android devices and emulators. This server implements the official MCP specification using JSON-RPC 2.0 over stdio transport.

## Features

- Lists all connected Android devices and emulators
- Provides detailed device information (name, model, architecture, Android version, SDK level)
- Follows the official MCP protocol specification
- Uses JSON-RPC 2.0 over stdio transport
- Proper error handling and protocol compliance

## How to use

### Build the server

```bash
go build
```

### Configure with MCP Clients

#### Cursor IDE - Step by Step Installation

1. **Build the server executable:**
   ```bash
   go build
   ```

2. **Copy the executable to a permanent location:**
   ```bash
   # Windows
   mkdir C:\tools\mcp_servers
   copy mcp_android_devices.exe C:\tools\mcp_servers\
   
   # macOS/Linux
   sudo mkdir -p /usr/local/bin/mcp_servers
   sudo cp mcp_android_devices /usr/local/bin/mcp_servers/
   ```

3. **Open Cursor IDE Settings:**
   - Press `Ctrl+,` (Windows/Linux) or `Cmd+,` (macOS) to open Settings
   - Or go to `File > Preferences > Settings`

4. **Navigate to MCP Settings:**
   - In the Settings search bar, type "MCP"
   - Look for "MCP Servers" section
   - Click "Edit in settings.json" or find the MCP configuration area

5. **Add the MCP server configuration:**
   
   **For Windows:**
   ```json
   {
       "mcp": {
           "servers": {
               "android_devices": {
                   "command": "C:\\tools\\mcp_servers\\mcp_android_devices.exe"
               }
           }
       }
   }
   ```
   
   **For macOS/Linux:**
   ```json
   {
       "mcp": {
           "servers": {
               "android_devices": {
                   "command": "/usr/local/bin/mcp_servers/mcp_android_devices"
               }
           }
       }
   }
   ```

6. **Alternative: Use relative path (if keeping in project folder):**
   ```json
   {
       "mcp": {
           "servers": {
               "android_devices": {
                   "command": "./mcp_android_devices"
               }
           }
       }
   }
   ```

7. **Save and restart Cursor IDE**

8. **Verify installation:**
   - Open Cursor IDE
   - Look for MCP status indicator in the status bar
   - Try asking: "List my Android devices" or "Show connected Android emulators"
   - The AI should now be able to use the Android devices tool

#### Troubleshooting Cursor IDE Integration

**Common Issues:**

1. **Server not found:**
   - Verify the executable path is correct
   - Use absolute paths instead of relative paths
   - Check file permissions (executable bit on Unix systems)

2. **ADB not found:**
   - Ensure Android SDK is installed
   - Add `adb` to your system PATH
   - Test with `adb devices` in terminal

3. **Permission issues (Windows):**
   - Run Cursor IDE as administrator (temporarily)
   - Or move executable to a user-accessible directory

4. **Settings not taking effect:**
   - Completely restart Cursor IDE
   - Check for syntax errors in settings.json
   - Look at Cursor IDE logs/console for error messages

#### Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
    "mcpServers": {
        "android_devices": {
            "command": "./mcp_android_devices"
        }
    }
}
```

### Test the server manually

The server communicates via JSON-RPC 2.0 over stdin/stdout. Here are some test examples:

1. **Initialize the connection:**
   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | ./mcp_android_devices
   ```

2. **List available tools:**
   ```bash
   echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | ./mcp_android_devices
   ```

3. **Call the get_android_devices tool:**
   ```bash
   echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_android_devices","arguments":{}}}' | ./mcp_android_devices
   ```

## MCP Protocol Examples

### Initialize Response
```json
{
    "jsonrpc": "2.0",
    "id": 1,
    "result": {
        "protocolVersion": "2024-11-05",
        "capabilities": {
            "tools": {
                "listChanged": true
            }
        },
        "serverInfo": {
            "name": "android-devices-mcp-server",
            "version": "1.0.0"
        }
    }
}
```

### Tools List Response
```json
{
    "jsonrpc": "2.0",
    "id": 2,
    "result": {
        "tools": [
            {
                "name": "get_android_devices",
                "description": "Get a list of connected Android devices and emulators",
                "inputSchema": {
                    "type": "object",
                    "properties": {}
                }
            }
        ]
    }
}
```

### Tool Call Response
```json
{
    "jsonrpc": "2.0",
    "id": 3,
    "result": {
        "content": [
            {
                "type": "text",
                "text": "[{\"name\":\"Pixel 2 API 30\",\"device\":\"emulator-5554\",\"model\":\"sdk_gphone_x86\",\"arch\":\"x86\",\"android_version\":\"11\",\"sdk_level\":\"30\",\"run_status\":\"device\"}]"
            }
        ],
        "isError": false
    }
}
```

## Requirements

- Go 1.19 or later
- Android SDK with `adb` in PATH
- Connected Android devices or running emulators