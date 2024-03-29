English | [简体中文](./README.zh-CN.md)

# ZbxTable

[![Build Status](https://drone.cactifans.org/api/badges/canghai908/zbxtable/status.svg?ref=refs/heads/2.1)](https://drone.cactifans.org/canghai908/zbxtable)

ZbxTable is a Zabbix report system developed using Go language.

## Features

- Custom drawing topology
- Device classification display and export
- Export Zabbix alert messages to xlsx within a specific time period
- Analyze the alarm messages in a specific time period, alarm Top 10, etc.

## Architecture

![1](/zbxtable.png)

## Component

ZbxTable: Backend written using [beego framework](https://github.com/astaxie/beego).

ZbxTable-Web: Front end written using [Vue](https://github.com/vuejs/vue).

MS-Agent: Installed on Zabbix Server, used to receive alarms generated by Zabbix Server and send to ZbxTable.

## Demo

[https://demo.zbxtable.com](https://demo.zbxtable.com)

## Compatibility

| Zabbix Version | Compatibility |
|:---------------| :------------ |
| 6.4.x          | ✅            |
| 6.2.x          | ✅            |
| 6.0.x          | ✅            |
| 5.4.x          | ✅            |
| 5.2.x          | ✅            |
| 5.0.x LTS      | ✅            |
| 4.4.x          | ✅            |
| 4.2.x          | ✅            |
| 4.0.x LTS      | ✅            |
| 3.4.x          | untested      |
| 3.2.x          | untested      |
| 3.0.x LTS      | untested      |

## Documentation

[ZbxTable Documentation](https://zbxtable.com)

## Code

**ZbxTable**: [https://github.com/canghai908/zbxtable](https://github.com/canghai908/zbxtable)

**ZbxTable-Web**: [https://github.com/canghai908/zbxtable-web](https://github.com/canghai908/zbxtable-web)

**MS-Agent**: [https://github.com/canghai908/ms-agent](https://github.com/canghai908/ms-agent)

## Compile

go >=1.21

```
mkdir -p $GOPATH/src/github.com/canghai908
cd $GOPATH/src/github.com/canghai908
git clone github.com/canghai908/zbxtable.git
cd zbxtable
wget -q -c https://dl.cactifans.com/stable/zbxtable/web-latest.tar.gz && tar xf web-latest.tar.gz
go install github.com/go-bindata/go-bindata/go-bindata@latest
./control build
./control pack
```

## Team

Back-end development

[canghai908](https://github.com/canghai908)

Front-end development

[ahyiru](https://github.com/ahyiru)

## License

ZbxTable is available under the Apache-2.0 license. See the [LICENSE](LICENSE) file for more info.
