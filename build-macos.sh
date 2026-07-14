#!/usr/bin/env bash
# 在 macOS 上构建 tsmerge。产出:
#   - tsmerge        通用二进制(Apple Silicon arm64 + Intel amd64)
#   - TSMerge.app    可双击的应用包(需要 fyne CLI 与 icon.png)
#
# 说明:Fyne 的 macOS 版必须开 CGO 并链接系统框架,因此只能在 macOS 上构建,
#      无法从 Linux 交叉编译(需要 Apple 的 macOS SDK,仅允许在苹果硬件上使用)。
set -euo pipefail
cd "$(dirname "$0")"

if [[ "$(uname)" != "Darwin" ]]; then
  echo "错误:此脚本需在 macOS 上运行。" >&2
  echo "若没有 Mac,可用仓库里的 GitHub Actions 工作流在云端 macOS 上构建。" >&2
  exit 1
fi

echo "==> 编译 arm64(Apple Silicon)"
CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w" -o tsmerge-arm64 .
echo "==> 编译 amd64(Intel)"
CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w" -o tsmerge-amd64 .

echo "==> lipo 合并为通用二进制"
lipo -create -output tsmerge tsmerge-arm64 tsmerge-amd64
rm -f tsmerge-arm64 tsmerge-amd64
file tsmerge

# 打包成 .app(需要 fyne CLI:go install fyne.io/tools/cmd/fyne@latest)
FYNE="$(command -v fyne || echo "$(go env GOPATH)/bin/fyne")"
if [[ -x "$FYNE" && -f icon.png ]]; then
  echo "==> 用 fyne 打包 TSMerge.app"
  "$FYNE" package -os darwin \
    --name TSMerge \
    --app-id com.example.tsmerge \
    --icon icon.png \
    --executable tsmerge
  echo "已生成:$(pwd)/TSMerge.app"
  echo "提示:未签名的 .app 首次打开会被 Gatekeeper 拦,右键→打开,或执行 xattr -cr TSMerge.app"
else
  echo "已生成通用二进制 tsmerge(命令行可直接运行)。"
  echo "如需 .app:先 go install fyne.io/tools/cmd/fyne@latest,并确保当前目录有 icon.png,再重跑本脚本。"
fi
