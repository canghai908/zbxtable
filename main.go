package main

import (
	"embed"
	"fmt"
	"github.com/json-iterator/go/extra"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"zbxtable/cmd"
	"zbxtable/utils"
)

//go:embed template
var f embed.FS

// AppVersion version
var (
	version   = "No Version Provided"
	gitHash   = "No GitHash Provided"
	buildTime = "No BuildTime Provided"
)

func customVersionPrinter(c *cli.Context) {
	fmt.Printf("%v Version:=%v\nGit Commit Hash:=%v\nUTC Build Time:=%v\n", filepath.Base(c.App.Name),
		version, gitHash, buildTime)
}
func init() {
	// RegisterFuzzyDecoders decode input from PHP with tolerance.
	extra.RegisterFuzzyDecoders()
	utils.GenTpl(f)
}

func main() {
	app := cli.NewApp()
	app.Name = "ZbxTable"
	app.Usage = "A Zabbix Table tools"
	cli.VersionPrinter = customVersionPrinter
	cmd.GetVersion(version, gitHash, buildTime)
	app.Version = version
	app.Commands = []*cli.Command{
		cmd.Web,
		cmd.Install,
		cmd.Update,
		//cmd.Init,
		cmd.Uninstall,
	}
	app.Run(os.Args)
}
