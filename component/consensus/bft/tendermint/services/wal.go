/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 3:19 下午
# @File : wal.go
# @Description :
# @Attention :
*/
package services

import (
	"fmt"
	v2 "github.com/hyperledger/fabric-droplib/base/log/v2"
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/types"
	"github.com/hyperledger/fabric-droplib/libs"
	auto "github.com/hyperledger/fabric-droplib/libs/autofiles"
	tmjson "github.com/hyperledger/fabric-droplib/libs/json"
	"io"
	"path/filepath"
	"time"
)

const (
	// time.Time + max consensus msg size
	maxMsgSizeBytes = types.MaxMsgSize + 24

	// how often the WAL should be sync'd during period sync'ing
	walDefaultFlushInterval = 2 * time.Second
)

func init() {
	tmjson.RegisterType(models.MsgInfo{}, "tendermint/wal/MsgInfo")
	tmjson.RegisterType(models.TimeoutInfo{}, "tendermint/wal/TimeoutInfo")
	tmjson.RegisterType(EndHeightMessage{}, "tendermint/wal/EndHeightMessage")
}


type EndHeightMessage struct {
	Height uint64 `json:"height"`
}
type TimedWALMessage struct {
	Time time.Time  `json:"time"`
	Msg  WALMessage `json:"msg"`
}

// WALSearchOptions are optional arguments to SearchForEndHeight.
type WALMessage interface{}

// WAL is an interface for any write-ahead logger.
type IWAL interface {
	Write(WALMessage) error
	WriteSync(WALMessage) error
	FlushAndSync() error

	SearchForEndHeight(height uint64, options *WALSearchOptions) (rd io.ReadCloser, found bool, err error)

	// service methods
	Start(flag services.START_FLAG) error
	Stop() error
	Wait()
}

type BaseWAL struct {
	*impl.BaseServiceImpl

	group *auto.Group

	enc *WALEncoder

	flushTicker   *time.Ticker
	flushInterval time.Duration
}

var _ IWAL = &BaseWAL{}

// NewWAL returns a new write-ahead logger based on `baseWAL`, which implements
// WAL. It's flushed and synced to disk every 2s and once when stopped.
func NewWAL(walFile string, groupOptions ...func(*auto.Group)) (*BaseWAL, error) {
	err := libs.EnsureDir(filepath.Dir(walFile), 0700)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure WAL directory is in place: %w", err)
	}

	group, err := auto.OpenGroup(walFile, groupOptions...)
	if err != nil {
		return nil, err
	}
	wal := &BaseWAL{
		group:         group,
		enc:           NewWALEncoder(group),
		flushInterval: walDefaultFlushInterval,
	}
	wal.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_WAL, wal)
	return wal, nil
}

// SetFlushInterval allows us to override the periodic flush interval for the WAL.
func (wal *BaseWAL) SetFlushInterval(i time.Duration) {
	wal.flushInterval = i
}

func (wal *BaseWAL) Group() *auto.Group {
	return wal.group
}

func (wal *BaseWAL) SetLogger(l v2.Logger) {
	wal.BaseServiceImpl.Logger = l
	wal.group.SetLogger(l)
}

func (wal *BaseWAL) OnStart() error {
	size, err := wal.group.Head.Size()
	if err != nil {
		return err
	} else if size == 0 {
		if err := wal.WriteSync(EndHeightMessage{0}); err != nil {
			return err
		}
	}
	err = wal.group.Start(services.ASYNC_START)
	if err != nil {
		return err
	}
	wal.flushTicker = time.NewTicker(wal.flushInterval)
	go wal.processFlushTicks()
	return nil
}

func (wal *BaseWAL) processFlushTicks() {
	for {
		select {
		case <-wal.flushTicker.C:
			if err := wal.FlushAndSync(); err != nil {
				wal.Logger.Error("Periodic WAL flush failed", "err", err)
			}
		case <-wal.Quit():
			return
		}
	}
}

// FlushAndSync flushes and fsync's the underlying group's data to disk.
// See auto#FlushAndSync
func (wal *BaseWAL) FlushAndSync() error {
	return wal.group.FlushAndSync()
}

