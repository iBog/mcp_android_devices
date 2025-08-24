package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"testing"
)

func TestMCPInitialize(t *testing.T) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      1,
		Method:  "initialize",
		Params: map[string]interface{}{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]interface{}{},
			"clientInfo": map[string]interface{}{
				"name":    "test-client",
				"version": "1.0.0",
			},
		},
	}

	// Capture the response
	var response JSONRPCResponse
	captureResponse := func(resp JSONRPCResponse) {
		response = resp
	}

	// Mock sendResponse
	originalSendResponse := sendResponse
	sendResponse = captureResponse

	handleRequest(request)

	// Reset sendResponse
	sendResponse = originalSendResponse

	if response.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	if response.ID != 1 {
		t.Errorf("expected id 1, got %v", response.ID)
	}

	result, ok := response.Result.(InitializeResult)
	if !ok {
		t.Fatal("expected InitializeResult")
	}

	if result.ProtocolVersion != "2024-11-05" {
		t.Errorf("expected protocol version 2024-11-05, got %s", result.ProtocolVersion)
	}

	if result.ServerInfo.Name != "android-devices-mcp-server" {
		t.Errorf("expected server name android-devices-mcp-server, got %s", result.ServerInfo.Name)
	}
}

func TestMCPToolsList(t *testing.T) {
	request := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      2,
		Method:  "tools/list",
	}

	// Capture the response
	var response JSONRPCResponse
	captureResponse := func(resp JSONRPCResponse) {
		response = resp
	}

	// Mock sendResponse
	originalSendResponse := sendResponse
	sendResponse = captureResponse

	handleRequest(request)

	// Reset sendResponse
	sendResponse = originalSendResponse

	if response.JSONRPC != "2.0" {
		t.Errorf("expected jsonrpc 2.0, got %s", response.JSONRPC)
	}

	result, ok := response.Result.(ToolsListResult)
	if !ok {
		t.Fatal("expected ToolsListResult")
	}

	if len(result.Tools) != 2 {
		t.Errorf("expected 2 tools, got %d", len(result.Tools))
	}

	// Check first tool
	tool1 := result.Tools[0]
	if tool1.Name != "get_android_devices" {
		t.Errorf("expected first tool name get_android_devices, got %s", tool1.Name)
	}

	// Check second tool
	tool2 := result.Tools[1]
	if tool2.Name != "get_android_screen" {
		t.Errorf("expected second tool name get_android_screen, got %s", tool2.Name)
	}
}

func TestMCPToolsCall(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		// Mock the exec.Command function
		originalExecCommand := execCommand
		execCommand = func(command string, args ...string) *exec.Cmd {
			cs := []string{"-test.run=TestHelperProcess", "--", command}
			cs = append(cs, args...)
			cmd := exec.Command(os.Args[0], cs...)
			cmd.Env = []string{"GO_WANT_HELPER_PROCESS=1"}
			return cmd
		}

		request := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      3,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name":      "get_android_devices",
				"arguments": map[string]interface{}{},
			},
		}

		// Capture the response
		var response JSONRPCResponse
		captureResponse := func(resp JSONRPCResponse) {
			response = resp
		}

		// Mock sendResponse
		originalSendResponse := sendResponse
		sendResponse = captureResponse

		handleRequest(request)

		// Reset mocks
		sendResponse = originalSendResponse
		execCommand = originalExecCommand

		if response.JSONRPC != "2.0" {
			t.Errorf("expected jsonrpc 2.0, got %s", response.JSONRPC)
		}

		result, ok := response.Result.(ToolsCallResult)
		if !ok {
			t.Fatal("expected ToolsCallResult")
		}

		if result.IsError {
			t.Error("expected isError to be false")
		}

		if len(result.Content) != 1 {
			t.Errorf("expected 1 content item, got %d", len(result.Content))
		}

		content := result.Content[0]
		if content.Type != "text" {
			t.Errorf("expected content type text, got %s", content.Type)
		}

		// Parse the JSON content to verify device structure
		var devices []Device
		err := json.Unmarshal([]byte(content.Text), &devices)
		if err != nil {
			t.Fatal(err)
		}

		if len(devices) != 1 {
			t.Errorf("expected 1 device, got %d", len(devices))
		}

		expectedDevice := Device{
			Name:           "Pixel 2 API 30",
			Device:         "emulator-5554",
			Model:          "sdk_gphone_x86",
			Arch:           "x86",
			AndroidVersion: "11",
			SDKLevel:       "30",
			RunStatus:      "device",
		}

		if !reflect.DeepEqual(devices[0], expectedDevice) {
			t.Errorf("unexpected device: got %+v want %+v", devices[0], expectedDevice)
		}
	})

	t.Run("UnknownTool", func(t *testing.T) {
		request := JSONRPCRequest{
			JSONRPC: "2.0",
			ID:      4,
			Method:  "tools/call",
			Params: map[string]interface{}{
				"name": "unknown_tool",
			},
		}

		// Capture the response
		var response JSONRPCResponse
		captureResponse := func(resp JSONRPCResponse) {
			response = resp
		}

		// Mock sendResponse
		originalSendResponse := sendResponse
		sendResponse = captureResponse

		handleRequest(request)

		// Reset sendResponse
		sendResponse = originalSendResponse

		if response.Error == nil {
			t.Fatal("expected error response")
		}

		if response.Error.Code != -32602 {
			t.Errorf("expected error code -32602, got %d", response.Error.Code)
		}

		if !strings.Contains(response.Error.Message, "Unknown tool") {
			t.Errorf("expected error message to contain 'Unknown tool', got %s", response.Error.Message)
		}
	})
}

