# 传奇私服游戏门户骨架

轻量门户方案：`Go + Echo + html/template + SQLite + Bootstrap + Nginx`

## 已包含

- 首页
- 游戏发布详情页
- 玩家留言板
- 后台入口页
- 分类管理
- 留言审核
- 后台图片上传
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
- 分类管理：`http://127.0.0.1:8080/admin/categories`
- 文章管理：`http://127.0.0.1:8080/admin/posts`
- 图片管理：`http://127.0.0.1:8080/admin/uploads`
- 留言审核：`http://127.0.0.1:8080/admin/messages`
- 站点设置：`http://127.0.0.1:8080/admin/settings`
- 修改密码：`http://127.0.0.1:8080/admin/password`

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
- 上传目录：`/Users/huangdandan/Documents/New project/storage/uploads`

## 存储配置

- 当前使用 `local` 本地存储
- 配置文件位置：`/Users/huangdandan/Documents/New project/configs/config.yaml`
- 已抽象存储接口，后续可切换到 OSS / S3 / R2

## 图片上传说明

- 文章编辑页支持直接上传图片
- 上传成功后返回本地 URL，可直接粘贴进正文
- 文章支持单独设置封面图
- 图片通过 `/uploads/...` 访问
- 已加基础 Referer 防盗链，外站带来源引用会被拦截

## 下一步建议

1. 做标签管理
2. 做首页栏目配置和推荐位管理
3. 给后台补登录日志和操作日志
4. 把 Bootstrap CDN 改成本地静态文件
5. 补评论频控和敏感词过滤
