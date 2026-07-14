//go:build !embedffmpeg

package merger

// 默认构建:不内置 ffmpeg(发行时才用 -tags embedffmpeg 打开)。
var embeddedFFmpeg []byte
var embeddedFFmpegName string
