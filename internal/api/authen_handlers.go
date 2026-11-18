package api // import "admin-console/internal/api"
import (
	"admin-console/internal/http/request"
	"admin-console/internal/http/response"
	"admin-console/internal/locale"
	"admin-console/internal/model"
	"admin-console/internal/utils"
	json_parser "encoding/json"
	"log/slog"
	"net/http"
	"time"
)

func (h *handler) Login(w http.ResponseWriter, r *http.Request) {
	var loginRequest model.LoginRequest

	// Decode the json data in request body.
	if err := json_parser.NewDecoder(r.Body).Decode(&loginRequest); err != nil {
		response.JSONBadRequest(w, r, err)
		return
	}

	ip := request.ClientIP(r)
	userAgent := r.UserAgent()

	loginInfoRequest := &model.LoginInfoCreationRequest{
		Username: loginRequest.Username,
	}
	if ip != "" {
		loginInfoRequest.IP = &ip
	}
	loginInfoRequest.Browser = &userAgent

	// Check captcha code.
	matched, err := h.store.VerifyCaptchaCode(r.Context(), loginRequest.Uuid, loginRequest.Code)
	if err != nil {
		apiResult := response.NewErrorResultWithMessage(locale.NewLocalizedError("error.api.invalid_captcha_code").Translate(request.GetLocale(r)))
		apiResult.SetError(&err)
		response.JSON(w, r, apiResult)
		return
	}
	if !matched {
		errorMessage := locale.NewLocalizedError("error.api.invalid_captcha_code").Translate(request.GetLocale(r))
		apiResult := response.NewErrorResultWithMessage(errorMessage)
		response.JSON(w, r, apiResult)
		// Recording login information asynchronously.
		login_at := time.Now().UTC()
		loginInfoRequest.Status = model.LoginFailed
		loginInfoRequest.Message = &errorMessage
		loginInfoRequest.LoginAt = &login_at
		h.EnqueueJob(loginInfoRequest)
		return
	}

	// Check password.
	if err := h.store.CheckPassword(loginRequest.Username, loginRequest.Password); err != nil {
		// slog.Warn("Incorrect username or password",
		// 	slog.Bool("authenication_failed", true),
		// 	slog.String("client_ip", clientIP),
		// 	slog.String("user_agent", r.UserAgent()),
		// 	slog.String("username", loginRequest.Username),
		// 	slog.Any("error", err),
		// )
		errorMessage := locale.NewLocalizedError("error.api.incorrect_username_or_password").Translate(request.GetLocale(r))
		apiResult := response.NewErrorResultWithMessage(errorMessage)
		apiResult.SetError(&err)
		response.JSON(w, r, apiResult)
		// Recording login information asynchronously.
		login_at := time.Now().UTC()
		loginInfoRequest.Status = model.LoginFailed
		loginInfoRequest.Message = &errorMessage
		loginInfoRequest.LoginAt = &login_at
		h.EnqueueJob(loginInfoRequest)
		return
	}

	// Retrieve the user detail info.
	user, err := h.store.UserByUsername(loginRequest.Username)
	if err != nil {
		// slog.Warn("Error When retrieving user info by username",
		// 	slog.String("username", loginRequest.Username),
		// 	slog.Any("error", err),
		// )
		apiResult := response.NewErrorResultWithMessage(locale.NewLocalizedError("error.api.incorrect_username_or_password").Translate(request.GetLocale(r)))
		apiResult.SetError(&err)
		response.JSON(w, r, apiResult)
		return
	}

	if user == nil {
		// slog.Warn("No user found",
		// 	slog.String("username", loginRequest.Username),
		// )
		apiResult := response.NewErrorResultWithMessage(locale.NewLocalizedError("error.api.incorrect_username_or_password").Translate(request.GetLocale(r)))
		response.JSON(w, r, apiResult)
		return
	}

	loginInfoRequest.UserID = user.UserID

	// Generate a token.
	uuid := utils.UUIDNewString()

	// Save login user info for verification in future.
	loginUser := model.NewLoginUser(user, uuid)
	if err := h.store.StoreLoginUser(r.Context(), *loginUser); err != nil {
		apiResult := response.NewErrorResultWithMessage(locale.NewLocalizedError("error.api.internal_server_error").Translate(request.GetLocale(r)))
		apiResult.SetError(&err)
		response.JSON(w, r, apiResult)
		return
	}

	tokenClaims := utils.TokenClaims{
		UUID: uuid,
	}

	token, err := utils.GenerateToken(tokenClaims)
	if err != nil {
		apiResult := response.NewErrorResultWithMessage(locale.NewLocalizedError("error.api.internal_server_error").Translate(request.GetLocale(r)))
		apiResult.SetError(&err)
		response.JSON(w, r, apiResult)
		return
	}

	tokenClaimsForValidation, err := utils.ValidateToken(token)
	if err != nil {
		apiResult := response.NewErrorResultWithMessage(locale.NewLocalizedError("error.api.internal_server_error").Translate(request.GetLocale(r)))
		response.JSON(w, r, apiResult)
		return
	}

	slog.Debug("The token including:",
		slog.String("user_id", tokenClaimsForValidation.UserID),
		slog.String("username", tokenClaimsForValidation.UserName),
		slog.String("uuid", tokenClaimsForValidation.UUID),
		slog.String("Subject", tokenClaimsForValidation.Subject),
		slog.String("Issuer", tokenClaimsForValidation.Issuer),
		slog.String("ID", tokenClaimsForValidation.ID))

	successMessage := locale.NewLocalizedError("api.success").Translate(request.GetLocale(r))
	// Recording login information asynchronously.
	login_at := time.Now().UTC()
	loginInfoRequest.Status = model.LoginSuccess
	loginInfoRequest.Message = &successMessage
	loginInfoRequest.LoginAt = &login_at
	go h.EnqueueJob(loginInfoRequest)

	apiResult := response.NewSucceesResultWithMessage(successMessage)
	apiResult.SetExtra("token", token)
	response.JSON(w, r, apiResult)
}
