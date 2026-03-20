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

type PasswordViewData struct {
	PageTitle string
	Error     string
	Success   string
}

type MessageListViewData struct {
	PageTitle string
	Messages  []model.GuestbookMessage
}

type SettingsViewData struct {
	PageTitle string
	Settings  model.SiteSettings
	Error     string
	Success   string
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
	protected.GET("/messages", h.MessageList)
	protected.POST("/messages/:id/approve", h.MessageApprove)
	protected.POST("/messages/:id/hide", h.MessageHide)
	protected.POST("/upload/image", h.UploadImage)
	protected.GET("/settings", h.SettingsPage)
	protected.POST("/settings", h.SettingsSubmit)
	protected.GET("/password", h.PasswordPage)
	protected.POST("/password", h.PasswordSubmit)
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

func (h *Handler) PasswordPage(c echo.Context) error {
	return c.Render(http.StatusOK, "admin/password.html", PasswordViewData{
		PageTitle: "修改密码",
		Success:   c.QueryParam("success"),
	})
}

func (h *Handler) PasswordSubmit(c echo.Context) error {
	adminIDValue := c.Get("admin_id")
	adminID, ok := adminIDValue.(int64)
	if !ok || adminID <= 0 {
		return c.Redirect(http.StatusFound, "/admin/login")
	}

	err := h.adminService.ChangePassword(
		adminID,
		c.FormValue("old_password"),
		c.FormValue("new_password"),
		c.FormValue("confirm_password"),
	)
	if err != nil {
		message := "修改失败，请稍后重试"
		switch err {
		case service.ErrPasswordRequired:
			message = "请完整填写密码字段"
		case service.ErrPasswordTooShort:
			message = "新密码至少 6 位"
		case service.ErrPasswordNotMatch:
			message = "两次输入的新密码不一致"
		case service.ErrOldPasswordWrong:
			message = "旧密码不正确"
		}
		return c.Render(http.StatusBadRequest, "admin/password.html", PasswordViewData{
			PageTitle: "修改密码",
			Error:     message,
		})
	}

	return c.Redirect(http.StatusFound, "/admin/password?success=密码修改成功")
}

func (h *Handler) MessageList(c echo.Context) error {
	messages, err := h.adminService.ListMessages(200)
	if err != nil {
		return err
	}
	return c.Render(http.StatusOK, "admin/messages_list.html", MessageListViewData{
		PageTitle: "留言审核",
		Messages:  messages,
	})
}

func (h *Handler) MessageApprove(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "参数错误")
	}
	if err := h.adminService.ApproveMessage(id, c.FormValue("reply_content")); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/admin/messages")
}

func (h *Handler) MessageHide(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return c.String(http.StatusBadRequest, "参数错误")
	}
	if err := h.adminService.HideMessage(id, c.FormValue("reply_content")); err != nil {
		return err
	}
	return c.Redirect(http.StatusFound, "/admin/messages")
}

func (h *Handler) UploadImage(c echo.Context) error {
	fileHeader, err := c.FormFile("image")
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "请选择图片文件",
		})
	}

	upload, err := h.adminService.UploadImage(fileHeader)
	if err != nil {
		message := "上传失败"
		switch err {
		case service.ErrUploadRequired:
			message = "请选择图片文件"
		case service.ErrUploadTooLarge:
			message = "图片大小不能超过 5MB"
		case service.ErrUploadType:
			message = "仅支持 jpg/png/gif/webp"
		}
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": message,
		})
	}

	return c.JSON(http.StatusOK, map[string]string{
		"url":  "/uploads/" + upload.Path,
		"path": upload.Path,
	})
}

func (h *Handler) SettingsPage(c echo.Context) error {
	settings, err := h.siteService.SiteSettings()
	if err != nil {
		return err
	}

	return c.Render(http.StatusOK, "admin/settings.html", SettingsViewData{
		PageTitle: "站点设置",
		Settings:  settings,
		Success:   c.QueryParam("success"),
	})
}

func (h *Handler) SettingsSubmit(c echo.Context) error {
	settings := model.SiteSettings{
		SiteName:        c.FormValue("site_name"),
		SiteTitle:       c.FormValue("site_title"),
		SiteKeywords:    c.FormValue("site_keywords"),
		SiteDescription: c.FormValue("site_description"),
		FooterText:      c.FormValue("footer_text"),
		ContactInfo:     c.FormValue("contact_info"),
	}

	if settings.SiteName == "" || settings.SiteTitle == "" {
		return c.Render(http.StatusBadRequest, "admin/settings.html", SettingsViewData{
			PageTitle: "站点设置",
			Settings:  settings,
			Error:     "站点名称和首页标题不能为空",
		})
	}

	if err := h.siteService.UpdateSiteSettings(settings); err != nil {
		return c.Render(http.StatusBadRequest, "admin/settings.html", SettingsViewData{
			PageTitle: "站点设置",
			Settings:  settings,
			Error:     "保存失败，请稍后重试",
		})
	}

	return c.Redirect(http.StatusFound, "/admin/settings?success=站点设置已保存")
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
