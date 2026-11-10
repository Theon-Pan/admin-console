# Auth Examples

JWT authentication examples using three popular Go web frameworks.

## Examples

| Framework | Port | File |
|-----------|------|------|
| Fiber     | 3001 | `fiber/main.go` |
| Gin       | 3002 | `gin/main.go` |
| Chi       | 3003 | `chi/main.go` |

## Running

```bash
# Run Fiber example
go run ./examples/auth/fiber

# Run Gin example (in another terminal)
go run ./examples/auth/gin

# Run Chi example (in another terminal)
go run ./examples/auth/chi
```

## Testing

### 1. Login to get JWT token

```bash
curl -X POST http://localhost:3001/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"password"}'
```

Response:
```json
{"token":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."}
```

### 2. Access protected route with token

```bash
# Get profile
curl http://localhost:3001/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# Get data
curl http://localhost:3001/api/data \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

### 3. Without token (should return 401)

```bash
curl http://localhost:3001/api/profile
# {"error":"missing authorization header"}
```

## Key Differences

| Feature | Fiber | Gin | Chi |
|---------|-------|-----|-----|
| Performance | Fastest (fasthttp) | Fast (net/http) | Standard (net/http) |
| Middleware | `c.Locals()` | `c.Set()` / `c.Get()` | `context.Context` |
| Routing | Express-like | Smart routing | Standard library style |
| Context | Custom `*fiber.Ctx` | Custom `*gin.Context` | Standard `http.Request` |

## Demo Credentials

- Username: `admin`
- Password: `password`
