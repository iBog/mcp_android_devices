package main

// Device represents an Android device
type Device struct {
	Name         string `json:"name"`
	Device       string `json:"device"`
	Model        string `json:"model"`
	Arch         string `json:"arch"`
	AndroidVersion string `json:"android_version"`
	SDKLevel     string `json:"sdk_level"`
	RunStatus    string `json:"run_status"`
}
