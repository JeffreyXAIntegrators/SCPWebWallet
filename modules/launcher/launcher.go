package launcher

import (
	"path/filepath"

	"github.com/georgemcarlson/lorca"
	"gitlab.com/scpcorp/webwallet/build"
)

var additionalArgs = []string{
	"--disable-background-networking",
	"--disable-background-timer-throttling",
	"--disable-backgrounding-occluded-windows",
	"--disable-breakpad",
	"--disable-default-apps",
	"--disable-dev-shm-usage",
	"--disable-extensions",
	"--disable-features=site-per-process",
	"--disable-hang-monitor",
	"--disable-ipc-flooding-protection",
	"--disable-popup-blocking",
	"--disable-prompt-on-repost",
	"--disable-renderer-backgrounding",
	"--disable-sync",
	"--disable-translate",
	"--disable-windows10-custom-titlebar",
	"--metrics-recording-only",
	"--no-first-run",
	"--no-default-browser-check",
	"--safebrowsing-disable-auto-update",
}

var width = 1366
var height = 768

// Launch will attempt to launch the application in the supplied browser. If that fails the
// launcher will attempt to launch the application in a series of fallback browsers. This
// allows the GUI head feel most like a native application.
func Launch(browserCfg string) chan struct{} {
	uiDone := launch(browserCfg)
	if uiDone != nil {
		return uiDone
	}
	uiDone = fallback(browserCfg)
	if uiDone != nil {
		return uiDone
	}
	return unsupported()
}

func launch(browserCfg string) chan struct{} {
	switch browserCfg {
	case "edge":
		return edge()
	case "chrome":
		return chrome()
	case "chromium":
		return chromium()
	}
	return nil
}

func fallback(browserCfg string) chan struct{} {
	if browserCfg != "chrome" {
		uiDone := chrome()
		if uiDone != nil {
			return uiDone
		}
	}
	if browserCfg != "chromium" {
		uiDone := chromium()
		if uiDone != nil {
			return uiDone
		}
	}
	if browserCfg != "edge" {
		uiDone := edge()
		if uiDone != nil {
			return uiDone
		}
	}
	return nil
}

func edge() chan struct{} {
	ui, err := lorca.NewEdge("http://localhost:4300", userProfileDir("edge"), width, height, additionalArgs...)
	if ui == nil || err != nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		<-ui.Done()
		close(done)
	}()
	return done
}

func chrome() chan struct{} {
	ui, err := lorca.NewGoogleChrome("http://localhost:4300", userProfileDir("chrome"), width, height, additionalArgs...)
	if ui == nil || err != nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		<-ui.Done()
		close(done)
	}()
	return done
}

func chromium() chan struct{} {
	ui, err := lorca.NewChromium("http://localhost:4300", userProfileDir("chromium"), width, height, additionalArgs...)
	if ui == nil || err != nil {
		return nil
	}
	done := make(chan struct{})
	go func() {
		<-ui.Done()
		close(done)
	}()
	return done
}

func unsupported() chan struct{} {
	go func() {
		lorca.PromptDownload()
	}()
	done := make(chan struct{})
	close(done)
	return done
}

func userProfileDir(dir string) string {
	return filepath.Join(build.ScPrimeWebWalletDir(), "browser", dir)
}
