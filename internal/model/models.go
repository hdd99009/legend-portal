package model

type SiteSettings struct {
	SiteName        string
	SiteTitle       string
	SiteKeywords    string
	SiteDescription string
	FooterText      string
	ContactInfo     string
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
}

type GuestbookMessage struct {
	ID        int64
	Nickname  string
	Contact   string
	Content   string
	Status    string
	CreatedAt string
}