// Stop the underlying autofile group.
// Use Wait() to ensure it's finished shutting down
// before cleaning up files.
func (wal *BaseWAL) OnStop() {
	wal.flushTicker.Stop()
	if err := wal.FlushAndSync(); err != nil {
		wal.Logger.Error("error on flush data to disk", "error", err)
	}
	if err := wal.group.Stop(); err != nil {
		wal.Logger.Error("error trying to stop wal", "error", err)
	}
	wal.group.Close()
}

// Wait for the underlying autofile group to finish shutting down
// so it's safe to cleanup files.
func (wal *BaseWAL) Wait() {
	wal.group.Wait()
}

// Write is called in newStep and for each receive on the
// peerMsgQueue and the timeoutTicker.
// NOTE: does not call fsync()
func (wal *BaseWAL) Write(msg WALMessage) error {
	if wal == nil {
		return nil
	}

	if err := wal.enc.Encode(&TimedWALMessage{libs.Now(), msg}); err != nil {
		wal.Logger.Error("Error writing msg to consensus wal. WARNING: recover may not be possible for the current height",
			"err", err, "msg", msg)
		return err
	}

	return nil
}

// WriteSync is called when we receive a msg from ourselves
// so that we write to disk before sending signed messages.
// NOTE: calls fsync()
func (wal *BaseWAL) WriteSync(msg WALMessage) error {
	if wal == nil {
		return nil
	}

	if err := wal.Write(msg); err != nil {
		return err
	}

	if err := wal.FlushAndSync(); err != nil {
		wal.Logger.Error(`WriteSync failed to flush consensus wal.
		WARNING: may result in creating alternative proposals / votes for the current height iff the node restarted`,
			"err", err)
		return err
	}

	return nil
}

// WALSearchOptions are optional arguments to SearchForEndHeight.
type WALSearchOptions struct {
	// IgnoreDataCorruptionErrors set to true will result in skipping data corruption errors.
	IgnoreDataCorruptionErrors bool
}

// SearchForEndHeight searches for the EndHeightMessage with the given height
// and returns an auto.GroupReader, whenever it was found or not and an error.
// Group reader will be nil if found equals false.
//
// CONTRACT: caller must close group reader.
func (wal *BaseWAL) SearchForEndHeight(
	height uint64,
	options *WALSearchOptions) (rd io.ReadCloser, found bool, err error) {
	var (
		msg *TimedWALMessage
		gr  *auto.GroupReader
	)
	lastHeightFound := int64(-1)

	// NOTE: starting from the last file in the group because we're usually
	// searching for the last height. See replay.go
	min, max := wal.group.MinIndex(), wal.group.MaxIndex()
	wal.Logger.Info("Searching for height", "height", height, "min", min, "max", max)
	for index := max; index >= min; index-- {
		gr, err = wal.group.NewReader(index)
		if err != nil {
			return nil, false, err
		}

		dec := NewWALDecoder(gr)
		for {
			msg, err = dec.Decode()
			if err == io.EOF {
				// OPTIMISATION: no need to look for height in older files if we've seen h < height
				if lastHeightFound > 0 && uint64(lastHeightFound) < height {
					gr.Close()
					return nil, false, nil
				}
				// check next file
				break
			}
			if options.IgnoreDataCorruptionErrors && IsDataCorruptionError(err) {
				wal.Logger.Error("Corrupted entry. Skipping...", "err", err)
				// do nothing
				continue
			} else if err != nil {
				gr.Close()
				return nil, false, err
			}

			if m, ok := msg.Msg.(EndHeightMessage); ok {
				lastHeightFound = int64(m.Height)
				if m.Height == height { // found
					wal.Logger.Info("Found", "height", height, "index", index)
					return gr, true, nil
				}
			}
		}
		gr.Close()
	}

	return nil, false, nil
}

// A WALEncoder writes custom-encoded WAL messages to an output stream.
//
// Format: 4 bytes CRC sum + 4 bytes length + arbitrary-length value
type WALEncoder struct {
	wr io.Writer
}

// NewWALEncoder returns a new encoder that writes to wr.
func NewWALEncoder(wr io.Writer) *WALEncoder {
	return &WALEncoder{wr}
}

