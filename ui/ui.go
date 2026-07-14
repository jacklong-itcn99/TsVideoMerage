// Package ui 构建 tsmerge 的界面。它刻意不导入 app / GL 驱动,
// 因此可以用 Fyne 的软件测试驱动无头渲染成图片来预览外观(见 preview_test.go)。
package ui

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"tsmerge/merger"
)

// BuildRoot 组装并返回主界面根节点。initialFiles 仅用于预览/测试时预填列表,
// 正常启动传 nil。th 为可切换深浅的主题,构建时会应用它。
func BuildRoot(w fyne.Window, initialFiles []string, th *appTheme) fyne.CanvasObject {
	if a := fyne.CurrentApp(); a != nil {
		a.Settings().SetTheme(th)
	}

	// 需要随主题一起换色的自定义画布元素(Fyne 不会自动刷新它们)。
	var themedTexts []*canvas.Text
	newMuted := func(s string, size float32) *canvas.Text {
		t := canvas.NewText(s, th.currentPalette().muted)
		t.TextSize = size
		themedTexts = append(themedTexts, t)
		return t
	}
	baseRect := canvas.NewRectangle(th.currentPalette().base)

	applyTheme := func() {
		p := th.currentPalette()
		baseRect.FillColor = p.base
		baseRect.Refresh()
		for _, t := range themedTexts {
			t.Color = p.muted
			t.Refresh()
		}
	}

	files := append([]string(nil), initialFiles...)
	selected := -1
	ffmpegReady := merger.HasFFmpeg()

	status := newMuted("请添加 .ts 分段,顺序很重要", 12)

	// 分段列表:每行 = 序号 + 视频图标 + 文件名。
	list := widget.NewList(
		func() int { return len(files) },
		func() fyne.CanvasObject {
			return container.NewHBox(widget.NewIcon(theme.MediaVideoIcon()), widget.NewLabel(""))
		},
		func(id widget.ListItemID, o fyne.CanvasObject) {
			row := o.(*fyne.Container)
			row.Objects[1].(*widget.Label).SetText(fmt.Sprintf("%02d   %s", id+1, filepath.Base(files[id])))
		},
	)
	list.OnSelected = func(id widget.ListItemID) { selected = id }
	list.OnUnselected = func(id widget.ListItemID) {
		if selected == id {
			selected = -1
		}
	}

	// 空状态提示,叠在列表之上;有文件时隐藏。
	emptyHint := container.NewCenter(container.NewVBox(
		container.NewCenter(widget.NewIcon(theme.FolderOpenIcon())),
		newMuted("还没有分段", 13),
		newMuted("点上方「添加文件夹」批量导入,或「添加文件」逐个添加", 12),
	))
	listStack := container.NewStack(list, emptyHint)

	refresh := func() {
		list.Refresh()
		if len(files) == 0 {
			emptyHint.Show()
			setText(status, "请添加 .ts 分段,顺序很重要")
			return
		}
		emptyHint.Hide()
		if total, err := merger.TotalSize(files); err == nil {
			setText(status, fmt.Sprintf("共 %d 个分段 · 约 %s", len(files), humanBytes(total)))
		} else {
			setText(status, fmt.Sprintf("共 %d 个分段", len(files)))
		}
	}

	// 顶部动作:添加文件 / 文件夹 / 清空。
	addFile := widget.NewButtonWithIcon("添加文件", theme.ContentAddIcon(), func() {
		d := dialog.NewFileOpen(func(rc fyne.URIReadCloser, err error) {
			if err != nil || rc == nil {
				return
			}
			defer rc.Close()
			files = append(files, rc.URI().Path())
			refresh()
		}, w)
		d.SetFilter(storage.NewExtensionFileFilter([]string{".ts"}))
		d.Show()
	})
	addFolder := widget.NewButtonWithIcon("添加文件夹", theme.FolderOpenIcon(), func() {
		dialog.ShowFolderOpen(func(lu fyne.ListableURI, err error) {
			if err != nil || lu == nil {
				return
			}
			var found []string
			_ = filepath.Walk(lu.Path(), func(p string, info os.FileInfo, err error) error {
				if err != nil || info.IsDir() {
					return nil
				}
				if strings.EqualFold(filepath.Ext(p), ".ts") {
					found = append(found, p)
				}
				return nil
			})
			sort.Slice(found, func(i, j int) bool { return merger.NaturalLess(found[i], found[j]) })
			files = append(files, found...)
			refresh()
		}, w)
	})
	clearBtn := widget.NewButtonWithIcon("清空", theme.ContentClearIcon(), func() {
		files = nil
		selected = -1
		list.UnselectAll()
		refresh()
	})

	// 列表右侧:上移 / 下移 / 删除(纯图标)。
	moveUp := widget.NewButtonWithIcon("", theme.MoveUpIcon(), func() {
		if selected > 0 {
			files[selected-1], files[selected] = files[selected], files[selected-1]
			selected--
			list.Select(selected)
			refresh()
		}
	})
	moveDown := widget.NewButtonWithIcon("", theme.MoveDownIcon(), func() {
		if selected >= 0 && selected < len(files)-1 {
			files[selected+1], files[selected] = files[selected], files[selected+1]
			selected++
			list.Select(selected)
			refresh()
		}
	})
	removeBtn := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
		if selected >= 0 && selected < len(files) {
			files = append(files[:selected], files[selected+1:]...)
			selected = -1
			list.UnselectAll()
			refresh()
		}
	})
	sideCtrls := container.NewVBox(moveUp, moveDown, removeBtn, layout.NewSpacer())

	// 输出选项。
	toMP4 := widget.NewCheck("合并后转成 MP4(兼容性更好)", nil)
	mp4Hint := newMuted("转 mp4 已就绪(程序已内置 ffmpeg)", 12)
	if !ffmpegReady {
		toMP4.Disable()
		mp4Hint = newMuted("未找到 ffmpeg,无法转 mp4(可把 ffmpeg 放到本程序旁)", 12)
	}

	progress := widget.NewProgressBar()
	progress.Hide()

	var mergeBtn *widget.Button
	mergeBtn = widget.NewButtonWithIcon("合并并保存", theme.DownloadIcon(), func() {
		if len(files) == 0 {
			dialog.ShowInformation("提示", "请先添加要合并的分段", w)
			return
		}
		defaultName := "merged.ts"
		if toMP4.Checked {
			defaultName = "merged.mp4"
		}
		save := dialog.NewFileSave(func(wc fyne.URIWriteCloser, err error) {
			if err != nil || wc == nil {
				return
			}
			outPath := wc.URI().Path()
			wc.Close() // 用自己的文件句柄写,关掉对话框给的空文件
			go runMerge(w, mergeBtn, progress, status, files, outPath, toMP4.Checked)
		}, w)
		save.SetFileName(defaultName)
		save.Show()
	})
	mergeBtn.Importance = widget.HighImportance

	// 主题切换:跟随系统 / 浅色 / 深色。
	modeNames := []string{"跟随系统", "浅色", "深色"}
	themeSelect := widget.NewSelect(modeNames, func(s string) {
		switch s {
		case "浅色":
			th.mode = ModeLight
		case "深色":
			th.mode = ModeDark
		default:
			th.mode = ModeAuto
		}
		if a := fyne.CurrentApp(); a != nil {
			a.Settings().SetTheme(th) // 触发所有控件刷新
		}
		applyTheme() // 自定义画布元素单独换色
	})
	themeSelect.Selected = modeNames[int(th.mode)] // 直接设值,避免触发回调

	// 自动模式下,系统深浅变化时实时跟随。
	if a := fyne.CurrentApp(); a != nil {
		ch := make(chan fyne.Settings, 1)
		a.Settings().AddChangeListener(ch)
		go func() {
			for range ch {
				fyne.Do(applyTheme)
			}
		}()
	}

	// 组装。
	titleBox := container.NewHBox(widget.NewIcon(theme.MediaVideoIcon()), headingText("TS 视频合并工具"))
	headerTop := container.NewBorder(nil, nil, titleBox, themeSelect)
	header := container.NewVBox(
		headerTop,
		newMuted("把多个 .ts 分段按顺序合并成一个可播放的视频", 12),
	)
	topActions := container.NewHBox(addFile, addFolder, layout.NewSpacer(), clearBtn)

	listInner := container.NewBorder(nil, nil, nil, sideCtrls, listStack)
	listCard := widget.NewCard("视频分段", "顺序即合并顺序,选中后用右侧按钮上移/下移/删除", listInner)

	outputCard := widget.NewCard("输出选项", "", container.NewVBox(toMP4, mp4Hint))
	statusRow := container.NewHBox(widget.NewIcon(theme.InfoIcon()), status)
	bottom := container.NewVBox(outputCard, progress, mergeBtn, statusRow)

	body := container.NewBorder(
		container.NewPadded(container.NewVBox(header, topActions)),
		container.NewPadded(bottom),
		nil, nil,
		container.NewPadded(listCard),
	)

	// 底层深色 + 上层内容,让卡片"浮起"形成层次感。
	root := container.NewStack(baseRect, body)
	applyTheme()
	refresh()
	return root
}

