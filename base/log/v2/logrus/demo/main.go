/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 10:13 上午
# @File : main.go
# @Description :
# @Attention :
*/
package main

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	logrusplugin "github.com/hyperledger/fabric-droplib/base/log/v2/logrus"
	"github.com/sirupsen/logrus"
)

var log = logrus.New()

func init() {
	log.Formatter = new(logrusplugin.TextFormatter)
	log.Level = logrus.DebugLevel
}

func main() {

	// diffModuleWithDiffColor()
	// disableColor()

	// withFieldsColor()
	//
	// withNoSpecific()
	// testWithFileds()
	testWith()
}
func testWith() {
	mm := make(map[string]logrusplugin.Color)
	mm["AAAA"] = logrusplugin.TextYellow
	logrusplugin.SetupDefaultWithInterest(mm)
	logger := logrusplugin.NewLogrusLogger(modules.NewModule("aaa", 1))
	newL := logger.With(map[string]interface{}{"TEST": "NEW_LOGGER"})
	newL.Info("123")
}
func testWithFileds() {
	mm := make(map[string]logrusplugin.Color)
	mm["AAAA"] = logrusplugin.TextYellow
	logrusplugin.SetupDefaultWithInterest(mm)
	logger := logrusplugin.NewLogrusLogger(modules.NewModule("aaa", 1))
	logger.Info("123", "AAAA", "1234", "asdddd")
}
func diffModuleWithDiffColor() {
	m := make(map[modules.Module]logrusplugin.Color)
	mod_1 := modules.NewModule("MODULE_EVENT", 1)
	mod_2 := modules.NewModule("MODULE_DEMO", 2)
	mod_3 := modules.NewModule("MODULE_ASDDD", 3)
	m[mod_1] = logrusplugin.TextRed
	m[mod_2] = logrusplugin.TextBlue
	m[mod_3] = logrusplugin.TextGreen
	logrusplugin.SetupDefaultWithInterestModule(m)

	l := logrus.New()
	l.SetFormatter(logrusplugin.NewTextFormmater())

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_1.String(),
	}).Info("这是mod_1")

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_2.String(),
	}).Info("这是mod_2")

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_3.String(),
	}).Info("这是mod_3")
}

func disableColor() {
	logrusplugin.SetupDisableColor()

	m := make(map[modules.Module]logrusplugin.Color)
	mod_1 := modules.NewModule("MODULE_EVENT", 1)
	mod_2 := modules.NewModule("MODULE_DEMO", 2)
	mod_3 := modules.NewModule("MODULE_ASDDD", 3)
	m[mod_1] = logrusplugin.TextRed
	m[mod_2] = logrusplugin.TextBlue
	m[mod_3] = logrusplugin.TextGreen
	logrusplugin.SetupDefaultWithInterestModule(m)

	l := logrus.New()
	l.SetFormatter(logrusplugin.NewTextFormmater())

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_1.String(),
	}).Info("这是mod_1")

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_2.String(),
	}).Info("这是mod_2")

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_3.String(),
	}).Info("这是mod_3")
}

func withFieldsColor() {
	m := make(map[modules.Module]logrusplugin.Color)
	mod_1 := modules.NewModule("MODULE_EVENT", 1)
	mod_2 := modules.NewModule("MODULE_DEMO", 2)
	mod_3 := modules.NewModule("MODULE_ASDDD", 3)
	m[mod_1] = logrusplugin.TextRed
	m[mod_2] = logrusplugin.TextBlue
	m[mod_3] = logrusplugin.TextGreen

	field_1 := "field_1"
	field_2 := "field_2"
	field_3 := "field_3"
	fm := make(map[string]logrusplugin.Color)
	fm[field_1] = logrusplugin.TextYellow
	fm[field_2] = logrusplugin.TextRed
	fm[field_3] = logrusplugin.TextBlue

	logrusplugin.SetupDefaultWithInterestModuleOrFields(m, fm)

	l := logrus.New()
	l.SetFormatter(logrusplugin.NewTextFormmater())

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE:           mod_1.String(),
		field_1:                       "11111",
		logrusplugin.CODE_LINE_NUMBER: "mmmmm",
	}).Info("aaaaaa")

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_2.String(),
		field_2:             "22222",
	}).Info("bbbbb")

	l.WithFields(logrus.Fields{
		logrusplugin.MODULE: mod_3.String(),
		field_3:             "33333",
	}).Info("ccccc")
}

func withNoSpecific() {
	l := logrus.New()
	l.SetFormatter(logrusplugin.NewTextFormmater())
	type s struct {
		name string
		age  int
	}
	ss := s{
		name: "joker",
		age:  123,
	}
	l.Info("传输失败", "key", "asdkkkkk", "value为", ss)
}
