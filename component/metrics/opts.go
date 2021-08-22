/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/5/19 9:24 上午
# @File : opts.go
# @Description :
# @Attention :
*/
package metrics

type MetricOpts func(opts *BaseOpts)

func Namespace(nameSpace string) MetricOpts {
	return func(opts *BaseOpts) {
		opts.Namespace = nameSpace
	}
}
func MetricName(name string) MetricOpts {
	return func(opts *BaseOpts) {
		opts.Name = name
	}
}
func SubSystem(subS string) MetricOpts {
	return func(opts *BaseOpts) {
		opts.Subsystem = subS
	}
}
func Help(help string) MetricOpts {
	return func(opts *BaseOpts) {
		opts.Help = help
	}
}

func LabelNames(LabelNames []string) MetricOpts {
	return func(opts *BaseOpts) {
		opts.LabelNames = LabelNames
	}
}

func StatsdFormat(StatsdFormat string) MetricOpts {
	return func(opts *BaseOpts) {
		opts.StatsdFormat = StatsdFormat
	}
}

func NewBaseOpts(opts ...MetricOpts) BaseOpts {
	r := &BaseOpts{}
	for _, opt := range opts {
		opt(r)
	}
	return *r
}
