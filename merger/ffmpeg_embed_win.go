//go:build embedffmpeg && windows

package merger

import _ "embed"

// Windows 发行版把(gzip 压缩的)ffmpeg.exe 内置进二进制,做到单文件分发。
// 构建前需由 CI 把文件放到 merger/ffmpeg-bin/ffmpeg.exe.gz。

//go:embed ffmpeg-bin/ffmpeg.exe.gz
var embeddedFFmpeg []byte

var embeddedFFmpegName = "ffmpeg.exe"
