#!/bin/bash

# Test script for Auth Example
BASE_URL="http://localhost:8080"

echo "========================================="
echo "Testing Auth Example"
echo "========================================="
echo

# Test 1: Public endpoint
echo "1. Testing public endpoint..."
curl -s "$BASE_URL/api/public" | jq .
echo
echo

# Test 2: Login as admin
echo "2. Login as admin..."
ADMIN_RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}')
echo "$ADMIN_RESPONSE" | jq .
ADMIN_TOKEN=$(echo "$ADMIN_RESPONSE" | jq -r '.token')
echo
echo

# Test 3: Login as regular user
echo "3. Login as regular user..."
USER_RESPONSE=$(curl -s -X POST "$BASE_URL/api/login" \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"user123"}')
echo "$USER_RESPONSE" | jq .
USER_TOKEN=$(echo "$USER_RESPONSE" | jq -r '.token')
echo
echo

# Test 4: Access user endpoint with admin token
echo "4. Access /api/user with admin token..."
curl -s "$BASE_URL/api/user" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo
echo

# Test 5: Access admin endpoint with admin token (should succeed)
echo "5. Access /api/admin with admin token (should succeed)..."
curl -s "$BASE_URL/api/admin" \
  -H "Authorization: Bearer $ADMIN_TOKEN" | jq .
echo
echo

# Test 6: Access admin endpoint with user token (should fail)
echo "6. Access /api/admin with user token (should fail with 403)..."
curl -s "$BASE_URL/api/admin" \
  -H "Authorization: Bearer $USER_TOKEN" | jq .
echo
echo

# Test 7: Access protected endpoint without token (should fail)
echo "7. Access /api/user without token (should fail with 401)..."
curl -s "$BASE_URL/api/user" | jq .
echo
echo

# Test 8: Test CORS preflight
echo "8. Testing CORS preflight (OPTIONS request)..."
curl -s -I -X OPTIONS "$BASE_URL/api/user" \
  -H "Origin: http://localhost:3000" \
  -H "Access-Control-Request-Method: GET"
echo

echo "========================================="
echo "Tests complete!"
echo "========================================="
