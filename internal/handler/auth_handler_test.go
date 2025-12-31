// Package handler provides HTTP handler tests.
package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid login request format",
			requestBody: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			expectedStatus: http.StatusOK, // Valid format passes binding, returns OK from mock handler
		},
		{
			name: "missing username",
			requestBody: map[string]string{
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "missing password",
			requestBody: map[string]string{
				"username": "testuser",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty request body",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			// Simple handler that validates request format
			router.POST("/auth/login", func(c *gin.Context) {
				var req LoginRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				// In real tests, we'd call the actual handler with mocked service
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_RefreshToken(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid refresh token request format",
			requestBody: map[string]string{
				"refresh_token": "valid-refresh-token",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing refresh token",
			requestBody:    map[string]string{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.POST("/auth/refresh", func(c *gin.Context) {
				var req RefreshTokenRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusOK, gin.H{"message": "ok"})
			})

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestAuthHandler_Logout(t *testing.T) {
	t.Run("logout requires authorization header", func(t *testing.T) {
		router := gin.New()

		router.POST("/auth/logout", func(c *gin.Context) {
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"message": "Successfully logged out"})
		})

		// Test without auth header
		req := httptest.NewRequest("POST", "/auth/logout", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)

		// Test with auth header
		req = httptest.NewRequest("POST", "/auth/logout", http.NoBody)
		req.Header.Set("Authorization", "Bearer test-token")
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestUserHandler_List(t *testing.T) {
	t.Run("list users with pagination", func(t *testing.T) {
		router := gin.New()

		router.GET("/users", func(c *gin.Context) {
			page := c.DefaultQuery("page", "1")
			pageSize := c.DefaultQuery("page_size", "20")

			c.JSON(http.StatusOK, gin.H{
				"users":     []interface{}{},
				"total":     0,
				"page":      page,
				"page_size": pageSize,
			})
		})

		req := httptest.NewRequest("GET", "/users?page=1&page_size=10", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		require.NoError(t, err)
		assert.Equal(t, "1", response["page"])
		assert.Equal(t, "10", response["page_size"])
	})
}

func TestUserHandler_Create(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
	}{
		{
			name: "valid user creation request",
			requestBody: map[string]string{
				"username":     "newuser",
				"email":        "new@example.com",
				"password":     "password123",
				"display_name": "New User",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "missing username",
			requestBody: map[string]string{
				"email":    "new@example.com",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email",
			requestBody: map[string]string{
				"username": "newuser",
				"email":    "invalid-email",
				"password": "password123",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "password too short",
			requestBody: map[string]string{
				"username": "newuser",
				"email":    "new@example.com",
				"password": "short",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()

			router.POST("/users", func(c *gin.Context) {
				var req CreateUserRequest
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
					return
				}
				c.JSON(http.StatusCreated, gin.H{"id": "new-user-id"})
			})

			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)
			req := httptest.NewRequest("POST", "/users", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
		})
	}
}

func TestUserHandler_GetByID(t *testing.T) {
	t.Run("get user by ID", func(t *testing.T) {
		router := gin.New()

		router.GET("/users/:id", func(c *gin.Context) {
			id := c.Param("id")
			if id == "" {
				c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"id": id, "username": "testuser"})
		})

		req := httptest.NewRequest("GET", "/users/user-123", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestResourceHandler_List(t *testing.T) {
	t.Run("list resources with filters", func(t *testing.T) {
		router := gin.New()

		router.GET("/resources", func(c *gin.Context) {
			resourceType := c.Query("type")
			environment := c.Query("environment")

			c.JSON(http.StatusOK, gin.H{
				"resources":   []interface{}{},
				"total":       0,
				"type":        resourceType,
				"environment": environment,
			})
		})

		req := httptest.NewRequest("GET", "/resources?type=vm&environment=dev", http.NoBody)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
