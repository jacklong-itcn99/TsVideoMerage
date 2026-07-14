package merger

import (
	"bytes"
	"compress/gzip"
	"os"
	"testing"
)

func gzipBytes(t *testing.T, data []byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	if _, err := gw.Write(data); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return buf.Bytes()
}

func TestUnpackGz(t *testing.T) {
	want := []byte("#!/bin/sh\necho fake-ffmpeg\n")
	gz := gzipBytes(t, want)

	path := unpackGz(gz, "ffmpeg")
	if path == "" {
		t.Fatal("unpackGz 返回空路径")
	}
	defer os.Remove(path)

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("解压内容不符\n got=%q\nwant=%q", got, want)
	}
	// 再次调用应复用同一文件(按 CRC 命名,路径一致)
	if p2 := unpackGz(gz, "ffmpeg"); p2 != path {
		t.Errorf("未复用缓存: %q != %q", p2, path)
	}
}

func TestExtractEmbeddedNoneByDefault(t *testing.T) {
	// 默认构建(无 embedffmpeg 标签)不应内置 ffmpeg
	if p := extractEmbeddedFFmpeg(); p != "" {
		t.Errorf("默认不应有内置 ffmpeg,却得到 %q", p)
	}
}
