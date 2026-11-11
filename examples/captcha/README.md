# Captcha Login Example

Complete captcha implementation with image generation, validation, and login flow.

## Features

- ✅ Image captcha generation (no external dependencies)
- ✅ In-memory captcha storage with expiration
- ✅ One-time use captcha validation
- ✅ Beautiful login UI
- ✅ Click-to-refresh captcha
- ✅ Mock JWT token generation

## Run

```bash
cd examples/captcha
go run main.go
```

Server starts on `http://localhost:8000`

## Test

Open `http://localhost:8000` in your browser.

**Test Credentials:**
- Username: `admin`
- Password: `admin123`

**Note:** The captcha code is logged to the server console for demo purposes.

## API Endpoints

### Get Captcha

```http
GET /captcha
```

Response:
```json
{
  "captcha_id": "abc123...",
  "image_url": "/captcha/image/abc123..."
}
```

### Get Captcha Image

```http
GET /captcha/image/{captcha_id}
```

Returns PNG image.

### Login

```http
POST /user/login
Content-Type: application/json

{
  "username": "admin",
  "password": "admin123",
  "captcha_id": "abc123...",
  "captcha_code": "a2b3"
}
```

Response (success):
```json
{
  "success": true,
  "message": "Login successful",
  "token": "mock-jwt-token-admin"
}
```

Response (failure):
```json
{
  "success": false,
  "message": "Invalid or expired captcha"
}
```

## How It Works

### 1. Captcha Generation

```go
// Generate random 4-character code
code := generateCaptchaCode() // e.g., "a3x9"

// Store in memory with expiration
captchaStore.Set(captchaID, code)

// Generate image with noise
imgBytes := generateCaptchaImage(code)
```

### 2. Captcha Storage

```go
type CaptchaStore struct {
    store map[string]CaptchaEntry
}

type CaptchaEntry struct {
    Code      string
    ExpiresAt time.Time
}
```

- Automatic cleanup of expired captchas
- One-time use (deleted after verification)
- Thread-safe with mutex

### 3. Validation

```go
func (s *CaptchaStore) Verify(id, code string) bool {
    entry, exists := s.store[id]
    if !exists {
        return false
    }
    
    defer delete(s.store, id) // One-time use
    
    if time.Now().After(entry.ExpiresAt) {
        return false // Expired
    }
    
    return entry.Code == code
}
```

### 4. Image Generation

- Creates RGBA image (120x40 pixels)
- Adds random noise dots
- Draws characters using pixel patterns
- Encodes to PNG

## Customization

### Change Captcha Length

```go
const CaptchaLength = 6 // Default is 4
```

### Change Expiration Time

```go
const CaptchaExpireMin = 10 // Default is 5 minutes
```

### Change Image Size

```go
const CaptchaWidth = 150
const CaptchaHeight = 50
```

## Production Enhancements

For production use, consider:

1. **Use real fonts** - Integrate `github.com/golang/freetype` with TTF fonts
2. **Add distortion** - Wave effects, rotation, scaling
3. **Color variation** - Random colors per character
4. **Line noise** - Add random lines through text
5. **Redis storage** - For distributed systems
6. **Rate limiting** - Prevent brute force attacks
7. **Don't log codes** - Remove the demo logging

## Example with Redis

```go
import "github.com/go-redis/redis/v8"

type RedisCaptchaStore struct {
    client *redis.Client
    ctx    context.Context
}

func (s *RedisCaptchaStore) Set(id, code string) {
    s.client.Set(s.ctx, "captcha:"+id, code, 5*time.Minute)
}

func (s *RedisCaptchaStore) Verify(id, code string) bool {
    key := "captcha:" + id
    stored, err := s.client.Get(s.ctx, key).Result()
    if err != nil {
        return false
    }
    s.client.Del(s.ctx, key) // One-time use
    return stored == code
}
```
