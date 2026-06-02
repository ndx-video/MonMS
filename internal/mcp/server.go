package mcp

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"

	"github.com/mark3labs/mcp-go/server"
	"github.com/monms/monms/internal/apikeys"
	"github.com/monms/monms/internal/content"
)

var (
	startedMu sync.Mutex
	started   bool
)

// Start launches the MCP HTTP listener when enabled in site config.
func Start(ctx context.Context, deps Deps) {
	if !deps.Config.MCP.Enabled {
		return
	}

	startedMu.Lock()
	if started {
		startedMu.Unlock()
		return
	}
	started = true
	startedMu.Unlock()

	host := strings.TrimSpace(deps.Config.MCP.Host)
	if host == "" {
		host = content.DefaultMCPConfig().Host
	}
	port := strings.TrimSpace(deps.Config.MCP.Port)
	if port == "" {
		port = content.DefaultMCPConfig().Port
	}
	addr := host + ":" + port

	pbBase := strings.TrimRight(deps.Config.SiteURL, "/")
	if pbBase == "" {
		pbBase = "http://127.0.0.1:8090"
	}

	mcpServer := server.NewMCPServer(
		"monms",
		"1.0.0",
		server.WithToolCapabilities(true),
	)
	registerTools(mcpServer, deps, pbBase)

	httpServer := server.NewStreamableHTTPServer(mcpServer, server.WithStateLess(true))
	handler := authMiddleware(deps, httpServer)

	go func() {
		slog.Info("mcp: listening", "addr", addr, "endpoint", "/mcp")
		srv := &http.Server{
			Addr:    addr,
			Handler: handler,
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("mcp: server stopped", "err", err)
		}
	}()

	go func() {
		<-ctx.Done()
		_ = httpServer.Shutdown(context.Background())
	}()
}

func authMiddleware(deps Deps, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok || strings.TrimSpace(token) == "" {
			http.Error(w, "authorization required", http.StatusUnauthorized)
			return
		}
		token = strings.TrimSpace(token)

		resolved, err := apikeys.Resolve(deps.App, deps.SiteAbs, token)
		if err != nil {
			http.Error(w, "invalid api key", http.StatusUnauthorized)
			return
		}

		ownerToken, err := resolved.Owner.NewAuthToken()
		if err != nil {
			http.Error(w, "auth failed", http.StatusInternalServerError)
			return
		}

		go apikeys.TouchLastUsed(deps.App, resolved.KeyRecord)

		sess := &session{
			Resolved:   resolved,
			OwnerToken: ownerToken,
		}
		ctx := withSession(r.Context(), sess)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ResetStartedForTests clears the start guard (tests only).
func ResetStartedForTests() {
	startedMu.Lock()
	started = false
	startedMu.Unlock()
}

// ListenAddr formats host:port from MCP config.
func ListenAddr(cfg content.MonmsConfig) string {
	host := strings.TrimSpace(cfg.MCP.Host)
	if host == "" {
		host = content.DefaultMCPConfig().Host
	}
	port := strings.TrimSpace(cfg.MCP.Port)
	if port == "" {
		port = content.DefaultMCPConfig().Port
	}
	return fmt.Sprintf("%s:%s", host, port)
}
