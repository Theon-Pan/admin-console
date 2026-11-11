# Math Captcha Example

Complete math-based captcha implementation using **base64Captcha** library.

## Features

- ✅ **Math expressions** (e.g., "3+7=?", "9-4=?")
- ✅ **Digit captchas** (random numbers)
- ✅ **String captchas** (random characters)
- ✅ **Chinese character support**
- ✅ **Built-in verification**
- ✅ **Base64 image output**
- ✅ **In-memory storage**

## Installation

```bash
cd examples/captcha-math
go mod tidy
```

## Run

```bash
go run main.go
```

Server starts on `http://localhost:8002`

## Test

Open `http://localhost:8002` in your browser.

**Test Credentials:**
- Username: `admin`
- Password: `admin123`

**How to solve:**
- Look at the math expression (e.g., "3+7=?")
- Enter the answer (e.g., "10")
- Click "Login"

## Captcha Types

### 1. Math Captcha (Default)

```go
driver := &base64Captcha.DriverMath{
    Height:          40,
    Width:           120,
    NoiseCount:      0,
    ShowLineOptions: 2 | 4,
    Length:          2,  // 2 numbers in expression
}
```

Generates: `3+7=?`, `9-4=?`, `2×5=?`

### 2. Digit Captcha

```go
driver := &base64Captcha.DriverDigit{
    Height:   40,
    Width:    120,
    Length:   4,
    MaxSkew:  0.5,
    DotCount: 3,
}
```

Generates: `a7k9`, `3m2x`

### 3. String Captcha

```go
driver := &base64Captcha.DriverString{
    Height:          40,
    Width:           120,
    NoiseCount:      0,
    ShowLineOptions: 2 | 4,
    Length:          4,
    Source:          "abcdefghjkmnpqrstuvwxyz23456789",
}
```

Generates: `hjkm`, `2345`

### 4. Chinese Captcha

```go
driver := &base64Captcha.DriverChinese{
    Height:          40,
    Width:           120,
    NoiseCount:      0,
    ShowLineOptions: 2 | 4,
    Length:          3,
    Source:          "们以我到他会作时要动国产的一是工就已在",
}
```

Generates: Chinese characters

## API Endpoints

### Get Captcha

```http
GET /api/captchaImage
```

Response:
```json
{
  "msg": "操作成功",
  "img": "iVBORw0KGgoAAAANSUhEUgAA...",
  "code": 200,
  "captchaEnabled": true,
  "uuid": "abc123..."
}
```

### Login

```http
POST /user/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "captcha_id": "abc123...",
  "captcha_code": "10"
}
```

## How It Works

### 1. Generate Captcha

```go
driver := createMathDriver()
captcha := base64Captcha.NewCaptcha(driver, captchaStore)
id, b64s, err := captcha.Generate()
```

- Creates math expression (e.g., "3+7=?")
- Generates image
- Stores answer in memory with UUID
- Returns base64 image + UUID

### 2. Verify Answer

```go
if !captchaStore.Verify(req.CaptchaID, req.CaptchaCode, true) {
    // Invalid or expired
}
```

- Checks UUID exists
- Compares answer (case-insensitive)
- Deletes after verification (one-time use)
- Checks expiration

## Customization

### Change Image Size

```go
driver := &base64Captcha.DriverMath{
    Height: 60,   // Increase height
    Width:  160,  // Increase width
}
```

### Add Noise

```go
driver := &base64Captcha.DriverMath{
    NoiseCount: 5,  // Add 5 noise dots
}
```

### Change Operations

Math driver automatically uses: `+`, `-`, `×`

### Set Expiration

```go
// In base64Captcha source, default is 5 minutes
// Or use custom store with your own expiration
```

## Production Enhancements

### Use Redis Store

```go
import "github.com/mojocn/base64Captcha"

store := base64Captcha.DefaultMemStore

// Or implement custom Redis store
type RedisStore struct {
    client *redis.Client
}

func (s *RedisStore) Set(id string, value string) error {
    return s.client.Set(context.Background(), "captcha:"+id, value, 5*time.Minute).Err()
}

func (s *RedisStore) Get(id string, clear bool) string {
    val, _ := s.client.Get(context.Background(), "captcha:"+id).Result()
    if clear {
        s.client.Del(context.Background(), "captcha:"+id)
    }
    return val
}

func (s *RedisStore) Verify(id, answer string, clear bool) bool {
    val := s.Get(id, clear)
    return val == answer
}
```

### Add Rate Limiting

```go
// Limit captcha generation per IP
rateLimiter := make(map[string]int)

func captchaHandler(w http.ResponseWriter, r *http.Request) {
    ip := r.RemoteAddr
    if rateLimiter[ip] > 10 {
        http.Error(w, "Too many requests", http.StatusTooManyRequests)
        return
    }
    rateLimiter[ip]++
    // ... generate captcha
}
```

## Comparison: base64Captcha vs Custom

| Feature | base64Captcha | Custom (previous example) |
|---------|--------------|---------------------------|
| Math expressions | ✅ Built-in | ❌ Manual |
| Image quality | ✅ Professional | ⚠️ Basic |
| Fonts | ✅ TTF support | ❌ Pixel art |
| Verification | ✅ Built-in | ⚠️ Manual |
| Multiple types | ✅ 4+ types | ❌ 1 type |
| Maintenance | ✅ Active | ⚠️ Your code |

## Other Go Captcha Libraries

| Library | Stars | Features |
|---------|-------|----------|
| **base64Captcha** | 1.5k+ | Math, digit, string, Chinese |
| **captcha** | 800+ | Simple, image-based |
| **go-captcha** | 500+ | Click-based, slider |

## Summary

**Use base64Captcha when:**
- ✅ You need multiple captcha types
- ✅ You want professional-looking images
- ✅ You need built-in verification
- ✅ You want active maintenance

**Use custom implementation when:**
- ✅ You need full control
- ✅ You have specific requirements
- ✅ You want zero dependencies
