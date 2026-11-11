# Auth Example

Complete JWT authentication example with CORS, role-based authorization, and middleware chaining.

## Features

- ✅ JWT-based stateless authentication
- ✅ CORS with credentials support
- ✅ Role-based authorization (admin/user)
- ✅ Middleware chaining
- ✅ Login endpoint
- ✅ Public, user, and admin endpoints

## Installation

```bash
cd examples
go mod tidy
```

## Run

```bash
go run auth_example.go
```

Server starts on `http://localhost:8080`

## Test Credentials

| Username | Password | Role  |
|----------|----------|-------|
| admin    | admin123 | admin |
| user     | user123  | user  |

## API Endpoints

### Public Endpoints (no auth required)

```bash
# Login
POST /api/login
{
  "username": "admin",
  "password": "admin123"
}

# Public endpoint
GET /api/public
```

### Protected Endpoints (require authentication)

```bash
# User endpoint (any authenticated user)
GET /api/user
Authorization: Bearer <token>
```

### Admin Endpoints (require admin role)

```bash
# Admin endpoint (admin only)
GET /api/admin
Authorization: Bearer <token>
```

## Quick Test

```bash
chmod +x test_auth.sh
./test_auth.sh
```

## Manual Testing

```bash
# 1. Login and get token
curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}'

# 2. Use token to access protected endpoint
curl http://localhost:8080/api/user \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"

# 3. Try admin endpoint with user token (should fail)
curl http://localhost:8080/api/admin \
  -H "Authorization: Bearer USER_TOKEN_HERE"
```

## Middleware Chain

```
Request
  ↓
LoggingMiddleware
  ↓
CORSMiddleware
  ↓
Router (mux)
  ↓
AuthMiddleware (validates JWT)
  ↓
RequireRole (checks role)
  ↓
Handler
```

## Customization

### Change JWT Secret

```go
const JWTSecret = "your-new-secret-key"
```

### Add More Roles

```go
// In RequireRole middleware, check multiple roles
func RequireRoles(roles ...string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            claims := r.Context().Value("claims").(*Claims)
            
            for _, role := range roles {
                if claims.Role == role {
                    next.ServeHTTP(w, r)
                    return
                }
            }
            
            http.Error(w, "Forbidden", http.StatusForbidden)
        })
    }
}

// Usage
mux.Handle("/api/manager", AuthMiddleware(RequireRoles("admin", "manager")(handler)))
```

### Change Allowed Origin

```go
const AllowedOrigin = "https://your-frontend.com"
```

### Allow Multiple Origins

```go
var allowedOrigins = map[string]bool{
    "http://localhost:3000": true,
    "https://admin.yoursite.com": true,
}

// In CORSMiddleware
origin := r.Header.Get("Origin")
if allowedOrigins[origin] {
    w.Header().Set("Access-Control-Allow-Origin", origin)
}
```
