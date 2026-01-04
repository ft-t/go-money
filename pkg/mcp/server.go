package mcp

import (
	"net/http"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
	"gorm.io/gorm"
)

type Server struct {
	mcpServer  *server.MCPServer
	httpServer *server.StreamableHTTPServer
	db         *gorm.DB
}

type ServerConfig struct {
	DB *gorm.DB
}

func NewServer(cfg *ServerConfig) *Server {
	mcpServer := server.NewMCPServer(
		"go-money",
		"1.0.0",
		server.WithToolCapabilities(true),
		server.WithRecovery(),
	)

	s := &Server{
		mcpServer: mcpServer,
		db:        cfg.DB,
	}

	s.registerTools()

	s.httpServer = server.NewStreamableHTTPServer(mcpServer)

	return s
}

func (s *Server) registerTools() {
	queryTool := mcp.NewTool(
		"query",
		mcp.WithDescription("Run a read-only SQL query against the Go Money database. Use the embedded documentation for schema reference."),
		mcp.WithString(
			"sql",
			mcp.Description("The SQL SELECT query to execute"),
			mcp.Required(),
		),
	)

	s.mcpServer.AddTool(queryTool, s.handleQuery)
}

func (s *Server) Handler() http.Handler {
	return s.httpServer
}

func (s *Server) MCPServer() *server.MCPServer {
	return s.mcpServer
}
