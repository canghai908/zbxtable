# Update all Go file import paths

$ErrorActionPreference = "Stop"
$projectRoot = "d:/canghai/zbxtable/zbxtable"

Write-Host "========================================"
Write-Host "Updating Import Paths"
Write-Host "========================================"
Write-Host ""

# Define path mappings
$importMappings = @{
    '"zbxtable/utils"' = '"zbxtable/pkg/utils"'
    '"zbxtable/middleware"' = '"zbxtable/internal/middleware"'
    '"zbxtable/controllers"' = '"zbxtable/internal/handler"'
    '"zbxtable/routers"' = '"zbxtable/api/v1"'
    '"zbxtable/models"' = '"zbxtable/internal/model"'
}

# Directories to update
$dirsToUpdate = @(
    "api",
    "internal",
    "pkg",
    "cmd"
)

$totalFiles = 0
$updatedFiles = 0

foreach ($dir in $dirsToUpdate) {
    $dirPath = Join-Path $projectRoot $dir
    if (Test-Path $dirPath) {
        Write-Host "Processing directory: $dir"
        
        Get-ChildItem -Path $dirPath -Filter "*.go" -Recurse | ForEach-Object {
            $totalFiles++
            $filePath = $_.FullName
            $content = Get-Content -Path $filePath -Raw -Encoding UTF8
            $originalContent = $content
            
            # Apply all mappings
            foreach ($oldPath in $importMappings.Keys) {
                $newPath = $importMappings[$oldPath]
                $content = $content -replace [regex]::Escape($oldPath), $newPath
            }
            
            # Write back if changed
            if ($content -ne $originalContent) {
                Set-Content -Path $filePath -Value $content -Encoding UTF8 -NoNewline
                $updatedFiles++
                Write-Host "  Updated: $($_.Name)"
            }
        }
    }
}

Write-Host ""
Write-Host "========================================"
Write-Host "Import Path Update Complete!"
Write-Host "========================================"
Write-Host ""
Write-Host "Statistics:"
Write-Host "  Total files: $totalFiles"
Write-Host "  Updated: $updatedFiles"
Write-Host "  Unchanged: $($totalFiles - $updatedFiles)"
Write-Host ""
