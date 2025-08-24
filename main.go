package main

import (
	"bufio"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var execCommand = exec.Command
var lookPath = exec.LookPath
var sendResponse = func(response JSONRPCResponse) {
	responseBytes, _ := json.Marshal(response)
	fmt.Println(string(responseBytes))
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}

		var request JSONRPCRequest
		if err := json.Unmarshal([]byte(line), &request); err != nil {
			sendError(nil, -32700, "Parse error", nil)
			continue
		}

		handleRequest(request)
	}

	if err := scanner.Err(); err != nil {
		log.Fatal("Error reading from stdin:", err)
	}
}

func handleRequest(request JSONRPCRequest) {
	switch request.Method {
	case "initialize":
		handleInitialize(request)
	case "tools/list":
		handleToolsList(request)
	case "tools/call":
		handleToolsCall(request)
	default:
		sendError(request.ID, -32601, "Method not found", nil)
	}
}

func handleInitialize(request JSONRPCRequest) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities: ServerCapabilities{
				Tools: &ToolsCapability{
					ListChanged: true,
				},
			},
			ServerInfo: ServerInfo{
				Name:    "android-devices-mcp-server",
				Version: "1.0.0",
			},
		},
	}
	sendResponse(response)
}

func handleToolsList(request JSONRPCRequest) {
	tools := []Tool{
		{
			Name:        "get_android_devices",
			Description: "Get a list of connected Android devices and emulators",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "get_android_screen",
			Description: "Capture a screenshot from an Android device",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"device": map[string]interface{}{
						"type":        "string",
						"description": "Device name/serial (e.g., 'emulator-5554'). If not provided, uses the first available device.",
					},
				},
			},
		},
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: ToolsListResult{
			Tools: tools,
		},
	}
	sendResponse(response)
}

func handleToolsCall(request JSONRPCRequest) {
	var params ToolsCallParams
	if request.Params != nil {
		paramsBytes, _ := json.Marshal(request.Params)
		if err := json.Unmarshal(paramsBytes, &params); err != nil {
			sendError(request.ID, -32602, "Invalid params", nil)
			return
		}
	}

	switch params.Name {
	case "get_android_devices":
		handleGetDevices(request, params)
	case "get_android_screen":
		handleGetScreen(request, params)
	default:
		sendError(request.ID, -32602, "Unknown tool: "+params.Name, nil)
	}
}

func handleGetDevices(request JSONRPCRequest, params ToolsCallParams) {
	devices, err := getDeviceList()
	if err != nil {
		sendError(request.ID, -32603, "Internal error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	devicesJSON, _ := json.Marshal(devices)
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: ToolsCallResult{
			Content: []ContentItem{
				{
					Type: "text",
					Text: string(devicesJSON),
				},
			},
			IsError: false,
		},
	}
	sendResponse(response)
}

func handleGetScreen(request JSONRPCRequest, params ToolsCallParams) {
	deviceName := ""
	if params.Arguments != nil {
		if device, exists := params.Arguments["device"]; exists {
			if deviceStr, ok := device.(string); ok {
				deviceName = deviceStr
			}
		}
	}

	// If no device specified, use the first available device
	if deviceName == "" {
		devices, err := getDeviceList()
		if err != nil {
			sendError(request.ID, -32603, "Internal error", map[string]interface{}{
				"error": "Failed to get device list: " + err.Error(),
			})
			return
		}

		if len(devices) == 0 {
			sendError(request.ID, -32603, "Internal error", map[string]interface{}{
				"error": "No Android devices found",
			})
			return
		}

		deviceName = devices[0].Device
	}

	// Capture screenshot
	base64Data, err := captureScreenshot(deviceName)
	if err != nil {
		sendError(request.ID, -32603, "Internal error", map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      request.ID,
		Result: ToolsCallResult{
			Content: []ContentItem{
				{
					Type:     "image",
					Data:     base64Data,
					MimeType: "image/png",
				},
			},
			IsError: false,
		},
	}
	sendResponse(response)
}

func sendError(id interface{}, code int, message string, data interface{}) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
			Data:    data,
		},
	}
	sendResponse(response)
}

