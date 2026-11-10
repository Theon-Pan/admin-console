package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/color"
	"image/jpeg"
	"io"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

// ============ Configuration ============

const (
	ServerPort         = ":8001"
	UpstreamCaptchaURL = "http://localhost:8000/api/captchaImage"
	CaptchaWidth       = 120
	CaptchaHeight      = 40
	CaptchaLength      = 4
	CaptchaExpireMin   = 5
)

// ============ Models ============

// Upstream captcha response (from localhost:8000)
type UpstreamCaptchaResponse struct {
	Msg            string `json:"msg"`
	Img            string `json:"img"`             // Base64 encoded image
	Code           int    `json:"code"`            // Status code
	CaptchaEnabled bool   `json:"captchaEnabled"`  // Whether captcha is enabled
	UUID           string `json:"uuid"`            // Captcha ID
}

// Our API response (matches upstream format)
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
	CaptchaID   string `json:"captcha_id"`   // UUID from captcha
	CaptchaCode string `json:"captcha_code"` // User input
	Code        string `json:"code"`         // Captcha code (for verification)
}

type LoginResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Token   string `json:"token,omitempty"`
	Code    int    `json:"code"`
}

// ============ Captcha Store ============

type CaptchaEntry struct {
	Code      string
	ExpiresAt time.Time
}

type CaptchaStore struct {
	mu    sync.RWMutex
	store map[string]CaptchaEntry
}

func NewCaptchaStore() *CaptchaStore {
	store := &CaptchaStore{
		store: make(map[string]CaptchaEntry),
	}
	go store.cleanupLoop()
	return store
}

func (s *CaptchaStore) Set(id, code string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[id] = CaptchaEntry{
		Code:      code,
		ExpiresAt: time.Now().Add(CaptchaExpireMin * time.Minute),
	}
}

func (s *CaptchaStore) Verify(id, code string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	entry, exists := s.store[id]
	if !exists {
		return false
	}

	defer delete(s.store, id) // Delete after verification (one-time use)

	if time.Now().After(entry.ExpiresAt) {
		return false // Expired
	}

	return entry.Code == code
}

func (s *CaptchaStore) cleanupLoop() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		s.mu.Lock()
		now := time.Now()
		for id, entry := range s.store {
			if now.After(entry.ExpiresAt) {
				delete(s.store, id)
			}
		}
		s.mu.Unlock()
	}
}

var captchaStore = NewCaptchaStore()

// ============ Captcha Generation (Fallback) ============

func generateCaptchaCode() string {
	const chars = "abcdefghjkmnpqrstuvwxyz23456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	code := make([]byte, CaptchaLength)
	for i := range code {
		code[i] = chars[r.Intn(len(chars))]
	}
	return string(code)
}

func generateCaptchaImage(code string) (string, error) {
	img := image.NewRGBA(image.Rect(0, 0, CaptchaWidth, CaptchaHeight))

	bgColor := color.RGBA{240, 240, 240, 255}
	for y := 0; y < CaptchaHeight; y++ {
		for x := 0; x < CaptchaWidth; x++ {
			img.Set(x, y, bgColor)
		}
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < 100; i++ {
		x := r.Intn(CaptchaWidth)
		y := r.Intn(CaptchaHeight)
		img.Set(x, y, color.RGBA{
			uint8(r.Intn(200)),
			uint8(r.Intn(200)),
			uint8(r.Intn(200)),
			255,
		})
	}

	drawText(img, code, r)

	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, img, nil); err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buf.Bytes()), nil
}

func drawText(img *image.RGBA, text string, r *rand.Rand) {
	colors := []color.RGBA{
		{255, 0, 0, 255},
		{0, 128, 0, 255},
		{0, 0, 255, 255},
		{128, 0, 128, 255},
		{255, 128, 0, 255},
	}

	xOffset := 20
	for i, char := range text {
		charColor := colors[r.Intn(len(colors))]
		yOffset := 10 + r.Intn(15)
		drawChar(img, int(char), xOffset+i*25, yOffset, charColor)
	}
}

