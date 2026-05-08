package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/willystout/vaultguard/pkg/mcp"
	"github.com/willystout/vaultguard/pkg/scanner"
)

func main() {
	// Create MCP server and register tools
	server := mcp.NewServer("vaultguard", "0.1.0")

	// Register the scan_repo tool
	server.RegisterTool(scanner.NewScanRepoTool())

	// TODO: Register remaining tools as you build them
	// server.RegisterTool(rotation.NewCheckRotationTool())
	// server.RegisterTool(envaudit.NewAuditEnvTool())
	// server.RegisterTool(report.NewHealthReportTool())

	// Read JSON-RPC messages from stdin, write responses to stdout
	input := bufio.NewScanner(os.Stdin)
	for input.Scan() {
		line := input.Bytes()

		response, err := server.HandleMessage(line)
		if err != nil {
			log.Printf("Error handling message: %v", err)
			continue
		}

		if response == nil {
			continue
		}

		// Write JSON-RPC response to stdout
		out, err := json.Marshal(response)
		if err != nil {
			log.Printf("Error marshaling response: %v", err)
			continue
		}
		fmt.Println(string(out))
	}

	if err := input.Err(); err != nil {
		log.Fatalf("Error reading stdin: %v", err)
	}
}
