package controller

import (
	"net/http"
	"time"

	"github.com/gengeo7/workmate/middleware"
	"github.com/gengeo7/workmate/repository"
	"github.com/gengeo7/workmate/service"
	"github.com/gengeo7/workmate/types"
	"github.com/gengeo7/workmate/utils"
)

type Controller struct {
	Repository repository.StatusRepository
	Queue      *service.TaskQueue
}

func NewController(repo repository.StatusRepository, queue *service.TaskQueue) *Controller {
	return &Controller{
		Repository: repo,
		Queue:      queue,
	}
}

func (c *Controller) RegisterController(mux *http.ServeMux) {
	mux.Handle(
		"POST /status",
		middleware.Chain(
			http.HandlerFunc(c.getStatus),
			middleware.ValidateJson[types.StatusReq](),
			middleware.Timeout(1*time.Second),
		),
	)

	mux.Handle(
		"POST /pdf",
		middleware.Chain(
			http.HandlerFunc(c.getPdf),
			middleware.ValidateJson[types.PdfReq](),
			middleware.Timeout(5*time.Second),
		),
	)
}

func (c *Controller) getStatus(w http.ResponseWriter, r *http.Request) {
	statusReq := middleware.DtoFromContext[types.StatusReq](r.Context())
	if statusReq == nil {
		utils.SendResponse(nil, utils.NewApiError(http.StatusBadRequest, "пустой запрос", nil), w, r)
		return
	}
	taskId := c.Queue.AddTask(statusReq.Links)
	if c.Queue.IsshuttingDown() {
		utils.SendResponse(nil, utils.NewApiError(http.StatusServiceUnavailable, "сервер не доступен в данный момент, запрос будет обработан, как только сервер станет доступен", nil), w, r)
		return
	}
	statusRes, err := service.CheckStatus(r.Context(), c.Repository, statusReq.Links)
	utils.SendResponse(&utils.Response{Data: statusRes, Status: http.StatusCreated}, err, w, r)
	c.Queue.CompleteTask(taskId)
}

func (c *Controller) getPdf(w http.ResponseWriter, r *http.Request) {
	if c.Queue.IsshuttingDown() {
		utils.SendResponse(nil, utils.NewApiError(http.StatusServiceUnavailable, "сервер не доступен в данный момент", nil), w, r)
		return
	}
	pdfReq := middleware.DtoFromContext[types.PdfReq](r.Context())
	if pdfReq == nil {
		utils.SendResponse(nil, utils.NewApiError(http.StatusBadRequest, "пустой запрос", nil), w, r)
		return
	}
	pdf, err := service.ProducePdf(r.Context(), c.Repository, pdfReq.LinksList)
	if err != nil {
		utils.SendResponse(nil, err, w, r)
		return
	}

	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"report.pdf\"")
	w.WriteHeader(http.StatusOK)
	w.Write(pdf)
}
