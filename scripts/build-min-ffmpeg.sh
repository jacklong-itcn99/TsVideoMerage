#!/usr/bin/env bash
# 从官方源码编译一个"极简 ffmpeg":只保留 TS→MP4 无损转封装(-c copy)所需的
# 解复用/复用/解析器/比特流过滤器,砍掉所有编解码器。产物 ~2MB(LGPL)。
#
# 用法: build-min-ffmpeg.sh <target> <输出路径>
#   target: linux | windows | macos-arm64 | macos-x64
set -euo pipefail

TARGET="${1:?需要 target}"
OUT="${2:?需要输出路径}"
VER=7.0.2
SRC="ffmpeg-${VER}"

# 只保留 mpegts 解复用 + mp4 复用 + 常见音视频的解析器/比特流过滤器。
# 需要 aac 解码器来补全 AAC 采样率等参数(否则 ADTS→MP4 复制会报 "sample rate not set")。
COMMON=(
  --disable-everything --disable-doc --disable-htmlpages --disable-manpages
  --disable-network --disable-autodetect --disable-debug
  --disable-programs --enable-ffmpeg
  --enable-protocol=file
  --enable-demuxer=mpegts,mov --enable-muxer=mp4,mov
  --enable-parser=h264,hevc,aac,aac_latm,ac3,mpegaudio
  --enable-decoder=aac
  --enable-bsf=h264_mp4toannexb,hevc_mp4toannexb,aac_adts_to_asc,extract_extradata,dump_extradata
  --enable-small --disable-asm --enable-static --disable-shared
  --disable-iconv --disable-sdl2 --disable-zlib --disable-bzlib --disable-lzma
)

# 下载源码(已存在则跳过)
if [ ! -d "$SRC" ]; then
  curl -fsSL -o "${SRC}.tar.xz" "https://ffmpeg.org/releases/${SRC}.tar.xz"
  tar xf "${SRC}.tar.xz"
  rm -f "${SRC}.tar.xz" # 删掉源码 tar,免得后续 *.tar.xz 通配到它
fi

CALLER="$PWD"
cd "$SRC"
make distclean >/dev/null 2>&1 || true

jobs() { command -v nproc >/dev/null && nproc || sysctl -n hw.ncpu; }

case "$TARGET" in
  linux)
    ./configure "${COMMON[@]}"
    make -j"$(jobs)" ffmpeg
    strip ffmpeg
    BIN=ffmpeg ;;
  windows)
    ./configure --cross-prefix=x86_64-w64-mingw32- --arch=x86_64 \
      --target-os=mingw32 --enable-cross-compile "${COMMON[@]}"
    make -j"$(jobs)" ffmpeg.exe
    x86_64-w64-mingw32-strip ffmpeg.exe
    BIN=ffmpeg.exe ;;
  macos-arm64)
    ./configure "${COMMON[@]}"                       # arm64 运行器上原生编译
    make -j"$(jobs)" ffmpeg
    strip ffmpeg
    BIN=ffmpeg ;;
  macos-x64)
    ./configure --enable-cross-compile --arch=x86_64 \
      --cc='clang -arch x86_64' --extra-cflags='-arch x86_64' \
      --extra-ldflags='-arch x86_64' "${COMMON[@]}"  # 在 arm64 mac 上交叉出 x86_64
    make -j"$(jobs)" ffmpeg
    strip ffmpeg
    BIN=ffmpeg ;;
  *)
    echo "未知 target: $TARGET" >&2; exit 2 ;;
esac

cp "$BIN" "$CALLER/$OUT"
echo "已生成极简 ffmpeg: $OUT ($(du -h "$CALLER/$OUT" | cut -f1))"
