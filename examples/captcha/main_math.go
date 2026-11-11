package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/mojocn/base64Captcha"
)

// ============ Configuration ============

const (
	ServerPort = ":8002"
)

// ============ Models ============

type CaptchaResponse struct {
	Msg            string `json:"msg"`
	Img            string `json:"img"`
	Code           int    `json:"code"`
	CaptchaEnabled bool   `json:"captchaEnabled"`
	UUID           string `json:"uuid"`
}

type LoginRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Code    int    `json:"code"`
}

// ============ Captcha Store ============

// Use base64Captcha's built-in store
var captchaStore = base64Captcha.DefaultMemStore

// ============ Captcha Drivers ============

// Math driver - generates math expressions like "3+7=?"
func createMathDriver() *base64Captcha.DriverMath {
	return &base64Captcha.DriverMath{
		Height:          40,
		Width:           120,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4, // Show lines
		Length:          2,     // 2 numbers in expression
	}
}

// Digit driver - generates random numbers
func createDigitDriver() *base64Captcha.DriverDigit {
	return &base64Captcha.DriverDigit{
		Height:   40,
		Width:    120,
		Length:   4,
		MaxSkew:  0.5,
		DotCount: 3,
	}
}

// String driver - generates random strings
func createStringDriver() *base64Captcha.DriverString {
	return &base64Captcha.DriverString{
		Height:          40,
		Width:           120,
		NoiseCount:      0,
		ShowLineOptions: 2 | 4,
		Length:          4,
		Source:          "abcdefghjkmnpqrstuvwxyz23456789",
	}
}

// ============ Handlers ============

func captchaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Choose driver type (math for this example)
	driver := createMathDriver()

	// Create captcha
	captcha := base64Captcha.NewCaptcha(driver, captchaStore)

	// Generate
	id, b64s, err := captcha.Generate()
	if err != nil {
		log.Printf("Error generating captcha: %v", err)
		http.Error(w, "Failed to generate captcha", http.StatusInternalServerError)
		return
	}

	// For math captcha, the answer is stored in the store
	// The "id" is the key to verify later

	resp := CaptchaResponse{
		Msg:            "操作成功",
		Img:            b64s,
		Code:           200,
		CaptchaEnabled: true,
		UUID:           id,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := LoginResponse{
			Success: false,
			Message: "Invalid request body",
			Code:    400,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Verify captcha
	// For math captcha, user should input the answer (e.g., "10" for "3+7=?")
	if !captchaStore.Verify(req.CaptchaID, req.CaptchaCode, true) {
		resp := LoginResponse{
			Success: false,
			Message: "验证码错误或已过期",
			Code:    400,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Verify credentials (mock)
	if req.Username == "admin" && req.Password == "admin123" {
		resp := LoginResponse{
			Success: true,
			Message: "登录成功",
			Token:   "mock-jwt-token-" + req.Username,
			Code:    200,
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	} else {
		resp := LoginResponse{
			Success: false,
			Message: "用户名或密码错误",
			Code:    401,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(resp)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/" {
		http.ServeFile(w, r, "examples/captcha/index_math.html")
		return
	}
	http.NotFound(w, r)
}

// ============ Main ============

func main() {
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/api/captchaImage", captchaHandler)
	http.HandleFunc("/user/login", loginHandler)

	log.Printf("Server starting on http://localhost%s", ServerPort)
	log.Printf("Open http://localhost%s in your browser", ServerPort)
	log.Printf("Captcha type: Math expressions (e.g., 3+7=?)")

	if err := http.ListenAndServe(ServerPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
