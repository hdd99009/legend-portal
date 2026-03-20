package admin

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"

	appmw "legend-portal/internal/middleware"
	"legend-portal/internal/model"
	"legend-portal/internal/service"
)

type Handler struct {
	siteService  *service.SiteService
	adminService *service.AdminService
	secret       string
}

type DashboardViewData struct {
	PageTitle      string
	PendingMessage int
}

type LoginViewData struct {
	PageTitle string
	Error     string
}

type PostListViewData struct {
	PageTitle string
	Posts     []model.Post
}

type PostFormViewData struct {
	PageTitle string
	Post      model.Post
	Error     string
	IsEdit    bool
}

func New(siteService *service.SiteService, adminService *service.AdminService, secret string) *Handler {
	return &Handler{
		siteService:  siteService,
		adminService: adminService,
		secret:       secret,
	}
}

func (h *Handler) Register(g *echo.Group) {
	g.GET("/login", h.LoginPage)
	g.POST("/login", h.LoginSubmit)
	g.GET("/logout", h.Logout)

	protected := g.Group("")
	protected.Use(appmw.RequireAdmin(h.secret))
	protected.GET("", h.Index)
	protected.GET("/", h.Index)
	protected.GET("/posts", h.PostList)
	protected.GET("/posts/new", h.PostNewPage)
	protected.POST("/posts/new", h.PostCreate)
	protected.GET("/posts/:id/edit", h.PostEditPage)
	protected.POST("/posts/:id/edit", h.PostUpdate)
}

func (h *Handler) LoginPage(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/login.html", LoginViewData{
		PageTitle: "后台登录",
		Error:     c.QueryParam("error"),
	})
}

func (h *Handler) LoginSubmit(c echo.Context) error {
	admin, err := h.adminService.Authenticate(c.FormValue("username"), c.FormValue("password"), c.RealIP())
	if err != nil {
		return c.Redirect(http.StatusFound, "/admin/login?error=用户名或密码错误")
	}

	c.SetCookie(&http.Cookie{
		Name:     appmw.AdminSessionCookie,
		Value:    appmw.SignAdminSession(admin.ID, h.secret),
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	return c.Redirect(http.StatusFound, "/admin")
}

func (h *Handler) Logout(c echo.Context) error {
	c.SetCookie(&http.Cookie{
		Name:     appmw.AdminSessionCookie,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		MaxAge:   -1,
	})
	return c.Redirect(http.StatusFound, "/admin/login")
}

func (h *Handler) Index(c echo.Context) error {
	pending, err := h.siteService.PendingMessageCount()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "admin/index.html", DashboardViewData{
		PageTitle:      "后台管理",
		PendingMessage: pending,
	})
}

func (h *Handler) PostList(c echo.Context) error {
	posts, err := h.adminService.ListPosts(100)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/posts_list.html", PostListViewData{
		PageTitle: "文章管理",
		Posts:     posts,
	})
}

func (h *Handler) PostNewPage(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/post_form.html", PostFormViewData{
		PageTitle: "新增文章",
		Post: model.Post{
			Type:   "game",
			Status: "draft",
		},
	})
}

func (h *Handler) PostCreate(c echo.Context) error {
	post := bindPostForm(c)
	if post.Title == "" || post.Content == "" {
		return c.Render(http.StatusBadRequest, "admin/post_form.html", PostFormViewData{
			PageTitle: "新增文章",
			Post:      post,
			Error:     "标题和正文不能为空",
		})
	}

	if err := h.adminService.CreatePost(post); err != nil {
		return c.Render(http.StatusBadRequest, "admin/post_form.html", PostFormViewData{
			PageTitle: "新增文章",
			Post:      post,
			Error:     "保存失败，请检查 slug 是否重复",
		})
	}

	return c.Redirect(http.StatusFound, "/admin/posts")
}

func (h *Handler) PostEditPage(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "参数错误")
	}

	post, err := h.adminService.GetPost(id)
	if err != nil {
		return c.String(http.StatusNotFound, "文章不存在")
	}

	return c.Render(http.StatusOK, "admin/post_form.html", PostFormViewData{
		PageTitle: "编辑文章",
		Post:      post,
		IsEdit:    true,
	})
}

func (h *Handler) PostUpdate(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "参数错误")
	}

	post := bindPostForm(c)
	post.ID = id
	if post.Title == "" || post.Content == "" {
		return c.Render(http.StatusBadRequest, "admin/post_form.html", PostFormViewData{
			PageTitle: "编辑文章",
			Post:      post,
			Error:     "标题和正文不能为空",
			IsEdit:    true,
		})
	}

	if err := h.adminService.UpdatePost(post); err != nil {
		return c.Render(http.StatusBadRequest, "admin/post_form.html", PostFormViewData{
			PageTitle: "编辑文章",
			Post:      post,
			Error:     "更新失败，请检查 slug 是否重复",
			IsEdit:    true,
		})
	}

	return c.Redirect(http.StatusFound, "/admin/posts")
}

func bindPostForm(c echo.Context) model.Post {
	categoryID, _ := strconv.ParseInt(c.FormValue("category_id"), 10, 64)
	isTop := 0
	if c.FormValue("is_top") == "1" {
		isTop = 1
	}
	isRecommend := 0
	if c.FormValue("is_recommend") == "1" {
		isRecommend = 1
	}

	return model.Post{
		Title:           c.FormValue("title"),
		Subtitle:        c.FormValue("subtitle"),
		Slug:            c.FormValue("slug"),
		CategoryID:      categoryID,
		Summary:         c.FormValue("summary"),
		Content:         c.FormValue("content"),
		ContentMarkdown: c.FormValue("content_markdown"),
		Type:            c.FormValue("type"),
		Status:          c.FormValue("status"),
		IsTop:           isTop,
		IsRecommend:     isRecommend,
		GameVersion:     c.FormValue("game_version"),
		ServerLine:      c.FormValue("server_line"),
		GameGenre:       c.FormValue("game_genre"),
		Region:          c.FormValue("region"),
		OfficialURL:     c.FormValue("official_url"),
		DownloadURL:     c.FormValue("download_url"),
		QQGroup:         c.FormValue("qq_group"),
		SEOTitle:        c.FormValue("seo_title"),
		SEOKeywords:     c.FormValue("seo_keywords"),
		SEODescription:  c.FormValue("seo_description"),
	}
}
