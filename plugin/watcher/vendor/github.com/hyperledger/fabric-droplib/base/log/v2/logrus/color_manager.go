/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 11:52 上午
# @File : color_manager.go
# @Description :
# @Attention :
*/
package logrusplugin

import (
	"github.com/hyperledger/fabric-droplib/base/log/common"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/mgutz/ansi"
	"sync"
	"sync/atomic"
	"time"
)

var (
	g *ColorManager
)

func init() {
	g = newColorManager()
}

type ColorSchemeProvider func() *ColorScheme

func SetupDisableColor() {
	g.Do(func() {
		g.colorCompiledScheme = &CompiledColorScheme{
			InfoLevelColor:     ansi.ColorFunc(""),
			WarnLevelColor:     ansi.ColorFunc(""),
			ErrorLevelColor:    ansi.ColorFunc(""),
			FatalLevelColor:    ansi.ColorFunc(""),
			PanicLevelColor:    ansi.ColorFunc(""),
			DebugLevelColor:    ansi.ColorFunc(""),
			TimestampColor:     ansi.ColorFunc(""),
			LineNumberColor:    ansi.ColorFunc(""),
			DefaultColor:       ansi.ColorFunc(""),
			ModuleColor:        make(map[string]func(string) string, 1),
			InterestFieldColor: make(map[string]func(string) string, 1),
		}
	})
}
func SetupDefaultWithInterestModule(modules map[modules.Module]Color) {
	g.Do(func() {
		for k, v := range modules {
			g.colorCompiledScheme.ModuleColor[k.String()] = getCompiledColor(g.colorStringM[v], "")
		}
	})
}
func SetupDefaultWithInterestModuleOrFields(modules map[modules.Module]Color, fields map[string]Color) {
	g.Do(func() {
		for k, v := range modules {
			g.colorCompiledScheme.ModuleColor[k.String()] = getCompiledColor(g.colorStringM[v], "")
		}

		for k, v := range fields {
			g.colorCompiledScheme.InterestFieldColor[k] = getCompiledColor(g.colorStringM[v], "")
		}
	})
}

func SetupDefaultWithInterest(fields map[string]Color) {
	g.Do(func() {
		for k, v := range fields {
			g.colorCompiledScheme.InterestFieldColor[k] = getCompiledColor(g.colorStringM[v], "")
		}
	})
}

func SetupColorScheme(provider ColorSchemeProvider) {
	g.Do(func() {
		g.compileColorScheme(provider())
	})
}

type ColorScheme struct {
	InfoLevelStyle  Color
	WarnLevelStyle  Color
	ErrorLevelStyle Color
	FatalLevelStyle Color
	PanicLevelStyle Color
	DebugLevelStyle Color
	LineNumberStyle Color
	DefaultStyle    Color
	TimestampStyle  Color

	// 当第一步初始化的时候调用
	ModuleStyle        map[modules.Module]Color
	InterestFieldStyle map[string]Color
}

type CompiledColorScheme struct {
	InfoLevelColor  func(string) string
	WarnLevelColor  func(string) string
	ErrorLevelColor func(string) string
	FatalLevelColor func(string) string
	PanicLevelColor func(string) string
	DebugLevelColor func(string) string
	TimestampColor  func(string) string
	LineNumberColor func(string) string
	DefaultColor    func(string) string

	ModuleColor map[string]func(string) string

	InterestFieldColor map[string]func(string) string
}

type ColorManager struct {
	sync.RWMutex
	status uint32

	colorStringM map[Color]string

	baseTimestamp time.Time

	colorCompiledScheme *CompiledColorScheme

	sync.Once
}

func newColorManager() *ColorManager {
	r := &ColorManager{}
	r.colorStringM = make(map[Color]string)
	r.colorStringM[TextBlack] = "black+h"
	r.colorStringM[TextRed] = "red"
	r.colorStringM[TextGreen] = "green"
	r.colorStringM[TextYellow] = "yellow"
	r.colorStringM[TextBlue] = "blue"
	// TextMagenta] = ""
	r.colorStringM[TextCyan] = "cyan"
	r.colorStringM[TextWhite] = ""

	r.baseTimestamp = time.Now()
	colorScheme := &ColorScheme{
		InfoLevelStyle:     TextGreen,
		WarnLevelStyle:     TextYellow,
		ErrorLevelStyle:    TextRed,
		FatalLevelStyle:    TextRed,
		PanicLevelStyle:    TextRed,
		DebugLevelStyle:    TextBlue,
		LineNumberStyle:    TextDefault,
		DefaultStyle:       TextCyan,
		TimestampStyle:     TextDefault,
		ModuleStyle:        make(map[modules.Module]Color, 1),
		InterestFieldStyle: make(map[string]Color, 1),
	}
	r.colorCompiledScheme = r.compileColorScheme(colorScheme)

	return r
}

// FIXME should not a recviver
func (this *ColorManager) compileColorScheme(s *ColorScheme) *CompiledColorScheme {
	r := &CompiledColorScheme{
		InfoLevelColor:  getCompiledColor(this.colorStringM[s.InfoLevelStyle], ""),
		WarnLevelColor:  getCompiledColor(this.colorStringM[s.WarnLevelStyle], ""),
		ErrorLevelColor: getCompiledColor(this.colorStringM[s.ErrorLevelStyle], ""),
		FatalLevelColor: getCompiledColor(this.colorStringM[s.FatalLevelStyle], ""),
		PanicLevelColor: getCompiledColor(this.colorStringM[s.PanicLevelStyle], ""),
		DebugLevelColor: getCompiledColor(this.colorStringM[s.DebugLevelStyle], ""),
		TimestampColor:  getCompiledColor(this.colorStringM[s.TimestampStyle], ""),
		LineNumberColor: getCompiledColor(this.colorStringM[s.LineNumberStyle], ""),
		DefaultColor:    getCompiledColor(this.colorStringM[s.DefaultStyle], ""),
	}
	r.InterestFieldColor = make(map[string]func(string) string)
	r.ModuleColor = make(map[string]func(string) string)
	for k, v := range s.InterestFieldStyle {
		// 为空则为默认的颜色
		r.InterestFieldColor[k] = getCompiledColor(this.colorStringM[v], "")
	}
	for k, v := range s.ModuleStyle {
		r.ModuleColor[k.String()] = getCompiledColor(this.colorStringM[v], "")
	}

	return r
}

func (this *ColorManager) getCompiledColor(main string, fallback string) func(string) string {
	var style string
	if main != "" {
		style = main
	} else {
		style = fallback
	}
	return ansi.ColorFunc(style)
}

func (this *ColorManager) Setup() {
	v := atomic.LoadUint32(&this.status)
	if v == common.READY {
		return
	}
	this.Lock()
	defer this.Unlock()
}
