package main

import (
	"embed"
	"fmt"
	"github.com/astaxie/beego/logs"
	"github.com/json-iterator/go/extra"
	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"zbxtable/cmd"
	"zbxtable/utils"
)

//go:embed template
var f embed.FS

const motd = `
$$$$$$$$\ $$$$$$$\  $$\   $$\ $$$$$$$$\  $$$$$$\  $$$$$$$\  $$\       $$$$$$$$\ 
\____$$  |$$  __$$\ $$ |  $$ |\__$$  __|$$  __$$\ $$  __$$\ $$ |      $$  _____|
    $$  / $$ |  $$ |\$$\ $$  |   $$ |   $$ /  $$ |$$ |  $$ |$$ |      $$ |      
   $$  /  $$$$$$$\ | \$$$$  /    $$ |   $$$$$$$$ |$$$$$$$\ |$$ |      $$$$$\    
  $$  /   $$  __$$\  $$  $$<     $$ |   $$  __$$ |$$  __$$\ $$ |      $$  __|   
 $$  /    $$ |  $$ |$$  /\$$\    $$ |   $$ |  $$ |$$ |  $$ |$$ |      $$ |      
$$$$$$$$\ $$$$$$$  |$$ /  $$ |   $$ |   $$ |  $$ |$$$$$$$  |$$$$$$$$\ $$$$$$$$\ 
\________|\_______/ \__|  \__|   \__|   \__|  \__|\_______/ \________|\________|`

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
	utils.GenTpl(f)
}

var updater = &selfupdate.Updater{
	CurrentVersion: version,                    // Manually update the const, or set it using `go build -ldflags="-X main.VERSION=<newver>" -o hello-updater src/hello-updater/main.go`
	ApiURL:         "http://dl.cactifans.com/", // The server hosting `$CmdName/$GOOS-$ARCH.json` which contains the checksum for the binary
	BinURL:         "http://dl.cactifans.com/", // The server hosting the zip file containing the binary application which is a fallback for the patch method
	DiffURL:        "http://dl.cactifans.com/", // The server hosting the binary patch diff for incremental updates
	Dir:            "update/",                  // The directory created by the app when run which stores the cktime file
	CmdName:        "zbxtable",                 // The app name which is appended to the ApiURL to look for an update
	ForceCheck:     true,                       // For this example, always check for an update unless the version is "dev"
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
		//cmd.Init,
		cmd.Uninstall,
	}
	logs.Info("Currently version %v", updater.CurrentVersion)
	updater.BackgroundRun()
	logs.Info("Next run, I should be %v", updater.Info.Version)
	app.Run(os.Args)
}
