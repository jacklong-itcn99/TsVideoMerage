package merger

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
)

// extractEmbeddedFFmpeg 若程序内置了 ffmpeg(gzip 压缩),将其解压到用户缓存目录并返回路径;
// 未内置返回 ""。按内容 CRC 命名:换了 ffmpeg 版本自动写新文件,已解压过则直接复用。
func extractEmbeddedFFmpeg() string {
	if len(embeddedFFmpeg) == 0 {
		return ""
	}
	return unpackGz(embeddedFFmpeg, embeddedFFmpegName)
}

// unpackGz 把 gzip 数据解压到缓存目录的一个稳定路径并返回它;失败返回 ""。
func unpackGz(gzData []byte, name string) string {
	base, err := os.UserCacheDir()
	if err != nil {
		base = os.TempDir()
	}
	dir := filepath.Join(base, "tsmerge")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return ""
	}
	out := filepath.Join(dir, fmt.Sprintf("ffmpeg-%08x%s", crc32.ChecksumIEEE(gzData), filepath.Ext(name)))
	if fi, err := os.Stat(out); err == nil && fi.Size() > 0 {
		return out // 已解压过,直接复用
	}

	gz, err := gzip.NewReader(bytes.NewReader(gzData))
	if err != nil {
		return ""
	}
	defer gz.Close()

	tmp := out + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o755)
	if err != nil {
		return ""
	}
	if _, err := io.Copy(f, gz); err != nil {
		f.Close()
		os.Remove(tmp)
		return ""
	}
	if err := f.Close(); err != nil {
		os.Remove(tmp)
		return ""
	}
	if err := os.Rename(tmp, out); err != nil { // 原子落盘,避免并发/中断产生半截文件
		os.Remove(tmp)
		return ""
	}
	_ = os.Chmod(out, 0o755)
	return out
}
