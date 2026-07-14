# tsmerge — TS 视频合并桌面工具

用 Go + [Fyne](https://fyne.io/) 写的小工具:把一堆 `.ts`(MPEG-TS 视频分段,
常见于 M3U8 / HLS 下载)**二进制拼接**成一个可播放的视频,带进度条;
若系统装了 ffmpeg,还能顺手转封装成 `.mp4`。

> ⚠ `.ts` 在这里指 **MPEG 传输流视频**,不是 TypeScript 源码。视频必须逐字节
> 拼接,不能当文本处理,否则会损坏数据、导致无法播放。

## 功能

- 添加单个文件或整个文件夹(自动扫描 `.ts`,并按 **自然顺序** 排序:seg2 在 seg10 前)
- 调整顺序:上移 / 下移 / 删除 / 清空(**拼接顺序 = 列表顺序,顺序错了视频会乱**)
- 合并时显示 **进度条**(按已处理字节数实时更新)
- 纯 Go 二进制拼接 → 输出 `.ts`,**零外部依赖**
- 勾选「转成 MP4」→ 用 ffmpeg 无损转封装(`-c copy`,不重新编码,秒级),兼容性最好
  - 未检测到 ffmpeg 时该选项自动禁用
- 右上角**主题切换**:跟随系统 / 浅色 / 深色,「跟随系统」会随 OS 深浅设置实时切换

## 目录结构

```
tsmerge/
├── main.go                    # 仅负责创建 app / 窗口
├── ui/
│   ├── ui.go                  # 界面构建 + 后台合并/进度(不导入 GL,可无头渲染)
│   ├── theme.go               # 暗色 + 紫色强调的自定义主题
│   └── preview_test.go        # 软件驱动无头渲染成 PNG,预览外观
├── merger/
│   ├── merger.go              # 二进制拼接 + ffmpeg 检测/转封装(纯逻辑,无 GUI 依赖)
│   ├── sort.go                # 文件名自然排序
│   ├── merger_test.go         # 拼接字节无损、进度回调等单元测试
│   ├── sort_test.go
│   └── integration_test.go    # 端到端:真实 TS 分段 → 合并 → ffprobe 验证可播放
├── icon.png                   # 应用图标(打包 macOS .app 用,可自行替换)
├── build-windows.sh           # 交叉编译成 Windows .exe(在 Linux/WSL 上)
├── build-macos.sh             # 在 macOS 上构建通用二进制 + .app
├── .github/workflows/
│   └── build-macos.yml        # 云端 macOS 构建(无需自己有 Mac)
└── go.mod
```

### 预览界面外观(无需显示器)

`ui` 包不依赖 OpenGL,可用 Fyne 的软件驱动把界面渲染成图片:

```bash
TSMERGE_PREVIEW=/tmp/preview.png go test ./ui/ -run TestRenderPreview
```

## 构建

### 交叉编译成 Windows .exe(在 Linux/WSL 上)

```bash
sudo apt-get install -y gcc-mingw-w64-x86-64   # 一次性安装交叉编译器
./build-windows.sh                             # 生成 tsmerge.exe
```

`-H windowsgui` 让程序在 Windows 上以纯 GUI 启动(不弹黑色控制台)。

### 构建 macOS 版

⚠ Fyne 的 macOS 版**必须在 macOS 上构建**——它要开 CGO 链接 Cocoa/OpenGL 框架,
依赖 Apple 的 macOS SDK,而该 SDK 只允许在苹果硬件上使用,**无法从 Linux/Windows 交叉编译**。
`merger`、`ui` 两层是跨平台纯 Go,已验证可为 darwin 编译;只有 app/GL 驱动那层需要 Mac。

有两条路:

**A. 有 Mac —— 直接构建(推荐)**

```bash
./build-macos.sh
```

产出通用二进制 `tsmerge`(Apple Silicon + Intel),并在装了 fyne CLI 时打包成 `TSMerge.app`:

```bash
go install fyne.io/tools/cmd/fyne@latest   # 一次性安装打包器
./build-macos.sh
```

**B. 没有 Mac —— 用 GitHub Actions 云端构建**

把项目推到 GitHub 后,到 Actions 页面手动运行 `build-macos` 工作流(或推一个 `v*` 标签),
它会在云端 macOS 机器上编译并把 `TSMerge.app` 作为产物供下载。工作流见
[.github/workflows/build-macos.yml](.github/workflows/build-macos.yml)。

> 未签名的 `.app` 首次打开会被 Gatekeeper 拦下。右键点图标 →「打开」,
> 或在终端执行 `xattr -cr TSMerge.app` 解除隔离标记即可。
> macOS 上启用「转成 MP4」同理:`brew install ffmpeg`,或把 `ffmpeg` 放到 .app 旁。

### 在 Windows 上启用 MP4 转换

工具会在 **PATH** 和 **自身所在目录** 里找 `ffmpeg.exe`。任选其一:

- 从 https://www.gyan.dev/ffmpeg/builds/ 下载,把 `ffmpeg.exe` 放到 `tsmerge.exe` 旁边;
- 或把 ffmpeg 加进系统 PATH。

放好后重启工具,「转成 MP4」选项即可勾选。

### 运行测试

合并逻辑是纯 Go、不依赖 GUI,可单独测试:

```bash
go test ./merger/            # 全部测试(含 ffmpeg 端到端,需已装 ffmpeg)
go test ./merger/ -short     # 跳过端到端测试
```

## 使用

1. 「添加文件夹」选中放分段的目录(或逐个「添加文件」)
2. 核对顺序,必要时用「上移/下移」调整
3. 需要 mp4 就勾「转成 MP4」
4. 「合并并保存」选择输出位置 → 看进度条跑完即可

## 说明与限制

- **二进制拼接适用于同一来源、同一编码参数的分段**(HLS 下载的分段正是如此)。
  若各段分辨率/编码不一致,拼出来可能在某些播放器上花屏——这种情况需要
  用 ffmpeg **重新编码**(`-c:v libx264 ...`)而非本工具的流拷贝。
- 输出 `.ts` 用 VLC / PotPlayer 基本都能放;Windows 自带播放器对 `.ts` 支持不好,
  这也是提供「转 MP4」的原因。
- 顺序完全由列表决定;来自不同命名规则的文件请手动核对顺序。
