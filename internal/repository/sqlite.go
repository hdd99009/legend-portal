package repository

import (
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"

	"legend-portal/internal/model"
)

type SQLiteRepository struct {
	DB *sql.DB
}

func NewSQLiteRepository(path string) (*SQLiteRepository, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(ON)", path)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	return &SQLiteRepository{DB: db}, nil
}

func (r *SQLiteRepository) Close() error {
	return r.DB.Close()
}

func (r *SQLiteRepository) Migrate(sqlPath string) error {
	content, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}

	statements := strings.Split(string(content), ";")
	for _, statement := range statements {
		stmt := strings.TrimSpace(statement)
		if stmt == "" {
			continue
		}
		if _, err := r.DB.Exec(stmt); err != nil {
			// Keep startup migrations idempotent for SQLite ALTER TABLE ADD COLUMN.
			if strings.Contains(err.Error(), "duplicate column name") {
				continue
			}
			return err
		}
	}

	return nil
}

func (r *SQLiteRepository) GetSiteSettings() (model.SiteSettings, error) {
	var settings model.SiteSettings
	row := r.DB.QueryRow(`
		SELECT site_name, site_title, site_keywords, site_description, footer_text, contact_info,
		       home_tech_title, home_tech_text, home_latest_title, home_latest_count,
		       home_recommend_title, home_recommend_count
		FROM site_settings
		ORDER BY id ASC
		LIMIT 1
	`)

	err := row.Scan(
		&settings.SiteName,
		&settings.SiteTitle,
		&settings.SiteKeywords,
		&settings.SiteDescription,
		&settings.FooterText,
		&settings.ContactInfo,
		&settings.HomeTechTitle,
		&settings.HomeTechText,
		&settings.HomeLatestTitle,
		&settings.HomeLatestCount,
		&settings.HomeRecommendTitle,
		&settings.HomeRecommendCount,
	)

	return settings, err
}

func (r *SQLiteRepository) UpdateSiteSettings(settings model.SiteSettings) error {
	_, err := r.DB.Exec(`
		UPDATE site_settings
		SET site_name = ?, site_title = ?, site_keywords = ?, site_description = ?,
		    footer_text = ?, contact_info = ?, home_tech_title = ?, home_tech_text = ?,
		    home_latest_title = ?, home_latest_count = ?, home_recommend_title = ?, home_recommend_count = ?,
		    updated_at = CURRENT_TIMESTAMP
		WHERE id = (
			SELECT id FROM site_settings ORDER BY id ASC LIMIT 1
		)
	`, settings.SiteName, settings.SiteTitle, settings.SiteKeywords, settings.SiteDescription, settings.FooterText, settings.ContactInfo, settings.HomeTechTitle, settings.HomeTechText, settings.HomeLatestTitle, settings.HomeLatestCount, settings.HomeRecommendTitle, settings.HomeRecommendCount)
	return err
}

func (r *SQLiteRepository) FindAdminByUsername(username string) (model.Admin, error) {
	var admin model.Admin
	row := r.DB.QueryRow(`
		SELECT id, username, password_hash, nickname, status
		FROM admins
		WHERE username = ?
		LIMIT 1
	`, strings.TrimSpace(username))

	err := row.Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Nickname, &admin.Status)
	return admin, err
}

func (r *SQLiteRepository) UpdateAdminLogin(adminID int64, ip string) error {
	_, err := r.DB.Exec(`
		UPDATE admins
		SET last_login_at = CURRENT_TIMESTAMP, last_login_ip = ?
		WHERE id = ?
	`, ip, adminID)
	return err
}

func (r *SQLiteRepository) FindAdminByID(id int64) (model.Admin, error) {
	var admin model.Admin
	row := r.DB.QueryRow(`
		SELECT id, username, password_hash, nickname, status
		FROM admins
		WHERE id = ?
		LIMIT 1
	`, id)

	err := row.Scan(&admin.ID, &admin.Username, &admin.PasswordHash, &admin.Nickname, &admin.Status)
	return admin, err
}

func (r *SQLiteRepository) UpdateAdminPassword(adminID int64, passwordHash string) error {
	_, err := r.DB.Exec(`
		UPDATE admins
		SET password_hash = ?
		WHERE id = ?
	`, passwordHash, adminID)
	return err
}

