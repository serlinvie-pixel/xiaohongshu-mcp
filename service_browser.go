package main

import (
	"os"
	"strings"
	"sync"

	"github.com/go-rod/rod"
	"github.com/xpzouying/headless_browser"
	"github.com/xpzouying/xiaohongshu-mcp/browser"
	"github.com/xpzouying/xiaohongshu-mcp/configs"
)

// mcpBrowser exposes only the methods used by service handlers.
type mcpBrowser interface {
	NewPage() *rod.Page
	Close()
}

// A single Chromium instance owns the persistent REDnote profile while the
// MCP server is running. Tools get separate pages. Closing a per-tool handle
// must never close the shared browser while another tool is using it.
type sharedBrowserHandle struct{ *headless_browser.Browser }

func (sharedBrowserHandle) Close() {}

var rednoteBrowserOnce sync.Once
var rednoteBrowser *headless_browser.Browser

func persistentProfileDir() string {
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("XHS_BACKEND")), "rednote") {
		return ""
	}
	return strings.TrimSpace(os.Getenv("XHS_BROWSER_PROFILE_DIR"))
}

func newBrowser() mcpBrowser {
	makeBrowser := func(opts ...browser.Option) *headless_browser.Browser {
		base := []browser.Option{
			browser.WithFingerprintSeed(configs.FingerprintSeed()),
			browser.WithProxy(configs.Proxy()),
		}
		return browser.NewBrowser(configs.IsHeadless(), append(base, opts...)...)
	}

	dir := persistentProfileDir()
	if dir == "" {
		return makeBrowser()
	}

	rednoteBrowserOnce.Do(func() {
		rednoteBrowser = makeBrowser(browser.WithProfileDir(dir))
	})
	return sharedBrowserHandle{rednoteBrowser}
}
