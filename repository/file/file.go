package file

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gengeo7/workmate/repository"
	"github.com/gengeo7/workmate/types"
)

type FileRepository struct {
	mu        sync.RWMutex
	serial    int
	statusDir string
	tasksDir  string
}

func NewFileRepository() (*FileRepository, error) {
	statusDir := "status_data"
	tasksDir := "tasks_data"
	if err := os.MkdirAll(statusDir, 0755); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(tasksDir, 0755); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(statusDir)
	serial := 0

	if err == nil && len(entries) > 0 {
		serial = len(entries) + 1
	}

	return &FileRepository{
		statusDir: statusDir,
		tasksDir:  tasksDir,
		serial:    serial,
	}, nil
}

func (fr *FileRepository) StatusCreate(links map[string]types.StatusEnum) (*types.StatusResp, error) {
	fr.mu.Lock()
	currentSerial := fr.serial
	fr.serial++
	fr.mu.Unlock()

	fname := filepath.Join(fr.statusDir, "links"+strconv.Itoa(fr.serial)+".json")
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(links); err != nil {
		return nil, err
	}

	statusResp := types.StatusResp{
		Links:    links,
		LinksNum: currentSerial,
	}

	return &statusResp, nil
}

func (fr *FileRepository) StatusGet(num int) (*types.StatusResp, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	fname := filepath.Join(fr.statusDir, "links"+strconv.Itoa(num)+".json")
	file, err := os.Open(fname)
	if err != nil {
		return nil, repository.ErrRepoNotFound
	}
	defer file.Close()

	var links map[string]types.StatusEnum
	if err := json.NewDecoder(file).Decode(&links); err != nil {
		return nil, err
	}

	statusResp := types.StatusResp{
		Links:    links,
		LinksNum: num,
	}

	return &statusResp, nil
}

func (fr *FileRepository) TasksCreate(ctx context.Context, links []string, id int) error {
	fr.mu.Lock()
	defer fr.mu.Unlock()

	fname := filepath.Join(fr.tasksDir, "task"+strconv.Itoa(id)+".json")
	file, err := os.OpenFile(fname, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	if err := json.NewEncoder(file).Encode(links); err != nil {
		return err
	}

	return nil
}

func (fr *FileRepository) TasksGet(ctx context.Context) ([][]string, error) {
	fr.mu.RLock()
	defer fr.mu.RUnlock()

	entries, err := os.ReadDir(fr.tasksDir)
	tasks := make([][]string, 0)
	if err != nil {
		return nil, err
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "task") && strings.HasSuffix(e.Name(), ".json") {
			fname := filepath.Join(fr.tasksDir, e.Name())
			f, err := os.Open(fname)
			if err != nil {
				continue
			}
			var links []string
			if err := json.NewDecoder(f).Decode(&links); err != nil {
				f.Close()
				continue
			}

			tasks = append(tasks, links)
			f.Close()
			os.Remove(fname)
		}
	}

	return tasks, nil
}
