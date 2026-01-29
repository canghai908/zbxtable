# 贡献指南

感谢您对 ZbxTable 项目的关注！我们欢迎任何形式的贡献。

## 目录

- [行为准则](#行为准则)
- [如何贡献](#如何贡献)
- [开发环境搭建](#开发环境搭建)
- [代码规范](#代码规范)
- [提交规范](#提交规范)
- [Pull Request 流程](#pull-request-流程)
- [问题反馈](#问题反馈)

## 行为准则

参与本项目即表示您同意遵守我们的行为准则：

- 尊重所有贡献者
- 使用友好和包容的语言
- 接受建设性的批评
- 关注对社区最有利的事情
- 对其他社区成员表示同理心

## 如何贡献

您可以通过以下方式为项目做出贡献：

### 1. 报告 Bug

如果您发现了 Bug，请：

1. 检查 [Issues](https://github.com/canghai908/zbxtable/issues) 确认问题是否已被报告
2. 如果没有，创建一个新的 Issue
3. 使用清晰的标题描述问题
4. 提供详细的复现步骤
5. 说明预期行为和实际行为
6. 提供环境信息（操作系统、版本等）
7. 如果可能，提供截图或日志

### 2. 提出新功能

如果您有新功能的想法：

1. 检查 [Issues](https://github.com/canghai908/zbxtable/issues) 确认功能是否已被提出
2. 创建一个 Feature Request Issue
3. 详细描述功能需求和使用场景
4. 说明为什么这个功能对用户有价值

### 3. 改进文档

文档改进包括：

- 修正拼写或语法错误
- 添加缺失的文档
- 改进现有文档的清晰度
- 添加示例和教程
- 翻译文档

### 4. 贡献代码

我们欢迎代码贡献！请参考下面的开发流程。

## 开发环境搭建

### 后端开发环境

1. **安装 Go**

```bash
# 下载并安装 Go 1.23+
wget https://go.dev/dl/go1.23.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.23.0.linux-amd64.tar.gz

# 配置环境变量
export PATH=$PATH:/usr/local/go/bin
export GOPATH=$HOME/go
export PATH=$PATH:$GOPATH/bin
```

2. **克隆项目**

```bash
git clone https://github.com/canghai908/zbxtable.git
cd zbxtable/zbxtable
```

3. **安装依赖**

```bash
go mod download
```

4. **配置数据库**

```bash
# 创建数据库
mysql -uroot -p -e "CREATE DATABASE zbxtable_dev DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"

# 复制配置文件
cp config/app.conf.example config/app.conf

# 编辑配置文件
vim config/app.conf
```

5. **运行项目**

```bash
# 开发模式运行
go run main.go web

# 或编译后运行
go build -o zbxtable main.go
./zbxtable web
```

### 前端开发环境

1. **安装 Node.js**

```bash
# 安装 Node.js 16+
curl -fsSL https://deb.nodesource.com/setup_16.x | sudo -E bash -
sudo apt-get install -y nodejs
```

2. **克隆前端项目**

```bash
git clone https://github.com/canghai908/zbxtable-web.git
cd zbxtable-web
```

3. **安装依赖**

```bash
npm install
```

4. **运行开发服务器**

```bash
npm run serve
```

5. **构建生产版本**

```bash
npm run build
```

## 代码规范

### Go 代码规范

遵循 [Effective Go](https://golang.org/doc/effective_go.html) 和 [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)。

**关键点**：

- 使用 `gofmt` 格式化代码
- 使用有意义的变量和函数名
- 添加必要的注释
- 错误处理要完善
- 避免使用全局变量
- 单元测试覆盖率 > 60%

**示例**：

```go
// GetHostByID 根据 ID 获取主机信息
// 参数：
//   id: 主机 ID
// 返回：
//   *Host: 主机对象
//   error: 错误信息
func GetHostByID(id int64) (*Host, error) {
    if id <= 0 {
        return nil, errors.New("invalid host id")
    }
    
    var host Host
    err := db.Where("id = ?", id).First(&host).Error
    if err != nil {
        return nil, fmt.Errorf("failed to get host: %w", err)
    }
    
    return &host, nil
}
```

### JavaScript/Vue 代码规范

遵循 [Vue Style Guide](https://vuejs.org/style-guide/) 和 [Airbnb JavaScript Style Guide](https://github.com/airbnb/javascript)。

**关键点**：

- 使用 ESLint 检查代码
- 组件名使用 PascalCase
- Props 定义要完整
- 使用 computed 而不是 watch（如果可能）
- 避免在模板中使用复杂表达式

### 提交信息规范

使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Type 类型**：

- `feat`: 新功能
- `fix`: Bug 修复
- `docs`: 文档更新
- `style`: 代码格式调整（不影响功能）
- `refactor`: 代码重构
- `perf`: 性能优化
- `test`: 测试相关
- `chore`: 构建/工具相关

**示例**：

```
feat(report): 添加 PDF 导出功能

- 实现 PDF 报表生成
- 添加自定义模板支持
- 优化导出性能

Closes #123
```

## Pull Request 流程

### 1. Fork 项目

点击 GitHub 页面右上角的 "Fork" 按钮。

### 2. 克隆到本地

```bash
git clone https://github.com/YOUR_USERNAME/zbxtable.git
cd zbxtable
```

### 3. 创建分支

```bash
# 从 main 分支创建新分支
git checkout -b feature/your-feature-name

# 或修复 bug
git checkout -b fix/your-bug-fix
```

### 4. 进行修改

- 编写代码
- 添加测试
- 更新文档
- 确保测试通过

```bash
# 运行测试
go test ./...

# 格式化代码
gofmt -w .

# 检查代码
go vet ./...
```

### 5. 提交更改

```bash
git add .
git commit -m "feat: add new feature"
```

### 6. 推送到 GitHub

```bash
git push origin feature/your-feature-name
```

### 7. 创建 Pull Request

1. 访问您的 Fork 仓库
2. 点击 "New Pull Request"
3. 选择您的分支
4. 填写 PR 描述：
   - 说明改动内容
   - 关联相关 Issue
   - 添加截图（如果适用）
5. 提交 PR

### 8. 代码审查

- 维护者会审查您的代码
- 根据反馈进行修改
- 保持沟通和耐心

### 9. 合并

代码审查通过后，维护者会合并您的 PR。

## 测试

### 单元测试

```bash
# 运行所有测试
go test ./...

# 运行特定包的测试
go test ./internal/model

# 查看测试覆盖率
go test -cover ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### 集成测试

```bash
# 运行集成测试
go test -tags=integration ./tests/integration/...
```

## 问题反馈

### 创建 Issue

访问 [Issues](https://github.com/canghai908/zbxtable/issues) 页面创建新 Issue。

**Bug 报告模板**：

```markdown
**描述问题**
简要描述遇到的问题。

**复现步骤**
1. 进入 '...'
2. 点击 '...'
3. 滚动到 '...'
4. 看到错误

**预期行为**
描述您期望发生什么。

**实际行为**
描述实际发生了什么。

**截图**
如果适用，添加截图帮助解释问题。

**环境信息**
- 操作系统: [如 Ubuntu 20.04]
- ZbxTable 版本: [如 2.1.0]
- Zabbix 版本: [如 6.0.10]
- 浏览器: [如 Chrome 120]

**日志**
粘贴相关日志信息。

**其他信息**
添加其他有助于解决问题的信息。
```

**功能请求模板**：

```markdown
**功能描述**
简要描述您想要的功能。

**使用场景**
描述这个功能的使用场景。

**建议的实现方式**
如果有想法，描述您认为应该如何实现。

**替代方案**
描述您考虑过的其他解决方案。

**其他信息**
添加其他相关信息或截图。
```

## 开发技巧

### 调试

```bash
# 使用 delve 调试
go install github.com/go-delve/delve/cmd/dlv@latest
dlv debug main.go -- web
```

### 性能分析

```bash
# CPU 性能分析
go test -cpuprofile=cpu.prof -bench=.
go tool pprof cpu.prof

# 内存性能分析
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

### 代码生成

```bash
# 生成 mock
go install github.com/golang/mock/mockgen@latest
mockgen -source=internal/model/host.go -destination=internal/model/mock/host_mock.go
```

## 许可证

贡献的代码将使用与项目相同的 [Apache-2.0](../LICENSE) 许可证。

## 联系方式

- GitHub Issues: https://github.com/canghai908/zbxtable/issues
- 邮箱: support@zbxtable.com

## 致谢

感谢所有为 ZbxTable 做出贡献的开发者！

您的贡献将被记录在 [CHANGELOG.md](../CHANGELOG.md) 中。

