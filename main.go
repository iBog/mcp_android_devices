package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

var execCommand = exec.Command
var lookPath = exec.LookPath

func main() {
	http.HandleFunc("/", rootHandler)
	http.HandleFunc("/devices", devicesHandler)
	log.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	root := NewRoot()
	json.NewEncoder(w).Encode(root)
}

func devicesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var requestBody map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Failed to decode request body", http.StatusBadRequest)
		return
	}

	toolName, ok := requestBody["tool"].(string)
	if !ok || toolName != "get_android_devices" {
		http.Error(w, "Invalid tool name", http.StatusBadRequest)
		return
	}

	devices, err := getDeviceList()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get device list: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(devices)
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