// Encode writes the custom encoding of v to the stream. It returns an error if
// the encoded size of v is greater than 1MB. Any error encountered
// during the write is also returned.
func (enc *WALEncoder) Encode(v *TimedWALMessage) error {
	// pbMsg, err := WALToProto(v.Msg)
	// if err != nil {
	// 	return err
	// }
	// pv := TimedWALMessage{
	// 	Time: v.Time,
	// 	Msg:  pbMsg,
	// }
	//
	// // data, err := proto.Marshal(&pv)
	// if err != nil {
	// 	panic(fmt.Errorf("encode timed wall message failure: %w", err))
	// }
	//
	// crc := crc32.Checksum(data, crc32c)
	// length := uint32(len(data))
	// if length > maxMsgSizeBytes {
	// 	return fmt.Errorf("msg is too big: %d bytes, max: %d bytes", length, maxMsgSizeBytes)
	// }
	// totalLength := 8 + int(length)
	//
	// msg := make([]byte, totalLength)
	// binary.BigEndian.PutUint32(msg[0:4], crc)
	// binary.BigEndian.PutUint32(msg[4:8], length)
	// copy(msg[8:], data)
	//
	// _, err = enc.wr.Write(msg)
	// return err
	return nil
}

// IsDataCorruptionError returns true if data has been corrupted inside WAL.
func IsDataCorruptionError(err error) bool {
	_, ok := err.(DataCorruptionError)
	return ok
}

// DataCorruptionError is an error that occures if data on disk was corrupted.
type DataCorruptionError struct {
	cause error
}

func (e DataCorruptionError) Error() string {
	return fmt.Sprintf("DataCorruptionError[%v]", e.cause)
}

func (e DataCorruptionError) Cause() error {
	return e.cause
}

// A WALDecoder reads and decodes custom-encoded WAL messages from an input
// stream. See WALEncoder for the format used.
//
// It will also compare the checksums and make sure data size is equal to the
// length from the header. If that is not the case, error will be returned.
type WALDecoder struct {
	Rd io.Reader
}

// NewWALDecoder returns a new decoder that reads from rd.
func NewWALDecoder(rd io.Reader) *WALDecoder {
	return &WALDecoder{rd}
}

// Decode reads the next custom-encoded value from its reader and returns it.
func (dec *WALDecoder) Decode() (*TimedWALMessage, error) {
	// b := make([]byte, 4)
	//
	// _, err := dec.rd.Read(b)
	// if errors.Is(err, io.EOF) {
	// 	return nil, err
	// }
	// if err != nil {
	// 	return nil, DataCorruptionError{fmt.Errorf("failed to read checksum: %v", err)}
	// }
	// crc := binary.BigEndian.Uint32(b)
	//
	// b = make([]byte, 4)
	// _, err = dec.rd.Read(b)
	// if err != nil {
	// 	return nil, DataCorruptionError{fmt.Errorf("failed to read length: %v", err)}
	// }
	// length := binary.BigEndian.Uint32(b)
	//
	// if length > maxMsgSizeBytes {
	// 	return nil, DataCorruptionError{fmt.Errorf(
	// 		"length %d exceeded maximum possible value of %d bytes",
	// 		length,
	// 		maxMsgSizeBytes)}
	// }
	//
	// data := make([]byte, length)
	// n, err := dec.rd.Read(data)
	// if err != nil {
	// 	return nil, DataCorruptionError{fmt.Errorf("failed to read data: %v (read: %d, wanted: %d)", err, n, length)}
	// }
	//
	// // check checksum before decoding data
	// actualCRC := crc32.Checksum(data, crc32c)
	// if actualCRC != crc {
	// 	return nil, DataCorruptionError{fmt.Errorf("checksums do not match: read: %v, actual: %v", crc, actualCRC)}
	// }
	//
	// var res = new(TimedWALMessage)
	// err = proto.Unmarshal(data, res)
	// if err != nil {
	// 	return nil, DataCorruptionError{fmt.Errorf("failed to decode data: %v", err)}
	// }
	//
	// walMsg, err := WALFromProto(res.Msg)
	// if err != nil {
	// 	return nil, DataCorruptionError{fmt.Errorf("failed to convert from proto: %w", err)}
	// }
	// tMsgWal := &TimedWALMessage{
	// 	Time: res.Time,
	// 	Msg:  walMsg,
	// }
	//
	// return tMsgWal, err
	return nil, nil
}

type NilWAL struct{}

var _ IWAL = NilWAL{}

