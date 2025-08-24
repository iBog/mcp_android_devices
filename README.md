# MCP Android Devices Server

A Model Context Protocol (MCP) server that provides information about connected Android devices and emulators. This server implements the official MCP specification using JSON-RPC 2.0 over stdio transport.

## Features

- Lists all connected Android devices and emulators
- Provides detailed device information (name, model, architecture, Android version, SDK level)
- Captures screenshots from Android devices and emulators
- Returns screenshots as Base64-encoded PNG images
- Follows the official MCP protocol specification
- Uses JSON-RPC 2.0 over stdio transport
- Proper error handling and protocol compliance
- Cross-platform support (Windows, macOS, Linux)

## How to use

### Install Dependencies

```bash
npm install
```

### Configure with MCP Clients

#### Cursor IDE - Step by Step Installation

1. **Open Cursor IDE Settings:**
   - Press `Ctrl+,` (Windows/Linux) or `Cmd+,` (macOS) to open Settings
   - Or go to `File > Preferences > Settings`

2. **Navigate to MCP Settings:**
   - In the Settings search bar, type "MCP"
   - Look for "MCP Servers" section
   - Click "Edit in settings.json" or find the MCP configuration area

3. **Add the MCP server configuration:**

   ```json
   {
       "mcp": {
           "servers": {
               "android_devices": {
                   "command": "node /path/to/your/project/index.js"
               }
           }
       }
   }
   ```

   **Note:** Replace `/path/to/your/project/` with the actual absolute path to the project directory.

4. **Save and restart Cursor IDE**

5. **Verify installation:**
   - Open Cursor IDE
   - Look for MCP status indicator in the status bar
   - Try asking: "List my Android devices" or "Show connected Android emulators"
   - Try asking: "Take a screenshot of my Android device" or "Capture the screen from my emulator"
   - The AI should now be able to use both the device listing and screenshot tools

#### Troubleshooting Cursor IDE Integration

**Common Issues:**

1. **Server not found:**
   - Verify the path to `index.js` is correct
   - Use absolute paths instead of relative paths

2. **ADB not found:**
   - Ensure Android SDK is installed
   - Add `adb` to your system PATH
   - Test with `adb devices` in terminal

3. **Permission issues (Windows):**
   - Run Cursor IDE as administrator (temporarily)

4. **Settings not taking effect:**
   - Completely restart Cursor IDE
   - Check for syntax errors in settings.json
   - Look at Cursor IDE logs/console for error messages

5. **Screenshot functionality not working:**
   - Ensure your Android device/emulator screen is unlocked
   - Verify the device is properly connected with `adb devices`
   - Check that the device has sufficient storage space
   - Some devices may require enabling "USB Debugging" and "Disable permission monitoring"

#### Claude Desktop

Add to your `claude_desktop_config.json`:

```json
{
    "mcpServers": {
        "android_devices": {
            "command": "node /path/to/your/project/index.js"
        }
    }
}

   **Note:** Replace `/path/to/your/project/` with the actual absolute path to the project directory.
```

### Test the server manually

The server communicates via JSON-RPC 2.0 over stdin/stdout. Here are some test examples:

1. **Initialize the connection:**

   ```bash
   echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | node index.js
   ```

2. **List available tools:**

   ```bash
   echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | node index.js
   ```

3. **Call the get_android_devices tool:**

   ```bash
   echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_android_devices","arguments":{}}}' | node index.js
   ```

4. **Capture a screenshot from an Android device:**

   ```bash
   echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_android_screen","arguments":{"device":"emulator-5554"}}}' | node index.js
   ```

   Or capture from the first available device:

   ```bash
   echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_android_screen","arguments":{}}}' | node index.js
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
            },
            {
                "name": "get_android_screen",
                "description": "Capture a screenshot from an Android device",
                "inputSchema": {
                    "type": "object",
                    "properties": {
                        "device": {
                            "type": "string",
                            "description": "Device name (e.g., 'emulator-5554'). If not provided, uses the first available device."
                        }
                    }
                }
            }
        ]
    }
}
```

### Device List Tool Call Response

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

### Screenshot Tool Call Response

```json
{
    "jsonrpc": "2.0",
    "id": 4,
    "result": {
        "content": [
            {
                "type": "image",
                "data": "iVBORw0KGgoAAAANSUhEUgAAAoAAAAHgCAYAAAA10dzkAAA...(base64 encoded PNG data)...==",
                "mimeType": "image/png"
            }
        ],
        "isError": false
    }
}
```

**Note:** The `data` field contains the complete Base64-encoded PNG image. The actual response will contain the full Base64 string, which has been truncated in this example for readability.

## Requirements

- Node.js v14 or later
- Android SDK with `adb` in PATH
- Connected Android devices or running emulators