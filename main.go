// tsmerge 是一个用 Fyne 写的桌面小工具:选择多个 .ts(MPEG-TS)视频分段,
// 二进制拼接成一个可播放的视频;若系统装了 ffmpeg,还可一并转封装成 mp4。
// 界面构建见 ui 包,合并逻辑见 merger 包。
package main

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/theme"

	"tsmerge/ui"
)

func main() {
	a := app.New()
	w := a.NewWindow("TS 视频合并工具")
	w.SetIcon(theme.MediaVideoIcon())
	w.Resize(fyne.NewSize(660, 700))
	w.SetContent(ui.BuildRoot(w, nil, ui.NewAppTheme()))
	w.ShowAndRun()
}
