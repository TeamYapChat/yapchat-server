package utils

type SuccessResponse struct {
	Success bool        `json:"success" example:"true"`
	Message string      `json:"message" example:"success"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorResponse struct {
	Success bool   `json:"success" example:"false"`
	Message string `json:"message" example:"error message"`
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
