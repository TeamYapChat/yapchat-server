package utils

type SuccessResponse struct {
	Success bool        `json:"success"        validate:"required" example:"true"`
	Message string      `json:"message"        validate:"required" example:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success" validate:"required" example:"false"`
	Message string `json:"message" validate:"required" example:"error message"`
}

func NewSuccessResponse(data interface{}) SuccessResponse {
	return SuccessResponse{
		Success: true,
		Message: "success",
		Data:    data,
	}
}

func NewErrorResponse(message string) ErrorResponse {
	return ErrorResponse{
		Success: false,
		Message: message,
	}
}

type Pagination struct {
	Page       int `json:"page"        validate:"required" example:"1"`
	PageSize   int `json:"page_size"   validate:"required" example:"10"`
	TotalRows  int `json:"total_rows"  validate:"required" example:"100"`
	TotalPages int `json:"total_pages" validate:"required" example:"10"`
	Data       any `json:"data"        validate:"required"`
}
