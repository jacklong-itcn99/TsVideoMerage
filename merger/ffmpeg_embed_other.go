//go:build embedffmpeg && !windows

package merger

// embedffmpeg 标签下的非 Windows 平台:当前不内置(macOS 装进 .app 包、
// Linux 随 tar 附带,均在程序同目录,由 FindFFmpeg 找到)。
var embeddedFFmpeg []byte
var embeddedFFmpegName string
