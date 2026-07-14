package merger

import (
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

// 端到端验证:用 ffmpeg 造出真实的 MPEG-TS 分段,经本工具拼接/转封装后,
// 再用 ffprobe 确认输出是合法可播放的视频。缺 ffmpeg 或 -short 时自动跳过。
func TestConcatProducesPlayableVideo(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过端到端测试(-short)")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("未安装 ffmpeg,跳过端到端测试")
	}
	dir := t.TempDir()

	// 生成 6 秒视频,强制每 2 秒一个关键帧,切成 3 个 mpegts 分段。
	gen := exec.Command(ffmpeg, "-y",
		"-f", "lavfi", "-i", "testsrc=duration=6:size=320x240:rate=25",
		"-f", "lavfi", "-i", "sine=frequency=440:duration=6",
		"-c:v", "libx264", "-g", "50", "-keyint_min", "50",
		"-force_key_frames", "expr:gte(t,n_forced*2)",
		"-c:a", "aac", "-f", "segment", "-segment_time", "2",
		"-segment_format", "mpegts", filepath.Join(dir, "seg%03d.ts"),
		"-loglevel", "error")
	if out, err := gen.CombinedOutput(); err != nil {
		t.Fatalf("生成测试分段失败: %v\n%s", err, out)
	}

	segs := []string{
		filepath.Join(dir, "seg000.ts"),
		filepath.Join(dir, "seg001.ts"),
		filepath.Join(dir, "seg002.ts"),
	}

	// 1) 二进制拼接
	mergedTS := filepath.Join(dir, "merged.ts")
	if _, err := ConcatToFile(segs, mergedTS, nil); err != nil {
		t.Fatalf("拼接失败: %v", err)
	}
	if d := probeDuration(t, mergedTS); d < 5.5 || d > 6.5 {
		t.Errorf("拼接后时长 %.2fs,期望约 6s(3×2s)", d)
	}

	// 2) 转封装成 mp4
	mp4 := filepath.Join(dir, "merged.mp4")
	if err := RemuxToMP4(ffmpeg, mergedTS, mp4); err != nil {
		t.Fatalf("转 mp4 失败: %v", err)
	}
	if d := probeDuration(t, mp4); d < 5.5 || d > 6.5 {
		t.Errorf("mp4 时长 %.2fs,期望约 6s", d)
	}
	if codecs := probeCodecs(t, mp4); !strings.Contains(codecs, "h264") || !strings.Contains(codecs, "aac") {
		t.Errorf("mp4 缺少预期音视频轨: %q", codecs)
	}
}

func probeDuration(t *testing.T, path string) float64 {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "format=duration", "-of", "csv=p=0", path).Output()
	if err != nil {
		t.Fatalf("ffprobe 读取时长失败(文件可能损坏): %v", err)
	}
	d, err := strconv.ParseFloat(strings.TrimSpace(string(out)), 64)
	if err != nil {
		t.Fatalf("解析时长失败: %v (%q)", err, out)
	}
	return d
}

func probeCodecs(t *testing.T, path string) string {
	t.Helper()
	out, err := exec.Command("ffprobe", "-v", "error",
		"-show_entries", "stream=codec_name", "-of", "csv=p=0", path).Output()
	if err != nil {
		t.Fatalf("ffprobe 读取编码失败: %v", err)
	}
	return string(out)
}
