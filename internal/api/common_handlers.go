package api // import "admin-console/internal/api"
import (
	"admin-console/internal/http/response"
	"fmt"
	"net/http"
	"strings"

	"github.com/mojocn/base64Captcha"
)

func (h *handler) captchaImg(w http.ResponseWriter, r *http.Request) {
	// Configure the math driver.
	driver := &base64Captcha.DriverMath{
		Height:          100,
		Width:           250,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowSlimeLine,
	}

	// Generate a math question (e.g., "3+7=?").
	id, question, answer := driver.GenerateIdQuestionAnswer()

	// Draw the captcha image.
	item, err := driver.DrawCaptcha(question)
	if err != nil {
		apiResult := response.NewErrorResultWithMessage(err.Error())
		response.JSON(w, r, apiResult)
		return
	}

	// TODO: Save the answer to a temporary store, like Redis/Valkey
	h.store.StoreCaptchaCode(r.Context(), id, answer)
	fmt.Printf("answer: %v\n", answer)

	// Return only the Base64 payload, without the data-URI prefix.
	b64s := item.EncodeB64string()
	if _, encoded, found := strings.Cut(b64s, ","); found {
		b64s = encoded
	}

	apiResult := response.NewSuccessResult().
		SetExtra("uuid", id).
		SetExtra("img", b64s)

	response.JSON(w, r, apiResult)
}
