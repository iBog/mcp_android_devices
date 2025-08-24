import { createInterface } from 'readline';
import { exec } from 'child_process';
import { z } from 'zod';

const rl = createInterface({
  input: process.stdin,
  output: process.stdout,
  terminal: false
});

const tools = {
  get_android_devices: {
    description: 'Get a list of connected Android devices and emulators',
    inputSchema: {
      type: 'object',
      properties: {},
    },
    execute: getAndroidDevices,
  },
  get_android_screen: {
    description: 'Capture a screenshot from an Android device',
    inputSchema: {
      type: 'object',
      properties: {
        device: {
          type: 'string',
          description: "Device name (e.g., 'emulator-5554'). If not provided, uses the first available device.",
        },
      },
    },
    execute: getAndroidScreen,
  },
};

rl.on('line', (line) => {
  try {
    const request = JSON.parse(line);
    handleRequest(request);
  } catch (e) {
    sendResponse({
      jsonrpc: '2.0',
      error: { code: -32700, message: 'Parse error' }
    });
  }
});

function handleRequest(request) {
  switch (request.method) {
    case 'initialize':
      handleInitialize(request);
      break;
    case 'tools/list':
      handleToolsList(request);
      break;
    case 'tools/call':
      handleToolsCall(request);
      break;
    default:
      sendResponse({
        jsonrpc: '2.0',
        id: request.id,
        error: { code: -32601, message: 'Method not found' }
      });
  }
}

function handleInitialize(request) {
  sendResponse({
    jsonrpc: '2.0',
    id: request.id,
    result: {
      protocolVersion: '2024-11-05',
      capabilities: {
        tools: {
          listChanged: true
        }
      },
      serverInfo: {
        name: 'android-devices-mcp-server',
        version: '1.0.0'
      }
    }
  });
}

function handleToolsList(request) {
  const toolList = Object.entries(tools).map(([name, tool]) => ({
    name,
    description: tool.description,
    inputSchema: tool.inputSchema,
  }));

  sendResponse({
    jsonrpc: '2.0',
    id: request.id,
    result: {
      tools: toolList,
    }
  });
}

function handleToolsCall(request) {
  const { name, arguments: args } = request.params;
  const tool = tools[name];

  if (!tool) {
    sendResponse({
      jsonrpc: '2.0',
      id: request.id,
      error: { code: -32601, message: 'Tool not found' }
    });
    return;
  }

  // We are not using zod anymore for parsing, so we can remove this try/catch block
  tool.execute(request.id, args || {});
}

function getAndroidDevices(id) {
  exec('adb devices -l', (error, stdout, stderr) => {
    if (error) {
      sendResponse({
        jsonrpc: '2.0',
        id,
        error: { code: -32000, message: `Error executing adb: ${stderr}` }
      });
      return;
    }

    const devices = stdout.trim().split('\n').slice(1).map(line => {
      const parts = line.trim().split(/\s+/);
      const device = parts[0];
      const model = parts.find(p => p.startsWith('model:'))?.replace('model:', '') || '';
      const product = parts.find(p => p.startsWith('product:'))?.replace('product:', '') || '';
      const transport_id = parts.find(p => p.startsWith('transport_id:'))?.replace('transport_id:', '') || '';
      return { name: product, device, model, run_status: parts[1], transport_id };
    });

    sendResponse({
      jsonrpc: '2.0',
      id,
      result: {
        content: [{
          type: 'text',
          text: JSON.stringify(devices)
        }],
        isError: false
      }
    });
  });
}

function getAndroidScreen(id, { device }) {
    const deviceIdentifier = device ? `-s ${device}` : '';
    exec(`adb ${deviceIdentifier} shell screencap -p`, { encoding: 'binary', maxBuffer: 1024 * 1024 * 10 }, (error, stdout, stderr) => {
        if (error) {
            sendResponse({
                jsonrpc: '2.0',
                id,
                error: { code: -32000, message: `Error executing adb: ${stderr}` }
            });
            return;
        }

        const base64Image = Buffer.from(stdout, 'binary').toString('base64');

        sendResponse({
            jsonrpc: '2.0',
            id,
            result: {
                content: [{
                    type: 'image',
                    data: base64Image,
                    mimeType: 'image/png'
                }],
                isError: false
            }
        });
    });
}

function sendResponse(response) {
  const responseString = JSON.stringify(response);
  process.stdout.write(`Content-Length: ${Buffer.byteLength(responseString)}\r\n\r\n${responseString}`);
}