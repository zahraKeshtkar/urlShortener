package workerpool

import (
	"errors"

	"url-shortner/log"
)

type Workerpool struct {
	maxWorker   int
	queuedTasks chan func() error
}

func NewWorkerpool(maxWorker int) (*Workerpool, error) {
	if maxWorker <= 0 {
		return nil, errors.New("the maxWorker value should be positive")
	}

	wp := &Workerpool{
		maxWorker:   maxWorker,
		queuedTasks: make(chan func() error),
	}

	return wp, nil
}

func (wp *Workerpool) Run() {
	for i := 0; i < wp.maxWorker; i++ {
		wID := i + 1
		go func(workerID int) {
			for task := range wp.queuedTasks {
				log.Debugf("Worker %d start processing task", wID)
				task()
				log.Debugf(" Worker %d finish processing task", wID)
			}
		}(wID)
	}
}

func (wp *Workerpool) AddTask(task func() error) {
	wp.queuedTasks <- task
}

func (wp *Workerpool) Close() {
	close(wp.queuedTasks)
}
