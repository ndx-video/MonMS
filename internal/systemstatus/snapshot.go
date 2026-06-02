package systemstatus

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/monms/monms/internal/content"
	"github.com/monms/monms/internal/daemon"
	"github.com/monms/monms/internal/logging"
	"github.com/monms/monms/internal/stop"
)

// View is operator-facing runtime status for the System console page.
type View struct {
	SiteAbs            string
	BuildMode          string
	TemplateMode       string
	PID                int
	DaemonChild        bool
	PIDFile            int
	PIDFileMatches     bool
	HTTPListen         string
	URLs               content.ServeURLs
	MCP                content.MCPConfig
	MCPListen          string
	ProductionURLSet   bool
	PublishTokenSet    bool
	PublisherCount     int
	LoggingLevels      string
	LoggingRotation    string
	LogsDir            string
	HealthOK           bool
	HealthDetail       string
	LifecycleSupported bool
	LifecycleNote      string
}

// Snapshot builds the system status view for the dashboard.
func Snapshot(siteAbs string, serveArgs []string, buildMode string, publishTokenSet bool) (View, error) {
	cfg, err := content.LoadMonmsConfig(siteAbs)
	if err != nil {
		return View{}, err
	}

	urls, err := content.ResolveServeURLs(siteAbs, serveArgs)
	if err != nil {
		return View{}, err
	}

	logCfg, err := logging.LoadConfig(siteAbs)
	if err != nil {
		return View{}, err
	}

	pid := os.Getpid()
	daemonChild := os.Getenv("MONMS_DAEMON") == "1"

	var pidFile int
	pidMatches := false
	if pidPath := daemon.PIDFilePath(siteAbs); pidPath != "" {
		if p, err := daemon.ReadPIDFile(pidPath); err == nil {
			pidFile = p
			pidMatches = p == pid
		}
	}

	mcpListen := ""
	if cfg.MCP.Enabled {
		host := strings.TrimSpace(cfg.MCP.Host)
		port := strings.TrimSpace(cfg.MCP.Port)
		if host == "" {
			host = content.DefaultMCPConfig().Host
		}
		if port == "" {
			port = content.DefaultMCPConfig().Port
		}
		mcpListen = host + ":" + port
	}

	levels := strings.Join(cfg.Logging, ", ")
	if levels == "" {
		levels = "(default from build mode)"
	}
	rot := logCfg.Rotation
	rotSummary := fmt.Sprintf("%d MB max, %d backups, %d days, compress=%v",
		rot.MaxSizeMB, rot.MaxBackups, rot.MaxAgeDays, rot.Compress)

	healthOK, healthDetail := probeHealth(content.ResolveHTTPListenAddr(serveArgs, cfg))

	lifecycleSupported := true
	lifecycleNote := ""
	if _, err := stop.Supported(); err != nil {
		lifecycleSupported = false
		lifecycleNote = err.Error()
	}

	templateMode := "development"
	if buildMode == "production" {
		templateMode = "production"
	}

	siteAbs, err = filepath.Abs(siteAbs)
	if err != nil {
		return View{}, fmt.Errorf("resolve site path: %w", err)
	}

	return View{
		SiteAbs:            siteAbs,
		BuildMode:          buildMode,
		TemplateMode:       templateMode,
		PID:                pid,
		DaemonChild:        daemonChild,
		PIDFile:            pidFile,
		PIDFileMatches:     pidMatches,
		HTTPListen:         content.ResolveHTTPListenAddr(serveArgs, cfg),
		URLs:               urls,
		MCP:                cfg.MCP,
		MCPListen:          mcpListen,
		ProductionURLSet:   strings.TrimSpace(cfg.ProductionURL) != "",
		PublishTokenSet:    publishTokenSet,
		PublisherCount:     len(cfg.PublisherEmails),
		LoggingLevels:      levels,
		LoggingRotation:    rotSummary,
		LogsDir:            urls.LogsDir,
		HealthOK:           healthOK,
		HealthDetail:       healthDetail,
		LifecycleSupported: lifecycleSupported,
		LifecycleNote:      lifecycleNote,
	}, nil
}

func probeHealth(listenAddr string) (ok bool, detail string) {
	host, port, err := splitListen(listenAddr)
	if err != nil {
		return false, err.Error()
	}
	url := fmt.Sprintf("http://%s:%s/api/health", host, port)
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false, fmt.Sprintf("unreachable (%v)", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusOK {
		return true, "OK"
	}
	return false, fmt.Sprintf("HTTP %d", resp.StatusCode)
}

func splitListen(addr string) (host, port string, err error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "127.0.0.1", "8090", nil
	}
	h, p, err := net.SplitHostPort(addr)
	if err != nil {
		if !strings.Contains(addr, ":") {
			return addr, "8090", nil
		}
		return "", "", err
	}
	if h == "" || h == "0.0.0.0" {
		h = "127.0.0.1"
	}
	if p == "" {
		p = "8090"
	}
	return h, p, nil
}