func (r *SQLiteRepository) ListCategories() ([]model.Category, error) {
	rows, err := r.DB.Query(`
		SELECT id, name, slug, parent_id, sort, seo_title, seo_keywords, seo_description, created_at
		FROM post_categories
		ORDER BY sort DESC, id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []model.Category
	for rows.Next() {
		var category model.Category
		if err := rows.Scan(
			&category.ID,
			&category.Name,
			&category.Slug,
			&category.ParentID,
			&category.Sort,
			&category.SEOTitle,
			&category.SEOKeywords,
			&category.SEODescription,
			&category.CreatedAt,
		); err != nil {
			return nil, err
		}
		categories = append(categories, category)
	}

	return categories, rows.Err()
}

func (r *SQLiteRepository) GetCategoryByID(id int64) (model.Category, error) {
	var category model.Category
	row := r.DB.QueryRow(`
		SELECT id, name, slug, parent_id, sort, seo_title, seo_keywords, seo_description, created_at
		FROM post_categories
		WHERE id = ?
		LIMIT 1
	`, id)

	err := row.Scan(
		&category.ID,
		&category.Name,
		&category.Slug,
		&category.ParentID,
		&category.Sort,
		&category.SEOTitle,
		&category.SEOKeywords,
		&category.SEODescription,
		&category.CreatedAt,
	)
	return category, err
}

func (r *SQLiteRepository) CreateCategory(category model.Category) error {
	_, err := r.DB.Exec(`
		INSERT INTO post_categories (name, slug, parent_id, sort, seo_title, seo_keywords, seo_description)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, category.Name, category.Slug, category.ParentID, category.Sort, category.SEOTitle, category.SEOKeywords, category.SEODescription)
	return err
}

