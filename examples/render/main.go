//go:build gl

// Video player using the mpv OpenGL render API in an IUP GLCanvas.
//
//	go run -tags gl . [file|url]
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"unsafe"

	"github.com/gen2brain/go-mpv"
	"github.com/gen2brain/iup-go/iup"
)

const maxRecent = 10

type player struct {
	canvas    iup.Ihandle
	slider    iup.Ihandle
	button    iup.Ihandle
	timeLabel iup.Ihandle
	controls  iup.Ihandle
	dlg       iup.Ihandle

	m  *mpv.Mpv
	rc *mpv.RenderContext

	w, h       int
	ready      bool
	loaded     bool
	paused     bool
	fullscreen bool
	toggleFS   bool
	savedSize  string
	duration   float64
	pending    string

	updatingSlider bool
}

var app = &player{}

func main() {
	iup.Open()
	defer iup.Close()
	iup.GLCanvasOpen()

	config := iup.Config()
	config.SetAttribute("APP_NAME", "player-iup")
	config.SetHandle("config")
	iup.ConfigLoad(config)

	app.canvas = iup.GLCanvas()
	app.canvas.SetAttributes(map[string]string{"BUFFER": "DOUBLE", "BORDER": "NO", "EXPAND": "YES", "RASTERSIZE": "640x360"})
	app.canvas.SetCallback("MAP_CB", iup.MapFunc(app.onMap))
	app.canvas.SetCallback("ACTION", iup.ActionFunc(app.onRedraw))
	app.canvas.SetCallback("BUTTON_CB", iup.ButtonFunc(app.onButton))
	app.canvas.SetCallback("RESIZE_CB", iup.ResizeFunc(func(_ iup.Ihandle, w, h int) int {
		app.w, app.h = w, h
		return iup.DEFAULT
	}))

	app.slider = iup.Val("HORIZONTAL")
	app.slider.SetAttributes(map[string]string{"EXPAND": "HORIZONTAL", "MIN": "0", "MAX": "1"})
	app.slider.SetCallback("VALUECHANGED_CB", iup.ValueChangedFunc(app.onSlider))

	app.button = iup.Button("Play")
	app.button.SetAttribute("PADDING", "DEFAULTBUTTONPADDING")
	app.button.SetCallback("ACTION", iup.ActionFunc(func(iup.Ihandle) int {
		app.togglePlay()
		return iup.DEFAULT
	}))

	app.timeLabel = iup.Label("0:00 / 0:00")
	app.timeLabel.SetAttributes(map[string]string{"SIZE": "70x", "ALIGNMENT": "ARIGHT"})

	recentMenu := iup.Menu()
	recentMenu.SetHandle("recentMenu")

	fileMenu := iup.Menu(
		iup.MenuItem("&Open...").SetCallback("ACTION", iup.ActionFunc(app.onOpen)),
		iup.Submenu("Open &Recent", recentMenu),
		iup.MenuItem("&Clear Recent").SetCallback("ACTION", iup.ActionFunc(app.clearRecent)),
		iup.MenuSeparator(),
		iup.MenuItem("E&xit").SetCallback("ACTION", iup.ActionFunc(func(iup.Ihandle) int { return iup.CLOSE })),
	)
	menu := iup.Menu(iup.Submenu("&File", fileMenu))
	menu.SetHandle("mainmenu")

	iup.ConfigRecentInit(config, recentMenu, app.onRecent, maxRecent)

	app.controls = iup.Hbox(app.button, app.slider, app.timeLabel).SetAttributes(map[string]string{"ALIGNMENT": "ACENTER", "GAP": "4", "MARGIN": "4x4"})
	app.dlg = iup.Dialog(iup.Vbox(app.canvas, app.controls).SetAttributes(map[string]string{"MARGIN": "0x0", "GAP": "0"}))
	app.dlg.SetAttributes(map[string]string{"TITLE": "player-iup", "MENU": "mainmenu"})
	app.dlg.SetCallback("K_ANY", iup.KAnyFunc(app.onKey))
	app.dlg.SetCallback("CLOSE_CB", iup.CloseFunc(func(iup.Ihandle) int {
		app.stop()
		return iup.DEFAULT
	}))

	iup.Timer().SetAttribute("TIME", 16).SetCallback("ACTION_CB", iup.TimerActionFunc(app.tick)).SetAttribute("RUN", "YES")

	if len(os.Args) > 1 {
		app.pending = os.Args[1]
	}

	iup.Show(app.dlg)
	iup.MainLoop()
}

