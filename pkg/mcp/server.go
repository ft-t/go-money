package mcp

import (
	"context"
	"fmt"
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gorm.io/gorm"
)

type Server struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	db         *gorm.DB
	cfg        *ServerConfig
}

type ServerConfig struct {
	DB   *gorm.DB
	Docs string
}

func NewServer(cfg *ServerConfig) *Server {
	mcpServer := server.NewMCPServer(
		"go-money",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithResourceCapabilities(false, false),
		server.WithRecovery(),
	)

	s := &Server{
		mcpServer: mcpServer,
		db:        cfg.DB,
		cfg:       cfg,
	}

	s.registerTools()
	s.registerResources()

	s.httpServer = server.NewStreamableHTTPServer(mcpServer)

	return s
}

func (s *Server) registerTools() {
	queryTool := mcp.NewTool(
		"query",
		mcp.WithDescription(fmt.Sprintf("Run a read-only SQL query against the Go Money database. Schema: %v", s.cfg.Docs)),
		mcp.WithString(
			"sql",
			mcp.Description("The SQL SELECT query to execute"),
			mcp.Required(),
		),
	)

	s.mcpServer.AddTool(queryTool, s.handleQuery)
}

func (s *Server) registerResources() {
	schemaResource := mcp.NewResource(
		"context://schema",
		"Database Schema",
		mcp.WithResourceDescription("Go Money database schema documentation"),
		mcp.WithMIMEType("text/markdown"),
	)

	s.mcpServer.AddResource(schemaResource, s.handleSchemaResource)
}

func (s *Server) handleSchemaResource(_ context.Context, req mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      req.Params.URI,
			MIMEType: "text/markdown",
			Text:     s.cfg.Docs,
		},
	}, nil
}

func (s *Server) Handler() http.Handler {
	return s.httpServer
}

func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}
