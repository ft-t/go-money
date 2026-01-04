package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	serverURL := flag.String("server", "http://localhost:8080/mcp/", "Go Money MCP server URL")
	token := flag.String("token", "", "Service token for authentication (required)")
	flag.Parse()

	if *token == "" {
		log.Fatal("service token is required: use -token flag")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sig
		cancel()
	}()

	httpTransport, err := transport.NewStreamableHTTP(
		*serverURL,
		transport.WithHTTPHeaders(map[string]string{
			"Authorization": "Bearer " + *token,
		}),
	)
	if err != nil {
		log.Fatalf("failed to create transport: %v", err)
	}

	mcpClient := client.NewClient(httpTransport)

	if err = mcpClient.Start(ctx); err != nil {
		log.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = mcpClient.Close() }()

	initResult, err := mcpClient.Initialize(ctx, mcp.InitializeRequest{
		Params: mcp.InitializeParams{
			ClientInfo: mcp.Implementation{
				Name:    "go-money-mcp-client",
				Version: "1.0.0",
			},
			ProtocolVersion: mcp.LATEST_PROTOCOL_VERSION,
		},
	})
	if err != nil {
		log.Fatalf("failed to initialize: %v", err)
	}

	tools, err := mcpClient.ListTools(ctx, mcp.ListToolsRequest{})
	if err != nil {
		log.Fatalf("failed to list tools: %v", err)
	}

	stdioServer := server.NewMCPServer(
		"go-money-mcp-client",
		"1.0.0",
		server.WithToolCapabilities(true),
	)

	for _, tool := range tools.Tools {
		t := tool
		stdioServer.AddTool(t, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			result, callErr := mcpClient.CallTool(ctx, mcp.CallToolRequest{
				Params: mcp.CallToolParams{
					Name:      req.Params.Name,
					Arguments: req.Params.Arguments,
				},
			})
			if callErr != nil {
				return nil, callErr
			}
			return result, nil
		})
	}

	fmt.Fprintf(os.Stderr, "Connected to %s (server: %s %s)\n",
		*serverURL, initResult.ServerInfo.Name, initResult.ServerInfo.Version)
	fmt.Fprintf(os.Stderr, "Available tools: %d\n", len(tools.Tools))

	if err = server.ServeStdio(stdioServer); err != nil {
		log.Fatalf("stdio server error: %v", err)
	}
}
