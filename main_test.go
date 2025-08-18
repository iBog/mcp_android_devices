package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"reflect"
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

	if len(result.Tools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(result.Tools))
	}

	tool := result.Tools[0]
	if tool.Name != "get_android_devices" {
		t.Errorf("expected tool name get_android_devices, got %s", tool.Name)
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
