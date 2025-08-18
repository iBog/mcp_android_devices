package main

import "encoding/json"

// MCP Tool
type Tool struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	// TODO: Add input and output schemas
}

// MCP Root
type Root struct {
	Tools []Tool `json:"tools"`
}

func NewRoot() *Root {
	return &Root{
		Tools: []Tool{
			{
				Name:        "get_android_devices",
				Description: "Get a list of connected Android devices",
			},
		},
	}
}

func (r *Root) ToJSON() ([]byte, error) {
	return json.Marshal(r)
}