func TestGetDeviceList(t *testing.T) {
	t.Run("AdbNotFound", func(t *testing.T) {
		// Mock exec.LookPath to return an error
		originalLookPath := lookPath
		lookPath = func(file string) (string, error) {
			return "", fmt.Errorf("adb not found")
		}

		_, err := getDeviceList()
		if err == nil {
			t.Error("expected an error, but got nil")
		}

		if err.Error() != "adb command not found: adb not found" {
			t.Errorf("unexpected error message: got %q", err.Error())
		}

		// Reset exec.LookPath
		lookPath = originalLookPath
	})
}

func TestHelperProcess(t *testing.T) {
	if os.Getenv("GO_WANT_HELPER_PROCESS") != "1" {
		return
	}

	args := os.Args[3:]
	cmd, args := args[0], args[1:]

	switch cmd {
	case "adb":
		switch args[0] {
		case "devices":
			fmt.Println("List of devices attached")
			fmt.Println(`emulator-5554	device`)
		case "-s":
			switch args[2] {
			case "shell":
				switch args[3] {
				case "getprop":
					switch args[4] {
					case "ro.build.version.release":
						fmt.Println("11")
					case "ro.build.version.sdk":
						fmt.Println("30")
					case "ro.product.model":
						fmt.Println("sdk_gphone_x86")
					case "ro.product.cpu.abi":
						fmt.Println("x86")
					case "ro.product.brand":
						fmt.Println("Google")
					case "ro.kernel.qemu":
						fmt.Println("1")
					case "ro.boot.qemu.avd_name":
						fmt.Println("Pixel_2_API_30")
					}
				}
			}
		}
	}
	os.Exit(0)
}

func TestFindEmulatorProcess(t *testing.T) {
	devices, err := getDeviceList()
	if err != nil {
		t.Fatalf("Failed to get device list: %v", err)
	}

	if len(devices) == 0 {
		t.Skip("No devices found, skipping test.")
	}

	// Get process list using platform-specific commands
	var psOutput []byte
	switch runtime.GOOS {
	case "windows":
		// Use tasklist on Windows
		psOutput, err = exec.Command("tasklist", "/fo", "csv").Output()
		if err != nil {
			t.Fatalf("Failed to run 'tasklist': %v", err)
		}
	case "linux", "darwin":
		// Use ps on Unix-like systems (Linux, macOS)
		psOutput, err = exec.Command("ps", "aux").Output()
		if err != nil {
			t.Fatalf("Failed to run 'ps aux': %v", err)
		}
	default:
		// Fallback for other Unix-like systems
		psOutput, err = exec.Command("ps", "aux").Output()
		if err != nil {
			t.Fatalf("Failed to run 'ps aux' on %s: %v", runtime.GOOS, err)
		}
	}

	foundProcess := false
	for _, device := range devices {
		// We are looking for emulator processes, which have a non-empty model
		if device.Model != "" {
			avdName := strings.ReplaceAll(device.Name, " ", "_")
			avdArg := "-avd " + avdName
			lines := strings.Split(string(psOutput), "\n")
			for _, line := range lines {
				// Check for emulator process with AVD name
				if strings.Contains(line, "emulator") && strings.Contains(line, avdArg) {
					t.Logf("Found process for AVD: %s", device.Name)
					t.Logf("Process details: %s", line)
					foundProcess = true

					// Take a screenshot
					screenshotPath := "/sdcard/screenshot.png"
					screenshotCmd := exec.Command("adb", "-s", device.Device, "shell", "screencap", screenshotPath)
					if err := screenshotCmd.Run(); err != nil {
						t.Logf("Failed to take screenshot: %v", err)
					} else {
						// Pull the screenshot from the device
						pullCmd := exec.Command("adb", "-s", device.Device, "pull", screenshotPath)
						if err := pullCmd.Run(); err != nil {
							t.Logf("Failed to pull screenshot: %v", err)
						} else {
							t.Logf("Screenshot saved to screenshot.png")
						}
						// Clean up the screenshot from the device
						rmCmd := exec.Command("adb", "-s", device.Device, "shell", "rm", screenshotPath)
						if err := rmCmd.Run(); err != nil {
							t.Logf("Failed to remove screenshot from device: %v", err)
						}
					}
				}
			}
		}
	}

	if !foundProcess {
		t.Error("Could not find a running emulator process.")
	}
}
