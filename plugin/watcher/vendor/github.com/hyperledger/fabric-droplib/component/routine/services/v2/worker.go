package v2

import (
	"runtime"
	"time"
)

type goWorker struct {
	// pool who owns this worker.
	pool *Pool

	// task is a job should be done.
	// task chan func()
	task TaskQueue

	// recycleTime will be update when putting a worker back into queue.
	recycleTime time.Time
}

// run starts a goroutine to repeat the process
// that performs the function calls.
func (w *goWorker) run() {
	w.pool.incRunning()
	go func() {
		defer func() {
			w.pool.decRunning()
			w.pool.workerCache.Put(w)
			if p := recover(); p != nil {
				if ph := w.pool.options.PanicHandler; ph != nil {
					ph(p)
				} else {
					w.pool.options.Logger.Printf("worker exits from a panic: %v\n", p)
					var buf [4096]byte
					n := runtime.Stack(buf[:], false)
					w.pool.options.Logger.Printf("worker exits from panic: %s\n", string(buf[:n]))
				}
			}
			// Call Signal() here in case there are goroutines waiting for available workers.
			w.pool.cond.Signal()
		}()
		var task ITask
		for {
			task = w.task.Take()
			if nil == task || !task.Status().available() {
				return
			}
			task.Execute()
			task = nil
			if ok := w.pool.revertWorker(w); !ok {
				return
			}
		}
		// for f := range w.task {
		// 	if f == nil {
		// 		return
		// 	}
		// 	f()
		// 	if ok := w.pool.revertWorker(w); !ok {
		// 		return
		// 	}
		// }
	}()
}
