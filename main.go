package main

import (
	"fmt"
	"github.com/json-iterator/go/extra"
	"github.com/urfave/cli/v2"
	"os"
	"path/filepath"
	"zbxtable/cmd"
)

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
	app.Run(os.Args)
}
