/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 2:11 下午
# @File : helper.go
# @Description :
# @Attention :
*/
package commonutils

import (
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/duration"
	"github.com/golang/protobuf/ptypes/timestamp"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
	"time"
)

func ConvtPBTime2GoogleTime(t *timestamppb.Timestamp) *timestamp.Timestamp {
	r := &timestamp.Timestamp{
		Seconds: t.Seconds,
		Nanos:   t.Nanos,
	}
	return r
}
func ConvtTime2GoogleTime(t time.Time) *timestamp.Timestamp {
	proto, err := ptypes.TimestampProto(t)
	if nil != err {
		panic(err)
	}
	return proto
}
func ConvGoogleTime2Time(t *timestamp.Timestamp) time.Time {
	t2, err := ptypes.Timestamp(t)
	if nil != err  {
		panic(err)
	}
	return t2
}

func ConvtDuration2TimeDuration(d *duration.Duration) time.Duration {
	t, err := ptypes.Duration(d)
	if nil != err {
		panic(err)
	}
	return t
}

func ConvTimeDuration2Duration(t time.Duration) *duration.Duration {
	return ptypes.DurationProto(t)
}

// Marshal serializes a protobuf message.
func Marshal(pb proto.Message) ([]byte, error) {
	return proto.Marshal(pb)
}
func UnMarshal(bytes []byte, pb proto.Message) error {
	return proto.Unmarshal(bytes, pb)
}

func ProtoSize(msg proto.Message) int {
	return proto.Size(msg)
}