func (p *player) onMap(ih iup.Ihandle) int {
	iup.GLMakeCurrent(ih)

	p.m = mpv.New()
	if err := p.m.SetOptionString("vo", "libmpv"); err != nil {
		fmt.Println("set vo:", err)
		return iup.DEFAULT
	}
	p.m.SetOptionString("idle", "yes")
	if err := p.m.Initialize(); err != nil {
		fmt.Println("mpv init:", err)
		return iup.DEFAULT
	}

	rc, err := p.m.NewRenderContextGL(func(name string) unsafe.Pointer {
		return iup.GLGetProcAddress(name)
	})
	if err != nil {
		fmt.Println("render context:", err)
		return iup.DEFAULT
	}
	p.rc = rc
	p.ready = true

	if p.pending != "" {
		p.openRecent(p.pending)
		p.pending = ""
	}

	return iup.DEFAULT
}

func (p *player) tick(iup.Ihandle) int {
	if !p.ready {
		return iup.DEFAULT
	}

	if p.toggleFS {
		p.toggleFS = false
		p.toggleFullscreen()
	}

	for {
		e := p.m.WaitEvent(0)
		if e.EventID == mpv.EventNone {
			break
		}
		switch e.EventID {
		case mpv.EventShutdown:
			return iup.CLOSE
		case mpv.EventEnd:
			if e.EndFile().Reason == mpv.EndFileEOF {
				p.paused = true
				p.loaded = false
				p.updateButton()
			}
		}
	}

	if p.rc.Update()&mpv.RenderUpdateFrame != 0 {
		p.render()
	}

	if p.loaded {
		p.updateProgress()
	}

	return iup.DEFAULT
}

func (p *player) onRedraw(iup.Ihandle) int {
	if p.ready {
		p.render()
	}

	return iup.DEFAULT
}

// onButton flags the toggle; it runs from tick, not this button dispatch.
func (p *player) onButton(_ iup.Ihandle, button, pressed, _, _ int, status string) int {
	if button == iup.BUTTON1 && pressed != 0 && iup.IsDouble(status) {
		p.toggleFS = true
	}

	return iup.DEFAULT
}

func (p *player) toggleFullscreen() {
	p.fullscreen = !p.fullscreen
	if p.fullscreen {
		p.savedSize = p.dlg.GetAttribute("RASTERSIZE")
		p.controls.SetAttributes(map[string]string{"FLOATING": "YES", "VISIBLE": "NO"})
		p.dlg.SetAttribute("MENU", "")
		p.dlg.SetAttribute("FULLSCREEN", "YES")
		return
	}

	p.dlg.SetAttribute("FULLSCREEN", "NO")
	p.dlg.SetAttribute("MENU", "mainmenu")
	p.controls.SetAttributes(map[string]string{"FLOATING": "NO", "VISIBLE": "YES"})
	p.dlg.SetAttribute("RASTERSIZE", p.savedSize)
	iup.Refresh(p.dlg)
	p.dlg.SetAttribute("RASTERSIZE", "")
}

func (p *player) clearRecent(iup.Ihandle) int {
	config := iup.GetHandle("config")
	for i := 1; i <= maxRecent; i++ {
		iup.ConfigSetVariableStrId(config, "Recent", "File", i, "")
	}
	iup.ConfigSave(config)
	iup.ConfigRecentInit(config, iup.GetHandle("recentMenu"), p.onRecent, maxRecent)

	return iup.DEFAULT
}

func (p *player) render() {
	if p.w <= 0 || p.h <= 0 {
		return
	}

	iup.GLMakeCurrent(p.canvas)
	_ = p.rc.RenderGL(0, p.w, p.h, true)
	iup.GLSwapBuffers(p.canvas)
	p.rc.ReportSwap()
}

