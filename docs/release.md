# 发布流程 (Release Guide)

本项目通过 GitHub Actions 自动构建并发布 Release，工作流定义在
[`.github/workflows/release.yml`](../.github/workflows/release.yml)。

## 触发方式

推送符合 `v*` 格式的 Git tag 即可自动触发构建与发布：

```bash
# 1. 确认代码已提交并推送到 main
git push origin main

# 2. 打一个版本 tag（遵循语义化版本）
git tag v1.0.0

# 3. 推送 tag，触发发布流水线
git push origin v1.0.0
```

推送后，前往仓库的 **Actions** 页面即可查看构建进度；完成后 Release
会出现在仓库的 **Releases** 页面。

## 版本号规范

- 正式版：`v1.0.0`、`v1.2.3`
- 预发布版：tag 名包含 `-` 时（如 `v1.0.0-beta.1`、`v1.0.0-rc.1`）会被自动
  标记为 **pre-release**，不会显示为 "Latest"。

## 流水线做了什么

工作流分为 `build` 和 `release` 两个 job：

**`build`（matrix，并行构建 amd64 与 arm64，均在 `windows-latest` 上）**

1. **Checkout** —— 检出仓库代码。
2. **Set up Go / Node.js** —— 安装 Go 1.25、Node 20，并启用依赖缓存。
3. **Install Wails CLI** —— 安装 `wails` 命令行工具。
4. **Build** —— 执行 `wails build -platform windows/<arch> -clean`，产物为
   `build/bin/VinylPlayer.exe`。
5. **Package artifacts** —— 将 exe 复制为带架构后缀的单文件、压缩为 zip，并为
   两者分别生成 `.sha256` 校验文件，放入 `dist/`。
6. **Upload build artifacts** —— 以 `dist-<arch>` 为名上传该架构的产物。

> `fail-fast: false`：即使 arm64 构建失败，也不影响 amd64 产物照常发布。

**`release`（`needs: build`，在 `windows-latest` 上）**

7. **Download artifacts** —— 合并下载两个架构的全部产物。
8. **Compose release body** —— 用 `gh api ... generate-notes` 自动生成
   "What's Changed" 更新日志，并拼接中英双语的下载指引。
9. **Create GitHub Release** —— 创建 Release 并上传所有产物。

## 产物说明

发布后的 Release 会为 **amd64（x64）** 和 **arm64** 两个架构各提供一组文件：

| 文件 | 说明 | 适合谁 |
| --- | --- | --- |
| `VinylPlayer-<ver>-windows-amd64.exe` | x64 绿色单文件，双击即运行，无需安装 | 大多数用户（Intel/AMD） |
| `VinylPlayer-<ver>-windows-amd64.zip` | x64 exe 的压缩包，体积更小，解压后运行 | x64 且网络较慢的用户 |
| `VinylPlayer-<ver>-windows-arm64.exe` | ARM64 绿色单文件 | ARM64 设备用户（如骁龙本） |
| `VinylPlayer-<ver>-windows-arm64.zip` | ARM64 exe 的压缩包 | ARM64 且网络较慢的用户 |
| `*.sha256` | SHA-256 校验文件，用于核对下载完整性 | 想验证文件完整性的用户 |

> **架构选择：** 除非你使用的是 ARM64 设备，否则请选择 **amd64（x64）**。
> 所有构建都需要 WebView2 运行时（大多数 Win10/11 已内置）。

### 校验下载完整性

Windows PowerShell 中：

```powershell
# 计算下载文件的 SHA-256，与 .sha256 文件内容比对
Get-FileHash .\VinylPlayer-<ver>-windows-amd64.exe -Algorithm SHA256
```

## 权限说明

工作流使用内置的 `GITHUB_TOKEN`（已声明 `permissions: contents: write`）
创建 Release，无需额外配置任何密钥。

## 常见问题

- **Actions 未触发**：确认推送的是 tag 而非普通分支，且 tag 名匹配 `v*`。
- **Go 版本拉取失败**：`go-version` 目前固定为 `1.25`，若该版本在 CI 上不可用，
  可改为 `1.25.x` 或 `stable`。
- **构建失败**：先在本地执行 `wails build` 验证能否成功编译，再重新打 tag。
- **需要重新发布同一版本**：先删除旧 tag 和对应 Release，再重新推送 tag。

  ```bash
  git tag -d v1.0.0
  git push origin :refs/tags/v1.0.0
  ```
