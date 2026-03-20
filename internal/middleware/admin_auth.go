package middleware

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"strings"

	"github.com/labstack/echo/v4"
)

const AdminSessionCookie = "legend_admin_session"

func SignAdminSession(adminID int64, secret string) string {
	payload := strconv.FormatInt(adminID, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	return payload + ":" + hex.EncodeToString(mac.Sum(nil))
}

func ParseAdminSession(token, secret string) (int64, bool) {
	parts := strings.Split(token, ":")
	if len(parts) != 2 {
		return 0, false
	}

	payload := parts[0]
	signature := parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(payload))
	expected := hex.EncodeToString(mac.Sum(nil))
	if !hmac.Equal([]byte(signature), []byte(expected)) {
		return 0, false
	}

	adminID, err := strconv.ParseInt(payload, 10, 64)
	if err != nil || adminID <= 0 {
		return 0, false
	}

	return adminID, true
}

func RequireAdmin(secret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cookie, err := c.Cookie(AdminSessionCookie)
			if err != nil {
				return c.Redirect(http.StatusFound, "/admin/login")
			}

			adminID, ok := ParseAdminSession(cookie.Value, secret)
			if !ok {
				return c.Redirect(http.StatusFound, "/admin/login")
			}

			c.Set("admin_id", adminID)
			return next(c)
		}
	}
}
