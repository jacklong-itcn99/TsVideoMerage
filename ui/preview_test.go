package ui

import (
	"image/png"
	"os"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"
)

// TestRenderPreview 用软件测试驱动(无需 OpenGL/显示器)把界面渲染成 PNG,
// 深色、浅色各一张,便于在无头环境下预览外观。
// 设置环境变量 TSMERGE_PREVIEW=<路径.png> 才会输出,生成 *_dark.png / *_light.png。
//
//	TSMERGE_PREVIEW=/tmp/preview.png go test ./ui/ -run TestRenderPreview
func TestRenderPreview(t *testing.T) {
	out := os.Getenv("TSMERGE_PREVIEW")
	if out == "" {
		t.Skip("未设置 TSMERGE_PREVIEW,跳过预览渲染")
	}
	test.NewApp()

	demo := []string{
		"/videos/ep01/seg000.ts", "/videos/ep01/seg001.ts", "/videos/ep01/seg002.ts",
		"/videos/ep01/seg003.ts", "/videos/ep01/seg004.ts", "/videos/ep01/seg005.ts",
	}
	cases := []struct {
		mode   ThemeMode
		suffix string
	}{
		{ModeDark, "dark"},
		{ModeLight, "light"},
	}
	for _, c := range cases {
		th := NewAppTheme()
		th.mode = c.mode

		w := test.NewWindow(BuildRoot(nil, demo, th))
		w.Resize(fyne.NewSize(660, 700))

		img := w.Canvas().Capture()
		path := strings.TrimSuffix(out, ".png") + "_" + c.suffix + ".png"
		f, err := os.Create(path)
		if err != nil {
			t.Fatal(err)
		}
		if err := png.Encode(f, img); err != nil {
			f.Close()
			t.Fatal(err)
		}
		f.Close()
		t.Logf("预览已写入 %s", path)
	}
}
