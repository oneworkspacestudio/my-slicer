package main

import (
	"github.com/oneworkspacestudio/my-slicer/internal/app"
	"github.com/oneworkspacestudio/my-slicer/internal/webui"
	"github.com/oneworkspacestudio/one-workspace/pkg/apphost"
)

var version = "0.1.0-preview.1"

func main() {
	apphost.Report(apphost.Main(apphost.Config{ID: "slicer", Name: "My Slicer", Version: version, Listen: "127.0.0.1:19020"}, webui.Files(), func(h *apphost.Host) (apphost.Application, error) { return app.New(h) }))
}
