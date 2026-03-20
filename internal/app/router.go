package app

import (
	"net/http"

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
	e.GET("/healthz", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	siteGroup := e.Group("")
	adminGroup := e.Group("/admin")

	sitehandler.New(siteService).Register(siteGroup)
	adminhandler.New(siteService, adminService, adminSecret).Register(adminGroup)

	return e
}