func (NilWAL) Write(m WALMessage) error     { return nil }
func (NilWAL) WriteSync(m WALMessage) error { return nil }
func (NilWAL) FlushAndSync() error          { return nil }
func (NilWAL) SearchForEndHeight(height uint64, options *WALSearchOptions) (rd io.ReadCloser, found bool, err error) {
	return nil, false, nil
}
func (NilWAL) Start(flag services.START_FLAG) error { return nil }
func (NilWAL) Stop() error  { return nil }
func (NilWAL) Wait()        {}

// WALToProto takes a WAL message and return a proto walMessage and error.
// func WALToProto(msg WALMessage) (*tmcons.WALMessage, error) {
// 	var pb tmcons.WALMessage
//
// 	switch msg := msg.(type) {
// 	case types.EventDataRoundState:
// 		pb = tmcons.WALMessage{
// 			Sum: &tmcons.WALMessage_EventDataRoundState{
// 				EventDataRoundState: &tmproto.EventDataRoundState{
// 					Height: msg.Height,
// 					Round:  msg.Round,
// 					Step:   msg.Step,
// 				},
// 			},
// 		}
// 	case msgInfo:
// 		consMsg, err := MsgToProto(msg.Msg)
// 		if err != nil {
// 			return nil, err
// 		}
// 		pb = tmcons.WALMessage{
// 			Sum: &tmcons.WALMessage_MsgInfo{
// 				MsgInfo: &tmcons.MsgInfo{
// 					Msg:    *consMsg,
// 					PeerID: string(msg.PeerID),
// 				},
// 			},
// 		}
//
// 	case timeoutInfo:
// 		pb = tmcons.WALMessage{
// 			Sum: &tmcons.WALMessage_TimeoutInfo{
// 				TimeoutInfo: &tmcons.TimeoutInfo{
// 					Duration: msg.Duration,
// 					Height:   msg.Height,
// 					Round:    msg.Round,
// 					Step:     uint32(msg.Step),
// 				},
// 			},
// 		}
//
// 	case EndHeightMessage:
// 		pb = tmcons.WALMessage{
// 			Sum: &tmcons.WALMessage_EndHeight{
// 				EndHeight: &tmcons.EndHeight{
// 					Height: msg.Height,
// 				},
// 			},
// 		}
//
// 	default:
// 		return nil, fmt.Errorf("to proto: wal message not recognized: %T", msg)
// 	}
//
// 	return &pb, nil
// }

// WALFromProto takes a proto wal message and return a consensus walMessage and
// error.
// func WALFromProto(msg *WALMessage) (WALMessage, error) {
// 	if msg == nil {
// 		return nil, errors.New("nil WAL message")
// 	}
//
// 	var pb WALMessage
//
// switch msg := msg.Sum.(type) {
// case *tmcons.WALMessage_EventDataRoundState:
// 	pb = types.EventDataRoundState{
// 		Height: msg.EventDataRoundState.Height,
// 		Round:  msg.EventDataRoundState.Round,
// 		Step:   msg.EventDataRoundState.Step,
// 	}
//
// case *tmcons.WALMessage_MsgInfo:
// 	walMsg, err := MsgFromProto(&msg.MsgInfo.Msg)
// 	if err != nil {
// 		return nil, fmt.Errorf("msgInfo from proto error: %w", err)
// 	}
// 	pb = msgInfo{
// 		Msg:    walMsg,
// 		PeerID: p2p.NodeID(msg.MsgInfo.PeerID),
// 	}
//
// case *tmcons.WALMessage_TimeoutInfo:
// 	tis, err := tmmath.SafeConvertUint8(int64(msg.TimeoutInfo.Step))
// 	// deny message based on possible overflow
// 	if err != nil {
// 		return nil, fmt.Errorf("denying message due to possible overflow: %w", err)
// 	}
//
// 	pb = timeoutInfo{
// 		Duration: msg.TimeoutInfo.Duration,
// 		Height:   msg.TimeoutInfo.Height,
// 		Round:    msg.TimeoutInfo.Round,
// 		Step:     cstypes.RoundStepType(tis),
// 	}
//
// 	return pb, nil
//
// case *tmcons.WALMessage_EndHeight:
// 	pb := EndHeightMessage{
// 		Height: msg.EndHeight.Height,
// 	}
//
// 	return pb, nil
// 	default:
// 		return nil, fmt.Errorf("from proto: wal message not recognized: %T", msg)
// 	}
//
// 	return pb, nil
// }
