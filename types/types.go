package types

type StatusEnum string

var (
	Available    StatusEnum = "available"
	NotAvailable StatusEnum = "not available"
)

type StatusReq struct {
	Links []string `json:"links"`
}

type StatusResp struct {
	Links    map[string]StatusEnum `json:"links"`
	LinksNum int                   `json:"links_num"`
}

type PdfReq struct {
	LinksList []int `json:"links_list"`
}

type ErrorRes struct {
	Error  string            `json:"error"`
	Fields map[string]string `json:"fields,omitempty"`
}

type MessageRes struct {
	Message string `json:"message"`
}
