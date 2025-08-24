#!/bin/bash

# Test initialize
echo '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2024-11-05","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' | node index.js

# Test tools/list
echo '{"jsonrpc":"2.0","id":2,"method":"tools/list"}' | node index.js

# Test tools/call with get_android_devices
echo '{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"get_android_devices","arguments":{}}}' | node index.js

# Test tools/call with get_android_screen
echo '{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"get_android_screen","arguments":{}}}' | node index.js
