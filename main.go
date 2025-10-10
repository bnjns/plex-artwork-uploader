package main

import "github.com/bnjns/plex-artwork-uploader/cmd"

var (
	version = "dev-master"
	commit  = "unknown"
)

func main() {
	cmd.SetVersionInfo(version, commit)
	cmd.Execute()
}
