package response // import "admin-console/internal/http/response"
import (
	"admin-console/internal/http/request"
	"admin-console/internal/locale"
	"encoding/json"
	"errors"
	"net/http"
)

const jsonContentTypeHeader = `application/json`

// JSON creates a new JSON response with a 200 status code.
func JSON(w http.ResponseWriter, r *http.Request, body any) {
	var bodyToMarshal any
	if result, ok := body.(*ApiResult); ok {
		bodyToMarshal = result.GetData()
	} else {
		bodyToMarshal = body
	}
	responseBody, err := json.Marshal(bodyToMarshal)
	if err != nil {
		JSONServerError(w, r, err)
		return
	}
	LogInfo(r, http.StatusOK)

	NewBuilder(w, r).
		WithHeader("Content-Type", jsonContentTypeHeader).
		WithBodyAsBytes(responseBody).
		Write()
}

// JSONServerError sends an internal error to the client.
func JSONServerError(w http.ResponseWriter, r *http.Request, err error) {
	LogError(r, http.StatusInternalServerError, err)
	errorForResp := errors.New(locale.NewLocalizedError("error.api.internal_server_error").Translate(request.GetLocale(r)))
	NewBuilder(w, r).
		WithStatus(http.StatusOK).
		WithHeader("Content-Type", jsonContentTypeHeader).
		WithBodyAsBytes(generateJSONError(errorForResp)).
		Write()
}

// JSONBadRequest sends a bad request error to the client.
func JSONBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	LogWarn(r, http.StatusBadRequest, err)
	errorForResp := errors.New(locale.NewLocalizedError("error.api.bad_request").Translate(request.GetLocale(r)))
	NewBuilder(w, r).
		WithStatus(http.StatusBadRequest).
		WithHeader("Content-Type", jsonContentTypeHeader).
		WithBodyAsBytes(gernerateJSONErrorWithStatusCode(BadRequest, errorForResp)).
		Write()
}

func generateJSONError(err error) []byte {
	result := NewErrorResultWithMessage(err.Error())
	encodeBody, _ := json.Marshal(result.data)

	return encodeBody
}

func gernerateJSONErrorWithStatusCode(code StatusCode, err error) []byte {
	return generateJSONErrorWithStatusCodeAndMessage(code, err.Error())
}

func generateJSONErrorWithStatusCodeAndMessage(code StatusCode, message string) []byte {
	result := NewErrorResultWithCodeAndMessage(code, message)
	encodeBody, _ := json.Marshal(result.data)

	return encodeBody
}
