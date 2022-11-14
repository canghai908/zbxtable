[{{.Status}}][{{.OccurTime.Format "15:04:05"}}]设备:{{.Hostname}}发生:{{.Message}}故障！
=========================================
告警主机: {{.Hostname}}
主机IP：{{.HostsIP}}
主机分组: {{.Hgroup}}
告警时间: {{.OccurTime.Format "2006-01-02 15:04:05"}}
告警等级: {{.Level}}
告警信息: {{.ItemName}}
告警项目: {{.Hkey}}
问题详情: {{.Detail}}
当前状态: {{.Status}}
事件ID: {{.EventID}}