func drawChar(img *image.RGBA, char int, x, y int, c color.RGBA) {
	patterns := map[int][]string{
		'a': {"0110", "1001", "1111", "1001", "1001"},
		'b': {"1110", "1001", "1110", "1001", "1110"},
		'c': {"0111", "1000", "1000", "1000", "0111"},
		'd': {"1110", "1001", "1001", "1001", "1110"},
		'e': {"1111", "1000", "1110", "1000", "1111"},
		'f': {"1111", "1000", "1110", "1000", "1000"},
		'g': {"0111", "1000", "1011", "1001", "0110"},
		'h': {"1001", "1001", "1111", "1001", "1001"},
		'j': {"0011", "0001", "0001", "1001", "0110"},
		'k': {"1001", "1010", "1100", "1010", "1001"},
		'm': {"10001", "11011", "10101", "10001", "10001"},
		'n': {"1001", "1101", "1011", "1001", "1001"},
		'p': {"1110", "1001", "1110", "1000", "1000"},
		'q': {"0110", "1001", "1001", "1010", "0101"},
		'r': {"1110", "1001", "1110", "1010", "1001"},
		's': {"0111", "1000", "0110", "0001", "1110"},
		't': {"1111", "0010", "0010", "0010", "0010"},
		'u': {"1001", "1001", "1001", "1001", "0110"},
		'v': {"1001", "1001", "1001", "0110", "0110"},
		'w': {"10001", "10001", "10101", "11011", "10001"},
		'x': {"1001", "0110", "0110", "0110", "1001"},
		'y': {"1001", "1001", "0110", "0010", "0100"},
		'z': {"1111", "0010", "0100", "1000", "1111"},
		'2': {"1110", "0001", "0110", "1000", "1111"},
		'3': {"1110", "0001", "0110", "0001", "1110"},
		'4': {"1001", "1001", "1111", "0001", "0001"},
		'5': {"1111", "1000", "1110", "0001", "1110"},
		'6': {"0111", "1000", "1110", "1001", "0110"},
		'7': {"1111", "0001", "0010", "0100", "0100"},
		'8': {"0110", "1001", "0110", "1001", "0110"},
		'9': {"0110", "1001", "0111", "0001", "0110"},
	}

	pattern, exists := patterns[char]
	if !exists {
		return
	}

	scale := 3
	for row, line := range pattern {
		for col, bit := range line {
			if bit == '1' {
				for dy := 0; dy < scale; dy++ {
					for dx := 0; dx < scale; dx++ {
						img.Set(x+col*scale+dx, y+row*scale+dy, c)
					}
				}
			}
		}
	}
}

// ============ Handlers ============

func captchaHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Try to fetch from upstream first
	upstreamResp, err := fetchUpstreamCaptcha()
	if err == nil {
		// Store the captcha code (for demo, we don't know the actual code)
		// In production, the upstream server validates it
		captchaStore.Set(upstreamResp.UUID, "unknown")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(upstreamResp)
		return
	}

	// Fallback: generate our own captcha
	log.Printf("Upstream captcha unavailable, generating locally: %v", err)

	captchaID := generateUUID()
	code := generateCaptchaCode()
	captchaStore.Set(captchaID, code)

	imgBase64, err := generateCaptchaImage(code)
	if err != nil {
		http.Error(w, "Failed to generate captcha", http.StatusInternalServerError)
		return
	}

	log.Printf("Generated fallback captcha: ID=%s, Code=%s", captchaID, code)

	resp := CaptchaResponse{
		Msg:            "操作成功",
		Img:            imgBase64,
		Code:           200,
		CaptchaEnabled: true,
		UUID:           captchaID,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func fetchUpstreamCaptcha() (*CaptchaResponse, error) {
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(UpstreamCaptchaURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var upstreamResp UpstreamCaptchaResponse
	if err := json.Unmarshal(body, &upstreamResp); err != nil {
		return nil, err
	}

	return &CaptchaResponse{
		Msg:            upstreamResp.Msg,
		Img:            upstreamResp.Img,
		Code:           upstreamResp.Code,
		CaptchaEnabled: upstreamResp.CaptchaEnabled,
		UUID:           upstreamResp.UUID,
	}, nil
}

func generateUUID() string {
	const chars = "abcdefghijklmnopqrstuvwxyz0123456789"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	id := make([]byte, 32)
	for i := range id {
		id[i] = chars[r.Intn(len(chars))]
	}
	return string(id)
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

	// Try to verify with upstream first (if upstream handles validation)
	// For now, we'll use local verification
	if req.CaptchaID != "" && req.CaptchaCode != "" {
		if !captchaStore.Verify(req.CaptchaID, req.CaptchaCode) {
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
		http.ServeFile(w, r, "examples/captcha/index.html")
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
	log.Printf("Upstream captcha URL: %s", UpstreamCaptchaURL)

	if err := http.ListenAndServe(ServerPort, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
