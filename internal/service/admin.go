package service

import (
	"database/sql"
	"errors"
	"strings"

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
