package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/mark3labs/mcp-go/client"
	"github.com/mark3labs/mcp-go/client/transport"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type headerFlags []string

func (h *headerFlags) String() string {
	return strings.Join(*h, ", ")
}

func (h *headerFlags) Set(value string) error {
	*h = append(*h, value)
	return nil
}

func main() {
	var headers headerFlags

	serverURL := flag.String("server", "http://localhost:8080/mcp/", "Go Money MCP server URL")
	token := flag.String("token", "", "Service token for authentication (required)")
	flag.Var(&headers, "header", "Additional HTTP header in 'Key: Value' format (can be specified multiple times)")
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

	httpHeaders := map[string]string{
		"Authorization": "Bearer " + *token,
	}

	for _, h := range headers {
		parts := strings.SplitN(h, ":", 2)
		if len(parts) != 2 {
			log.Fatalf("invalid header format %q, expected 'Key: Value'", h)
		}
		httpHeaders[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
	}

	httpTransport, err := transport.NewStreamableHTTP(
		*serverURL,
		transport.WithHTTPHeaders(httpHeaders),
	)
	if err != nil {
		log.Fatalf("failed to create transport: %v", err)
	}

	mcpClient := client.NewClient(httpTransport)

	if err = mcpClient.Start(ctx); err != nil {
		log.Fatalf("failed to start client: %v", err)
	}
	defer func() { _ = mcpClient.Close() }()

	_, err = mcpClient.Initialize(ctx, mcp.InitializeRequest{
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
		server.WithResourceCapabilities(true, true),
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

	resources, err := mcpClient.ListResources(ctx, mcp.ListResourcesRequest{})
	if err != nil {
		log.Fatalf("failed to list resources: %v", err)
	}

	for _, resource := range resources.Resources {
		r := resource
		stdioServer.AddResource(r, func(ctx context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
			result, readErr := mcpClient.ReadResource(ctx, mcp.ReadResourceRequest{
				Params: mcp.ReadResourceParams{
					URI: req.Params.URI,
				},
			})
			if readErr != nil {
				return nil, readErr
			}
			return result.Contents, nil
		})
	}

	if err = server.ServeStdio(stdioServer); err != nil {
		log.Fatalf("stdio server error: %v", err)
	}
}
