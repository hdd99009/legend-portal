package app

import (
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	adminhandler "legend-portal/internal/handler/admin"
	sitehandler "legend-portal/internal/handler/site"
	"legend-portal/internal/service"
)

func NewServer(renderer *TemplateRenderer, siteService *service.SiteService, adminService *service.AdminService, adminSecret string) *echo.Echo {
	e := echo.New()
	e.HideBanner = true
	e.Renderer = renderer

	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Use(middleware.Gzip())

	e.Static("/assets", "assets")
	e.GET("/uploads/*", serveProtectedUpload)
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	siteGroup := e.Group("")
	adminGroup := e.Group("/admin")

	sitehandler.New(siteService).Register(siteGroup)
	adminhandler.New(siteService, adminService, adminSecret).Register(adminGroup)

	return e
}

func serveProtectedUpload(c echo.Context) error {
	relative := strings.TrimPrefix(c.Param("*"), "/")
	clean := filepath.Clean(relative)
	if clean == "." || strings.Contains(clean, "..") {
		return c.String(http.StatusBadRequest, "invalid path")
	}

	referer := c.Request().Header.Get("Referer")
	if referer != "" {
		if refURL, err := url.Parse(referer); err == nil {
			requestHost := c.Request().Host
			if host := refURL.Host; host != "" && !sameHost(host, requestHost) {
				return c.String(http.StatusForbidden, "forbidden")
			}
		}
	}

	target := filepath.Join("storage", "uploads", clean)
	if _, err := os.Stat(target); err != nil {
		return c.String(http.StatusNotFound, "not found")
	}

	return c.File(target)
}

func sameHost(a, b string) bool {
	return stripPort(strings.ToLower(a)) == stripPort(strings.ToLower(b))
}

func stripPort(host string) string {
	if idx := strings.Index(host, ":"); idx >= 0 {
		return host[:idx]
	}
	return host
}
