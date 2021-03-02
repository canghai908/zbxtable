package cmd

import (
	"github.com/astaxie/beego/logs"
	"github.com/urfave/cli/v2"
)

var (
	// Install cli
	Un = &cli.Command{
		Name:   "un",
		Usage:  "Uninstall action",
		Action: UninstallAction,
	}
)

func UninstallAction(*cli.Context) error {
	//CheckConfExist
	CheckConfExist()
	//login zabbix to get version
	_, err := CheckZabbixAPI(InitConfig("zabbix_web"), InitConfig("zabbix_user"), InitConfig("zabbix_pass"))
	if err != nil {
		logs.Error(err)
		return err
	}
	//action delete
	actionid, err := GetActionID(MSAction)
	if err != nil {
		logs.Error(err)
		return err
	}
	actionmap := make(map[string]string, 0)
	actionmap["params"] = actionid
	_, err = API.CallWithError("action.delete", actionmap)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Delete " + MSAction + " action successfully!")
	//user delete
	userinfo, err := GetUserID(MSUser)
	if err != nil {
		logs.Error(err)
		return err
	}
	usermap := make(map[string]string, 0)
	usermap["params"] = userinfo.Userid
	_, err = API.CallWithError("user.delete", usermap)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Delete " + MSUser + " user successfully!")
	//usergroup delete
	usergroupmap := make(map[string]string, 0)
	usergroupmap["params"] = userinfo.Usrgrps[0].Usrgrpid
	_, err = API.CallWithError("usergroup.delete", usergroupmap)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Delete " + MSGroup + " group successfully!")
	//	mediatype delete
	mediatypemap := make(map[string]interface{}, 0)
	mepar := []string{userinfo.Mediatypes[0].Mediatypeid}
	mediatypemap["params"] = mepar

	_, err = API.CallWithError("mediatype.delete", mepar)
	if err != nil {
		logs.Error(err)
		return err
	}
	logs.Info("Delete " + MSMedia + " mediatype successfully!")
	return nil
}
