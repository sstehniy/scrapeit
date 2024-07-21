package taskqueue

import (
	"context"
	"sync"
)

type TaskJob func() error

type TaskQueue struct {
	tasks      chan TaskJob
	semaphore  chan struct{}
	wg         sync.WaitGroup
	numWorkers int
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewTaskQueue(maxConcurrentTasks, numWorkers int) *TaskQueue {
	ctx, cancel := context.WithCancel(context.Background())
	tq := &TaskQueue{
		tasks:      make(chan TaskJob),
		semaphore:  make(chan struct{}, maxConcurrentTasks),
		numWorkers: numWorkers,
		ctx:        ctx,
		cancel:     cancel,
	}

	tq.start()
	return tq
}

func (tq *TaskQueue) start() {
	for i := 0; i < tq.numWorkers; i++ {
		tq.wg.Add(1)
		go tq.worker()
	}
}

func (tq *TaskQueue) worker() {
	defer tq.wg.Done()
	for {
		select {
		case task := <-tq.tasks:
			tq.semaphore <- struct{}{}
			func() {
				defer func() { <-tq.semaphore }()
				_ = task() // Ignore error
			}()
		case <-tq.ctx.Done():
			return
		}
	}
}

func (tq *TaskQueue) AddTask(task TaskJob) {
	select {
	case tq.tasks <- task:
	case <-tq.ctx.Done():
		return
	}
}

func (tq *TaskQueue) Stop() {
	tq.cancel()
	tq.wg.Wait()
}
