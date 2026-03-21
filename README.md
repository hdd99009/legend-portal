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

## 生产部署

推荐部署方式：`GitHub Actions + SSH 自动发布 + systemd + Nginx`

### 服务器目录

约定生产环境使用：

- 代码目录：`/srv/legend-portal/app`
- 共享目录：`/srv/legend-portal/shared`
- SQLite：`/srv/legend-portal/shared/data/site.db`
- 上传目录：`/srv/legend-portal/shared/uploads`
- 运行配置：`/srv/legend-portal/shared/config.production.yaml`

### 应用配置

程序支持通过环境变量 `APP_CONFIG_PATH` 指定配置文件。

- 本地默认仍使用：`configs/config.yaml`
- 生产模板已提供：`configs/config.production.yaml`

首次部署时，`deploy/deploy.sh` 会在共享目录里自动生成一份配置文件模板：

`/srv/legend-portal/shared/config.production.yaml`

请在服务器上至少修改这两项：

- `app.base_url`
- `app.session_secret`

### 首次部署准备

1. 安装依赖：

```bash
sudo apt update
sudo apt install -y git nginx curl
```

2. 安装 Go，并确认服务器能执行：

```bash
go version
```

3. 准备目录并拉代码：

```bash
sudo mkdir -p /srv/legend-portal
sudo chown -R $USER:$USER /srv/legend-portal
git clone https://github.com/hdd99009/legend-portal.git /srv/legend-portal/app
chmod +x /srv/legend-portal/app/deploy/deploy.sh
```

4. 安装 systemd 服务：

```bash
sudo cp /srv/legend-portal/app/deploy/legend-portal.service /etc/systemd/system/legend-portal.service
sudo systemctl daemon-reload
sudo systemctl enable legend-portal
```

5. 安装 Nginx 配置：

```bash
sudo cp /srv/legend-portal/app/deploy/nginx.conf /etc/nginx/sites-available/legend-portal
sudo ln -sf /etc/nginx/sites-available/legend-portal /etc/nginx/sites-enabled/legend-portal
sudo nginx -t
sudo systemctl reload nginx
```

6. 首次手动发布一次：

```bash
cd /srv/legend-portal/app
DEPLOY_PATH=/srv/legend-portal/app bash deploy/deploy.sh
```

查看状态：

```bash
sudo systemctl status legend-portal
curl http://127.0.0.1:8080/healthz
```

### GitHub Actions 自动发布

仓库已包含工作流：

- `.github/workflows/deploy.yml`

你需要在 GitHub 仓库里配置这些 Secrets：

- `SERVER_HOST`
- `SERVER_PORT`
- `SERVER_USER`
- `SERVER_SSH_KEY`
- `DEPLOY_PATH`

建议值：

- `DEPLOY_PATH=/srv/legend-portal/app`
- `SERVER_PORT=22`
- `SERVER_USER=root` 或具备 `systemctl` 权限的部署用户

配置完成后，每次 push 到 `main`，GitHub Actions 都会：

1. SSH 登录服务器
2. 进入 `/srv/legend-portal/app`
3. 执行 `deploy/deploy.sh`
4. 拉最新代码、编译、重启服务
5. 检查 `http://127.0.0.1:8080/healthz`

### 常用运维命令

查看服务状态：

```bash
sudo systemctl status legend-portal
```

查看运行日志：

```bash
sudo journalctl -u legend-portal -n 100 --no-pager
```

手动重启：

```bash
sudo systemctl restart legend-portal
```

手动发布：

```bash
cd /srv/legend-portal/app
DEPLOY_PATH=/srv/legend-portal/app bash deploy/deploy.sh
```
