package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/canghai908/zbxtable/cmd"
	_ "github.com/canghai908/zbxtable/packfile"
	"github.com/json-iterator/go/extra"
	"github.com/urfave/cli/v2"
)

//AppVersion version
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
}
func main() {
	app := cli.NewApp()
	app.Name = "ZbxTable"
	app.Usage = "A Zabbix Table tools"
	cli.VersionPrinter = customVersionPrinter
	app.Version = version
	app.Commands = []*cli.Command{
		cmd.Web,
		cmd.Install,
		cmd.Init,
		//cmd.Uc,
		cmd.Ua,
		cmd.Un,
	}
	app.Run(os.Args)
}
