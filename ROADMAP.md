# Roadmap

## TMM 功能拆解

### 已覆盖的 MVP

- 电影库路径配置。
- 电影/电视剧媒体库类型。
- 多 data source 路径配置。
- Web 目录选择器。
- 递归扫描视频文件。
- 扫描结果 bbolt 嵌入式持久化。
- 扫描后台任务与进度展示。
- 扫描时发现一条即时进入内存列表，并批量写入数据库。
- 重启后从缓存加载扫描结果。
- TMM 风格 skip folders。
- TMM 风格媒体文件分类基础版。
- TMM 风格电视剧季集解析基础版：`S01E01`、`1x02`、`102`、纯数字集、日期集、季目录推断。
- TMM 习惯 UI 基础版：电影/电视剧模块、data source 列表、表格、详情操作区。
- 文件名标题/年份猜测。
- TMDb 电影搜索。
- TMDb 电影详情刮削。
- Kodi 风格 `movie.nfo` 写入。
- `poster.jpg`、`fanart.jpg` 下载。
- 重命名预览。
- 重命名执行。
- Go API + Vue WebUI。

### P0：让它能替代 TMM 的核心日常流程

- 完善 bbolt 存储模型，保留 media entity、data source、scan state、scrape state。
- 任务队列：扫描已有后台任务和基础进度；刮削、图片下载、重命名还要补进度、取消、失败重试。
- 增量扫描：
  - 保留上次扫描快照。
  - 新增/删除/移动文件差异更新。
  - 跳过 `.tmmignore` 目录。
  - 记录每个 data source 的最后扫描时间。
- 更强文件名解析：
  - 电影标题里的年份数字不能误判，例如 `Blade Runner 2049 (2017)`。
  - 识别 edition、resolution、source、codec、audio、release group。
  - 支持中文片名、英文片名、点/下划线/括号混合。
- 批量刮削：
  - 自动选择最高相似度候选。
  - 低置信度进入人工确认队列。
- 批量重命名：
  - dry-run。
  - 冲突检测。
  - 回滚记录。
- 本地 NFO 读取：
  - 识别已有 `movie.nfo`。
  - 从 NFO 回填 UI。
- 图片命名规则：
  - Kodi/Emby/Jellyfin 常见命名。
  - `poster.jpg`、`folder.jpg`、`fanart.jpg`、`clearlogo.png`。

### P1：电视剧支持

- TV show datasource 已有基础类型，继续补齐 TMM 级别模型。
- 剧集文件解析：`S01E01`、`1x01`、纯数字集、日期集、多集、特别篇。
- 动漫特殊规则：OVA/OAV/SP、hash、Specials 目录。
- `tvshow.nfo`。
- season NFO 和 season poster。
- episode NFO 和 episode thumb。
- TVDb/TMDb TV scraper。

### P2：TMM 高级能力

- MediaInfo 扫描。
- 字幕搜索/下载。
- 预告片下载。
- 演员图片下载。
- 合集/movie set 管理。
- 标签、评分、认证分级。
- 导出模板：HTML、CSV、XML。
- 后处理命令。
- 多语言和多刮削器优先级。
- Web 认证和权限。

## 技术边界

TMM 原项目是 Java/Swing 桌面应用。这个项目不会直接复用 Swing UI 或 Java 内部类，而是按功能重新实现：

- Go 后端负责文件系统、任务、刮削、NFO、重命名。
- Vue 前端负责扫描结果、候选确认、批量操作、设置。
- Docker 只暴露 HTTP WebUI，不需要 X11/VNC/noVNC。