func getDeviceList() ([]Device, error) {
	// Check if adb command exists
	_, err := lookPath("adb")
	if err != nil {
		return nil, fmt.Errorf("adb command not found: %w", err)
	}

	cmd := execCommand("adb", "devices", "-l")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running adb command: %w, output: %s", err, string(output))
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var devices []Device

	for _, line := range lines {
		if strings.HasPrefix(line, "List of devices") || line == "" {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}

		device := Device{
			Device:    parts[0],
			RunStatus: parts[1],
		}

		// Get Android Version, SDK Level, Model, Arch and Brand
		name, androidVersion, sdkLevel, model, arch, err := getDeviceDetails(device.Device)
		if err != nil {
			log.Printf("Failed to get details for device %s: %v", device.Device, err)
		}
		device.Name = name
		device.AndroidVersion = androidVersion
		device.SDKLevel = sdkLevel
		device.Model = model
		device.Arch = arch

		devices = append(devices, device)
	}

	return devices, nil
}

func getDeviceDetails(deviceName string) (string, string, string, string, string, error) {
	// Check if the device is an emulator
	isEmulatorCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.kernel.qemu")
	isEmulatorOutput, err := isEmulatorCmd.CombinedOutput()
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to check if device is an emulator: %w, output: %s", err, string(isEmulatorOutput))
	}
	isEmulator := strings.TrimSpace(string(isEmulatorOutput)) == "1"

	var name string
	if isEmulator {
		avdNameCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.boot.qemu.avd_name")
		avdNameOutput, err := avdNameCmd.CombinedOutput()
		if err != nil {
			log.Printf("failed to get avd name for device %s: %v", deviceName, err)
		} else {
			name = strings.ReplaceAll(strings.TrimSpace(string(avdNameOutput)), "_", " ")
		}
	}

	if name == "" {
		brandCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.product.brand")
		brandOutput, err := brandCmd.CombinedOutput()
		if err != nil {
			return "", "", "", "", "", fmt.Errorf("failed to get device brand: %w, output: %s", err, string(brandOutput))
		}
		brand := strings.TrimSpace(string(brandOutput))

		modelCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.product.model")
		modelOutput, err := modelCmd.CombinedOutput()
		if err != nil {
			return "", "", "", "", "", fmt.Errorf("failed to get device model: %w, output: %s", err, string(modelOutput))
		}
		model := strings.TrimSpace(string(modelOutput))
		name = fmt.Sprintf("%s %s", brand, model)
	}

	androidVersionCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.build.version.release")
	androidVersionOutput, err := androidVersionCmd.CombinedOutput()
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to get android version: %w, output: %s", err, string(androidVersionOutput))
	}
	androidVersion := strings.TrimSpace(string(androidVersionOutput))

	sdkLevelCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.build.version.sdk")
	sdkLevelOutput, err := sdkLevelCmd.CombinedOutput()
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to get sdk level: %w, output: %s", err, string(sdkLevelOutput))
	}
	sdkLevel := strings.TrimSpace(string(sdkLevelOutput))

	modelCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.product.model")
	modelOutput, err := modelCmd.CombinedOutput()
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to get device model: %w, output: %s", err, string(modelOutput))
	}
	model := strings.TrimSpace(string(modelOutput))

	archCmd := execCommand("adb", "-s", deviceName, "shell", "getprop", "ro.product.cpu.abi")
	archOutput, err := archCmd.CombinedOutput()
	if err != nil {
		return "", "", "", "", "", fmt.Errorf("failed to get device architecture: %w, output: %s", err, string(archOutput))
	}
	arch := strings.TrimSpace(string(archOutput))

	return name, androidVersion, sdkLevel, model, arch, nil
}

func captureScreenshot(deviceName string) (string, error) {
	// Use exec-out to stream screenshot data directly from device to PC
	// This avoids creating temporary files on the Android device
	screenshotCmd := execCommand("adb", "-s", deviceName, "exec-out", "screencap", "-p")
	imageData, err := screenshotCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to capture screenshot from device %s: %w", deviceName, err)
	}

	// Encode to base64
	base64Data := base64.StdEncoding.EncodeToString(imageData)
	return base64Data, nil
}
