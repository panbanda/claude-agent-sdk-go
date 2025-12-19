// Example: mcp-server
// Demonstrates how to configure MCP servers that Claude can use as tools.
//
// This example uses the filesystem MCP server to give Claude access to
// read and write files in /tmp. You can configure any MCP server by
// modifying the mcp-config.json file.
//
// Prerequisites:
//   - Node.js and npx must be installed for the filesystem server
//   - The MCP server package will be downloaded automatically via npx
package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/panbanda/claude-agent-sdk-go/claude"
)

func main() {
	ctx := context.Background()

	// Get the path to our MCP config file (relative to this example)
	configPath := filepath.Join(getExampleDir(), "mcp-config.json")

	// Create a test file for Claude to read
	testFile := "/tmp/sdk-mcp-test.txt"
	if err := os.WriteFile(testFile, []byte("Hello from the MCP server example!"), 0600); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating test file: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("MCP Server Example")
	fmt.Println("==================")
	fmt.Printf("Config: %s\n", configPath)
	fmt.Printf("Test file: %s\n\n", testFile)

	// Query Claude with MCP server configured
	msgs, err := claude.Query(ctx,
		fmt.Sprintf("Read the file at %s and tell me what it says.", testFile),
		claude.WithMCPConfig(configPath),
		claude.WithMaxTurns(5),
	)
	if err != nil {
		os.Remove(testFile)
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Response:")
	for msg := range msgs {
		switch m := msg.(type) {
		case *claude.AssistantMessage:
			for _, block := range m.Content {
				if block.IsText() {
					fmt.Print(block.Text)
				}
			}
		case *claude.ResultMessage:
			fmt.Printf("\n\nCost: $%.4f\n", m.TotalCostUSD)
		}
	}

	// Clean up test file
	os.Remove(testFile)
}

// getExampleDir returns the directory containing this example.
func getExampleDir() string {
	// When running with 'go run', we need to find the source directory
	exe, err := os.Executable()
	if err != nil {
		// Fallback to current directory
		return "."
	}

	// Check if we're in the examples directory already
	cwd, _ := os.Getwd()
	configInCwd := filepath.Join(cwd, "mcp-config.json")
	if _, err := os.Stat(configInCwd); err == nil {
		return cwd
	}

	// Check relative to executable
	dir := filepath.Dir(exe)
	configInExeDir := filepath.Join(dir, "mcp-config.json")
	if _, err := os.Stat(configInExeDir); err == nil {
		return dir
	}

	// Default: assume running from repo root
	return "examples/mcp-server"
}
