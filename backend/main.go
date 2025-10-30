package main

import (
	"github.com/pramodksahoo/kubechat/backend/cmd"
)

// version specify version of application using ldflags
var version = "dev"
var commit = "unknown"

func main() {
	cmd.Version = version
	cmd.Commit = commit
	cmd.Execute()
}
