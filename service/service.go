package service

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gengeo7/workmate/repository"
	"github.com/gengeo7/workmate/types"
	"github.com/gengeo7/workmate/utils"
	"github.com/signintech/gopdf"
)

type StatusCreater interface {
	StatusCreate(links map[string]types.StatusEnum) (*types.StatusResp, error)
}

type StatusGetter interface {
	StatusGet(num int) (*types.StatusResp, error)
}

func ProducePdf(ctx context.Context, statusGetter StatusGetter, linksList []int) ([]byte, error) {
	pdfData := make(map[int]map[string]types.StatusEnum, 0)

	for _, n := range linksList {
		statusResp, err := statusGetter.StatusGet(n)
		if err != nil {
			return nil, utils.TestDbErr(err, &utils.ErrDbCase{Func: repository.IsErrNotFound, Creator: utils.NotFound, CheckErr: false})
		}
		pdfData[n] = statusResp.Links
	}

	pdf, err := generatePDF(pdfData)
	return pdf, err
}

func CheckStatus(ctx context.Context, statusCreater StatusCreater, links []string) (*types.StatusResp, error) {
	if len(links) == 0 {
		return nil, utils.NewApiError(400, "links list is empty", nil)
	}
	linksWithStatus := make(map[string]types.StatusEnum, 0)
	var wg sync.WaitGroup
	var mu sync.Mutex
	jobChan := make(chan string, len(links))

	numWorkers := min(5, len(links))
	for range numWorkers {
		wg.Add(1)
		go worker(ctx, &wg, jobChan, linksWithStatus, &mu)
	}

	for _, link := range links {
		jobChan <- link
	}

	close(jobChan)

	wg.Wait()

	status, err := statusCreater.StatusCreate(linksWithStatus)
	if err != nil {
		return nil, utils.TestDbErr(err)
	}
	return status, nil
}

func generatePDF(data map[int]map[string]types.StatusEnum) ([]byte, error) {
	pdf := gopdf.GoPdf{}
	pdf.Start(gopdf.Config{PageSize: gopdf.Rect{W: 210, H: 297}})

	pdf.AddPage()

	err := pdf.AddTTFFont("Roboto", "Roboto-VariableFont_wdth,wght.ttf")
	if err != nil {
		return nil, err
	}
	err = pdf.SetFont("Roboto", "", 10)
	if err != nil {
		return nil, err
	}

	yPos := 25.0
	for id, links := range data {
		if yPos > 270 {
			pdf.AddPage()
			yPos = 10.0
		}

		pdf.SetXY(10, yPos)
		pdf.Cell(nil, "ID: "+strconv.Itoa(id))
		yPos += 7

		for link, status := range links {
			if yPos > 270 {
				pdf.AddPage()
				yPos = 10.0
			}

			pdf.SetXY(15, yPos)
			pdf.Cell(nil, "- "+link+" - "+string(status))
			yPos += 7
		}

		yPos += 3
	}

	var buf gopdf.Buff
	pdf.WriteTo(&buf)
	return buf.Bytes(), nil
}

func worker(
	ctx context.Context,
	wg *sync.WaitGroup,
	jobsChan chan string,
	results map[string]types.StatusEnum,
	mu *sync.Mutex,
) {
	defer wg.Done()

	for link := range jobsChan {
		status := checkLink(ctx, link)

		mu.Lock()
		results[link] = status
		mu.Unlock()
	}
}

func checkLink(ctx context.Context, link string) types.StatusEnum {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	if !strings.Contains(link, "://") {
		link = "https://" + link
	}

	url, err := url.Parse(link)
	if err != nil {
		return types.NotAvailable
	}

	fmt.Println(url.String())

	req, err := http.NewRequestWithContext(ctx, "GET", url.String(), nil)
	if err != nil {
		return types.NotAvailable
	}

	resp, err := client.Do(req)
	if err != nil {
		return types.NotAvailable
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return types.Available
	}

	return types.NotAvailable
}
