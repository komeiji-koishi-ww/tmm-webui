# TMM WebUI

一个以 tinyMediaManager 功能面为参考、用 Go 后端和 Vue WebUI 重写的轻量媒体管理服务。目标是替代“桌面版 TMM + VNC”的使用方式，让 NAS 上可以直接运行 Web 服务。

## 对 TMM 的源码/容器调研

TMM 官方源码在 GitLab，整体是 Java/Maven 单体应用：

- `core/movie`、`core/tvshow`：电影/剧集实体、列表、扫描、改数据源、重命名、刮削任务。
- `core/movie/connector`、`core/tvshow/connector`：Kodi、Emby、Jellyfin、XBMC 等 NFO 读写 connector。
- `core/movie/filenaming`、`core/tvshow/*FileNaming`：海报、fanart、NFO、预告片等落盘命名规则。
- `scraper/*`：TMDb、TVDb、IMDb、OMDb、fanart.tv、AniDB、OpenSubtitles、Trakt 等 provider。
- `core/mediainfo`：MediaInfo 集成。
- `core/tasks`：下载、字幕、预告片、图片缓存、导出、队列任务。
- `templates`：HTML、CSV、XML、M3U 等导出模板。
- `ui/*`：Swing 桌面 UI、表格、对话框、批量编辑、设置页。

常见 TMM Docker 镜像不是服务端版，而是：

- `jlesage/baseimage-gui`
- 下载 TMM Linux tar 包
- 运行 Java Swing 桌面程序
- 用 VNC/noVNC 暴露 GUI

所以“脱离 VNC 的 WebUI 版”不能简单改 Dockerfile，需要重新实现服务端和 Web 前端。

## 当前 MVP 功能

- 添加电影媒体库路径。
- 递归扫描视频文件。
- 从文件名猜测标题和年份。
- 使用 TMDb 搜索候选。
- 写入 `movie.nfo`。
- 下载 `poster.jpg` 和 `fanart.jpg`。
- 生成重命名预览。
- 执行文件重命名。
- Go 单二进制提供 API 和 Vue 静态页面。

## 后续需要补齐的 TMM 功能

- 电视剧/季/集扫描和 `tvshow.nfo`、episode NFO。
- 多数据源：TVDb、IMDb、OMDb、fanart.tv、AniDB。
- 本地 NFO 反向解析。
- 本地图片命名规则兼容 Kodi/Emby/Jellyfin。
- MediaInfo。
- 字幕搜索和下载。
- 预告片下载。
- 导出模板。
- 批量编辑、标签、合集、演员图。
- 任务队列、进度、取消、历史日志。
- 权限和用户认证。

## 运行

需要 TMDb API Key：

```bash
export TMDB_API_KEY=your_key
go run ./cmd/tmmweb
```

打开：

```text
http://localhost:8080
```

Docker：

```bash
TMDB_API_KEY=your_key docker compose up -d --build
```

## API 摘要

- `GET /api/libraries`
- `POST /api/libraries`
- `POST /api/scan`
- `GET /api/search?itemId=...`
- `POST /api/scrape`
- `POST /api/rename/preview`
- `POST /api/rename/apply`