func (p *player) open(path string) {
	if !p.ready {
		return
	}

	if err := p.m.Command([]string{"loadfile", path}); err != nil {
		fmt.Println("loadfile:", err)
		return
	}

	p.loaded = true
	p.paused = false
	p.duration = 0
	p.updateButton()
	p.dlg.SetAttribute("TITLE", "player-iup - "+filepath.Base(path))
}

func (p *player) togglePlay() {
	if !p.loaded {
		return
	}

	p.paused = !p.paused
	_ = p.m.SetProperty("pause", mpv.FormatFlag, p.paused)
	p.updateButton()
}

func (p *player) updateButton() {
	if p.paused {
		p.button.SetAttribute("TITLE", "Play")
	} else {
		p.button.SetAttribute("TITLE", "Pause")
	}
}

func (p *player) updateProgress() {
	if p.duration <= 0 {
		p.duration = p.propDouble("duration")
		if p.duration > 0 {
			p.slider.SetAttribute("MAX", fmt.Sprintf("%f", p.duration))
		}
	}

	pos := p.propDouble("time-pos")
	p.setSliderValue(pos)
	p.timeLabel.SetAttribute("TITLE", formatTime(pos)+" / "+formatTime(p.duration))
}

func (p *player) onSlider(iup.Ihandle) int {
	if !p.updatingSlider && p.loaded {
		_ = p.m.Command([]string{"seek", fmt.Sprintf("%f", p.slider.GetDouble("VALUE")), "absolute"})
	}

	return iup.DEFAULT
}

func (p *player) onKey(_ iup.Ihandle, c int) int {
	switch c {
	case iup.K_q, iup.K_ESC:
		return iup.CLOSE
	case iup.K_SP:
		p.togglePlay()
	case iup.K_RIGHT:
		_ = p.m.Command([]string{"seek", "5"})
	case iup.K_LEFT:
		_ = p.m.Command([]string{"seek", "-5"})
	}

	return iup.DEFAULT
}

func (p *player) onOpen(iup.Ihandle) int {
	dlg := iup.FileDlg()
	dlg.SetAttributes(map[string]string{
		"DIALOGTYPE": "OPEN",
		"TITLE":      "Open",
		"EXTFILTER":  "Media|*.mp4;*.mkv;*.webm;*.avi;*.mov;*.mp3;*.flac|All Files|*.*",
	})
	iup.Popup(dlg, iup.CENTERPARENT, iup.CENTERPARENT)
	if dlg.GetInt("STATUS") >= 0 {
		p.openRecent(dlg.GetAttribute("VALUE"))
	}
	dlg.Destroy()

	return iup.DEFAULT
}

func (p *player) onRecent(ih iup.Ihandle) int {
	if path := ih.GetAttribute("RECENTFILENAME"); path != "" {
		p.openRecent(path)
	}

	return iup.DEFAULT
}

func (p *player) openRecent(path string) {
	config := iup.GetHandle("config")
	iup.ConfigRecentUpdate(config, path)
	iup.ConfigSave(config)
	p.open(path)
}

func (p *player) propDouble(name string) float64 {
	v, err := p.m.GetProperty(name, mpv.FormatDouble)
	if err != nil {
		return 0
	}
	d, _ := v.(float64)

	return d
}

func (p *player) setSliderValue(t float64) {
	p.updatingSlider = true
	p.slider.SetAttribute("VALUE", fmt.Sprintf("%f", t))
	p.updatingSlider = false
}

func (p *player) stop() {
	if p.rc != nil {
		p.rc.Free()
		p.rc = nil
	}
	if p.m != nil {
		p.m.TerminateDestroy()
		p.m = nil
	}
	p.ready = false
}

func formatTime(sec float64) string {
	if sec < 0 {
		sec = 0
	}
	s := int(sec)
	if h := s / 3600; h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, (s%3600)/60, s%60)
	}

	return fmt.Sprintf("%d:%02d", s/60, s%60)
}
