#!/usr/bin/env bash
# 在 Linux/WSL 上把 tsmerge 交叉编译成 Windows 64 位 .exe。
# 依赖:go、mingw-w64(sudo apt-get install -y gcc-mingw-w64-x86-64)
set -euo pipefail
cd "$(dirname "$0")"

CGO_ENABLED=1 \
GOOS=windows \
GOARCH=amd64 \
CC=x86_64-w64-mingw32-gcc \
go build -ldflags "-H windowsgui -s -w" -o tsmerge.exe .

echo "已生成: $(pwd)/tsmerge.exe"
file tsmerge.exe
