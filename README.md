# tsmerge — TS 视频合并桌面工具

用 Go + [Fyne](https://fyne.io/) 写的小工具:把一堆 `.ts`(MPEG-TS 视频分段,
常见于 M3U8 / HLS 下载)**二进制拼接**成一个可播放的视频,带进度条;
还能顺手转封装成 `.mp4`(发行包已内置 ffmpeg,开箱即用)。

> ⚠ `.ts` 在这里指 **MPEG 传输流视频**,不是 TypeScript 源码。视频必须逐字节
> 拼接,不能当文本处理,否则会损坏数据、导致无法播放。

## 功能

- 添加单个文件或整个文件夹(自动扫描 `.ts`,并按 **自然顺序** 排序:seg2 在 seg10 前)
- 调整顺序:上移 / 下移 / 删除 / 清空(**拼接顺序 = 列表顺序,顺序错了视频会乱**)
- 合并时显示 **进度条**(按已处理字节数实时更新)
- 纯 Go 二进制拼接 → 输出 `.ts`,**零外部依赖**
- 勾选「转成 MP4」→ 用 ffmpeg 无损转封装(`-c copy`,不重新编码,秒级),兼容性最好
  - **发行包(Release)已内置 ffmpeg,开箱即用,无需另装**:Windows 直接编进单个 exe(首次用到时
    自动解压到缓存),macOS 装在 `.app` 包内,Linux 随 tar 附带;都找不到才回落到系统 PATH,再没有则禁用该选项
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
├── icon.png                   # 应用图标源(窗口/任务栏、mac .app、Linux .desktop 都用它)
├── icon.ico                   # 由 icon.png 生成的多分辨率图标(Windows 用)
├── rsrc_windows_amd64.syso    # Windows exe 文件图标资源,go build 自动链入(由 icon.ico 生成)
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

### 一键三端发布(推荐)—— 打个标签就行

仓库带了统一的发布流水线 [.github/workflows/release.yml](.github/workflows/release.yml)。
**推一个 `v*` 标签**,GitHub Actions 会并行构建三端并自动创建 Release 挂上产物:

```bash
git tag -a v0.2.0 -m "发布说明" && git push origin v0.2.0
```

| 平台 | 产物 | 怎么构建的 |
|---|---|---|
| Windows 64 位 | `tsmerge-windows-amd64.exe`(单文件,ffmpeg 已编入) | Linux runner + mingw 交叉编译,`-tags embedffmpeg` |
| macOS(通用) | `TSMerge-macos-universal.dmg` | macOS runner,arm64+amd64 合并,打 .dmg |
| Linux 64 位 | `tsmerge-linux-amd64.tar.xz` | Linux runner,fyne 打包 |

构建完在仓库 **Releases** 页面直接下载。也可在 Actions 页面手动运行(只出产物、不发 Release)。

> ⚠ Fyne 的 macOS/Linux 版必须在对应系统上构建(要 CGO + 系统框架),**无法从别的系统交叉编译**;
> 只有 Windows 版能用 mingw 交叉编。这也是用云端 CI 的原因。
> macOS 未签名,首次打开右键→「打开」或 `xattr -cr TSMerge.app`(会一并解除内置 ffmpeg 的隔离标记)。

### 本地手动构建(单平台)

```bash
./build-windows.sh    # 在 Linux/WSL 上交叉编译出 tsmerge.exe(需 gcc-mingw-w64-x86-64)
./build-macos.sh      # 在 macOS 上构建通用二进制 + TSMerge.app(需 fyne CLI)
```

### 关于内置 ffmpeg

发行包三端都**内置了 ffmpeg**(Windows 编进 exe、macOS 装在 `.app` 内、Linux 随 tar),转 mp4 开箱即用。

内置的不是现成的全功能包,而是 CI 里**从官方源码现编的"极简 ffmpeg"**:只保留 TS→MP4 无损转封装
(`-c copy`)所需的 mpegts 解复用 + mp4 复用 + 少量解析器/比特流过滤器,砍掉全部编解码器。
因此它只有 **~2MB(压缩后 <1MB)**,且是 **LGPL**(禁用了 x264/x265 等 GPL 组件)。
构建配方见 [scripts/build-min-ffmpeg.sh](scripts/build-min-ffmpeg.sh),许可见 `THIRD-PARTY-ffmpeg.txt`。

从**源码本地构建**(`build-*.sh`)默认不带 ffmpeg——此时程序回落到系统 PATH 的 ffmpeg,
自己 `brew/apt install ffmpeg` 或把 `ffmpeg` 放到可执行文件旁即可。

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
