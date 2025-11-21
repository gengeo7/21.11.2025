package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"
	"time"

	"github.com/gengeo7/workmate/controller"
	"github.com/gengeo7/workmate/logger"
	"github.com/gengeo7/workmate/middleware"
	"github.com/gengeo7/workmate/repository/file"
	"github.com/gengeo7/workmate/service"
)

func main() {
	logger.Init()
	defer func() {
		if err := recover(); err != nil {
			logger.Error("PANIC", "error", err, "stacktrace", string(debug.Stack()))
		}
	}()

	repo, err := file.NewFileRepository()
	if err != nil {
		logger.Error("repository error", "error", err)
		return
	}

	queue := service.NewTaskQueue(repo, repo)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()
	err = queue.LoadTasks(ctx)
	if err != nil {
		logger.Error("could not load tasks", "error", err)
	}
	finishedTasks, errorTasks, err := queue.FinishTasks(ctx)
	logger.Info("succesfully finished tasks", "count", finishedTasks)
	if err != nil {
		logger.Error("failed to finish all tasks", "error", err, "count", errorTasks)
	}

	mux := http.NewServeMux()
	c := controller.NewController(repo, queue)
	c.RegisterController(mux)

	handler := middleware.Recoverer(
		mux,
	)

	server := http.Server{
		Addr:         "localhost:3000",
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  600 * time.Second,
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup
	wg.Go(func() {
		<-sigChan

		logger.Info("server is shutting down...")
		queue.SetShuttingDown()
		time.Sleep(5 * time.Second)
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()
		savedTasks, errorTasks := queue.DumpTasks(ctx)
		logger.Info("succesfully saved tasks", "count", savedTasks)
		logger.Warn("error while saving tasks", "count", errorTasks)

		if err := server.Shutdown(ctx); err != nil {
			logger.Error("shutdown error", "error", err)
		}
		logger.Info("server is down")
	})

	logger.Info("ready to start", "addr", server.Addr)
	server.ListenAndServe()
	wg.Wait()
	logger.Info("server stopped")
}
