package service

import (
	"database/sql"
	"errors"

	"legend-portal/internal/model"
	"legend-portal/internal/repository"
)

type SiteService struct {
	repo *repository.SQLiteRepository
}

func NewSiteService(repo *repository.SQLiteRepository) *SiteService {
	return &SiteService{repo: repo}
}

func (s *SiteService) SiteSettings() (model.SiteSettings, error) {
	return s.repo.GetSiteSettings()
}

func (s *SiteService) HomePosts(limit int) ([]model.Post, error) {
	return s.repo.ListPublishedPosts(limit)
}

func (s *SiteService) PostBySlug(slug string) (model.Post, error) {
	post, err := s.repo.GetPostBySlug(slug)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Post{}, ErrNotFound
	}
	return post, err
}

func (s *SiteService) ApprovedMessages(limit int) ([]model.GuestbookMessage, error) {
	return s.repo.ListApprovedMessages(limit)
}

func (s *SiteService) CreateMessage(nickname, contact, content, ip string) error {
	return s.repo.CreateGuestbookMessage(nickname, contact, content, ip)
}

func (s *SiteService) PendingMessageCount() (int, error) {
	return s.repo.CountPendingMessages()
}

func (s *SiteService) UpdateSiteSettings(settings model.SiteSettings) error {
	return s.repo.UpdateSiteSettings(settings)
}

var ErrNotFound = errors.New("resource not found")