func (r *SQLiteRepository) ListTags() ([]model.Tag, error) {
	rows, err := r.DB.Query(`
		SELECT id, name, slug, created_at
		FROM post_tags
		ORDER BY id DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

func (r *SQLiteRepository) GetTagByID(id int64) (model.Tag, error) {
	var tag model.Tag
	row := r.DB.QueryRow(`
		SELECT id, name, slug, created_at
		FROM post_tags
		WHERE id = ?
		LIMIT 1
	`, id)

	err := row.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt)
	return tag, err
}

func (r *SQLiteRepository) GetTagBySlug(slug string) (model.Tag, error) {
	var tag model.Tag
	row := r.DB.QueryRow(`
		SELECT id, name, slug, created_at
		FROM post_tags
		WHERE slug = ?
		LIMIT 1
	`, slug)

	err := row.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt)
	return tag, err
}

func (r *SQLiteRepository) CreateTag(tag model.Tag) error {
	_, err := r.DB.Exec(`
		INSERT INTO post_tags (name, slug)
		VALUES (?, ?)
	`, tag.Name, tag.Slug)
	return err
}

func (r *SQLiteRepository) UpdateTag(tag model.Tag) error {
	_, err := r.DB.Exec(`
		UPDATE post_tags
		SET name = ?, slug = ?
		WHERE id = ?
	`, tag.Name, tag.Slug, tag.ID)
	return err
}

func (r *SQLiteRepository) UpdateCategory(category model.Category) error {
	_, err := r.DB.Exec(`
		UPDATE post_categories
		SET name = ?, slug = ?, parent_id = ?, sort = ?, seo_title = ?, seo_keywords = ?, seo_description = ?
		WHERE id = ?
	`, category.Name, category.Slug, category.ParentID, category.Sort, category.SEOTitle, category.SEOKeywords, category.SEODescription, category.ID)
	return err
}

func (r *SQLiteRepository) ListPublishedPosts(limit int) ([]model.Post, error) {
	rows, err := r.DB.Query(`
		SELECT p.id, p.title, p.subtitle, p.slug, p.category_id, COALESCE(c.name, ''), p.cover_image, p.summary, p.content, p.content_markdown, p.type, p.status, p.is_top, p.is_recommend, p.game_version, p.server_line,
		       p.game_genre, p.region, p.official_url, p.download_url, p.qq_group, p.seo_title, p.seo_keywords,
		       p.seo_description, p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_categories c ON c.id = p.category_id
		WHERE p.status = 'published'
		ORDER BY p.is_top DESC, p.published_at DESC, p.id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Subtitle,
			&post.Slug,
			&post.CategoryID,
			&post.CategoryName,
			&post.CoverImage,
			&post.Summary,
			&post.Content,
			&post.ContentMarkdown,
			&post.Type,
			&post.Status,
			&post.IsTop,
			&post.IsRecommend,
			&post.GameVersion,
			&post.ServerLine,
			&post.GameGenre,
			&post.Region,
			&post.OfficialURL,
			&post.DownloadURL,
			&post.QQGroup,
			&post.SEOTitle,
			&post.SEOKeywords,
			&post.SEODescription,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.populatePostTags(posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *SQLiteRepository) ListRecommendedPosts(limit int) ([]model.Post, error) {
	rows, err := r.DB.Query(`
		SELECT p.id, p.title, p.subtitle, p.slug, p.category_id, COALESCE(c.name, ''), p.cover_image, p.summary, p.content, p.content_markdown, p.type, p.status, p.is_top, p.is_recommend, p.game_version, p.server_line,
		       p.game_genre, p.region, p.official_url, p.download_url, p.qq_group, p.seo_title, p.seo_keywords,
		       p.seo_description, p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_categories c ON c.id = p.category_id
		WHERE p.status = 'published' AND p.is_recommend = 1
		ORDER BY p.is_top DESC, p.published_at DESC, p.id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Subtitle,
			&post.Slug,
			&post.CategoryID,
			&post.CategoryName,
			&post.CoverImage,
			&post.Summary,
			&post.Content,
			&post.ContentMarkdown,
			&post.Type,
			&post.Status,
			&post.IsTop,
			&post.IsRecommend,
			&post.GameVersion,
			&post.ServerLine,
			&post.GameGenre,
			&post.Region,
			&post.OfficialURL,
			&post.DownloadURL,
			&post.QQGroup,
			&post.SEOTitle,
			&post.SEOKeywords,
			&post.SEODescription,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.populatePostTags(posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *SQLiteRepository) GetPostBySlug(slug string) (model.Post, error) {
	var post model.Post
	row := r.DB.QueryRow(`
		SELECT p.id, p.title, p.subtitle, p.slug, p.category_id, COALESCE(c.name, ''), p.cover_image, p.summary, p.content, p.content_markdown, p.type, p.status, p.is_top, p.is_recommend, p.game_version, p.server_line,
		       p.game_genre, p.region, p.official_url, p.download_url, p.qq_group, p.seo_title, p.seo_keywords,
		       p.seo_description, p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_categories c ON c.id = p.category_id
		WHERE p.slug = ? AND p.status = 'published'
		LIMIT 1
	`, slug)

	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Subtitle,
		&post.Slug,
		&post.CategoryID,
		&post.CategoryName,
		&post.CoverImage,
		&post.Summary,
		&post.Content,
		&post.ContentMarkdown,
		&post.Type,
		&post.Status,
		&post.IsTop,
		&post.IsRecommend,
		&post.GameVersion,
		&post.ServerLine,
		&post.GameGenre,
		&post.Region,
		&post.OfficialURL,
		&post.DownloadURL,
		&post.QQGroup,
		&post.SEOTitle,
		&post.SEOKeywords,
		&post.SEODescription,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return post, err
	}
	post.Tags, err = r.ListTagsByPostID(post.ID)
	if err != nil {
		return post, err
	}
	post.TagIDs = make([]int64, 0, len(post.Tags))
	for _, tag := range post.Tags {
		post.TagIDs = append(post.TagIDs, tag.ID)
	}

	return post, err
}

func (r *SQLiteRepository) ListAdminPosts(limit int) ([]model.Post, error) {
	rows, err := r.DB.Query(`
		SELECT p.id, p.title, p.subtitle, p.slug, p.category_id, COALESCE(c.name, ''), p.cover_image, p.summary, p.content, p.content_markdown, p.type, p.status, p.is_top, p.is_recommend, p.game_version, p.server_line,
		       p.game_genre, p.region, p.official_url, p.download_url, p.qq_group, p.seo_title, p.seo_keywords,
		       p.seo_description, p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_categories c ON c.id = p.category_id
		ORDER BY p.updated_at DESC, p.id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Subtitle,
			&post.Slug,
			&post.CategoryID,
			&post.CategoryName,
			&post.CoverImage,
			&post.Summary,
			&post.Content,
			&post.ContentMarkdown,
			&post.Type,
			&post.Status,
			&post.IsTop,
			&post.IsRecommend,
			&post.GameVersion,
			&post.ServerLine,
			&post.GameGenre,
			&post.Region,
			&post.OfficialURL,
			&post.DownloadURL,
			&post.QQGroup,
			&post.SEOTitle,
			&post.SEOKeywords,
			&post.SEODescription,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.populatePostTags(posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *SQLiteRepository) GetAdminPostByID(id int64) (model.Post, error) {
	var post model.Post
	row := r.DB.QueryRow(`
		SELECT p.id, p.title, p.subtitle, p.slug, p.category_id, COALESCE(c.name, ''), p.cover_image, p.summary, p.content, p.content_markdown, p.type, p.status, p.is_top, p.is_recommend, p.game_version, p.server_line,
		       p.game_genre, p.region, p.official_url, p.download_url, p.qq_group, p.seo_title, p.seo_keywords,
		       p.seo_description, p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_categories c ON c.id = p.category_id
		WHERE p.id = ?
		LIMIT 1
	`, id)

	err := row.Scan(
		&post.ID,
		&post.Title,
		&post.Subtitle,
		&post.Slug,
		&post.CategoryID,
		&post.CategoryName,
		&post.CoverImage,
		&post.Summary,
		&post.Content,
		&post.ContentMarkdown,
		&post.Type,
		&post.Status,
		&post.IsTop,
		&post.IsRecommend,
		&post.GameVersion,
		&post.ServerLine,
		&post.GameGenre,
		&post.Region,
		&post.OfficialURL,
		&post.DownloadURL,
		&post.QQGroup,
		&post.SEOTitle,
		&post.SEOKeywords,
		&post.SEODescription,
		&post.PublishedAt,
		&post.CreatedAt,
		&post.UpdatedAt,
	)
	if err != nil {
		return post, err
	}
	post.Tags, err = r.ListTagsByPostID(post.ID)
	if err != nil {
		return post, err
	}
	post.TagIDs = make([]int64, 0, len(post.Tags))
	for _, tag := range post.Tags {
		post.TagIDs = append(post.TagIDs, tag.ID)
	}
	return post, err
}

func (r *SQLiteRepository) CreatePost(post model.Post) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		INSERT INTO posts (
			title, subtitle, slug, cover_image, summary, content, content_markdown, type, category_id, status,
			is_top, is_recommend, game_version, server_line, game_genre, region, official_url,
			download_url, qq_group, seo_title, seo_keywords, seo_description, published_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
	`,
		post.Title, post.Subtitle, post.Slug, post.CoverImage, post.Summary, post.Content, post.ContentMarkdown,
		post.Type, post.CategoryID, post.Status, post.IsTop, post.IsRecommend, post.GameVersion,
		post.ServerLine, post.GameGenre, post.Region, post.OfficialURL, post.DownloadURL,
		post.QQGroup, post.SEOTitle, post.SEOKeywords, post.SEODescription,
	)
	if err != nil {
		return err
	}

	postID, err := result.LastInsertId()
	if err != nil {
		return err
	}
	if err := r.replacePostTagsTx(tx, postID, post.TagIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *SQLiteRepository) UpdatePost(post model.Post) error {
	tx, err := r.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE posts SET
			title = ?, subtitle = ?, slug = ?, cover_image = ?, summary = ?, content = ?, content_markdown = ?,
			type = ?, category_id = ?, status = ?, is_top = ?, is_recommend = ?, game_version = ?,
			server_line = ?, game_genre = ?, region = ?, official_url = ?, download_url = ?,
			qq_group = ?, seo_title = ?, seo_keywords = ?, seo_description = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`,
		post.Title, post.Subtitle, post.Slug, post.CoverImage, post.Summary, post.Content, post.ContentMarkdown,
		post.Type, post.CategoryID, post.Status, post.IsTop, post.IsRecommend, post.GameVersion,
		post.ServerLine, post.GameGenre, post.Region, post.OfficialURL, post.DownloadURL,
		post.QQGroup, post.SEOTitle, post.SEOKeywords, post.SEODescription, post.ID,
	)
	if err != nil {
		return err
	}
	if err := r.replacePostTagsTx(tx, post.ID, post.TagIDs); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *SQLiteRepository) ListTagsByPostID(postID int64) ([]model.Tag, error) {
	rows, err := r.DB.Query(`
		SELECT t.id, t.name, t.slug, t.created_at
		FROM post_tags t
		INNER JOIN post_tag_relations rel ON rel.tag_id = t.id
		WHERE rel.post_id = ?
		ORDER BY t.id DESC
	`, postID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}

	return tags, rows.Err()
}

func (r *SQLiteRepository) ListPublishedPostsByTag(tagID int64, limit int) ([]model.Post, error) {
	rows, err := r.DB.Query(`
		SELECT p.id, p.title, p.subtitle, p.slug, p.category_id, COALESCE(c.name, ''), p.cover_image, p.summary, p.content, p.content_markdown, p.type, p.status, p.is_top, p.is_recommend, p.game_version, p.server_line,
		       p.game_genre, p.region, p.official_url, p.download_url, p.qq_group, p.seo_title, p.seo_keywords,
		       p.seo_description, p.published_at, p.created_at, p.updated_at
		FROM posts p
		LEFT JOIN post_categories c ON c.id = p.category_id
		INNER JOIN post_tag_relations rel ON rel.post_id = p.id
		WHERE rel.tag_id = ? AND p.status = 'published'
		ORDER BY p.is_top DESC, p.published_at DESC, p.id DESC
		LIMIT ?
	`, tagID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(
			&post.ID,
			&post.Title,
			&post.Subtitle,
			&post.Slug,
			&post.CategoryID,
			&post.CategoryName,
			&post.CoverImage,
			&post.Summary,
			&post.Content,
			&post.ContentMarkdown,
			&post.Type,
			&post.Status,
			&post.IsTop,
			&post.IsRecommend,
			&post.GameVersion,
			&post.ServerLine,
			&post.GameGenre,
			&post.Region,
			&post.OfficialURL,
			&post.DownloadURL,
			&post.QQGroup,
			&post.SEOTitle,
			&post.SEOKeywords,
			&post.SEODescription,
			&post.PublishedAt,
			&post.CreatedAt,
			&post.UpdatedAt,
		); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if err := r.populatePostTags(posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func (r *SQLiteRepository) populatePostTags(posts []model.Post) error {
	for i := range posts {
		tags, err := r.ListTagsByPostID(posts[i].ID)
		if err != nil {
			return err
		}
		posts[i].Tags = tags
	}
	return nil
}

func (r *SQLiteRepository) replacePostTagsTx(tx *sql.Tx, postID int64, tagIDs []int64) error {
	if _, err := tx.Exec(`DELETE FROM post_tag_relations WHERE post_id = ?`, postID); err != nil {
		return err
	}

	seen := make(map[int64]struct{}, len(tagIDs))
	for _, tagID := range tagIDs {
		if tagID <= 0 {
			continue
		}
		if _, ok := seen[tagID]; ok {
			continue
		}
		seen[tagID] = struct{}{}
		if _, err := tx.Exec(`
			INSERT INTO post_tag_relations (post_id, tag_id)
			VALUES (?, ?)
		`, postID, tagID); err != nil {
			return err
		}
	}

	return nil
}

func (r *SQLiteRepository) ListApprovedMessages(limit int) ([]model.GuestbookMessage, error) {
	rows, err := r.DB.Query(`
		SELECT id, nickname, contact, content, ip, status, reply_content, created_at, updated_at
		FROM guestbook_messages
		WHERE status = 'approved'
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.GuestbookMessage
	for rows.Next() {
		var message model.GuestbookMessage
		if err := rows.Scan(
			&message.ID,
			&message.Nickname,
			&message.Contact,
			&message.Content,
			&message.IP,
			&message.Status,
			&message.ReplyContent,
			&message.CreatedAt,
			&message.UpdatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (r *SQLiteRepository) CreateGuestbookMessage(nickname, contact, content, ip string) error {
	nickname = strings.TrimSpace(nickname)
	contact = strings.TrimSpace(contact)
	content = strings.TrimSpace(content)

	_, err := r.DB.Exec(`
		INSERT INTO guestbook_messages (nickname, contact, content, ip, status)
		VALUES (?, ?, ?, ?, 'pending')
	`, nickname, contact, content, ip)

	return err
}

func (r *SQLiteRepository) CountPendingMessages() (int, error) {
	var total int
	err := r.DB.QueryRow(`SELECT COUNT(1) FROM guestbook_messages WHERE status = 'pending'`).Scan(&total)
	return total, err
}

func (r *SQLiteRepository) ListAdminMessages(limit int) ([]model.GuestbookMessage, error) {
	rows, err := r.DB.Query(`
		SELECT id, nickname, contact, content, ip, status, reply_content, created_at, updated_at
		FROM guestbook_messages
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []model.GuestbookMessage
	for rows.Next() {
		var message model.GuestbookMessage
		if err := rows.Scan(
			&message.ID,
			&message.Nickname,
			&message.Contact,
			&message.Content,
			&message.IP,
			&message.Status,
			&message.ReplyContent,
			&message.CreatedAt,
			&message.UpdatedAt,
		); err != nil {
			return nil, err
		}
		messages = append(messages, message)
	}

	return messages, rows.Err()
}

func (r *SQLiteRepository) UpdateGuestbookStatus(id int64, status, replyContent string) error {
	_, err := r.DB.Exec(`
		UPDATE guestbook_messages
		SET status = ?, reply_content = ?, updated_at = CURRENT_TIMESTAMP
		WHERE id = ?
	`, status, strings.TrimSpace(replyContent), id)
	return err
}

func (r *SQLiteRepository) CreateUpload(upload model.Upload) error {
	_, err := r.DB.Exec(`
		INSERT INTO uploads (origin_name, saved_name, path, mime_type, size)
		VALUES (?, ?, ?, ?, ?)
	`, upload.OriginName, upload.SavedName, upload.Path, upload.MimeType, upload.Size)
	return err
}

func (r *SQLiteRepository) ListUploads(limit int) ([]model.Upload, error) {
	rows, err := r.DB.Query(`
		SELECT id, origin_name, saved_name, path, mime_type, size, created_at
		FROM uploads
		ORDER BY id DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var uploads []model.Upload
	for rows.Next() {
		var upload model.Upload
		if err := rows.Scan(
			&upload.ID,
			&upload.OriginName,
			&upload.SavedName,
			&upload.Path,
			&upload.MimeType,
			&upload.Size,
			&upload.CreatedAt,
		); err != nil {
			return nil, err
		}
		uploads = append(uploads, upload)
	}

	return uploads, rows.Err()
}

func (r *SQLiteRepository) GetUploadByID(id int64) (model.Upload, error) {
	var upload model.Upload
	row := r.DB.QueryRow(`
		SELECT id, origin_name, saved_name, path, mime_type, size, created_at
		FROM uploads
		WHERE id = ?
		LIMIT 1
	`, id)

	err := row.Scan(
		&upload.ID,
		&upload.OriginName,
		&upload.SavedName,
		&upload.Path,
		&upload.MimeType,
		&upload.Size,
		&upload.CreatedAt,
	)
	return upload, err
}

func (r *SQLiteRepository) DeleteUploadByID(id int64) error {
	_, err := r.DB.Exec(`DELETE FROM uploads WHERE id = ?`, id)
	return err
}

func (r *SQLiteRepository) CountPostReferencesByUploadURL(url string) (int, error) {
	var total int
	err := r.DB.QueryRow(`
		SELECT COUNT(1)
		FROM posts
		WHERE cover_image = ? OR content LIKE ?
	`, url, "%"+url+"%").Scan(&total)
	return total, err
}

func SaveUploadedFile(src io.Reader, targetPath string) error {
	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return err
	}

	dst, err := os.Create(targetPath)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
