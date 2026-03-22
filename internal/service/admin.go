package service

import (
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/bcrypt"

	"legend-portal/internal/model"
	"legend-portal/internal/repository"
	appstorage "legend-portal/internal/storage"
	"legend-portal/internal/util"
)

type AdminService struct {
	repo    *repository.SQLiteRepository
	storage appstorage.FileStorage
}

func NewAdminService(repo *repository.SQLiteRepository, storage appstorage.FileStorage) *AdminService {
	return &AdminService{repo: repo, storage: storage}
}

func (s *AdminService) Authenticate(username, password, ip string) (model.Admin, error) {
	admin, err := s.repo.FindAdminByUsername(username)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Admin{}, ErrInvalidCredentials
	}
	if err != nil {
		return model.Admin{}, err
	}
	if admin.Status != 1 {
		return model.Admin{}, ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(password)); err != nil {
		return model.Admin{}, ErrInvalidCredentials
	}
	if err := s.repo.UpdateAdminLogin(admin.ID, ip); err != nil {
		return model.Admin{}, err
	}
	return admin, nil
}

func (s *AdminService) ListPosts(limit int) ([]model.Post, error) {
	return s.repo.ListAdminPosts(limit)
}

func (s *AdminService) ListCategories() ([]model.Category, error) {
	return s.repo.ListCategories()
}

func (s *AdminService) ListTags() ([]model.Tag, error) {
	return s.repo.ListTags()
}

func (s *AdminService) GetTag(id int64) (model.Tag, error) {
	tag, err := s.repo.GetTagByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Tag{}, ErrNotFound
	}
	return tag, err
}

func (s *AdminService) CreateTag(tag model.Tag) error {
	tag = normalizeTag(tag)
	return s.repo.CreateTag(tag)
}

func (s *AdminService) UpdateTag(tag model.Tag) error {
	tag = normalizeTag(tag)
	if tag.ID <= 0 {
		return ErrNotFound
	}
	return s.repo.UpdateTag(tag)
}

func (s *AdminService) GetCategory(id int64) (model.Category, error) {
	category, err := s.repo.GetCategoryByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Category{}, ErrNotFound
	}
	return category, err
}

func (s *AdminService) CreateCategory(category model.Category) error {
	category = normalizeCategory(category)
	return s.repo.CreateCategory(category)
}

func (s *AdminService) UpdateCategory(category model.Category) error {
	category = normalizeCategory(category)
	if category.ID <= 0 {
		return ErrNotFound
	}
	return s.repo.UpdateCategory(category)
}

func (s *AdminService) GetPost(id int64) (model.Post, error) {
	post, err := s.repo.GetAdminPostByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Post{}, ErrNotFound
	}
	return post, err
}

func (s *AdminService) CreatePost(post model.Post) error {
	post = normalizePost(post)
	return s.repo.CreatePost(post)
}

func (s *AdminService) UpdatePost(post model.Post) error {
	post = normalizePost(post)
	if post.ID <= 0 {
		return ErrNotFound
	}
	return s.repo.UpdatePost(post)
}

func (s *AdminService) ChangePassword(adminID int64, oldPassword, newPassword, confirmPassword string) error {
	if strings.TrimSpace(oldPassword) == "" || strings.TrimSpace(newPassword) == "" || strings.TrimSpace(confirmPassword) == "" {
		return ErrPasswordRequired
	}
	if len(newPassword) < 6 {
		return ErrPasswordTooShort
	}
	if newPassword != confirmPassword {
		return ErrPasswordNotMatch
	}

	admin, err := s.repo.FindAdminByID(adminID)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrOldPasswordWrong
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	return s.repo.UpdateAdminPassword(adminID, string(newHash))
}

func (s *AdminService) ListMessages(limit int) ([]model.GuestbookMessage, error) {
	return s.repo.ListAdminMessages(limit)
}

func (s *AdminService) ApproveMessage(id int64, replyContent string) error {
	return s.repo.UpdateGuestbookStatus(id, "approved", replyContent)
}

func (s *AdminService) HideMessage(id int64, replyContent string) error {
	return s.repo.UpdateGuestbookStatus(id, "hidden", replyContent)
}

