 [English](./README.md) | 简体中文

# ZbxTable

ZbxTable 是使用 Go 语言开发的一个 Zabbix 报表系统。

## 主要功能：

* 导出监控指标特定时间段内的详情数据与趋势数据到 xlsx
* 导出特定时间段内 Zabbix 的告警消息到 xlsx
* 对特定时间段内的告警消息进行分析，告警 Top10 等
* 按照主机组导出巡检报告
* 对 Zabbix 图形按照数类型进行显示和查看并支持导出到 pdf
* 主机未恢复告警显示和查询

## 系统架构

![1](https://img.cactifans.com/wp-content/uploads/2020/07/zbxtable.png)

## 组件介绍

ZbxTable: 使用 beego 框架编写的后端程序

ZbxTable-Web: 使用 React 编写的前端

MS-Agent: 安装在 Zabbix Server 上, 用于接收 Zabbix Server 产生的告警，并发送到 ZbxTable 平台

## 在线体验

直接点击登录即可

[https://zbx.cactifans.com](https://zbx.cactifans.com)

## 兼容性

| zabbix 版本 | 兼容性            |
| :---------- | :---------------- |
| 5.0.x LTS   | ✅                |
| 4.4.x       | ✅                |
| 4.2.x       | ✅                |
| 4.0.x LTS   | ✅                |
| 3.4.x       | ❓ 理论支持未测试 |
| 3.2.x       | ❓ 理论支持未测试 |
| 3.0.x LTS   | ❓ 理论支持未测试 |

## 文档

[ZbxTable 使用说明](https://zbxtable.cactifans.com)

## 源码

ZbxTable: [https://github.com/canghai908/zbxtable](https://github.com/canghai908/zbxtable)

ZbxTable-Web: [https://github.com/canghai908/zbxtable-web](https://github.com/canghai908/zbxtable-web)

MS-Agent: [https://github.com/canghai908/ms-agent](https://github.com/canghai908/ms-agent)

## 编译

``` 
mkdir -p $GOPATH/src/github.com/canghai908
cd $GOPATH/src/github.com/canghai908
git clone https://github.com/canghai908/zbxtable.git
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
