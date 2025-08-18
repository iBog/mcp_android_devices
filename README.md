# MCP Android Emulators Server

This is a simple MCP server that provides a list of currently running Android emulators and devices.

## How to use

### Build the server

1.  **Build the server:**

    ```bash
    go build
    ```

### Configure Cursor IDE

1.  **Open Cursor IDE.**
2.  **Go to `File > Settings > Open User Settings (JSON)`**
3.  **Add the following configuration to your `mcp.json` file:**

    ```json
    {
        "mcpServers": {
            "android_emulators": {
                "command": "./mcp_android_emulators" 
            }
        }
    }
    ```

    **Note:** You might need to use the full path to the `mcp_android_emulators` executable.

4.  **Save the configuration.**

Now you can use the `Android Emulators` MCP in Cursor to get a list of running devices.

### Test the server

1.  **Run the server:**

    ```bash
    ./mcp_android_emulators
    ```

2.  **Test the root endpoint:**

    ```bash
    curl http://localhost:8080/
    ```

    You should see the following output:

    ```json
    {"tools":[{"name":"get_android_devices","description":"Get a list of connected Android devices"}]}
    ```

3.  **Test the devices endpoint:**

    ```bash
    curl -X POST -d '{"tool": "get_android_devices"}' http://localhost:8080/devices
    ```

    You should see a JSON array of connected devices.

## Example JSON output

### Root endpoint (`/`)

```json
{
    "tools": [
        {
            "name": "get_android_devices",
            "description": "Get a list of connected Android devices"
        }
    ]
}
```

### Devices endpoint (`/devices`)

To execute the `get_android_devices` tool, you need to send a POST request to `http://localhost:8080/devices` with the following JSON body:

```json
{
    "tool": "get_android_devices"
}
```

The server will respond with a JSON array of devices:

```json
[
    {
        "name": "Pixel 2 API 30",
        "device": "emulator-5554",
        "model": "sdk_gphone_x86",
        "arch": "x86",
        "android_version": "11",
        "sdk_level": "30",
        "run_status": "device"
    }
]
```