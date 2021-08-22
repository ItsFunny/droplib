/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/8 9:09 上午
# @File : ticker.go
# @Description :
# @Attention :
*/
package ticker

import (
	"github.com/hyperledger/fabric-droplib/base/log/modules"
	"github.com/hyperledger/fabric-droplib/base/services/impl"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/models"
	"github.com/hyperledger/fabric-droplib/component/consensus/bft/tendermint/services"
	"time"
)

var (
	_ services.ITimeoutTicker = (*timeoutTicker)(nil)
)

var (
	tickTockBufferSize = 10
)

type timeoutTicker struct {
	*impl.BaseServiceImpl
	timer    *time.Timer
	tickChan chan models.TimeoutInfo // for scheduling timeouts
	tockChan chan models.TimeoutInfo // for notifying about them
}

// NewTimeoutTicker returns a new TimeoutTicker.
func NewTimeoutTicker() services.ITimeoutTicker {
	tt := &timeoutTicker{
		timer:    time.NewTimer(0),
		tickChan: make(chan models.TimeoutInfo, tickTockBufferSize),
		tockChan: make(chan models.TimeoutInfo, tickTockBufferSize),
	}
	tt.BaseServiceImpl = impl.NewBaseService(nil, modules.MODULE_TIMEOUT_TICKER, tt)
	tt.stopTimer() // don't want to fire until the first scheduled timeout
	return tt
}

// OnStart implements service.Service. It starts the timeout routine.
func (t *timeoutTicker) OnStart() error {

	go t.timeoutRoutine()

	return nil
}

// OnStop implements service.Service. It stops the timeout routine.
func (t *timeoutTicker) OnStop() {
	t.BaseServiceImpl.OnStop()
	t.stopTimer()
}

// Chan returns a channel on which timeouts are sent.
func (t *timeoutTicker) Chan() <-chan models.TimeoutInfo {
	return t.tockChan
}

// ScheduleTimeout schedules a new timeout by sending on the internal tickChan.
// The timeoutRoutine is always available to read from tickChan, so this won't block.
// The scheduling may fail if the timeoutRoutine has already scheduled a timeout for a later height/round/step.
func (t *timeoutTicker) ScheduleTimeout(ti models.TimeoutInfo) {
	t.tickChan <- ti
}

func (t *timeoutTicker) timeoutRoutine() {
	t.Logger.Debug("Starting timeout routine")
	var ti models.TimeoutInfo
	for {
		select {
		case newti := <-t.tickChan:
			t.Logger.Debug("Received tick", "old_ti", ti, "new_ti", newti)

			// ignore tickers for old height/round/step
			if newti.Height < ti.Height {
				continue
			} else if newti.Height == ti.Height {
				if newti.Round < ti.Round {
					continue
				} else if newti.Round == ti.Round {
					if ti.Step > 0 && newti.Step <= ti.Step {
						continue
					}
				}
			}

			// stop the last timer
			t.stopTimer()

			// update timeoutInfo and reset timer
			// NOTE time.Timer allows duration to be non-positive
			ti = newti
			t.timer.Reset(ti.Duration)
			t.Logger.Debug("Scheduled timeout", "dur", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)
		case <-t.timer.C:
			t.Logger.Info("Timed out", "dur", ti.Duration, "height", ti.Height, "round", ti.Round, "step", ti.Step)
			// go routine here guarantees timeoutRoutine doesn't block.
			// Determinism comes from playback in the receiveRoutine.
			// We can eliminate it by merging the timeoutRoutine into receiveRoutine
			//  and managing the timeouts ourselves with a millisecond ticker
			go func(toi models.TimeoutInfo) { t.tockChan <- toi }(ti)
		case <-t.Quit():
			return
		}
	}
}

func (t *timeoutTicker) stopTimer() {
	// Stop() returns false if it was already fired or was stopped
	if !t.timer.Stop() {
		select {
		case <-t.timer.C:
		default:
			t.Logger.Debug("Timer already stopped")
		}
	}
}
