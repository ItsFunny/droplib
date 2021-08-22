/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2020-11-27 14:08
# @File : debug.go
# @Description :
# @Attention :
*/
package libutils

import (
	"fmt"
	lg "github.com/hyperledger/fabric-droplib/base/log"
	"runtime"
	"strings"
)

var (
	// logg          *zap.SugaredLogger
	// consoleLogger *flogging.FabricLogger
	islog = true
	logg  lg.Logger
)

var (
	defaultBlackList = []string{"debugutil/debug", "utils/log"}
)

func init() {
	logg = lg.GlobalLogger("ALL")
}
// func RegisterBlackList(bl ...string) {
// 	for _, v := range bl {
//
// 		config.RegisterBlackList(v)
// 	}
// }

func UpdateLogValue(logAble bool) {
	islog = logAble
}
func init() {
	// config.RegisterBlackList("debugutil/debug")
	// config.RegisterBlackList("utils/log")
	// config.InitViper(nil, "core")
	// viper.ReadInConfig()
	// peerName := viper.GetString("peer.id")
	// logg = flogging.MustFileLogger("debugUtil", "/Users/joker/go/src/github.com/hyperledger/fabric/logs/"+peerName+".log")
	// consoleLogger = flogging.MustGetLogger("consoleLogger")
}

func InfoLogging(msg ...string) {
	if islog {
		sb := strings.Builder{}
		s, ok := FindCaller(1, defaultBlackList)
		if ok {
			sb.WriteString(s)
		}
		for _, s := range msg {
			sb.WriteString(decorate(s))
		}
		consoleInfo(sb.String())
	}
}
func OneLineWarning(msg ...string) {
	sb := strings.Builder{}
	for _, s := range msg {
		sb.WriteString(decorate(s))
	}
	str := "当前协程:[%d]  %s"
	logg.Info(fmt.Sprintf(str, CurGoroutineID(), msg))
	// consoleLogger.Warn(fmt.Sprintf(str, CurGoroutineID(), msg))
}
func ErrorLogging(msg ...string) {
	sb := strings.Builder{}
	for _, s := range msg {
		sb.WriteString(decorate(s))
	}
	m := "\n" +
		"<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n" +
		" 当前协程:[%d]  %s  \n" +
		">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> "
	// consoleLogger.Error(fmt.Sprintf(m, CurGoroutineID(), sb.String()))
	logg.Error(fmt.Sprintf(m, CurGoroutineID(), sb.String()))
}
func ImportantLogging(msg ...string) {
	sb := strings.Builder{}
	s, ok := FindCaller(1, defaultBlackList)
	if ok {
		sb.WriteString(s + "   ")
	}
	for _, s := range msg {
		sb.WriteString(decorate(s))
	}
	m := "\n" +
		"<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n" +
		" 当前协程:[%d]  %s  \n" +
		">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> "
	logg.Info(fmt.Sprintf(m, CurGoroutineID(), sb.String()))
	// consoleLogger.Warn(fmt.Sprintf(m, CurGoroutineID(), sb.String()))
	// warn(logg, sb.String())
}
func MsgLogging(msg ...string) {
	if islog {
		sb := strings.Builder{}
		for _, s := range msg {
			sb.WriteString(decorate(s))
		}
		// warn(logg, sb.String())
		consoleInfo(sb.String())
	}
}
func MultiWarning(msg ...string) {
	sb := strings.Builder{}
	for _, s := range msg {
		sb.WriteString(decorate(s))
	}
	// warn(logg, sb.String())
	consoleInfo(sb.String())
}

func FileWarning(msg ...string) {
	MultiWarning(msg...)
	// sb := strings.Builder{}
	// for _, s := range msg {
	// 	sb.WriteString(decorate(s))
	// }
	// warn(logg, sb.String())
}
func decorate(str string) string {
	return str + "\n"
}

func Warning(msg string) {
	// warn(logg, msg)
	consoleInfo(msg)
}

func Warningf(template string, data ...interface{}) {
	sprintf := fmt.Sprintf(template, data...)
	consoleInfo(sprintf)
}

func consoleInfo(msg string) {
	if islog {
		pc, _, line, ok := runtime.Caller(3)
		if ok {
			msg = fmt.Sprintf("[%s]:[%d]", runtime.FuncForPC(pc).Name(), line) + msg
		}
		m := "\n" +
			"<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n" +
			" 当前协程:[%d]  %s  \n" +
			">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> "
		// loggg.Info(fmt.Sprintf(m, CurGoroutineID(), msg))
		logg.Info("ALL", fmt.Sprintf(m, CurGoroutineID(), msg))
	}
}
func warn(msg string) {
	// base64.StdEncoding.EncodeToString(sig)
	m := "\n" +
		"<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<\n" +
		" 当前协程:[%d]  %s  \n" +
		">>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>>> "
	logg.Warn(fmt.Sprintf(m, CurGoroutineID(), msg))
}
