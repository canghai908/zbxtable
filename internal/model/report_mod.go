package model

import "time"

// Reports
type Report struct {
	ID              int        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Name            string     `gorm:"column:name;size:255" json:"name"`
	ReportType      string     `gorm:"column:report_type;size:255" json:"report_type"`
	Instance        int        `gorm:"column:instance;default:0" json:"instance"`
	Items           string     `gorm:"column:items;size:200" json:"items"`
	LinkBandWidth   string     `gorm:"column:link_band_width;size:200" json:"linkbandwidth"`
	HostIds         string     `gorm:"column:host_ids;type:text" json:"host_ids"`
	ResolvedHostIds string     `gorm:"column:resolved_host_ids;type:text" json:"resolved_host_ids"`
	ItemIds         string     `gorm:"column:item_ids;type:text" json:"item_ids"`
	Cycle           string     `gorm:"column:cycle;size:200" json:"cycle"`
	Desc            string     `gorm:"column:desc;size:200" json:"desc"`
	Emails          string     `gorm:"column:emails;size:240" json:"emails"`
	Status          string     `gorm:"column:status;size:50" json:"status"`
	ExecStatus      string     `gorm:"column:exec_status;default:0" json:"exec_status"`
	ReportMode      string     `gorm:"column:report_mode;size:50;default:scheduled" json:"report_mode"`
	Start           *time.Time `gorm:"column:start;type:timestamp" json:"start"`
	End             *time.Time `gorm:"column:end;type:timestamp" json:"end"`
	StartAt         *time.Time `gorm:"column:start_at;type:timestamp" json:"start_at"`
	EndAt           *time.Time `gorm:"column:end_at;type:timestamp" json:"end_at"`
	CreatedAt       time.Time  `gorm:"column:created_at;autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time  `gorm:"column:updated_at;autoUpdateTime" json:"updated_at"`
}

// SystemList struct
type ReportRes struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items interface{} `json:"items"`
		Total int64       `json:"total"`
	} `json:"data"`
}

var htmlReport = `<div>
<includetail>
        <div style="font: Verdana normal 14px; color: #000">
            <div style="position: relative">
                <div class="eml-w eml-w-sys-layout">
                    <div style="font-size: 0px">
                        <div class="eml-w-sys-line">
                            <div class="eml-w-sys-line-left"></div>
                            <div class="eml-w-sys-line-right"></div>
                        </div>
                    </div>
                    <div class="eml-w-sys-content">
                        <div class="dragArea gen-group-list">
                            <div class="gen-item">
                                <div style="padding: 0px">
                                    <div class="eml-w-phase-normal-16">
                                        链路名称:{{.LinkName}}
                                    </div>
                                    <div class="eml-w-phase-normal-16">
                                        报表周期：{{.Start}}--{{.End}}
                                    </div>
                                </div>
                            </div>
                            <div class="gen-item">
                                <table border="1">
                                    <caption>
                                        {{.Title}}
                                    </caption>
                                    <tr>
										<th>设备名称</th>
										<th>IP地址</th>
                                        <th>指标名称</th>
                                        <th>带宽</th>
                                        <th>平均流量</th>
                                        <th>平均使用率</th>
                                    </tr>
                                    {{range .TableInfo}}
                                    <tr>
                                        <td>{{.Host}}</td>
										<td>{{.IP}}</td>
										<td>{{.ItemName}}</td>
                                        <td>{{.LinkBinWith}}</td>
                                        <td>{{.Avg}}</td>
                                        <td>{{.AVgPre }}</td>
                                    </tr>
                                    {{end}}
                                </table>
                            </div>
                        </div>
                    </div>
                    <div class="eml-w-sys-footer">{{.EndLine}}</div>
                </div>
            </div>
        </div>
</includetail>
</div>
<style>
    .eml-w .eml-w-phase-normal-16 {
        color: #2b2b2b;
        font-size: 16px;
        line-height: 1.75;
    }

    .eml-w .eml-w-phase-bold-16 {
        font-size: 16px;
        color: #2b2b2b;
        font-weight: 500;
        line-height: 1;
    }

    .eml-w-title-level1 {
        font-size: 20px;
        font-weight: 500;
        padding: 15px 0;
    }

    .eml-w-title-level3 {
        font-size: 16px;
        font-weight: 500;
        padding-bottom: 10px;
    }

    .eml-w-title-level3.center {
        text-align: center;
    }

    .eml-w-phase-small-normal {
        font-size: 14px;
        color: #2b2b2b;
        line-height: 1.75;
    }

    .eml-w-picture-wrap {
        padding: 10px 0;
        width: 100%;
        overflow: hidden;
    }

    .eml-w-picture-full-img {
        display: block;
        width: auto;
        max-width: 100%;
        margin: 0 auto;
    }

    .eml-w-sys-layout {
        background: #fff;
        box-shadow: 0 2px 8px 0 rgba(0, 0, 0, 0.2);
        border-radius: 4px;
        margin: 50px auto;
        max-width: 800px;
        overflow: hidden;
    }

    .eml-w-sys-line-left {
        display: inline-block;
        width: 88%;
        background: #2984ef;
        height: 3px;
    }

    .eml-w-sys-line-right {
        display: inline-block;
        width: 11.5%;
        height: 3px;
        background: #8bd5ff;
        margin-left: 1px;
    }

    .eml-w-sys-logo {
        text-align: right;
    }

    .eml-w-sys-logo img {
        display: inline-block;
        margin: 30px 50px 0 0;
    }

    .eml-w-sys-content {
        position: relative;
        padding: 20px 50px 0;
        min-height: 16px;
        word-break: break-all;
    }

    .eml-w-sys-footer {
        font-weight: 500;
        font-size: 12px;
        color: #bebebe;
        letter-spacing: 0.5px;
        padding: 0 0 30px 50px;
        margin-top: 60px;
    }

    .eml-w {
        font-family: Helvetica Neue, Arial, PingFang SC, Hiragino Sans GB, STHeiti,
            Microsoft YaHei, sans-serif;
        -webkit-font-smoothing: antialiased;
        color: #2b2b2b;
        font-size: 14px;
        line-height: 1.75;
    }

    .eml-w a {
        text-decoration: none;
    }

    .eml-w a,
    .eml-w a:active {
        color: #186fd5;
    }

    .eml-w h1,
    .eml-w h2,
    .eml-w h3,
    .eml-w h4,
    .eml-w h5,
    .eml-w h6,
    .eml-w li,
    .eml-w p,
    .eml-w ul {
        margin: 0;
        padding: 0;
    }
</style>`
