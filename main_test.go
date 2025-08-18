package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"reflect"
	"testing"
)

func TestGetDevices(t *testing.T) {
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

		// Create a request body
		requestBody := map[string]interface{}{
			"tool": "get_android_devices",
		}
		jsonBody, _ := json.Marshal(requestBody)

		req, err := http.NewRequest("POST", "/devices", bytes.NewBuffer(jsonBody))
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		http.HandlerFunc(devicesHandler).ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		// Check the response body
		var devices []Device
		err = json.Unmarshal(rr.Body.Bytes(), &devices)
		if err != nil {
			t.Fatal(err)
		}

		if len(devices) != 1 {
			t.Errorf("expected 1 device, got %d", len(devices))
		}

		expectedDevice := Device{
			Name:         "Pixel 2 API 30",
			Device:       "emulator-5554",
			Model:        "sdk_gphone_x86",
			Arch:         "x86",
			AndroidVersion: "11",
			SDKLevel:     "30",
			RunStatus:    "device",
		}

		if !reflect.DeepEqual(devices[0], expectedDevice) {
			t.Errorf("unexpected device: got %+v want %+v", devices[0], expectedDevice)
		}

		// Reset execCommand
		execCommand = originalExecCommand
	})

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
