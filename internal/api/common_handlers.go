package api // import "admin-console/internal/api"
import (
	"admin-console/internal/http/response"
	"fmt"
	"net/http"

	"github.com/mojocn/base64Captcha"
)

func (h *handler) captchaImg(w http.ResponseWriter, r *http.Request) {
	// Configure the math driver.
	driver := &base64Captcha.DriverMath{
		Height:          50,
		Width:           200,
		NoiseCount:      0,
		ShowLineOptions: base64Captcha.OptionShowSlimeLine | base64Captcha.OptionShowSineLine,
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
	fmt.Printf("answer: %v\n", answer)

	// Get base 64 data-URI string (e.g. "data:image/png;base64, iVBOR...")
	b64s := item.EncodeB64string()

	apiResult := response.NewSuccessResult().
		SetExtra("uuid", id).
		SetExtra("img", b64s)

	response.JSON(w, r, apiResult)
}
