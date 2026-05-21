package router

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pocketbase/pocketbase/core"
)

var errForbidden = errors.New("path outside assets root")

// AssetsHandler serves files from workspace/assets with root-jail validation (D-29).
func AssetsHandler(wsAbs string) func(*core.RequestEvent) error {
	assetsRoot := filepath.Join(wsAbs, "assets")
	return func(e *core.RequestEvent) error {
		requestPath := e.Request.PathValue("path")
		absPath, err := safeAssetPath(assetsRoot, requestPath)
		if err != nil {
			if errors.Is(err, errForbidden) {
				return e.HTML(http.StatusForbidden, "Forbidden")
			}
			return err
		}
		return serveAssetFile(e, absPath)
	}
}

func safeAssetPath(assetsRoot, requestPath string) (string, error) {
	clean := filepath.Clean(filepath.Join(assetsRoot, requestPath))
	absRoot, err := filepath.Abs(assetsRoot)
	if err != nil {
		return "", err
	}
	absClean, err := filepath.Abs(clean)
	if err != nil {
		return "", err
	}
	if absClean != absRoot && !strings.HasPrefix(absClean, absRoot+string(os.PathSeparator)) {
		return "", errForbidden
	}
	return absClean, nil
}

func serveAssetFile(e *core.RequestEvent, absPath string) error {
	fi, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return e.HTML(http.StatusNotFound, "Not Found")
		}
		return err
	}
	if fi.IsDir() {
		return e.HTML(http.StatusNotFound, "Not Found")
	}
	http.ServeFile(e.Response, e.Request, absPath)
	return nil
}
