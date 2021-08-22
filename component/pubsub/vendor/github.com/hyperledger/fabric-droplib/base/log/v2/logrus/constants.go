/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/10 12:02 下午
# @File : constants.go
# @Description :
# @Attention :
*/
package logrusplugin

type Color int

const (
	TextDefault Color = 0
	TextBlack   Color = iota + 30
	TextRed
	TextGreen
	TextYellow
	TextBlue
	TextMagenta
	TextCyan
	TextWhite
)


