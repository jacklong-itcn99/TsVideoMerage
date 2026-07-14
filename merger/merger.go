// Package merger 负责把多个 MPEG-TS(.ts)视频分段合并成一个可播放的视频。
//
// .ts 是传输流(Transport Stream)容器,设计上支持首尾直接拼接,所以核心做法是
// 严格的二进制拼接(绝不能当文本处理,否则会破坏视频数据)。若系统装了 ffmpeg,
// 还可把拼接结果重新封装(remux)成 .mp4 以获得最佳播放兼容性。
package merger

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// ProgressFunc 在拼接过程中被反复调用,报告已写入 / 总字节数。
type ProgressFunc func(done, total int64)

// TotalSize 返回所有分段的总字节数,用于进度条封顶值。
func TotalSize(paths []string) (int64, error) {
	var total int64
	for _, p := range paths {
		fi, err := os.Stat(p)
		if err != nil {
			return 0, fmt.Errorf("读取 %s 失败: %w", filepath.Base(p), err)
		}
		if fi.IsDir() {
			return 0, fmt.Errorf("%s 是文件夹", filepath.Base(p))
		}
		total += fi.Size()
	}
	return total, nil
}

// ConcatToFile 把 paths 按顺序二进制拼接写入 outPath,期间通过 onProgress 报告进度。
// 返回写入的总字节数。这是纯 Go 实现,无任何外部依赖。
func ConcatToFile(paths []string, outPath string, onProgress ProgressFunc) (int64, error) {
	total, err := TotalSize(paths)
	if err != nil {
		return 0, err
	}
	out, err := os.Create(outPath)
	if err != nil {
		return 0, fmt.Errorf("创建输出文件失败: %w", err)
	}
	defer out.Close()

	written, err := concat(paths, out, total, onProgress)
	if err != nil {
		return written, err
	}
	if err := out.Sync(); err != nil {
		return written, err
	}
	return written, nil
}

// concat 是实际的拷贝循环,按 1MB 块读写并回报进度。
func concat(paths []string, out io.Writer, total int64, onProgress ProgressFunc) (int64, error) {
	buf := make([]byte, 1<<20) // 1MB
	var done int64
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil {
			return done, fmt.Errorf("打开 %s 失败: %w", filepath.Base(p), err)
		}
		for {
			n, rerr := f.Read(buf)
			if n > 0 {
				if _, werr := out.Write(buf[:n]); werr != nil {
					f.Close()
					return done, fmt.Errorf("写入失败: %w", werr)
				}
				done += int64(n)
				if onProgress != nil {
					onProgress(done, total)
				}
			}
			if rerr == io.EOF {
				break
			}
			if rerr != nil {
				f.Close()
				return done, fmt.Errorf("读取 %s 失败: %w", filepath.Base(p), rerr)
			}
		}
		f.Close()
	}
	return done, nil
}

// FindFFmpeg 查找 ffmpeg 可执行文件;找不到返回空串。
// 优先用与程序捆绑在同一目录的 ffmpeg(发行包里带的那份,保证可用),
// 其次回落到系统 PATH(用户自己装的)。
func FindFFmpeg() string {
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		for _, name := range []string{"ffmpeg.exe", "ffmpeg"} {
			cand := filepath.Join(dir, name)
			if fi, err := os.Stat(cand); err == nil && !fi.IsDir() {
				return cand
			}
		}
	}
	if p, err := exec.LookPath("ffmpeg"); err == nil {
		return p
	}
	return ""
}

// HasFFmpeg 便宜地判断是否有可用的 ffmpeg(内置的或系统/同目录的),供 UI 决定
// 是否启用 mp4 选项。它不做解压,因此可在启动时反复调用而无副作用。
func HasFFmpeg() bool {
	return len(embeddedFFmpeg) > 0 || FindFFmpeg() != ""
}

// ResolveFFmpeg 返回一个可执行的 ffmpeg 路径,供转封装前调用:
// 优先解压内置的那份(单文件发行版),否则回落到同目录 / 系统 PATH。
// 内置存在时会在首次调用时把 ffmpeg 解压到缓存目录(之后复用)。
func ResolveFFmpeg() string {
	if p := extractEmbeddedFFmpeg(); p != "" {
		return p
	}
	return FindFFmpeg()
}

// RemuxToMP4 用 ffmpeg 把已拼接好的 .ts 无损重封装成 mp4(-c copy,不重新编码,秒级完成)。
// 现代 ffmpeg 会在需要时自动插入 AAC 比特流过滤器,因此一般无需额外参数。
func RemuxToMP4(ffmpeg, inTS, outMP4 string) error {
	cmd := exec.Command(ffmpeg,
		"-y",       // 覆盖已存在的输出
		"-i", inTS, // 输入
		"-c", "copy", // 流拷贝,不转码
		"-movflags", "+faststart", // 便于边下边播 / 拖动
		outMP4,
	)
	var stderr writerCapture
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("ffmpeg 转封装失败: %w\n%s", err, stderr.tail())
	}
	return nil
}

// writerCapture 只保留 stderr 末尾若干字节,便于出错时给出有用信息而不撑爆内存。
type writerCapture struct{ buf []byte }

func (w *writerCapture) Write(p []byte) (int, error) {
	const max = 4096
	w.buf = append(w.buf, p...)
	if len(w.buf) > max {
		w.buf = w.buf[len(w.buf)-max:]
	}
	return len(p), nil
}

func (w *writerCapture) tail() string { return string(w.buf) }