func (s *AdminService) UploadImage(fileHeader *multipart.FileHeader) (model.Upload, error) {
	if fileHeader == nil {
		return model.Upload{}, ErrUploadRequired
	}
	if fileHeader.Size <= 0 {
		return model.Upload{}, ErrUploadRequired
	}
	if fileHeader.Size > 5*1024*1024 {
		return model.Upload{}, ErrUploadTooLarge
	}

	file, err := fileHeader.Open()
	if err != nil {
		return model.Upload{}, err
	}
	defer file.Close()

	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && !errors.Is(err, io.EOF) {
		return model.Upload{}, err
	}

	contentType := http.DetectContentType(buffer[:n])
	allowedExts := map[string]string{
		"image/jpeg": ".jpg",
		"image/png":  ".png",
		"image/gif":  ".gif",
		"image/webp": ".webp",
	}
	ext, ok := allowedExts[contentType]
	if !ok {
		return model.Upload{}, ErrUploadType
	}

	if _, err := file.Seek(0, 0); err != nil {
		return model.Upload{}, err
	}

	stored, err := s.storage.SaveImage(file, ext)
	if err != nil {
		return model.Upload{}, err
	}

	upload := model.Upload{
		OriginName: fileHeader.Filename,
		SavedName:  filepath.Base(stored.Key),
		Path:       stored.Key,
		URL:        stored.URL,
		MimeType:   contentType,
		Size:       fileHeader.Size,
	}

	if err := s.repo.CreateUpload(upload); err != nil {
		_ = s.storage.Delete(stored.Key)
		return model.Upload{}, err
	}

	return upload, nil
}

func (s *AdminService) ListUploads(limit int) ([]model.Upload, error) {
	uploads, err := s.repo.ListUploads(limit)
	if err != nil {
		return nil, err
	}
	for i := range uploads {
		uploads[i].URL = s.storage.PublicURL(uploads[i].Path)
	}
	return uploads, nil
}

func (s *AdminService) DeleteUpload(id int64) error {
	upload, err := s.repo.GetUploadByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}

	url := s.storage.PublicURL(upload.Path)
	refCount, err := s.repo.CountPostReferencesByUploadURL(url)
	if err != nil {
		return err
	}
	if refCount > 0 {
		return ErrUploadReferenced
	}

	if err := s.storage.Delete(upload.Path); err != nil {
		return err
	}

	return s.repo.DeleteUploadByID(id)
}

func normalizePost(post model.Post) model.Post {
	post.Title = strings.TrimSpace(post.Title)
	post.Subtitle = strings.TrimSpace(post.Subtitle)
	post.Summary = strings.TrimSpace(post.Summary)
	post.Content = strings.TrimSpace(post.Content)
	post.ContentMarkdown = strings.TrimSpace(post.ContentMarkdown)
	post.CoverImage = strings.TrimSpace(post.CoverImage)
	post.Type = strings.TrimSpace(post.Type)
	post.Status = strings.TrimSpace(post.Status)
	post.GameVersion = strings.TrimSpace(post.GameVersion)
	post.ServerLine = strings.TrimSpace(post.ServerLine)
	post.GameGenre = strings.TrimSpace(post.GameGenre)
	post.Region = strings.TrimSpace(post.Region)
	post.OfficialURL = strings.TrimSpace(post.OfficialURL)
	post.DownloadURL = strings.TrimSpace(post.DownloadURL)
	post.QQGroup = strings.TrimSpace(post.QQGroup)
	post.SEOTitle = strings.TrimSpace(post.SEOTitle)
	post.SEOKeywords = strings.TrimSpace(post.SEOKeywords)
	post.SEODescription = strings.TrimSpace(post.SEODescription)

	if strings.TrimSpace(post.Slug) == "" {
		post.Slug = util.Slugify(post.Title)
	} else {
		post.Slug = util.Slugify(post.Slug)
	}
	if post.Type == "" {
		post.Type = "game"
	}
	if post.Status == "" {
		post.Status = "draft"
	}
	return post
}

func normalizeCategory(category model.Category) model.Category {
	category.Name = strings.TrimSpace(category.Name)
	category.Slug = strings.TrimSpace(category.Slug)
	category.SEOTitle = strings.TrimSpace(category.SEOTitle)
	category.SEOKeywords = strings.TrimSpace(category.SEOKeywords)
	category.SEODescription = strings.TrimSpace(category.SEODescription)

	if category.Slug == "" {
		category.Slug = util.Slugify(category.Name)
	} else {
		category.Slug = util.Slugify(category.Slug)
	}

	return category
}

func normalizeTag(tag model.Tag) model.Tag {
	tag.Name = strings.TrimSpace(tag.Name)
	tag.Slug = strings.TrimSpace(tag.Slug)

	if tag.Slug == "" {
		tag.Slug = util.Slugify(tag.Name)
	} else {
		tag.Slug = util.Slugify(tag.Slug)
	}

	return tag
}

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrOldPasswordWrong = errors.New("old password wrong")
var ErrPasswordNotMatch = errors.New("password confirm not match")
var ErrPasswordTooShort = errors.New("password too short")
var ErrPasswordRequired = errors.New("password required")
var ErrUploadRequired = errors.New("upload required")
var ErrUploadTooLarge = errors.New("upload too large")
var ErrUploadType = errors.New("upload type invalid")
var ErrUploadReferenced = errors.New("upload referenced by posts")
