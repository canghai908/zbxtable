package cmd

import (
	"github.com/astaxie/beego/logs"
	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/urfave/cli/v2"
	"zbxtable/models"
)

var updater = &selfupdate.Updater{
	// Manually update the const, or set it using `go build -ldflags="-X main.VERSION=<newver>" -o hello-updater src/hello-updater/main.go`
	ApiURL:     "http://dl.cactifans.com/stable/", // The server hosting `$CmdName/$GOOS-$ARCH.json` which contains the checksum for the binary
	BinURL:     "http://dl.cactifans.com/stable/", // The server hosting the zip file containing the binary application which is a fallback for the patch method
	DiffURL:    "http://dl.cactifans.com/stable/", // The server hosting the binary patch diff for incremental updates
	Dir:        "update/",                         // The directory created by the app when run which stores the cktime file
	CmdName:    "zbxtable",                        // The app name which is appended to the ApiURL to look for an update
	ForceCheck: true,                              // For this example, always check for an update unless the version is "dev"
}

func GetVersion(version, gitHash, buildTime string) selfupdate.Updater {
	models.Version = version
	models.GitHash = gitHash
	models.BuildTime = buildTime
	updater.CurrentVersion = version
	return *updater
}

var (
	// Install cli
	Update = &cli.Command{
		Name:   "update",
		Usage:  "update zbxtable",
		Action: update,
	}
)

func update(*cli.Context) error {
	//获取更新信息
	UpdateVersion, err := updater.UpdateAvailable()
	//更新出错
	if err != nil {
		logs.Error("Update failed!", err)
		return err
	}
	//版本需要更新
	if UpdateVersion != "" {
		logs.Info("Start updating zbxtable!")
		err := updater.BackgroundRun()
		if err != nil {
			logs.Error("Update failed!", err)
			return err
		}
		//更新成功
		logs.Info("Successfully updated to version %s", updater.Info.Version)
		logs.Info("Update Next run, I should be %v", updater.Info.Version)
		return nil
	} else {
		logs.Info("The current version is %v ", updater.CurrentVersion)
		logs.Info("The latest version is %v ", updater.Info.Version)
		logs.Info("The current version is the latest version, no need to update!")
		return nil
	}
	return nil
}
