package middleware

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	RequestIDHeader = "X-Request-ID"
	RequestIDKey    = "request_id"

	// maxRequestIDLen caps client-supplied request IDs. Uncapped values are
	// reflected into the response header and stored in the request context
	// (logs), so overlong input risks header bloat and log injection.
	maxRequestIDLen = 64
)

func isValidRequestID(s string) bool {
	if s == "" || len(s) > maxRequestIDLen {
		return false
	}
	for i := range len(s) {
		c := s[i]
		if (c < 'a' || c > 'z') && (c < 'A' || c > 'Z') && (c < '0' || c > '9') && c != '-' {
			return false
		}
	}
	return true
}

func RequestID() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			requestID := c.Request().Header.Get(RequestIDHeader)
			if !isValidRequestID(requestID) {
				requestID = uuid.New().String() // 4c90fc3f-39cc-4b04-af21-c83ee64aa67e
			}

			c.Set(RequestIDKey, requestID)
			c.Response().Header().Set(RequestIDHeader, requestID)

			return next(c)
		}
	}
}

func GetRequestID(c echo.Context) string {
	if requestID, ok := c.Get(RequestIDKey).(string); ok {
		return requestID
	}
	return ""
}
