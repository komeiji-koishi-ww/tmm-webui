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

- 按 TMM 的 data source 思路添加媒体库。
- 一个媒体库支持多个数据源路径。
- 支持电影和电视剧两类媒体库。
- WebUI 目录选择器，不需要手动记路径。
- 递归扫描视频文件，并持久化扫描结果。
- 扫描任务后台运行，切换媒体库不会中断当前扫描。
- 扫描时按发现顺序实时加入列表，同时批量落盘，避免大库每个文件一次磁盘事务。
- 重启后直接加载上次扫描结果，不要求每次打开都重新扫描。
- TMM 风格跳过目录：`@eaDir`、`Plex Versions`、`CERTIFICATE`、回收站、隐藏目录等。
- TMM 风格媒体文件分类：`VIDEO`、`TRAILER`、`SAMPLE`、`SUBTITLE`、`NFO`、`POSTER`、`FANART` 等。
- 电视剧季集解析覆盖 `S01E01`、`1x02`、`102`、纯数字集、日期集和季目录推断。
- 从文件名猜测标题和年份。
- 使用 TMDb 搜索候选。
- 写入 `movie.nfo`。
- 下载 `poster.jpg` 和 `fanart.jpg`。
- 生成重命名预览。
- 执行文件重命名。
- Go 单二进制提供 API 和 Vue 静态页面。
- WebUI 采用 TMM 习惯的电影/电视剧模块切换、数据源列表、表格列表和详情操作区。

当前持久化使用嵌入式 bbolt 数据库：

- `tmmweb.db`：媒体库、扫描条目、任务记录。
- buckets：`libraries`、`items`、`tasks`。

这对应 TMM 的核心策略：先配置电影/电视剧 data source，执行 update data sources 扫描导入内部库，之后 UI 主要从本地库读取。TMM 源码里使用 H2 MVStore/MVMap + Jackson 序列化保存实体；本项目用 bbolt 的 bucket/kv 结构实现相近的嵌入式持久化模型。旧版 `libraries.json` / `items.json` 会在首次启动时自动迁移进 `tmmweb.db`。

## 后续需要补齐的 TMM 功能

- 完整电视剧 `tvshow.nfo`、season NFO、episode NFO。
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
- `GET /api/items?libraryId=...`
- `GET /api/browse?path=...`
- `POST /api/scan`
- `GET /api/search?itemId=...`
- `POST /api/scrape`
- `POST /api/rename/preview`
- `POST /api/rename/apply`
