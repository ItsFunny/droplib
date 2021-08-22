/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/3 11:04 上午
# @File : queue.go
# @Description :
# @Attention :
*/
package impl

import (
	"github.com/hyperledger/fabric-droplib/common/models"
	"sync"
)

type Queue interface {
	// enqueue returns a channel for submitting envelopes.
	Enqueue() chan<- models.Envelope

	// dequeue returns a channel ordered according to some queueing policy.
	Dequeue() <-chan models.Envelope

	// close closes the queue. After this call enqueue() will block, so the
	// caller must select on closed() as well to avoid blocking forever. The
	// enqueue() and dequeue() channels will not be closed.
	Close()

	// closed returns a channel that's closed when the scheduler is closed.
	Closed() <-chan struct{}

}

// fifoQueue is a simple unbuffered lossless queue that passes messages through
// in the order they were received, and blocks until message is received.
type fifoQueue struct {
	queueCh   chan models.Envelope
	closeCh   chan struct{}
	closeOnce sync.Once
}

var _ Queue = (*fifoQueue)(nil)

func NewFIFOQueue(size int) *fifoQueue {
	return &fifoQueue{
		queueCh: make(chan models.Envelope, size),
		closeCh: make(chan struct{}),
	}
}

func (q *fifoQueue) Enqueue() chan<- models.Envelope {
	return q.queueCh
}

func (q *fifoQueue) Dequeue() <-chan models.Envelope {
	return q.queueCh
}

func (q *fifoQueue) Close() {
	q.closeOnce.Do(func() {
		close(q.closeCh)
	})
}

func (q *fifoQueue) Closed() <-chan struct{} {
	return q.closeCh
}


// 内部反转使用,既原先为
func (q *fifoQueue) InternalChangeSizeQueue() chan models.Envelope {
	return q.queueCh
}
