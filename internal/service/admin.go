package service

import (
	"database/sql"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"legend-portal/internal/model"
	"legend-portal/internal/repository"
	"legend-portal/internal/util"
)

type AdminService struct {
	repo *repository.SQLiteRepository
}

func NewAdminService(repo *repository.SQLiteRepository) *AdminService {
	return &AdminService{repo: repo}
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

	datePath := time.Now().Format("2006/01")
	savedName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	relativePath := filepath.ToSlash(filepath.Join(datePath, savedName))
	storagePath := filepath.Join("storage", "uploads", datePath, savedName)

	if err := repository.SaveUploadedFile(file, storagePath); err != nil {
		return model.Upload{}, err
	}

	upload := model.Upload{
		OriginName: fileHeader.Filename,
		SavedName:  savedName,
		Path:       relativePath,
		MimeType:   contentType,
		Size:       fileHeader.Size,
	}

	if err := s.repo.CreateUpload(upload); err != nil {
		_ = os.Remove(storagePath)
		return model.Upload{}, err
	}

	return upload, nil
}

func normalizePost(post model.Post) model.Post {
	post.Title = strings.TrimSpace(post.Title)
	post.Subtitle = strings.TrimSpace(post.Subtitle)
	post.Summary = strings.TrimSpace(post.Summary)
	post.Content = strings.TrimSpace(post.Content)
	post.ContentMarkdown = strings.TrimSpace(post.ContentMarkdown)
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

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrOldPasswordWrong = errors.New("old password wrong")
var ErrPasswordNotMatch = errors.New("password confirm not match")
var ErrPasswordTooShort = errors.New("password too short")
var ErrPasswordRequired = errors.New("password required")
var ErrUploadRequired = errors.New("upload required")
var ErrUploadTooLarge = errors.New("upload too large")
var ErrUploadType = errors.New("upload type invalid")
