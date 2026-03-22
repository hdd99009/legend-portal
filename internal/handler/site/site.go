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
	SiteName           string
	SiteTitle          string
	SiteKeywords       string
	SiteDescription    string
	FooterText         string
	ContactInfo        string
	HomeTechTitle      string
	HomeTechText       string
	HomeLatestTitle    string
	HomeRecommendTitle string
	PageTitle          string
	CurrentPath        string
	Posts              interface{}
	RecommendedPosts   interface{}
	Post               interface{}
	Tag                interface{}
	Messages           interface{}
	Notice             string
}

func New(service *service.SiteService) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/", h.Home)
	g.GET("/posts/:slug", h.PostDetail)
	g.GET("/tags/:slug", h.TagDetail)
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

	latestCount := settings.HomeLatestCount
	if latestCount <= 0 {
		latestCount = 12
	}
	recommendCount := settings.HomeRecommendCount
	if recommendCount <= 0 {
		recommendCount = 6
	}

	posts, err := h.service.HomePosts(latestCount)
	if err != nil {
		return err
	}

	if settings.HomeLatestTitle == "" {
		settings.HomeLatestTitle = "最新发布"
	}
	if settings.HomeRecommendTitle == "" {
		settings.HomeRecommendTitle = "推荐内容"
	}
	if settings.HomeTechTitle == "" {
		settings.HomeTechTitle = "首页模块"
	}

	recommendedPosts, err := h.service.RecommendedPosts(recommendCount)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "site/index.html", ViewData{
		SiteName:           settings.SiteName,
		SiteTitle:          settings.SiteTitle,
		SiteKeywords:       settings.SiteKeywords,
		SiteDescription:    settings.SiteDescription,
		FooterText:         settings.FooterText,
		ContactInfo:        settings.ContactInfo,
		HomeTechTitle:      settings.HomeTechTitle,
		HomeTechText:       settings.HomeTechText,
		HomeLatestTitle:    settings.HomeLatestTitle,
		HomeRecommendTitle: settings.HomeRecommendTitle,
		PageTitle:          settings.SiteTitle,
		CurrentPath:        c.Request().URL.Path,
		Posts:              posts,
		RecommendedPosts:   recommendedPosts,
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
	keywords := settings.SiteKeywords
	description := settings.SiteDescription
	if post.SEOTitle != "" {
		title = post.SEOTitle
	}
	if post.SEOKeywords != "" {
		keywords = post.SEOKeywords
	}
	if post.SEODescription != "" {
		description = post.SEODescription
	}

	return c.Render(http.StatusOK, "site/post_detail.html", ViewData{
		SiteName:        settings.SiteName,
		SiteTitle:       settings.SiteTitle,
		SiteKeywords:    keywords,
		SiteDescription: description,
		FooterText:      settings.FooterText,
		ContactInfo:     settings.ContactInfo,
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
		SiteName:        settings.SiteName,
		SiteTitle:       settings.SiteTitle,
		SiteKeywords:    settings.SiteKeywords,
		SiteDescription: settings.SiteDescription,
		FooterText:      settings.FooterText,
		ContactInfo:     settings.ContactInfo,
		PageTitle:       "玩家留言板",
		CurrentPath:     c.Request().URL.Path,
		Messages:        messages,
		Notice:          c.QueryParam("notice"),
	})
}

func (h *Handler) TagDetail(c echo.Context) error {
	settings, err := h.service.SiteSettings()
	if err != nil {
		return err
	}

	tag, err := h.service.TagBySlug(c.Param("slug"))
	if err != nil {
		if err == service.ErrNotFound {
			return c.String(http.StatusNotFound, "页面不存在")
		}
		return err
	}

	posts, err := h.service.PostsByTag(tag.ID, 100)
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "site/tag_detail.html", ViewData{
		SiteName:        settings.SiteName,
		SiteTitle:       settings.SiteTitle,
		SiteKeywords:    settings.SiteKeywords,
		SiteDescription: settings.SiteDescription,
		FooterText:      settings.FooterText,
		ContactInfo:     settings.ContactInfo,
		PageTitle:       tag.Name + " - 标签内容",
		CurrentPath:     c.Request().URL.Path,
		Tag:             tag,
		Posts:           posts,
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
