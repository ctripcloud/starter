package dto

//Base base dto will be embeded into other dtos
type Base struct {
	Code    int
	Message string
	Success bool
	Data    interface{} `json:"Data,omitempty"`
}

//NewSuccess wrapper for filling success response
func NewSuccess(code int, message string, data interface{}) *Base {
	return &Base{
		Code:    code,
		Message: message,
		Success: true,
		Data:    data,
	}
}

//Success wrapper for filling success response
func Success(b *Base, code int, message string, data interface{}) {
	b.Code = code
	b.Data = data
	b.Message = message
	b.Success = true
}

//NewSuccessOK wrapper for filling 200 response
func NewSuccessOK(data interface{}) *Base {
	return NewSuccess(200, "OK", data)
}

//SuccessOK wrapper for filling 200 response
func SuccessOK(b *Base, data interface{}) {
	Success(b, 200, "OK", data)
}
