package model

type SiteSettings struct {
	SiteName           string
	SiteTitle          string
	SiteKeywords       string
	SiteDescription    string
	FooterText         string
	ContactInfo        string
	HomeTechTitle      string
	HomeTechText       string
	HomeLatestTitle    string
	HomeLatestCount    int
	HomeRecommendTitle string
	HomeRecommendCount int
}

type Admin struct {
	ID           int64
	Username     string
	PasswordHash string
	Nickname     string
	Status       int
}

type Post struct {
	ID              int64
	Title           string
	Subtitle        string
	Slug            string
	CategoryID      int64
	CategoryName    string
	CoverImage      string
	Summary         string
	Content         string
	ContentMarkdown string
	Type            string
	Status          string
	IsTop           int
	IsRecommend     int
	GameVersion     string
	ServerLine      string
	GameGenre       string
	Region          string
	OfficialURL     string
	DownloadURL     string
	QQGroup         string
	SEOTitle        string
	SEOKeywords     string
	SEODescription  string
	PublishedAt     string
	CreatedAt       string
	UpdatedAt       string
	Tags            []Tag
	TagIDs          []int64
}

type GuestbookMessage struct {
	ID           int64
	Nickname     string
	Contact      string
	Content      string
	IP           string
	Status       string
	ReplyContent string
	CreatedAt    string
	UpdatedAt    string
}

type Upload struct {
	ID         int64
	OriginName string
	SavedName  string
	Path       string
	URL        string
	MimeType   string
	Size       int64
	CreatedAt  string
}

type Category struct {
	ID             int64
	Name           string
	Slug           string
	ParentID       int64
	Sort           int
	SEOTitle       string
	SEOKeywords    string
	SEODescription string
	CreatedAt      string
}

type Tag struct {
	ID        int64
	Name      string
	Slug      string
	CreatedAt string
}
