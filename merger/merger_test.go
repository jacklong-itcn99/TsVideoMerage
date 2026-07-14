package merger

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// writeSeg 在临时目录里写一个分段文件,返回路径。
func writeSeg(t *testing.T, dir, name string, data []byte) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatalf("写入分段失败: %v", err)
	}
	return p
}

func TestConcatIsByteExact(t *testing.T) {
	dir := t.TempDir()
	// 用含 0x00 / 0x0A 等字节的二进制数据,验证不会被当文本处理而损坏。
	a := []byte{0x47, 0x00, 0x10, 0x0A, 0xFF, 0x47}
	b := []byte{0x00, 0x01, 0x02, 0x0A, 0x0A, 0x47, 0xAB}
	c := []byte{0x47, 0xFE, 0x0D, 0x0A}
	pa := writeSeg(t, dir, "seg0.ts", a)
	pb := writeSeg(t, dir, "seg1.ts", b)
	pc := writeSeg(t, dir, "seg2.ts", c)

	out := filepath.Join(dir, "merged.ts")
	n, err := ConcatToFile([]string{pa, pb, pc}, out, nil)
	if err != nil {
		t.Fatalf("拼接失败: %v", err)
	}

	want := append(append(append([]byte{}, a...), b...), c...)
	if int(n) != len(want) {
		t.Fatalf("返回字节数 %d,期望 %d", n, len(want))
	}
	got, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("输出与逐字节拼接不一致\n got=%v\nwant=%v", got, want)
	}
}

func TestProgressReportsMonotonicToTotal(t *testing.T) {
	dir := t.TempDir()
	pa := writeSeg(t, dir, "a.ts", bytes.Repeat([]byte{0x47}, 3<<20)) // 3MB
	pb := writeSeg(t, dir, "b.ts", bytes.Repeat([]byte{0x11}, 2<<20)) // 2MB

	var last int64
	var total int64
	calls := 0
	_, err := ConcatToFile([]string{pa, pb}, filepath.Join(dir, "o.ts"),
		func(done, tot int64) {
			calls++
			if done < last {
				t.Fatalf("进度回退: %d < %d", done, last)
			}
			last, total = done, tot
		})
	if err != nil {
		t.Fatal(err)
	}
	if calls == 0 {
		t.Fatal("进度回调从未被调用")
	}
	if total != 5<<20 || last != 5<<20 {
		t.Fatalf("最终进度 %d/%d,期望 5MB/5MB", last, total)
	}
}

func TestTotalSize(t *testing.T) {
	dir := t.TempDir()
	pa := writeSeg(t, dir, "a.ts", make([]byte, 100))
	pb := writeSeg(t, dir, "b.ts", make([]byte, 250))
	got, err := TotalSize([]string{pa, pb})
	if err != nil {
		t.Fatal(err)
	}
	if got != 350 {
		t.Fatalf("总大小 %d,期望 350", got)
	}
}

func TestConcatMissingFileErrors(t *testing.T) {
	dir := t.TempDir()
	_, err := ConcatToFile([]string{filepath.Join(dir, "nope.ts")}, filepath.Join(dir, "o.ts"), nil)
	if err == nil {
		t.Fatal("缺失文件应报错")
	}
}
