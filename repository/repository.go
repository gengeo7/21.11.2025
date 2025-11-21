package repository

import (
	"context"
	"errors"

	"github.com/gengeo7/workmate/types"
)

var (
	ErrRepoNotFound error = errors.New("not found in repository")
)

func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrRepoNotFound)
}

func IsErrDeadline(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

type StatusRepository interface {
	StatusCreate(links map[string]types.StatusEnum) (*types.StatusResp, error)
	StatusGet(num int) (*types.StatusResp, error)
}

type TaskRepository interface {
	TasksCreate(ctx context.Context, links []string, id int) error
	TasksGet(ctx context.Context) ([][]string, error)
}
