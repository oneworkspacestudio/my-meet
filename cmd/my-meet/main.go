package main

import (
	"github.com/oneworkspacestudio/my-meet/internal/app"
	"github.com/oneworkspacestudio/my-meet/internal/webui"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
	"os"
)

var version = "0.1.0-preview.2"

func main() {
	if e := apphost.Main(apphost.Config{ID: "meet", Name: "My Meet", Version: version, Listen: "127.0.0.1:9016", MediaCapture: true, Realtime: true}, webui.Files(), func(h *apphost.Host) (apphost.Application, error) { return app.New(h) }); e != nil {
		apphost.Report(e)
		os.Exit(1)
	}
}
