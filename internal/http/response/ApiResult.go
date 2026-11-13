package response // import "admin-console/internal/http/response"

type StatusCode int
type Tag string
type ApiResultData map[Tag]any

const (
	// Status codes
	Success             StatusCode = 200
	BadRequest          StatusCode = 400
	Unauthorized        StatusCode = 401
	Forbidden           StatusCode = 403
	NotFound            StatusCode = 404
	ServerInternalError StatusCode = 500
)

const (
	// Tag Types for API response fields
	code_tag    Tag = "code"
	message_tag Tag = "msg"
	data_tag    Tag = "data"
)

type ApiResult struct {
	data ApiResultData
}

func NewResult(code StatusCode, message string) *ApiResult {
	data := make(ApiResultData)
	data[code_tag] = code
	data[message_tag] = message

	return &ApiResult{data: data}
}

func NewResultWithData(code StatusCode, message string, resultData any) *ApiResult {
	data := make(ApiResultData)
	data[code_tag] = code
	data[message_tag] = message
	if resultData != nil {
		data[data_tag] = resultData
	}

	return &ApiResult{data: data}
}

func NewSuccessResult() *ApiResult {
	return NewSucceesResultWithMessage("success")
}

func NewSucceesResultWithMessage(message string) *ApiResult {
	data := make(ApiResultData)
	data[code_tag] = Success
	data[message_tag] = message

	return &ApiResult{data: data}
}

func NewErrorResult() *ApiResult {
	return NewSucceesResultWithMessage("")
}

func NewErrorResultWithMessage(message string) *ApiResult {
	data := make(ApiResultData)
	data[code_tag] = ServerInternalError
	data[message_tag] = message

	return &ApiResult{data: data}
}

func NewErrorResultWithCodeAndMessage(code StatusCode, message string) *ApiResult {
	data := make(ApiResultData)
	data[code_tag] = code
	data[message_tag] = message

	return &ApiResult{data: data}
}

func (r *ApiResult) SetExtra(key string, value any) *ApiResult {
	if r.data == nil {
		r.data = make(ApiResultData)
	}
	r.data[Tag(key)] = value

	return r
}

func (r *ApiResult) GetData() ApiResultData {
	return r.data
}
