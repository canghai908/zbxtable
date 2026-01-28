# ZbxTable 项目重构迁移脚本

$ErrorActionPreference = "Stop"
$projectRoot = "d:/canghai/zbxtable/zbxtable"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "ZbxTable 项目结构重构" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# 步骤1: 移动 utils → pkg/utils
Write-Host "[1/9] 移动 utils → pkg/utils..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/utils") {
    Copy-Item -Path "$projectRoot/utils/*" -Destination "$projectRoot/pkg/utils/" -Recurse -Force
    Write-Host "  ✓ utils 文件已复制到 pkg/utils" -ForegroundColor Green
}

# 步骤2: 移动 middleware → internal/middleware
Write-Host "[2/9] 移动 middleware → internal/middleware..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/middleware") {
    Copy-Item -Path "$projectRoot/middleware/*" -Destination "$projectRoot/internal/middleware/" -Recurse -Force
    Write-Host "  ✓ middleware 文件已复制到 internal/middleware" -ForegroundColor Green
}

# 步骤3: 移动 conf → config
Write-Host "[3/9] 移动 conf → config..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/conf") {
    Copy-Item -Path "$projectRoot/conf/*" -Destination "$projectRoot/config/" -Recurse -Force
    Write-Host "  ✓ conf 文件已复制到 config" -ForegroundColor Green
}

# 步骤4: 移动 controllers → internal/handler (去掉_gin后缀)
Write-Host "[4/9] 移动 controllers → internal/handler..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/controllers") {
    Get-ChildItem -Path "$projectRoot/controllers" -Filter "*.go" | ForEach-Object {
        $newName = $_.Name -replace '_gin\.go$', '.go'
        $destPath = Join-Path "$projectRoot/internal/handler" $newName
        Copy-Item -Path $_.FullName -Destination $destPath -Force
    }
    Write-Host "  ✓ controllers 文件已复制到 internal/handler (已去掉_gin后缀)" -ForegroundColor Green
}

# 步骤5: 移动 routers → api/v1 (去掉_gin后缀)
Write-Host "[5/9] 移动 routers → api/v1..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/routers") {
    Get-ChildItem -Path "$projectRoot/routers" -Filter "*.go" | ForEach-Object {
        $newName = $_.Name -replace '_gin\.go$', '.go'
        $newName = $newName -replace 'router_gin\.go$', 'router.go'
        $destPath = Join-Path "$projectRoot/api/v1" $newName
        Copy-Item -Path $_.FullName -Destination $destPath -Force
    }
    Write-Host "  ✓ routers 文件已复制到 api/v1 (已去掉_gin后缀)" -ForegroundColor Green
}

# 步骤6: 复制 models → internal/model (暂时整体复制，后续需要手动拆分)
Write-Host "[6/9] 复制 models → internal/model..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/models") {
    Copy-Item -Path "$projectRoot/models/*" -Destination "$projectRoot/internal/model/" -Recurse -Force
    Write-Host "  ✓ models 文件已复制到 internal/model" -ForegroundColor Green
    Write-Host "  ⚠ 注意: models 需要后续手动拆分为 model/repository/service" -ForegroundColor Yellow
}

# 步骤7: 移动 main.go → cmd/zbxtable/main.go
Write-Host "[7/9] 移动 main.go → cmd/zbxtable/main.go..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/main.go") {
    Copy-Item -Path "$projectRoot/main.go" -Destination "$projectRoot/cmd/zbxtable/main.go" -Force
    Write-Host "  ✓ main.go 已复制到 cmd/zbxtable/" -ForegroundColor Green
}

# 步骤8: 重组 cmd 目录
Write-Host "[8/9] 重组 cmd 目录..." -ForegroundColor Yellow
if (Test-Path "$projectRoot/cmd") {
    $cmdFiles = @("install.go", "update.go", "web.go", "uninstallaction.go", "updateconf.go", "init.go")
    foreach ($file in $cmdFiles) {
        $srcPath = Join-Path "$projectRoot/cmd" $file
        if (Test-Path $srcPath) {
            Copy-Item -Path $srcPath -Destination "$projectRoot/cmd/commands/$file" -Force
        }
    }
    Write-Host "  ✓ cmd 命令文件已复制到 cmd/commands/" -ForegroundColor Green
}

# 步骤9: 创建 pkg/logger 包装
Write-Host "[9/9] 创建 pkg/logger 包装..." -ForegroundColor Yellow
$loggerContent = @"
package logger

import (
    "zbxtable/pkg/utils"
)

// Log 全局日志实例 (兼容旧代码)
var Log = utils.Log

// InitLogger 初始化日志
func InitLogger(logPath string, level, maxday, maxlines, maxsize int, daily bool) error {
    return utils.InitLogger(logPath, level, maxday, maxlines, maxsize, daily)
}
"@
Set-Content -Path "$projectRoot/pkg/logger/logger.go" -Value $loggerContent -Encoding UTF8
Write-Host "  ✓ pkg/logger 包装已创建" -ForegroundColor Green

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "文件迁移完成！" -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "下一步操作:" -ForegroundColor Yellow
Write-Host "1. 运行 update_imports.ps1 更新所有 import 路径" -ForegroundColor White
Write-Host "2. 手动拆分 internal/model 为 model/repository/service" -ForegroundColor White
Write-Host "3. 运行 go mod tidy" -ForegroundColor White
Write-Host "4. 运行测试验证" -ForegroundColor White
Write-Host ""

