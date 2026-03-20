PRAGMA journal_mode = WAL;

CREATE TABLE IF NOT EXISTS admins (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  nickname TEXT NOT NULL DEFAULT '',
  last_login_at TEXT NOT NULL DEFAULT '',
  last_login_ip TEXT NOT NULL DEFAULT '',
  status INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS admin_login_logs (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  admin_id INTEGER NOT NULL DEFAULT 0,
  ip TEXT NOT NULL DEFAULT '',
  user_agent TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS post_categories (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  parent_id INTEGER NOT NULL DEFAULT 0,
  sort INTEGER NOT NULL DEFAULT 0,
  seo_title TEXT NOT NULL DEFAULT '',
  seo_keywords TEXT NOT NULL DEFAULT '',
  seo_description TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  title TEXT NOT NULL,
  subtitle TEXT NOT NULL DEFAULT '',
  slug TEXT NOT NULL UNIQUE,
  cover_image TEXT NOT NULL DEFAULT '',
  summary TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL DEFAULT '',
  content_markdown TEXT NOT NULL DEFAULT '',
  type TEXT NOT NULL DEFAULT 'game',
  category_id INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'draft',
  is_top INTEGER NOT NULL DEFAULT 0,
  is_recommend INTEGER NOT NULL DEFAULT 0,
  game_version TEXT NOT NULL DEFAULT '',
  server_line TEXT NOT NULL DEFAULT '',
  game_genre TEXT NOT NULL DEFAULT '',
  region TEXT NOT NULL DEFAULT '',
  official_url TEXT NOT NULL DEFAULT '',
  download_url TEXT NOT NULL DEFAULT '',
  qq_group TEXT NOT NULL DEFAULT '',
  seo_title TEXT NOT NULL DEFAULT '',
  seo_keywords TEXT NOT NULL DEFAULT '',
  seo_description TEXT NOT NULL DEFAULT '',
  view_count INTEGER NOT NULL DEFAULT 0,
  comment_count INTEGER NOT NULL DEFAULT 0,
  published_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_posts_status ON posts(status);
CREATE INDEX IF NOT EXISTS idx_posts_category_id ON posts(category_id);
CREATE INDEX IF NOT EXISTS idx_posts_published_at ON posts(published_at DESC);

CREATE TABLE IF NOT EXISTS post_tags (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  slug TEXT NOT NULL UNIQUE,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS post_tag_relations (
  post_id INTEGER NOT NULL,
  tag_id INTEGER NOT NULL,
  PRIMARY KEY (post_id, tag_id)
);

CREATE TABLE IF NOT EXISTS guestbook_messages (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  nickname TEXT NOT NULL,
  contact TEXT NOT NULL DEFAULT '',
  content TEXT NOT NULL,
  ip TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'pending',
  reply_content TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_guestbook_status ON guestbook_messages(status);
CREATE INDEX IF NOT EXISTS idx_guestbook_created_at ON guestbook_messages(created_at DESC);

CREATE TABLE IF NOT EXISTS site_settings (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  site_name TEXT NOT NULL,
  site_title TEXT NOT NULL DEFAULT '',
  site_keywords TEXT NOT NULL DEFAULT '',
  site_description TEXT NOT NULL DEFAULT '',
  logo TEXT NOT NULL DEFAULT '',
  footer_text TEXT NOT NULL DEFAULT '',
  contact_info TEXT NOT NULL DEFAULT '',
  icp_no TEXT NOT NULL DEFAULT '',
  analytics_code TEXT NOT NULL DEFAULT '',
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS friendly_links (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  sort INTEGER NOT NULL DEFAULT 0,
  status INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS ad_slots (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  image_url TEXT NOT NULL DEFAULT '',
  target_url TEXT NOT NULL DEFAULT '',
  sort INTEGER NOT NULL DEFAULT 0,
  status INTEGER NOT NULL DEFAULT 1,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS uploads (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  origin_name TEXT NOT NULL,
  saved_name TEXT NOT NULL,
  path TEXT NOT NULL,
  mime_type TEXT NOT NULL DEFAULT '',
  size INTEGER NOT NULL DEFAULT 0,
  created_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO site_settings (site_name, site_title, site_keywords, site_description, footer_text, contact_info)
SELECT
  '传奇私服游戏门户',
  '传奇私服游戏门户 - 游戏发布与玩家社区',
  '传奇私服,传奇发布网,传奇开服表,传奇攻略',
  '轻量、SEO 友好的传奇私服游戏门户站点骨架。',
  'Copyright 2026 传奇私服游戏门户',
  '站长QQ：123456'
WHERE NOT EXISTS (SELECT 1 FROM site_settings WHERE id = 1);

INSERT INTO admins (username, password_hash, nickname)
SELECT 'admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', '站点管理员'
WHERE NOT EXISTS (SELECT 1 FROM admins WHERE username = 'admin');

INSERT INTO post_categories (name, slug, sort, seo_title, seo_keywords, seo_description)
SELECT '传奇发布', 'legend-release', 10, '传奇发布', '传奇发布,传奇私服', '最新传奇私服发布信息'
WHERE NOT EXISTS (SELECT 1 FROM post_categories WHERE slug = 'legend-release');

INSERT INTO posts (
  title, subtitle, slug, summary, content, content_markdown, type, category_id, status,
  is_top, is_recommend, game_version, server_line, game_genre, region,
  official_url, download_url, qq_group, seo_title, seo_keywords, seo_description
)
SELECT
  '经典传奇三区今日首发',
  '轻量门户骨架示例文章',
  'classic-legend-launch',
  '这是系统初始化时写入的一篇演示内容，后续可直接在后台替换。',
  '<p>欢迎使用基于 Go + Echo + html/template + SQLite 的轻量门户站点骨架。</p><p>这一版重点是低资源、SEO 友好和后台可运营。</p>',
  '欢迎使用基于 Go + Echo + html/template + SQLite 的轻量门户站点骨架。',
  'game',
  1,
  'published',
  1,
  1,
  '1.76',
  '三区',
  '复古',
  '华东',
  'https://example.com',
  'https://example.com/download',
  '12345678',
  '经典传奇三区今日首发',
  '经典传奇,传奇首发,传奇三区',
  '经典传奇三区今日首发示例内容'
WHERE NOT EXISTS (SELECT 1 FROM posts WHERE slug = 'classic-legend-launch');
