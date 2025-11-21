package service

import (
	"context"
	"errors"
	"sync"

	"github.com/gengeo7/workmate/repository"
)

type TaskQueue struct {
	mu             sync.Mutex
	queue          map[int][]string
	serial         int
	repository     repository.TaskRepository
	statusCreater  StatusCreater
	isShuttingDown bool
}

func NewTaskQueue(repo repository.TaskRepository, statusCreater StatusCreater) *TaskQueue {
	serial := 1
	return &TaskQueue{
		repository:    repo,
		serial:        serial,
		statusCreater: statusCreater,
		queue:         make(map[int][]string, 0),
	}
}

// LoadTasks is not thread safe
func (tq *TaskQueue) LoadTasks(ctx context.Context) error {
	queueqLinks, err := tq.repository.TasksGet(ctx)
	if err != nil {
		return err
	}
	for _, links := range queueqLinks {
		tq.queue[tq.serial] = links
		tq.serial++
	}
	return nil
}

// FinishTasks is not thread safe
func (tq *TaskQueue) FinishTasks(ctx context.Context) (finishedTasks int, errorTasks int, err error) {
	var errs []error
	for _, links := range tq.queue {
		_, err := CheckStatus(ctx, tq.statusCreater, links)
		if err != nil {
			errorTasks++
			errs = append(errs, err)
		} else {
			finishedTasks++
		}
	}

	tq.queue = make(map[int][]string, 0)

	err = errors.Join(errs...)

	return
}

// DumpTasks is not thread safe
func (tq *TaskQueue) DumpTasks(ctx context.Context) (savedTasks int, errorTasks int) {
	for id, links := range tq.queue {
		err := tq.repository.TasksCreate(ctx, links, id)
		if err != nil {
			errorTasks++
		} else {
			savedTasks++
		}
	}

	tq.queue = make(map[int][]string, 0)

	return
}

// AddTask is thread safe
func (tq *TaskQueue) AddTask(links []string) int {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if len(links) == 0 {
		return 0
	}

	id := tq.serial
	tq.queue[id] = links
	tq.serial++
	return id
}

// CompleteTask is thread safe
func (tq *TaskQueue) CompleteTask(id int) error {
	tq.mu.Lock()
	defer tq.mu.Unlock()
	if _, ok := tq.queue[id]; !ok {
		return errors.New("taks not found")
	}
	delete(tq.queue, id)
	return nil
}

func (tq *TaskQueue) SetShuttingDown() {
	tq.isShuttingDown = true
}

func (tq *TaskQueue) IsshuttingDown() bool {
	return tq.isShuttingDown
}
