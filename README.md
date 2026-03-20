# 传奇私服游戏门户骨架

轻量门户方案：`Go + Echo + html/template + SQLite + Bootstrap + Nginx`

## 已包含

- 首页
- 游戏发布详情页
- 玩家留言板
- 后台入口页
- SQLite 自动初始化
- `robots.txt`
- `sitemap.xml`

## 本地启动

```bash
env GOCACHE=/tmp/go-build GOMODCACHE=/tmp/go-mod-cache go run ./cmd/web
```

默认地址：

- 前台首页：`http://127.0.0.1:8080/`
- 留言板：`http://127.0.0.1:8080/guestbook`
- 后台入口：`http://127.0.0.1:8080/admin`
- 后台登录：`http://127.0.0.1:8080/admin/login`
- 文章管理：`http://127.0.0.1:8080/admin/posts`

## 初始化数据

首次启动会自动执行 `/migrations/001_init.sql`，并写入：

- 默认站点配置
- 演示分类
- 演示文章
- 默认管理员

默认管理员账号：

- 用户名：`admin`
- 密码：`password`

## 数据文件

- SQLite 数据库：`/Users/huangdandan/Documents/New project/storage/data/site.db`
- 上传目录：`/Users/huangdandan/Documents/New project/assets/uploads`

## 下一步建议

1. 做后台登录和 Session 鉴权
2. 做留言审核页
3. 做站点设置和 SEO 编辑页
4. 做文件上传和封面管理
5. 把 Bootstrap CDN 改成本地静态文件
