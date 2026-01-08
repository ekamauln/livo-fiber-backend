# API Documentation Setup

## What Was Fixed

### 1. Auth Controller Fixes
- ✅ Fixed type mismatch: `user.ID` (uint) → `string` conversion for TokenClaims
- ✅ Fixed context key from `user_id` to `userId` (camelCase consistency)
- ✅ Changed JSON request body from `refresh_token` to `refreshToken` (camelCase)
- ✅ Added `fmt` import for string conversion

### 2. Token Utils Fixes
- ✅ Changed TokenClaims JSON tags to camelCase (`userId`, not `user_id`)
- ✅ Fixed token claim keys from `user_id` to `userId` in both access and refresh tokens
- ✅ Now token creation and validation use the same keys

### 3. Swagger/OpenAPI Documentation
- ✅ Added Swaggo annotations to all auth controller methods
- ✅ Configured main.go with API metadata
- ✅ Added swagger endpoint at `/docs/*`
- ✅ Added RapiDoc UI at `/rapidoc`

## Generate Swagger Documentation

After making changes to controller annotations, regenerate the docs:

```bash
swag init
```

This will create/update:
- `docs/docs.go`
- `docs/swagger.json`
- `docs/swagger.yaml`

## API Documentation URLs

Once your server is running:

- **Swagger UI**: http://localhost:3000/docs/index.html
- **RapiDoc UI**: http://localhost:3000/rapidoc (Modern, dark-themed UI)
- **Swagger JSON**: http://localhost:3000/docs/swagger.json
- **Swagger YAML**: http://localhost:3000/docs/swagger.yaml

## Testing Authentication

### 1. Register a User
```bash
POST http://localhost:3000/api/v1/auth/register
Content-Type: application/json

{
  "username": "testuser",
  "password": "SecurePass123",
  "fullName": "Test User",
  "email": "test@example.com"
}
```

### 2. Login
```bash
POST http://localhost:3000/api/v1/auth/login
Content-Type: application/json

{
  "username": "testuser",
  "password": "SecurePass123"
}
```

Response will include `accessToken` and user data.

### 3. Use Protected Endpoints
```bash
GET http://localhost:3000/api/v1/users
Authorization: Bearer <your_access_token>
```

## RapiDoc Features

The RapiDoc UI provides:
- 🎨 Modern dark theme
- 🔐 Built-in authentication testing
- 📝 Interactive API testing
- 🌳 Schema visualization in tree format
- 📱 Responsive design
- ⚡ Fast and lightweight

## Next Steps

To add documentation to other controllers:
1. Add swaggo annotations above each handler function
2. Run `swag init` to regenerate docs
3. Refresh the documentation page

Example annotation structure:
```go
// @Summary Short description
// @Description Detailed description
// @Tags Category
// @Accept json
// @Produce json
// @Param paramName paramType dataType required "description"
// @Success 200 {object} ResponseType
// @Failure 400 {object} ErrorType
// @Router /api/path [method]
```
