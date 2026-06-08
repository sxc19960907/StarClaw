package tools

import (
	"log"
	"time"
)

func HandBrowserOff(browser *BrowserTool, backstop time.Duration, cleanup func() error) {
	if browser == nil {
		return
	}
	browser.MarkDeprecated()
	if cleanupBrowserIfNoOwners(browser, cleanup) {
		return
	}
	time.AfterFunc(backstop, func() {
		if cleanupBrowserIfNoOwners(browser, cleanup) {
			return
		}
		log.Printf("browser handoff: deprecated browser still has %d active lease(s)", BrowserOwnerActiveCount(browser))
	})
}

func cleanupBrowserIfNoOwners(browser *BrowserTool, cleanup func() error) bool {
	globalBrowserTracker.mu.Lock()
	defer globalBrowserTracker.mu.Unlock()
	if globalBrowserTracker.owners != nil && globalBrowserTracker.owners[browser] != 0 {
		return false
	}
	if cleanup != nil {
		_ = cleanup()
	}
	return true
}
