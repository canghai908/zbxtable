# ZbxTable 文档中心

欢迎来到 ZbxTable 文档中心！这里包含了所有您需要的文档和指南。

## 📚 文档导航

### 快速开始

- **[快速开始指南](./QUICKSTART.md)** - 5分钟快速部署和使用 ZbxTable
- **[README](../README.md)** - 项目概述和基本信息
- **[中文 README](../README.zh-CN.md)** - 中文版项目说明

### 部署指南

- **[Docker 部署指南](../deployments/docker/README.md)** - 使用 Docker 和 Docker Compose 部署
- **[系统说明文档](./SYSTEM.md)** - 完整的系统文档，包含安装、配置、使用等

### 开发文档

- **[贡献指南](../CONTRIBUTING.md)** - 如何为项目做出贡献
- **[API 文档](./api/README.md)** - RESTful API 接口文档
- **[更新日志](../CHANGELOG.md)** - 版本更新记录

### 配置文件

- **[配置文件示例](../config/app.conf.example)** - 详细的配置文件说明
- **[配置文件](../config/app.conf)** - 实际使用的配置文件

## 🚀 快速链接

### 新手入门

1. [环境要求](./QUICKSTART.md#环境要求)
2. [Docker 快速部署](./QUICKSTART.md#docker-快速部署)
3. [首次配置](./QUICKSTART.md#首次配置)
4. [基本使用](./QUICKSTART.md#基本使用)

### 常见任务

- [生成主机报表](./SYSTEM.md#生成报表)
- [配置定时任务](./SYSTEM.md#配置定时任务)
- [创建拓扑图](./SYSTEM.md#创建拓扑图)
- [配置邮件通知](./QUICKSTART.md#配置邮件通知)
- [配置企业微信](./QUICKSTART.md#配置企业微信通知)

### 运维管理

- [日志管理](./SYSTEM.md#日志配置说明)
- [数据库备份](../deployments/docker/README.md#备份与恢复)
- [性能优化](./SYSTEM.md#性能优化)
- [安全建议](./SYSTEM.md#安全建议)

### 故障排查

- [常见问题](./SYSTEM.md#常见问题)
- [Docker 故障排查](../deployments/docker/README.md#故障排查)
- [查看日志](./QUICKSTART.md#查看日志)

## 📖 文档结构

```
zbxtable/
├── README.md                          # 项目说明（英文）
├── README.zh-CN.md                    # 项目说明（中文）
├── CHANGELOG.md                       # 更新日志
├── CONTRIBUTING.md                    # 贡献指南
├── LICENSE                            # 许可证
├── Dockerfile                         # Docker 构建文件
├── docker-compose.yml                 # Docker Compose 配置
├── .dockerignore                      # Docker 忽略文件
│
├── config/
│   ├── app.conf                       # 配置文件
│   └── app.conf.example               # 配置文件示例
│
├── docs/                              # 文档目录
│   ├── README.md                      # 文档索引（本文件）
│   ├── QUICKSTART.md                  # 快速开始指南
│   ├── SYSTEM.md                      # 系统说明文档
│   └── api/
│       └── README.md                  # API 文档
│
└── deployments/
    └── docker/
        ├── README.md                  # Docker 部署指南
        └── mysql/
            └── init.sql               # MySQL 初始化脚本
```

## 🎯 按角色查找文档

### 系统管理员

1. [Docker 部署指南](../deployments/docker/README.md) - 部署和运维
2. [系统说明文档](./SYSTEM.md) - 系统配置和管理
3. [配置文件说明](../config/app.conf.example) - 配置参数详解

### 普通用户

1. [快速开始指南](./QUICKSTART.md) - 快速上手
2. [基本使用](./QUICKSTART.md#基本使用) - 日常操作
3. [常见问题](./SYSTEM.md#常见问题) - 问题解答

### 开发者

1. [贡献指南](../CONTRIBUTING.md) - 参与开发
2. [API 文档](./api/README.md) - API 接口
3. [系统架构](./SYSTEM.md#系统架构) - 架构设计

## 🔍 按主题查找文档

### 安装部署

- [Docker 部署](../deployments/docker/README.md)
- [二进制部署](./QUICKSTART.md#二进制快速部署)
- [源码编译](../README.md#compile-from-source)
- [环境要求](./QUICKSTART.md#环境要求)

### 配置管理

- [应用配置](./SYSTEM.md#应用配置-configappconf)
- [日志配置](./SYSTEM.md#日志配置说明)
- [数据库配置](./QUICKSTART.md#准备数据库)
- [Zabbix 配置](./QUICKSTART.md#zabbix-配置)

### 功能使用

- [主机管理](./SYSTEM.md#查看主机列表)
- [报表管理](./SYSTEM.md#生成报表)
- [告警分析](./SYSTEM.md#查看告警统计)
- [拓扑管理](./SYSTEM.md#创建拓扑图)
- [定时任务](./SYSTEM.md#配置定时任务)

### 系统管理

- [用户管理](./QUICKSTART.md#用户管理)
- [权限管理](./SYSTEM.md#权限管理)
- [实例管理](./QUICKSTART.md#添加-zabbix-实例可选)
- [系统配置](./QUICKSTART.md#配置邮件通知)

### 运维监控

- [日志查看](./QUICKSTART.md#查看日志)
- [健康检查](../deployments/docker/README.md#查看健康检查状态)
- [性能优化](./SYSTEM.md#性能优化)
- [备份恢复](../deployments/docker/README.md#备份与恢复)

### 开发相关

- [开发环境](../CONTRIBUTING.md#开发环境搭建)
- [代码规范](../CONTRIBUTING.md#代码规范)
- [测试指南](../CONTRIBUTING.md#测试)
- [API 开发](./api/README.md)

## 💡 学习路径

### 初级用户

1. 阅读 [README](../README.md) 了解项目
2. 按照 [快速开始指南](./QUICKSTART.md) 部署系统
3. 完成 [首次配置](./QUICKSTART.md#首次配置)
4. 学习 [基本使用](./QUICKSTART.md#基本使用)

### 中级用户

1. 深入学习 [系统说明文档](./SYSTEM.md)
2. 配置 [定时任务](./SYSTEM.md#配置定时任务)
3. 创建 [自定义拓扑](./SYSTEM.md#创建拓扑图)
4. 了解 [性能优化](./SYSTEM.md#性能优化)

### 高级用户

1. 学习 [Docker 部署](../deployments/docker/README.md)
2. 掌握 [API 使用](./api/README.md)
3. 参与 [项目贡献](../CONTRIBUTING.md)
4. 进行 [二次开发](../CONTRIBUTING.md#开发环境搭建)

## 🌟 推荐阅读顺序

### 第一次使用

1. [README](../README.md) - 了解项目
2. [快速开始指南](./QUICKSTART.md) - 快速部署
3. [基本使用](./QUICKSTART.md#基本使用) - 学习操作

### 生产环境部署

1. [环境要求](./QUICKSTART.md#环境要求) - 确认环境
2. [Docker 部署指南](../deployments/docker/README.md) - 部署系统
3. [系统说明文档](./SYSTEM.md) - 深入了解
4. [安全建议](./SYSTEM.md#安全建议) - 加固系统

### 开发贡献

1. [贡献指南](../CONTRIBUTING.md) - 了解流程
2. [开发环境搭建](../CONTRIBUTING.md#开发环境搭建) - 准备环境
3. [代码规范](../CONTRIBUTING.md#代码规范) - 遵循规范
4. [API 文档](./api/README.md) - 接口开发

## 📞 获取帮助

### 文档问题

如果您在文档中发现错误或有改进建议：

1. 提交 [Issue](https://github.com/canghai908/zbxtable/issues)
2. 或直接提交 [Pull Request](https://github.com/canghai908/zbxtable/pulls)

### 使用问题

如果您在使用过程中遇到问题：

1. 查看 [常见问题](./SYSTEM.md#常见问题)
2. 搜索 [Issues](https://github.com/canghai908/zbxtable/issues)
3. 创建新的 [Issue](https://github.com/canghai908/zbxtable/issues/new)

### 联系我们

- **GitHub**: https://github.com/canghai908/zbxtable
- **官网**: https://zbxtable.com
- **邮箱**: support@zbxtable.com

## 🔄 文档更新

文档会随着项目更新而持续改进。最后更新时间：2024-01-29

查看 [CHANGELOG](../CHANGELOG.md) 了解最新变更。

## 📝 文档贡献

我们欢迎文档贡献！您可以：

- 修正错误
- 改进表达
- 添加示例
- 翻译文档
- 补充缺失内容

详见 [贡献指南](../CONTRIBUTING.md#改进文档)。

## ⭐ 相关资源

### 官方资源

- [GitHub 仓库](https://github.com/canghai908/zbxtable)
- [前端仓库](https://github.com/canghai908/zbxtable-web)
- [MS-Agent](https://github.com/canghai908/ms-agent)
- [在线演示](https://demo.zbxtable.com)

### 社区资源

- [Zabbix 官网](https://www.zabbix.com)
- [Zabbix 中文社区](https://www.zabbix.org.cn)
- [Go 语言官网](https://golang.org)
- [Vue.js 官网](https://vuejs.org)

---

**提示**：建议将本页面加入书签，方便随时查阅！

