package site

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"

	"legend-portal/internal/service"
)

type Handler struct {
	service *service.SiteService
}

type ViewData struct {
	SiteTitle       string
	SiteKeywords    string
	SiteDescription string
	PageTitle       string
	CurrentPath     string
	Posts           interface{}
	Post            interface{}
	Messages        interface{}
	Notice          string
}

func New(service *service.SiteService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/", h.Home)
	g.GET("/posts/:slug", h.PostDetail)
	g.GET("/guestbook", h.Guestbook)
	g.POST("/guestbook", h.CreateGuestbook)
	g.GET("/robots.txt", h.Robots)
	g.GET("/sitemap.xml", h.Sitemap)
}

func (h *Handler) Home(c echo.Context) error {
	settings, err := h.service.SiteSettings()
	if err != nil {
		return err
	}

	posts, err := h.service.HomePosts(12)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "site/index.html", ViewData{
		SiteTitle:       settings.SiteTitle,
		SiteKeywords:    settings.SiteKeywords,
		SiteDescription: settings.SiteDescription,
		PageTitle:       settings.SiteTitle,
		CurrentPath:     c.Request().URL.Path,
		Posts:           posts,
	})
}

func (h *Handler) PostDetail(c echo.Context) error {
	settings, err := h.service.SiteSettings()
	if err != nil {
		return err
	}

	post, err := h.service.PostBySlug(c.Param("slug"))
	if err != nil {
		if err == service.ErrNotFound {
			return c.String(http.StatusNotFound, "页面不存在")
		}
		return err
	}

	title := post.Title
	if post.SEOTitle != "" {
		title = post.SEOTitle
	}

	return c.Render(http.StatusOK, "site/post_detail.html", ViewData{
		SiteTitle:       settings.SiteTitle,
		SiteKeywords:    settings.SiteKeywords,
		SiteDescription: settings.SiteDescription,
		PageTitle:       title,
		CurrentPath:     c.Request().URL.Path,
		Post:            post,
	})
}

func (h *Handler) Guestbook(c echo.Context) error {
	settings, err := h.service.SiteSettings()
	if err != nil {
		return err
	}

	messages, err := h.service.ApprovedMessages(50)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "site/guestbook.html", ViewData{
		SiteTitle:       settings.SiteTitle,
		SiteKeywords:    settings.SiteKeywords,
		SiteDescription: settings.SiteDescription,
		PageTitle:       "玩家留言板",
		CurrentPath:     c.Request().URL.Path,
		Messages:        messages,
		Notice:          c.QueryParam("notice"),
	})
}

func (h *Handler) CreateGuestbook(c echo.Context) error {
	nickname := c.FormValue("nickname")
	contact := c.FormValue("contact")
	content := c.FormValue("content")

	if nickname == "" || content == "" {
		return c.Redirect(http.StatusFound, "/guestbook?notice=昵称和内容不能为空")
	}

	if err := h.service.CreateMessage(nickname, contact, content, c.RealIP()); err != nil {
		return err
	}

	return c.Redirect(http.StatusFound, "/guestbook?notice=留言已提交，等待审核")
}

func (h *Handler) Robots(c echo.Context) error {
	return c.String(http.StatusOK, "User-agent: *\nAllow: /\nSitemap: /sitemap.xml\n")
}

func (h *Handler) Sitemap(c echo.Context) error {
	posts, err := h.service.HomePosts(200)
	if err != nil {
		return err
	}

	var builder strings.Builder
	builder.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	builder.WriteString(`<urlset xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">` + "\n")
	builder.WriteString("  <url><loc>/</loc></url>\n")
	builder.WriteString("  <url><loc>/guestbook</loc></url>\n")
	for _, post := range posts {
		builder.WriteString("  <url><loc>/posts/" + post.Slug + "</loc></url>\n")
	}
	builder.WriteString(`</urlset>`)

	return c.Blob(http.StatusOK, "application/xml; charset=utf-8", []byte(builder.String()))
}
