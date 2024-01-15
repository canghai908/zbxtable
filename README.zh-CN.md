[English](./README.md) | 简体中文

# ZbxTable

[![Build Status](https://drone.cactifans.org/api/badges/canghai908/zbxtable/status.svg?ref=refs/heads/2.1)](https://drone.cactifans.org/canghai908/zbxtable)

ZbxTable 是使用 Go 语言开发的一个 Zabbix 报表系统。

## 主要功能：

- 开箱即用的微信和邮件告警
- 多租户告警消息分发、屏蔽、统计、到处
- 接口流量报表
- 可视化绘制拓扑图
- 设备分类展示及导出

## 系统架构

![1](/zbxtable.png)

## 组件介绍

ZbxTable: 使用 beego 框架编写的后端程序

ZbxTable-Web: 使用 Vue 编写的前端

MS-Agent: 安装在 Zabbix Server 上, 用于接收 Zabbix Server 产生的告警，并发送到 ZbxTable 平台

## Demo

[https://demo.zbxtable.com](https://demo.zbxtable.com)

## 兼容性

| zabbix 版本 | 兼容性 |
|:----------| :----- |
| 6.4.x     | ✅     |
| 6.2.x     | ✅     |
| 6.0.x     | ✅     |
| 5.4.x     | ✅     |
| 5.2.x     | ✅     |
| 5.0.x LTS | ✅     |
| 4.4.x     | ✅     |
| 4.2.x     | ✅     |
| 4.0.x LTS | ✅     |
| 3.4.x     | 未测试 |
| 3.2.x     | 未测试 |
| 3.0.x LTS | 未测试 |

## 文档

[ZbxTable 文档](https://zbxtable.com)

## 源码

**ZbxTable**: [https://github.com/canghai908/zbxtable](https://github.com/canghai908/zbxtable)

**ZbxTable-Web**: [https://github.com/canghai908/zbxtable-web](https://github.com/canghai908/zbxtable-web)

**MS-Agent**: [https://github.com/canghai908/ms-agent](https://github.com/canghai908/ms-agent)

## 编译

环境：go >=1.18

```
mkdir -p $GOPATH/src/github.com/canghai908
cd $GOPATH/src/github.com/canghai908
git clone github.com/canghai908/zbxtable.git
cd zbxtable
./control build
./control pack
```

## Team

后端

[canghai908](https://github.com/canghai908)

前端

[ahyiru](https://github.com/ahyiru)

## License

<img alt="Apache-2.0 license" src="https://s3-gz01.didistatic.com/n9e-pub/image/apache.jpeg" width="128">

Nightingale is available under the Apache-2.0 license. See the [LICENSE](LICENSE) file for more info.
