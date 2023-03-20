package api

type ApiResponseStatus string

const (
	ApiResponseStatusOk ApiResponseStatus = "OK"
)

type ApiResponse struct {
	Status ApiResponseStatus `json:"status"`
	Data   any               `json:"data"`
}

func NewApiResponse(status ApiResponseStatus, data any) ApiResponse {
	return ApiResponse{
		Status: status,
		Data:   data,
	}
}
