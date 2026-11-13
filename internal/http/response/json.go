package response // import "admin-console/internal/http/response"
import (
	"admin-console/internal/http/request"
	"encoding/json"
	"log/slog"
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

	NewBuilder(w, r).
		WithHeader("Content-Type", jsonContentTypeHeader).
		WithBodyAsBytes(responseBody).
		Write()
}

// JSONServerError sends an internal error to the client.
func JSONServerError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error(http.StatusText(http.StatusInternalServerError),
		slog.Any("error", err),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusInternalServerError),
		),
	)

	NewBuilder(w, r).
		WithStatus(http.StatusOK).
		WithHeader("Content-Type", jsonContentTypeHeader).
		WithBodyAsBytes(generateJSONError(err)).
		Write()
}

// JSONBadRequest sends a bad request error to the client.
func JSONBadRequest(w http.ResponseWriter, r *http.Request, err error) {
	slog.Warn(http.StatusText(http.StatusBadRequest),
		slog.Any("error", err),
		slog.String("client_ip", request.ClientIP(r)),
		slog.Group("request",
			slog.String("method", r.Method),
			slog.String("uri", r.RequestURI),
			slog.String("user_agent", r.UserAgent()),
		),
		slog.Group("response",
			slog.Int("status_code", http.StatusBadRequest),
		),
	)

	NewBuilder(w, r).
		WithStatus(http.StatusBadRequest).
		WithHeader("Content-Type", jsonContentTypeHeader).
		WithBodyAsBytes(gernerateJSONErrorWithStatusCode(BadRequest, err)).
		Write()
}

func generateJSONError(err error) []byte {
	result := NewErrorResultWithMessage(err.Error())
	encodeBody, _ := json.Marshal(result.data)

	return encodeBody
}

func gernerateJSONErrorWithStatusCode(code StatusCode, err error) []byte {
	result := NewErrorResultWithCodeAndMessage(code, err.Error())
	encodeBody, _ := json.Marshal(result.data)

	return encodeBody
}
