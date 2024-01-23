package cmd

import (
	"github.com/pterm/pterm"
	"github.com/sanbornm/go-selfupdate/selfupdate"
	"github.com/urfave/cli/v2"
	"os"
	"zbxtable/models"
)

const UpdateURL = "http://dl.cactifans.com/stable/"

var updater = &selfupdate.Updater{
	// Manually update the const, or set it using `go build -ldflags="-X main.VERSION=<newver>" -o hello-updater src/hello-updater/main.go`
	ApiURL:     UpdateURL,  // The server hosting `$CmdName/$GOOS-$ARCH.json` which contains the checksum for the binary
	BinURL:     UpdateURL,  // The server hosting the zip file containing the binary application which is a fallback for the patch method
	DiffURL:    UpdateURL,  // The server hosting the binary patch diff for incremental updates
	Dir:        "update/",  // The directory created by the app when run which stores the cktime file
	CmdName:    "zbxtable", // The app name which is appended to the ApiURL to look for an update
	ForceCheck: true,       // For this example, always check for an update unless the version is "dev"
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

// update 更新
func update(*cli.Context) error {
	//获取更新信息
	UpdateVersion, err := updater.UpdateAvailable()
	//更新出错
	if err != nil {
		pterm.Error.Println("Update failed!", err)
		return err
	}
	//版本需要更新
	if UpdateVersion != "" {
		pterm.Info.Println("A new version is available!")
		//提示用户升级会删除web目录
		pterm.Println(pterm.Red("Warning: The upgrade operation will delete the old web directory!!!"))
		options := []string{"yes", "no"}
		prompt := pterm.DefaultInteractiveContinue.WithDefaultText("Do you want to update now?").WithOptions(options)
		result, _ := prompt.Show()
		// Print a blank line for better readability
		pterm.Println()
		// Print the user's input with an info prefix
		// As this is a continue prompt, the input should be empty
		switch result {
		case "yes":
			// 删除旧的web备份目录，
			// 备份web目录
			_ = os.RemoveAll("./.web")
			_ = os.Rename("./web", "./.web")
			spinnerInfo, _ := pterm.DefaultSpinner.Start("Start updating ...")
			err := updater.Update()
			if err != nil {
				//升级失败
				spinnerInfo.Fail("Update failed!", err)
				pterm.Println()
				return err
			}
			spinnerInfo.Success("Successfully updated to version ", updater.Info.Version)
			spinnerInfo.Success("Next run, I should be ", updater.Info.Version)
			pterm.Println()
			return nil
		default:
			return nil
		}

	} else {
		pterm.Info.Println("The current version is ", updater.CurrentVersion)
		pterm.Info.Println("The latest version is ", updater.Info.Version)
		pterm.Info.Println("The current version is the latest version, no need to update!")
		return nil
	}
	return nil
}
