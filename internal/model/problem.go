package models

import (
	"github.com/astaxie/beego/logs"
)

//ProblemsRes rest
type ProblemsRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items []Problems `json:"items"`
		Total int64      `json:"total"`
	} `json:"data"`
}

//Problems struct
type Problems struct {
	Acknowledged  string `json:"acknowledged"`
	Clock         string `json:"clock"`
	Correlationid string `json:"correlationid"`
	Eventid       string `json:"eventid"`
	Name          string `json:"name"`
	Objectid      string `json:"objectid"`
	Severity      string `json:"severity"`
}

//GetProblems get porblems
func GetProblems() ([]Problems, int64, error) {
	par := []string{"eventid"}
	problems, err := API.CallWithError("problem.get", Params{"output": "extend",
		"sortfield": par,
		"sortorder": "DESC"})
	if err != nil {
		logs.Error(err)
		return []Problems{}, 0, err
	}
	hba, err := json.Marshal(problems.Result)
	if err != nil {
		logs.Error(err)
		return []Problems{}, 0, err
	}
	var hb []Problems
	err = json.Unmarshal(hba, &hb)
	if err != nil {
		logs.Error(err)
		return []Problems{}, 0, err
	}
	return hb, int64(len(hb)), err
}