// runMerge 在后台执行拼接(以及可选的 mp4 转封装),并把进度/结果安全地刷回 UI。
func runMerge(w fyne.Window, btn *widget.Button, progress *widget.ProgressBar, status *canvas.Text,
	files []string, outPath string, wantMP4 bool) {

	fyne.Do(func() {
		btn.Disable()
		progress.SetValue(0)
		progress.Show()
		setText(status, "正在合并…")
	})

	lastPct := -1
	onProgress := func(done, total int64) {
		if total <= 0 {
			return
		}
		pct := int(done * 100 / total)
		if pct == lastPct {
			return // 节流:百分比没变就不刷 UI
		}
		lastPct = pct
		frac := float64(done) / float64(total)
		fyne.Do(func() {
			progress.SetValue(frac)
			setText(status, fmt.Sprintf("正在合并… %d%%", pct))
		})
	}

	// mp4:先二进制拼接到临时 .ts,再用 ffmpeg 无损转封装。
	concatTarget := outPath
	var tmpTS string
	if wantMP4 {
		tf, err := os.CreateTemp("", "tsmerge-*.ts")
		if err != nil {
			finish(w, btn, progress, status, "", err)
			return
		}
		tmpTS = tf.Name()
		tf.Close()
		concatTarget = tmpTS
		defer os.Remove(tmpTS)
	}

	n, err := merger.ConcatToFile(files, concatTarget, onProgress)
	if err != nil {
		finish(w, btn, progress, status, "", err)
		return
	}

	if wantMP4 {
		fyne.Do(func() { setText(status, "拼接完成,正在用 ffmpeg 转封装 MP4…") })
		ffmpeg := merger.ResolveFFmpeg() // 首次会把内置 ffmpeg 解压到缓存目录
		if ffmpeg == "" {
			finish(w, btn, progress, status, "", fmt.Errorf("找不到 ffmpeg,无法转 mp4"))
			return
		}
		if err := merger.RemuxToMP4(ffmpeg, tmpTS, outPath); err != nil {
			finish(w, btn, progress, status, "", err)
			return
		}
	}

	finish(w, btn, progress, status, fmt.Sprintf("%s(%s)", outPath, humanBytes(n)), nil)
}

// finish 统一在主线程收尾:恢复按钮、隐藏进度条、弹出结果或错误。
func finish(w fyne.Window, btn *widget.Button, progress *widget.ProgressBar, status *canvas.Text,
	okMsg string, err error) {
	fyne.Do(func() {
		btn.Enable()
		progress.Hide()
		if err != nil {
			setText(status, "合并失败")
			dialog.ShowError(err, w)
			return
		}
		setText(status, "完成")
		dialog.ShowInformation("合并完成", "已保存:\n"+okMsg, w)
	})
}

// headingText 返回加粗的大号标题文字。
func headingText(s string) *widget.Label {
	l := widget.NewLabel(s)
	l.TextStyle = fyne.TextStyle{Bold: true}
	l.SizeName = theme.SizeNameHeadingText
	return l
}

// setText 更新 canvas.Text 并刷新(须在主线程或 fyne.Do 内调用)。
func setText(t *canvas.Text, s string) {
	t.Text = s
	t.Refresh()
}

func humanBytes(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for x := n / unit; x >= unit; x /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
