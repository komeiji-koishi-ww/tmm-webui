const { createApp } = Vue;

createApp({
  data() {
    return {
      libraries: [],
      activeModule: "movie",
      selectedLibrary: null,
      items: [],
      selectedItem: null,
      candidates: [],
      scrapeSearch: {
        query: "",
        year: "",
      },
      rename: {
        pattern: "{title} ({year}) {tmdb-{tmdbid}}",
        title: "",
        year: "",
        tmdbId: 0,
      },
      renamePreview: null,
      newLibrary: {
        name: "电影",
        paths: [],
        type: "movie",
      },
      scraperSettings: {
        tmdbApiKey: "",
        tmdbConfigured: false,
        tmdbKeySource: "none",
        proxyEnabled: false,
        proxyHost: "",
        proxyPort: 7890,
        proxyUsername: "",
        proxyPassword: "",
        proxyPasswordConfigured: false,
        clearProxyPassword: false,
        movieScrapeMetadata: true,
        movieScrapeNfo: true,
        movieScrapeImages: true,
        movieScrapeOverwrite: false,
        tvShowScrapeMetadata: true,
        tvShowEpisodeMetadata: true,
        tvShowScrapeNfo: true,
        tvShowScrapeImages: true,
        tvShowScrapeOverwrite: false,
        movieRenamerPathname: "{title} ({year})",
        movieRenamerFilename: "{title} ({year})",
        tvShowRenamerShowFolder: "{showTitle}",
        tvShowRenamerSeason: "Season {seasonNr2}",
        tvShowRenamerFilename: "{showTitle} - S{seasonNr2}E{episodeNr2} - {title}",
        moviePosterName: "poster.jpg",
        movieFanartName: "fanart.jpg",
        moviePosterNames: "poster.jpg\nfolder.jpg\n{filename}-poster.jpg",
        movieFanartNames: "fanart.jpg\n{filename}-fanart.jpg",
        tvShowPosterName: "poster.jpg",
        tvShowFanartName: "fanart.jpg",
        tvShowPosterNames: "poster.jpg\nfolder.jpg",
        tvShowFanartNames: "fanart.jpg\nbackdrop.jpg",
        saving: false,
      },
      settingsOpen: false,
      settingsFilter: "",
      settingsActiveSection: "movie-source",
      pendingPath: "",
      browser: {
        open: false,
        path: "/Volumes",
        parent: "/",
        entries: [],
      },
      query: "",
      filters: [],
      filterEditor: {
        open: false,
        tab: "details",
      },
      layout: {
        browserWidth: 0,
        filterNavWidth: 180,
        resizing: null,
      },
      sortKey: "title",
      sortDirection: "asc",
      detailTab: "info",
      selectedEntity: null,
      expandedShows: {},
      expandedSeasons: {},
      contextMenu: {
        open: false,
        x: 0,
        y: 0,
        scope: "",
        payload: null,
      },
      chooser: {
        open: false,
        scope: "",
        mediaType: "movie",
        targetItem: null,
        targetShow: null,
        targetSeason: null,
        path: "",
        query: "",
        year: "",
        candidates: [],
        selected: null,
        detail: null,
        loading: false,
        scraping: false,
        error: "",
        options: {
          metadata: true,
          nfo: true,
          artwork: true,
          overwrite: false,
          showMetadata: true,
          episodeMetadata: true,
        },
      },
      tasks: {},
      poller: null,
      status: "正在初始化",
      busy: false,
    };
  },
  computed: {
    filteredLibraries() {
      return this.libraries.filter((library) => library.type === this.activeModule);
    },
    moduleTitle() {
      return this.activeModule === "tvshow" ? "电视剧" : "电影";
    },
    renamePreviewText() {
      if (!this.renamePreview) return "";
      return `文件:\n${this.renamePreview.sourceFile}\n=>\n${this.renamePreview.targetFile}\n\n目录:\n${this.renamePreview.sourceDir}\n=>\n${this.renamePreview.targetDir}`;
    },
    selectedTypeText() {
      if (!this.selectedLibrary) return "";
      return this.selectedLibrary.type === "tvshow" ? "电视剧" : "电影";
    },
    tmdbStatusText() {
      if (!this.scraperSettings.tmdbConfigured) return "未配置";
      if (this.scraperSettings.tmdbKeySource === "environment") return "已通过环境变量配置";
      if (this.scraperSettings.tmdbKeySource === "settings") return "已配置";
      return "已启用";
    },
    proxyStatusText() {
      if (!this.scraperSettings.proxyEnabled) return "未启用代理";
      const host = this.scraperSettings.proxyHost || "未设置主机";
      const port = this.scraperSettings.proxyPort || 80;
      return `HTTP 代理 ${host}:${port}`;
    },
    settingsSections() {
      return [
        {
          id: "general",
          title: "通用",
          children: [
            { id: "general-system", title: "系统 / 代理" },
            { id: "general-services", title: "外部服务" },
          ],
        },
        {
          id: "movie",
          title: "电影",
          children: [
            { id: "movie-ui", title: "UI" },
            { id: "movie-source", title: "数据源" },
            { id: "movie-nfo", title: "NFO" },
            {
              id: "movie-scraper",
              title: "元数据刮削",
              children: [{ id: "movie-scraper-options", title: "高级选项" }],
            },
            {
              id: "movie-artwork",
              title: "图片",
              children: [
                { id: "movie-artwork-options", title: "高级选项" },
                { id: "movie-artwork-naming", title: "图片命名" },
              ],
            },
            { id: "movie-renamer", title: "重命名" },
          ],
        },
        {
          id: "tvshow",
          title: "电视剧",
          children: [
            { id: "tvshow-ui", title: "UI" },
            { id: "tvshow-source", title: "数据源" },
            { id: "tvshow-nfo", title: "NFO" },
            {
              id: "tvshow-scraper",
              title: "元数据刮削",
              children: [{ id: "tvshow-scraper-options", title: "高级选项" }],
            },
            {
              id: "tvshow-artwork",
              title: "图片",
              children: [
                { id: "tvshow-artwork-options", title: "高级选项" },
                { id: "tvshow-artwork-naming", title: "图片命名" },
              ],
            },
            { id: "tvshow-renamer", title: "重命名" },
          ],
        },
      ];
    },
    activeSettingsTitle() {
      const node = this.findSettingsNode(this.settingsActiveSection);
      return node ? node.title : "设置";
    },
    selectedTask() {
      if (!this.selectedLibrary) return null;
      return this.tasks[this.selectedLibrary.id] || null;
    },
    selectedScanning() {
      return this.selectedTask && this.selectedTask.state === "running";
    },
    workbenchStyle() {
      if (!this.layout.browserWidth) return {};
      return {
        gridTemplateColumns: `minmax(420px, ${this.layout.browserWidth}px) 6px minmax(300px, 1fr)`,
      };
    },
    filterEditorStyle() {
      return {
        gridTemplateColumns: `${this.layout.filterNavWidth}px 6px minmax(0, 1fr)`,
      };
    },
    filterGroups() {
      const groups = [
        {
          id: "details",
          label: "详情",
          filters: [
            { id: "allInOne", label: "全部字段", type: "text", available: true },
            { id: "title", label: "片名", type: "text", available: true },
            { id: "originalTitle", label: "原始片名", type: "text", available: true },
            { id: "datasource", label: "数据源", type: "choice", available: true },
            { id: "dateAdded", label: "添加日期", type: "date", available: true },
            { id: "filename", label: "文件名", type: "text", available: true },
            { id: "path", label: "路径", type: "text", available: true },
            { id: "new", label: this.activeModule === "tvshow" ? "新剧集" : "新电影", type: "boolean", available: true },
            { id: "duplicate", label: this.activeModule === "tvshow" ? "重复剧集" : "重复电影", type: "boolean", available: false },
            { id: "watched", label: "已观看", type: "boolean", available: false },
            { id: "locked", label: "已锁定", type: "boolean", available: false },
          ],
        },
        {
          id: "metadata",
          label: "元数据",
          filters: [
            { id: "year", label: "年份", type: "number", available: true },
            { id: "decade", label: "年代", type: "choice", available: true },
            { id: "rating", label: "评分", type: "number", available: true },
            { id: "runtime", label: "片长", type: "number", available: this.activeModule === "movie" },
            { id: "genre", label: "类型", type: "choice", available: true },
            { id: "tmdbId", label: "TMDb ID", type: "number", available: true },
            { id: "imdbId", label: "IMDb ID", type: "text", available: true },
            { id: "missingMetadata", label: "缺少元数据", type: "boolean", available: true },
            { id: "missingArtwork", label: "缺少图片", type: "boolean", available: true },
            { id: "missingSubtitles", label: "缺少字幕", type: "boolean", available: true },
            { id: "cast", label: "演员", type: "choice", available: false },
            { id: "country", label: "国家", type: "choice", available: false },
            { id: "certification", label: "分级", type: "choice", available: false },
            { id: "tag", label: "标签", type: "choice", available: false },
            { id: "note", label: "备注", type: "text", available: false },
            { id: "episodeCount", label: "集数", type: "number", available: this.activeModule === "tvshow" },
            { id: "seasonCount", label: "季数", type: "number", available: this.activeModule === "tvshow" },
          ],
        },
        {
          id: "video",
          label: "视频",
          filters: [
            { id: "videoFormat", label: "视频格式", type: "choice", available: false },
            { id: "videoCodec", label: "视频编码", type: "choice", available: false },
            { id: "videoBitrate", label: "视频码率", type: "number", available: false },
            { id: "videoBitdepth", label: "视频位深", type: "number", available: false },
            { id: "videoContainer", label: "容器", type: "choice", available: false },
            { id: "aspectRatio", label: "宽高比", type: "choice", available: false },
            { id: "frameRate", label: "帧率", type: "number", available: false },
            { id: "hdrFormat", label: "HDR", type: "choice", available: false },
            { id: "videoFilesize", label: "文件大小", type: "number", available: false },
          ],
        },
        {
          id: "audio",
          label: "音频",
          filters: [
            { id: "audioCodec", label: "音频编码", type: "choice", available: false },
            { id: "audioChannels", label: "声道", type: "choice", available: false },
            { id: "audioLanguage", label: "音频语言", type: "choice", available: false },
            { id: "audioTitle", label: "音轨标题", type: "text", available: false },
            { id: "audioStreamCount", label: "音轨数量", type: "number", available: false },
          ],
        },
        {
          id: "subtitles",
          label: "字幕",
          filters: [
            { id: "subtitleCount", label: "字幕数量", type: "number", available: false },
            { id: "subtitleFormat", label: "字幕格式", type: "choice", available: false },
            { id: "subtitleLanguage", label: "字幕语言", type: "choice", available: false },
          ],
        },
        {
          id: "artwork",
          label: "图片",
          filters: [
            { id: "poster", label: "海报", type: "boolean", available: true },
            { id: "fanart", label: "同人画", type: "boolean", available: true },
            { id: "posterSize", label: "海报尺寸", type: "number", available: false },
            { id: "fanartSize", label: "同人画尺寸", type: "number", available: false },
            { id: "bannerSize", label: "横幅尺寸", type: "number", available: false },
            { id: "clearLogoSize", label: "ClearLogo 尺寸", type: "number", available: false },
            { id: "discArtSize", label: "DiscArt 尺寸", type: "number", available: false },
          ],
        },
      ];
      if (this.activeModule === "movie") {
        groups[1].filters.push({ id: "inMovieSet", label: "属于合集", type: "boolean", available: false });
      } else {
        groups[1].filters.push({ id: "status", label: "剧集状态", type: "choice", available: false });
        groups[1].filters.push({ id: "missingEpisodes", label: "缺少集", type: "boolean", available: false });
        groups[1].filters.push({ id: "uncategorizedEpisodes", label: "未分类集", type: "boolean", available: false });
      }
      return groups;
    },
    filterDefinitions() {
      const map = new Map();
      for (const group of this.filterGroups) {
        for (const filter of group.filters) {
          map.set(filter.id, { ...filter, group: group.id, groupLabel: group.label });
        }
      }
      return map;
    },
    activeFilters() {
      return this.filters.filter((filter) => {
        const definition = this.filterDefinitions.get(filter.id);
        return definition && definition.available && filter.enabled !== false;
      });
    },
    sortOptions() {
      const base = [
        { id: "title", label: "片名" },
        { id: "originalTitle", label: "原始片名" },
        { id: "year", label: "年份" },
        { id: "rating", label: "评分" },
        { id: "dateAdded", label: "添加日期" },
        { id: "metadata", label: "元数据" },
        { id: "artwork", label: "图片" },
        { id: "datasource", label: "数据源" },
      ];
      if (this.activeModule === "movie") {
        base.push({ id: "runtime", label: "片长" });
      } else {
        base.push({ id: "season", label: "季" }, { id: "episode", label: "集" }, { id: "episodeCount", label: "集数" });
      }
      return base;
    },
    datasourceOptions() {
      return (this.selectedLibrary ? this.selectedLibrary.paths || [this.selectedLibrary.path] : []).filter(Boolean);
    },
    genreOptions() {
      const values = new Set();
      for (const item of this.items) {
        for (const genre of item.genres || []) {
          if (genre) values.add(genre);
        }
      }
      return Array.from(values).sort((a, b) => a.localeCompare(b, "zh-CN"));
    },
    decadeOptions() {
      const values = new Set();
      for (const item of this.items) {
        const year = Number(item.yearGuess || 0);
        if (year > 0) values.add(`${Math.floor(year / 10) * 10}s`);
      }
      return Array.from(values).sort();
    },
    visibleItems() {
      const query = this.query.trim().toLowerCase();
      const tvStats = this.activeModule === "tvshow" ? this.buildTVStats(this.items) : null;
      const filtered = this.items.filter((item) => {
        const haystack = [item.titleGuess, item.showGuess, item.fileName, item.path, item.yearGuess, item.matchedName]
          .filter(Boolean)
          .join(" ")
          .toLowerCase();
        if (query && !haystack.includes(query)) return false;
        return this.activeFilters.every((filter) => this.filterAcceptsItem(filter, item, tvStats));
      });
      return filtered;
    },
    movieRows() {
      return this.sortItems(this.visibleItems.slice());
    },
    tvTree() {
      const shows = new Map();
      for (const item of this.visibleItems) {
        const showName = item.showGuess || item.titleGuess || "未知剧集";
        if (!shows.has(showName)) {
          shows.set(showName, { key: showName, title: showName, episodes: 0, seasons: new Map(), items: [] });
        }
        const show = shows.get(showName);
        show.items.push(item);
        const seasonNumber = item.season || 0;
        const seasonKey = `${showName}::${seasonNumber}`;
        if (!show.seasons.has(seasonKey)) {
          show.seasons.set(seasonKey, {
            key: seasonKey,
            showKey: showName,
            showTitle: showName,
            season: seasonNumber,
            title: seasonNumber ? `Season ${String(seasonNumber).padStart(2, "0")}` : "未识别季",
            items: [],
          });
        }
        show.seasons.get(seasonKey).items.push(item);
        show.episodes += 1;
      }
      return this.sortShows(Array.from(shows.values()))
        .map((show) => ({
          ...show,
          seasons: Array.from(show.seasons.values())
            .map((season) => ({
              ...season,
              items: this.sortEpisodes(season.items.slice()),
            }))
            .sort((a, b) => a.season - b.season),
        }));
    },
    allTasks() {
      return Object.values(this.tasks)
        .filter(Boolean)
        .sort((a, b) => (a.startedAt < b.startedAt ? 1 : -1));
    },
    selectedCountText() {
      if (!this.selectedItem) return "未选择";
      return `已选择 1 / ${this.visibleItems.length}`;
    },
    scanProgressText() {
      const task = this.selectedTask;
      if (!task) return "";
      if (task.state === "running") return `已检查 ${task.visitedFiles || 0} 个文件，发现 ${task.foundItems || 0} 个视频`;
      if (task.state === "completed") return `扫描完成，共 ${task.resultCount || task.foundItems || 0} 个视频`;
      if (task.state === "failed") return `扫描失败：${task.error || "未知错误"}`;
      return "";
    },
    filteredSettingsSections() {
      const query = this.settingsFilter.trim().toLowerCase();
      if (!query) return this.settingsSections;
      return this.settingsSections
        .map((section) => {
          const children = this.filterSettingsChildren(section.children, `${section.title} `, query);
          if (section.title.toLowerCase().includes(query) || children.length) {
            return { ...section, children };
          }
          return null;
        })
        .filter(Boolean);
    },
    activeSettingsPage() {
      return this.settingsActiveSection;
    },
    detailTitle() {
      if (!this.selectedSummary) return "";
      return this.selectedSummary.title;
    },
    selectedSummary() {
      if (this.selectedEntity && this.selectedEntity.kind === "show") {
        const show = this.selectedEntity.payload;
        const first = this.firstTVItem(show);
        return this.buildSummary(first, {
          entityType: "show",
          title: show.title,
          subtitle: `${show.seasons.length} 季 / ${show.episodes} 集`,
          year: first ? first.yearGuess : "",
          itemCount: show.episodes,
        });
      }
      if (this.selectedEntity && this.selectedEntity.kind === "season") {
        const season = this.selectedEntity.payload;
        const first = season.items[0];
        return this.buildSummary(first, {
          entityType: "season",
          title: `${season.showTitle || (first && first.showGuess) || ""} ${season.title}`.trim(),
          subtitle: `${season.items.length} 集`,
          year: first ? first.yearGuess : "",
          season: season.season,
          itemCount: season.items.length,
        });
      }
      if (!this.selectedItem) return null;
      return this.buildSummary(this.selectedItem, {
        entityType: this.selectedItem.kind === "tvshow" ? "episode" : "movie",
        title: this.selectedItem.kind === "tvshow" ? this.selectedItem.showGuess || this.selectedItem.titleGuess : this.selectedItem.titleGuess,
        subtitle: this.selectedItem.kind === "tvshow" ? this.itemSeasonText(this.selectedItem) : this.selectedItem.originalTitle || this.selectedItem.original || "",
        year: this.selectedItem.yearGuess,
      });
    },
  },
  async mounted() {
    await this.loadSettings();
    this.loadLayoutSettings();
    await this.loadLibraries();
    this.startPolling();
    window.addEventListener("click", this.closeContextMenu);
    window.addEventListener("keydown", this.handleKeydown);
    window.addEventListener("pointermove", this.handleResizeMove);
    window.addEventListener("pointerup", this.stopResize);
    this.status = "就绪";
  },
  beforeUnmount() {
    if (this.poller) clearInterval(this.poller);
    window.removeEventListener("click", this.closeContextMenu);
    window.removeEventListener("keydown", this.handleKeydown);
    window.removeEventListener("pointermove", this.handleResizeMove);
    window.removeEventListener("pointerup", this.stopResize);
  },
  methods: {
    loadLayoutSettings() {
      const browserWidth = Number(localStorage.getItem("tmmweb.browserWidth") || 0);
      const filterNavWidth = Number(localStorage.getItem("tmmweb.filterNavWidth") || 0);
      if (browserWidth >= 420) this.layout.browserWidth = browserWidth;
      if (filterNavWidth >= 140) this.layout.filterNavWidth = filterNavWidth;
    },
    startWorkbenchResize(event) {
      const rect = this.$refs.workbench ? this.$refs.workbench.getBoundingClientRect() : null;
      if (!rect) return;
      const currentWidth = this.layout.browserWidth || Math.round(rect.width - 380);
      this.layout.resizing = {
        type: "workbench",
        startX: event.clientX,
        startWidth: currentWidth,
        containerWidth: rect.width,
      };
      event.preventDefault();
    },
    startFilterNavResize(event) {
      this.layout.resizing = {
        type: "filterNav",
        startX: event.clientX,
        startWidth: this.layout.filterNavWidth,
      };
      event.preventDefault();
    },
    handleResizeMove(event) {
      const resizing = this.layout.resizing;
      if (!resizing) return;
      const delta = event.clientX - resizing.startX;
      if (resizing.type === "workbench") {
        const max = Math.max(420, resizing.containerWidth - 300 - 6);
        this.layout.browserWidth = Math.min(max, Math.max(420, resizing.startWidth + delta));
        return;
      }
      if (resizing.type === "filterNav") {
        this.layout.filterNavWidth = Math.min(320, Math.max(140, resizing.startWidth + delta));
      }
    },
    stopResize() {
      if (!this.layout.resizing) return;
      if (this.layout.browserWidth) localStorage.setItem("tmmweb.browserWidth", String(Math.round(this.layout.browserWidth)));
      localStorage.setItem("tmmweb.filterNavWidth", String(Math.round(this.layout.filterNavWidth)));
      this.layout.resizing = null;
    },
    filterDefinition(id) {
      return this.filterDefinitions.get(id) || null;
    },
    openFilterEditor() {
      this.filterEditor.open = true;
      this.filterEditor.tab = this.filterGroups[0] ? this.filterGroups[0].id : "details";
    },
    closeFilterEditor() {
      this.filterEditor.open = false;
    },
    isFilterActive(id) {
      return this.filters.some((filter) => filter.id === id);
    },
    addFilter(id) {
      const definition = this.filterDefinition(id);
      if (!definition || !definition.available || this.isFilterActive(id)) return;
      this.filters.push(this.defaultFilter(definition));
    },
    removeFilter(index) {
      this.filters.splice(index, 1);
    },
    clearFilters() {
      this.query = "";
      this.filters = [];
    },
    defaultFilter(definition) {
      const filter = {
        id: definition.id,
        enabled: true,
        invert: false,
        operator: "contains",
        value: "",
        valueHigh: "",
      };
      if (definition.type === "number") {
        filter.operator = "between";
        if (definition.id === "year") {
          const year = new Date().getFullYear();
          filter.value = String(year);
          filter.valueHigh = String(year);
        }
      } else if (definition.type === "date") {
        filter.operator = "is";
        filter.value = this.todayValue();
      } else if (definition.type === "boolean") {
        filter.operator = "is";
        filter.value = "true";
      } else if (definition.type === "choice") {
        filter.operator = "is";
      }
      return filter;
    },
    filterOperatorOptions(filter) {
      const definition = this.filterDefinition(filter.id);
      if (!definition) return [];
      if (definition.type === "number") {
        return [
          { id: "between", label: "介于" },
          { id: "equals", label: "等于" },
          { id: "gte", label: "大于等于" },
          { id: "lte", label: "小于等于" },
        ];
      }
      if (definition.type === "date") {
        return [
          { id: "is", label: "等于" },
          { id: "after", label: "之后" },
          { id: "before", label: "之前" },
          { id: "lastDays", label: "最近 N 天" },
        ];
      }
      if (definition.type === "boolean") {
        return [
          { id: "is", label: "是" },
          { id: "not", label: "不是" },
        ];
      }
      if (definition.type === "choice") {
        return [
          { id: "is", label: "是" },
          { id: "contains", label: "包含" },
        ];
      }
      return [
        { id: "contains", label: "包含" },
        { id: "equals", label: "等于" },
        { id: "starts", label: "开头是" },
        { id: "ends", label: "结尾是" },
      ];
    },
    filterInputType(filter) {
      const definition = this.filterDefinition(filter.id);
      if (!definition) return "text";
      if (definition.type === "number") return "number";
      if (definition.type === "date" && filter.operator !== "lastDays") return "date";
      return "text";
    },
    filterValueOptions(filter) {
      if (filter.id === "datasource") return this.datasourceOptions;
      if (filter.id === "genre") return this.genreOptions;
      if (filter.id === "decade") return this.decadeOptions;
      return [];
    },
    filterAcceptsItem(filter, item, tvStats = null) {
      const definition = this.filterDefinition(filter.id);
      if (!definition || !definition.available) return true;
      let accepted = true;
      if (definition.type === "number") {
        accepted = this.matchNumber(this.filterFieldValue(filter.id, item, tvStats), filter);
      } else if (definition.type === "date") {
        accepted = this.matchDate(item.dateAdded, filter);
      } else if (definition.type === "boolean") {
        accepted = this.matchBoolean(this.filterFieldValue(filter.id, item, tvStats), filter);
      } else if (definition.type === "choice") {
        accepted = this.matchChoice(this.filterFieldValue(filter.id, item, tvStats), filter);
      } else {
        accepted = this.matchText(this.filterFieldValue(filter.id, item, tvStats), filter);
      }
      return filter.invert ? !accepted : accepted;
    },
    filterFieldValue(id, item, tvStats = null) {
      switch (id) {
        case "allInOne":
          return [item.titleGuess, item.showGuess, item.originalTitle || item.original, item.yearGuess, item.matchedName, item.fileName, item.path, item.imdbId, item.matchedId, ...(item.genres || [])].join(" ");
        case "title":
          return item.kind === "tvshow" ? item.showGuess || item.titleGuess : item.titleGuess;
        case "originalTitle":
          return item.originalTitle || item.original || "";
        case "datasource":
          return item.sourcePath || "";
        case "filename":
          return item.fileName || "";
        case "path":
          return item.path || item.dir || "";
        case "new":
          return this.isNewItem(item);
        case "year":
          return Number(item.yearGuess || 0);
        case "decade": {
          const year = Number(item.yearGuess || 0);
          return year > 0 ? `${Math.floor(year / 10) * 10}s` : "";
        }
        case "rating":
          return Number(item.rating || 0);
        case "runtime":
          return Number(item.runtime || 0);
        case "genre":
          return item.genres || [];
        case "tmdbId":
          return Number(item.matchedId || 0);
        case "imdbId":
          return item.imdbId || "";
        case "missingMetadata":
          return !item.hasNfo || !item.matchedName;
        case "missingArtwork":
          return !item.hasPoster || !item.hasFanart;
        case "missingSubtitles":
          return !item.hasSubtitle;
        case "poster":
          return !!item.hasPoster;
        case "fanart":
          return !!item.hasFanart;
        case "episodeCount":
          return this.tvStatForItem(item, tvStats).episodes;
        case "seasonCount":
          return this.tvStatForItem(item, tvStats).seasons;
        default:
          return "";
      }
    },
    buildTVStats(items) {
      const stats = new Map();
      for (const item of items) {
        const key = item.showGuess || item.titleGuess || "";
        if (!key) continue;
        if (!stats.has(key)) stats.set(key, { episodes: 0, seasons: new Set() });
        const stat = stats.get(key);
        stat.episodes += 1;
        stat.seasons.add(item.season || 0);
      }
      return stats;
    },
    tvStatForItem(item, stats) {
      const key = item.showGuess || item.titleGuess || "";
      const stat = stats && stats.get(key);
      if (!stat) return { episodes: 0, seasons: 0 };
      return { episodes: stat.episodes, seasons: stat.seasons.size };
    },
    matchText(value, filter) {
      const actual = String(value || "").toLowerCase();
      const expected = String(filter.value || "").toLowerCase();
      if (!expected) return true;
      if (filter.operator === "equals") return actual === expected;
      if (filter.operator === "starts") return actual.startsWith(expected);
      if (filter.operator === "ends") return actual.endsWith(expected);
      return actual.includes(expected);
    },
    matchNumber(value, filter) {
      const actual = Number(value || 0);
      const low = Number(filter.value || 0);
      const high = Number(filter.valueHigh || filter.value || 0);
      if (filter.operator === "equals") return actual === low;
      if (filter.operator === "gte") return actual >= low;
      if (filter.operator === "lte") return actual <= low;
      return actual >= Math.min(low, high) && actual <= Math.max(low, high);
    },
    matchChoice(value, filter) {
      const expected = String(filter.value || "").toLowerCase();
      if (!expected) return true;
      const values = Array.isArray(value) ? value : [value];
      return values.some((entry) => {
        const actual = String(entry || "").toLowerCase();
        return filter.operator === "contains" ? actual.includes(expected) : actual === expected;
      });
    },
    matchBoolean(value, filter) {
      const expected = filter.operator === "not" ? false : String(filter.value || "true") !== "false";
      return Boolean(value) === expected;
    },
    matchDate(value, filter) {
      const actual = this.dateOnly(value);
      if (!actual) return false;
      if (filter.operator === "lastDays") {
        const days = Math.max(1, Number(filter.value || 1));
        const since = new Date();
        since.setHours(0, 0, 0, 0);
        since.setDate(since.getDate() - days + 1);
        return actual >= since;
      }
      const expected = this.dateOnly(filter.value);
      if (!expected) return true;
      if (filter.operator === "after") return actual >= expected;
      if (filter.operator === "before") return actual <= expected;
      return actual.getTime() === expected.getTime();
    },
    dateOnly(value) {
      if (!value) return null;
      const date = new Date(value);
      if (Number.isNaN(date.getTime())) return null;
      date.setHours(0, 0, 0, 0);
      return date;
    },
    todayValue() {
      return new Date().toISOString().slice(0, 10);
    },
    isNewItem(item) {
      const date = this.dateOnly(item.dateAdded);
      if (!date) return false;
      const since = new Date();
      since.setHours(0, 0, 0, 0);
      since.setDate(since.getDate() - 30);
      return date >= since;
    },
    sortItems(items) {
      const direction = this.sortDirection === "desc" ? -1 : 1;
      return items.sort((a, b) => {
        const result = this.compareSortValue(a, b, this.sortKey);
        if (result !== 0) return result * direction;
        return this.compareText(a.titleGuess || a.fileName, b.titleGuess || b.fileName);
      });
    },
    sortEpisodes(items) {
      if (this.sortKey === "dateAdded" || this.sortKey === "rating" || this.sortKey === "year") return this.sortItems(items);
      return items.sort((a, b) => (a.season || 0) - (b.season || 0) || (a.episode || 0) - (b.episode || 0) || this.compareText(a.fileName, b.fileName));
    },
    sortShows(shows) {
      const direction = this.sortDirection === "desc" ? -1 : 1;
      return shows.sort((a, b) => {
        const firstA = this.firstTVItem(a);
        const firstB = this.firstTVItem(b);
        if (this.sortKey === "episodeCount") return (a.episodes - b.episodes) * direction;
        const result = this.compareSortValue(firstA || {}, firstB || {}, this.sortKey);
        if (result !== 0) return result * direction;
        return this.compareText(a.title, b.title);
      });
    },
    compareSortValue(a, b, key) {
      if (["year", "rating", "runtime", "season", "episode"].includes(key)) {
        return Number(a[this.sortFieldName(key)] || 0) - Number(b[this.sortFieldName(key)] || 0);
      }
      if (key === "dateAdded") return this.compareDate(a.dateAdded, b.dateAdded);
      if (key === "metadata") return Number(!!a.hasNfo) - Number(!!b.hasNfo);
      if (key === "artwork") return this.artworkScore(a) - this.artworkScore(b);
      if (key === "datasource") return this.compareText(a.sourcePath, b.sourcePath);
      if (key === "originalTitle") return this.compareText(a.originalTitle || a.original, b.originalTitle || b.original);
      return this.compareText(a.kind === "tvshow" ? a.showGuess || a.titleGuess : a.titleGuess, b.kind === "tvshow" ? b.showGuess || b.titleGuess : b.titleGuess);
    },
    sortFieldName(key) {
      if (key === "year") return "yearGuess";
      if (key === "rating") return "rating";
      if (key === "runtime") return "runtime";
      if (key === "season") return "season";
      if (key === "episode") return "episode";
      return key;
    },
    compareText(a, b) {
      return String(a || "").localeCompare(String(b || ""), "zh-CN", { numeric: true, sensitivity: "base" });
    },
    compareDate(a, b) {
      return (this.dateOnly(a)?.getTime() || 0) - (this.dateOnly(b)?.getTime() || 0);
    },
    artworkScore(item) {
      return Number(!!item.hasPoster) + Number(!!item.hasFanart);
    },
    formatDate(value) {
      const date = this.dateOnly(value);
      if (!date) return "-";
      return date.toLocaleDateString("zh-CN");
    },
    formatRating(value) {
      const rating = Number(value || 0);
      return rating > 0 ? rating.toFixed(1) : "-";
    },
    artworkURL(item, type) {
      if (!item || (type === "poster" && !item.hasPoster) || (type === "fanart" && !item.hasFanart)) return "";
      return `/api/artwork?id=${encodeURIComponent(item.id)}&type=${encodeURIComponent(type)}&v=${encodeURIComponent(item.dateAdded || item.matchedId || "")}`;
    },
    buildSummary(item, overrides = {}) {
      const fallback = item || {};
      return {
        item: fallback,
        entityType: overrides.entityType || fallback.kind || "movie",
        title: overrides.title || fallback.titleGuess || fallback.showGuess || "未命名",
        subtitle: overrides.subtitle || fallback.originalTitle || fallback.original || "",
        year: overrides.year || fallback.yearGuess || "",
        season: overrides.season || fallback.season || 0,
        itemCount: overrides.itemCount || 1,
        poster: this.artworkURL(fallback, "poster"),
        fanart: this.artworkURL(fallback, "fanart"),
        rating: fallback.rating || 0,
        genres: fallback.genres || [],
        overview: fallback.overview || "",
        dateAdded: fallback.dateAdded || "",
        matchedName: fallback.matchedName || "",
        matchedId: fallback.matchedId || 0,
        imdbId: fallback.imdbId || "",
        hasNfo: !!fallback.hasNfo,
        hasPoster: !!fallback.hasPoster,
        hasFanart: !!fallback.hasFanart,
        hasSubtitle: !!fallback.hasSubtitle,
        sourcePath: fallback.sourcePath || "",
      };
    },
    async api(path, options = {}) {
      const response = await fetch(path, {
        headers: { "Content-Type": "application/json" },
        ...options,
      });
      if (!response.ok) throw new Error(await response.text());
      return response.json();
    },
    async loadLibraries() {
      this.libraries = await this.api("/api/libraries");
      if (this.libraries.length && !this.selectedLibrary) {
        const first = this.filteredLibraries[0] || this.libraries[0];
        this.activeModule = first.type || "movie";
        await this.selectLibrary(first);
      }
    },
    async loadSettings() {
      try {
        const settings = await this.api("/api/settings");
        this.scraperSettings.tmdbConfigured = !!settings.tmdbConfigured;
        this.scraperSettings.tmdbKeySource = settings.tmdbKeySource || "none";
        this.scraperSettings.proxyEnabled = !!settings.proxyEnabled;
        this.scraperSettings.proxyHost = settings.proxyHost || "";
        this.scraperSettings.proxyPort = settings.proxyPort || 7890;
        this.scraperSettings.proxyUsername = settings.proxyUsername || "";
        this.scraperSettings.proxyPassword = "";
        this.scraperSettings.proxyPasswordConfigured = !!settings.proxyPassword;
        this.scraperSettings.clearProxyPassword = false;
        this.scraperSettings.movieScrapeMetadata = settings.movieScrapeMetadata !== false;
        this.scraperSettings.movieScrapeNfo = settings.movieScrapeNfo !== false;
        this.scraperSettings.movieScrapeImages = settings.movieScrapeImages !== false;
        this.scraperSettings.movieScrapeOverwrite = !!settings.movieScrapeOverwrite;
        this.scraperSettings.tvShowScrapeMetadata = settings.tvShowScrapeMetadata !== false;
        this.scraperSettings.tvShowEpisodeMetadata = settings.tvShowEpisodeMetadata !== false;
        this.scraperSettings.tvShowScrapeNfo = settings.tvShowScrapeNfo !== false;
        this.scraperSettings.tvShowScrapeImages = settings.tvShowScrapeImages !== false;
        this.scraperSettings.tvShowScrapeOverwrite = !!settings.tvShowScrapeOverwrite;
        this.scraperSettings.movieRenamerPathname = settings.movieRenamerPathname || "{title} ({year})";
        this.scraperSettings.movieRenamerFilename = settings.movieRenamerFilename || "{title} ({year})";
        this.scraperSettings.tvShowRenamerShowFolder = settings.tvShowRenamerShowFolder || "{showTitle}";
        this.scraperSettings.tvShowRenamerSeason = settings.tvShowRenamerSeason || "Season {seasonNr2}";
        this.scraperSettings.tvShowRenamerFilename = settings.tvShowRenamerFilename || "{showTitle} - S{seasonNr2}E{episodeNr2} - {title}";
        this.scraperSettings.moviePosterName = settings.moviePosterName || "poster.jpg";
        this.scraperSettings.movieFanartName = settings.movieFanartName || "fanart.jpg";
        this.scraperSettings.moviePosterNames = settings.moviePosterNames || "poster.jpg\nfolder.jpg\n{filename}-poster.jpg";
        this.scraperSettings.movieFanartNames = settings.movieFanartNames || "fanart.jpg\n{filename}-fanart.jpg";
        this.scraperSettings.tvShowPosterName = settings.tvShowPosterName || "poster.jpg";
        this.scraperSettings.tvShowFanartName = settings.tvShowFanartName || "fanart.jpg";
        this.scraperSettings.tvShowPosterNames = settings.tvShowPosterNames || "poster.jpg\nfolder.jpg";
        this.scraperSettings.tvShowFanartNames = settings.tvShowFanartNames || "fanart.jpg\nbackdrop.jpg";
      } catch (error) {
        this.status = error.message;
      }
    },
    openSettings() {
      this.settingsOpen = true;
      this.loadSettings();
    },
    selectSettingsSection(id) {
      this.settingsActiveSection = id;
    },
    findSettingsNode(id, nodes = this.settingsSections) {
      for (const node of nodes) {
        if (node.id === id) return node;
        if (node.children) {
          const found = this.findSettingsNode(id, node.children);
          if (found) return found;
        }
      }
      return null;
    },
    filterSettingsChildren(children, prefix, query) {
      return children
        .map((child) => {
          const nested = child.children ? this.filterSettingsChildren(child.children, `${prefix}${child.title} `, query) : [];
          const text = `${prefix}${child.title}`.toLowerCase();
          if (text.includes(query) || nested.length) return { ...child, children: nested };
          return null;
        })
        .filter(Boolean);
    },
    librariesByType(type) {
      return this.libraries.filter((library) => library.type === type);
    },
    async saveSettings() {
      this.scraperSettings.saving = true;
      this.status = "正在保存设置";
      try {
        const body = {
          proxyEnabled: this.scraperSettings.proxyEnabled,
          proxyHost: this.scraperSettings.proxyHost,
          proxyPort: Number(this.scraperSettings.proxyPort) || 0,
          proxyUsername: this.scraperSettings.proxyUsername,
          clearProxyPassword: this.scraperSettings.clearProxyPassword,
          movieScrapeMetadata: !!this.scraperSettings.movieScrapeMetadata,
          movieScrapeNfo: !!this.scraperSettings.movieScrapeNfo,
          movieScrapeImages: !!this.scraperSettings.movieScrapeImages,
          movieScrapeOverwrite: !!this.scraperSettings.movieScrapeOverwrite,
          tvShowScrapeMetadata: !!this.scraperSettings.tvShowScrapeMetadata,
          tvShowEpisodeMetadata: !!this.scraperSettings.tvShowEpisodeMetadata,
          tvShowScrapeNfo: !!this.scraperSettings.tvShowScrapeNfo,
          tvShowScrapeImages: !!this.scraperSettings.tvShowScrapeImages,
          tvShowScrapeOverwrite: !!this.scraperSettings.tvShowScrapeOverwrite,
          movieRenamerPathname: this.scraperSettings.movieRenamerPathname,
          movieRenamerFilename: this.scraperSettings.movieRenamerFilename,
          tvShowRenamerShowFolder: this.scraperSettings.tvShowRenamerShowFolder,
          tvShowRenamerSeason: this.scraperSettings.tvShowRenamerSeason,
          tvShowRenamerFilename: this.scraperSettings.tvShowRenamerFilename,
          moviePosterName: this.scraperSettings.moviePosterName,
          movieFanartName: this.scraperSettings.movieFanartName,
          moviePosterNames: this.scraperSettings.moviePosterNames,
          movieFanartNames: this.scraperSettings.movieFanartNames,
          tvShowPosterName: this.scraperSettings.tvShowPosterName,
          tvShowFanartName: this.scraperSettings.tvShowFanartName,
          tvShowPosterNames: this.scraperSettings.tvShowPosterNames,
          tvShowFanartNames: this.scraperSettings.tvShowFanartNames,
        };
        if (this.scraperSettings.tmdbApiKey) body.tmdbApiKey = this.scraperSettings.tmdbApiKey;
        if (this.scraperSettings.proxyPassword) body.proxyPassword = this.scraperSettings.proxyPassword;
        const settings = await this.api("/api/settings", {
          method: "PUT",
          body: JSON.stringify(body),
        });
        this.scraperSettings.tmdbConfigured = !!settings.tmdbConfigured;
        this.scraperSettings.tmdbKeySource = settings.tmdbKeySource || "none";
        this.scraperSettings.proxyEnabled = !!settings.proxyEnabled;
        this.scraperSettings.proxyHost = settings.proxyHost || "";
        this.scraperSettings.proxyPort = settings.proxyPort || 7890;
        this.scraperSettings.proxyUsername = settings.proxyUsername || "";
        this.scraperSettings.proxyPasswordConfigured = !!settings.proxyPassword;
        this.scraperSettings.movieScrapeMetadata = settings.movieScrapeMetadata !== false;
        this.scraperSettings.movieScrapeNfo = settings.movieScrapeNfo !== false;
        this.scraperSettings.movieScrapeImages = settings.movieScrapeImages !== false;
        this.scraperSettings.movieScrapeOverwrite = !!settings.movieScrapeOverwrite;
        this.scraperSettings.tvShowScrapeMetadata = settings.tvShowScrapeMetadata !== false;
        this.scraperSettings.tvShowEpisodeMetadata = settings.tvShowEpisodeMetadata !== false;
        this.scraperSettings.tvShowScrapeNfo = settings.tvShowScrapeNfo !== false;
        this.scraperSettings.tvShowScrapeImages = settings.tvShowScrapeImages !== false;
        this.scraperSettings.tvShowScrapeOverwrite = !!settings.tvShowScrapeOverwrite;
        this.scraperSettings.movieRenamerPathname = settings.movieRenamerPathname || this.scraperSettings.movieRenamerPathname;
        this.scraperSettings.movieRenamerFilename = settings.movieRenamerFilename || this.scraperSettings.movieRenamerFilename;
        this.scraperSettings.tvShowRenamerShowFolder = settings.tvShowRenamerShowFolder || this.scraperSettings.tvShowRenamerShowFolder;
        this.scraperSettings.tvShowRenamerSeason = settings.tvShowRenamerSeason || this.scraperSettings.tvShowRenamerSeason;
        this.scraperSettings.tvShowRenamerFilename = settings.tvShowRenamerFilename || this.scraperSettings.tvShowRenamerFilename;
        this.scraperSettings.moviePosterName = settings.moviePosterName || this.scraperSettings.moviePosterName;
        this.scraperSettings.movieFanartName = settings.movieFanartName || this.scraperSettings.movieFanartName;
        this.scraperSettings.moviePosterNames = settings.moviePosterNames || this.scraperSettings.moviePosterNames;
        this.scraperSettings.movieFanartNames = settings.movieFanartNames || this.scraperSettings.movieFanartNames;
        this.scraperSettings.tvShowPosterName = settings.tvShowPosterName || this.scraperSettings.tvShowPosterName;
        this.scraperSettings.tvShowFanartName = settings.tvShowFanartName || this.scraperSettings.tvShowFanartName;
        this.scraperSettings.tvShowPosterNames = settings.tvShowPosterNames || this.scraperSettings.tvShowPosterNames;
        this.scraperSettings.tvShowFanartNames = settings.tvShowFanartNames || this.scraperSettings.tvShowFanartNames;
        this.scraperSettings.tmdbApiKey = "";
        this.scraperSettings.proxyPassword = "";
        this.scraperSettings.clearProxyPassword = false;
        this.status = "设置已保存";
        return true;
      } catch (error) {
        this.status = error.message;
        return false;
      } finally {
        this.scraperSettings.saving = false;
      }
    },
    async closeSettings() {
      if (this.scraperSettings.saving) return;
      const saved = await this.saveSettings();
      if (saved) this.settingsOpen = false;
    },
    async clearTmdbKey() {
      this.scraperSettings.saving = true;
      this.status = "正在清除 TMDb Key";
      try {
        const settings = await this.api("/api/settings", {
          method: "PUT",
          body: JSON.stringify({
            clearTmdbKey: true,
            proxyEnabled: this.scraperSettings.proxyEnabled,
            proxyHost: this.scraperSettings.proxyHost,
            proxyPort: Number(this.scraperSettings.proxyPort) || 0,
            proxyUsername: this.scraperSettings.proxyUsername,
            clearProxyPassword: false,
            movieScrapeMetadata: !!this.scraperSettings.movieScrapeMetadata,
            movieScrapeNfo: !!this.scraperSettings.movieScrapeNfo,
            movieScrapeImages: !!this.scraperSettings.movieScrapeImages,
            movieScrapeOverwrite: !!this.scraperSettings.movieScrapeOverwrite,
            tvShowScrapeMetadata: !!this.scraperSettings.tvShowScrapeMetadata,
            tvShowEpisodeMetadata: !!this.scraperSettings.tvShowEpisodeMetadata,
            tvShowScrapeNfo: !!this.scraperSettings.tvShowScrapeNfo,
            tvShowScrapeImages: !!this.scraperSettings.tvShowScrapeImages,
            tvShowScrapeOverwrite: !!this.scraperSettings.tvShowScrapeOverwrite,
            movieRenamerPathname: this.scraperSettings.movieRenamerPathname,
            movieRenamerFilename: this.scraperSettings.movieRenamerFilename,
            tvShowRenamerShowFolder: this.scraperSettings.tvShowRenamerShowFolder,
            tvShowRenamerSeason: this.scraperSettings.tvShowRenamerSeason,
            tvShowRenamerFilename: this.scraperSettings.tvShowRenamerFilename,
            moviePosterName: this.scraperSettings.moviePosterName,
            movieFanartName: this.scraperSettings.movieFanartName,
            moviePosterNames: this.scraperSettings.moviePosterNames,
            movieFanartNames: this.scraperSettings.movieFanartNames,
            tvShowPosterName: this.scraperSettings.tvShowPosterName,
            tvShowFanartName: this.scraperSettings.tvShowFanartName,
            tvShowPosterNames: this.scraperSettings.tvShowPosterNames,
            tvShowFanartNames: this.scraperSettings.tvShowFanartNames,
          }),
        });
        this.scraperSettings.tmdbConfigured = !!settings.tmdbConfigured;
        this.scraperSettings.tmdbKeySource = settings.tmdbKeySource || "none";
        this.scraperSettings.tmdbApiKey = "";
        this.status = "TMDb Key 已清除";
      } catch (error) {
        this.status = error.message;
      } finally {
        this.scraperSettings.saving = false;
      }
    },
    async switchModule(module) {
      this.activeModule = module;
      this.newLibrary.type = module;
      this.newLibrary.name = module === "tvshow" ? "电视剧" : "电影";
      this.query = "";
      this.filters = [];
      this.sortKey = "title";
      this.sortDirection = "asc";
      const first = this.filteredLibraries[0];
      if (first) {
        await this.selectLibrary(first);
      } else {
        this.selectedLibrary = null;
        this.items = [];
        this.selectedItem = null;
        this.selectedEntity = null;
        this.status = `未配置${this.moduleTitle}数据源`;
      }
    },
    addPendingPath() {
      const path = this.pendingPath.trim();
      if (!path || this.newLibrary.paths.includes(path)) return;
      this.newLibrary.paths.push(path);
      this.pendingPath = "";
      if (!this.newLibrary.name || this.newLibrary.name === "电影" || this.newLibrary.name === "电视剧") {
        this.newLibrary.name = this.newLibrary.type === "tvshow" ? "电视剧" : "电影";
      }
    },
    removePath(path) {
      this.newLibrary.paths = this.newLibrary.paths.filter((item) => item !== path);
    },
    prepareDatasource(type) {
      this.newLibrary.type = type;
      this.newLibrary.name = type === "tvshow" ? "电视剧" : "电影";
    },
    async browseDatasource(type) {
      this.prepareDatasource(type);
      await this.openBrowser();
    },
    async openBrowser(start = this.pendingPath || "/Volumes") {
      this.browser.open = true;
      await this.browse(start);
    },
    async browse(path) {
      this.busy = true;
      try {
        const result = await this.api(`/api/browse?path=${encodeURIComponent(path)}`);
        this.browser.path = result.path;
        this.browser.parent = result.parent;
        this.browser.entries = result.entries || [];
        this.status = `正在浏览 ${result.path}`;
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    chooseBrowserPath() {
      this.pendingPath = this.browser.path;
      this.addPendingPath();
      this.browser.open = false;
    },
    async addLibrary() {
      this.addPendingPath();
      this.busy = true;
      this.status = "正在添加媒体库";
      try {
        const library = await this.api("/api/libraries", {
          method: "POST",
          body: JSON.stringify({
            name: this.newLibrary.name,
            type: this.newLibrary.type,
            paths: this.newLibrary.paths,
          }),
        });
        this.libraries.push(library);
        this.activeModule = library.type;
        await this.selectLibrary(library);
        this.newLibrary.paths = [];
        this.pendingPath = "";
        this.status = "媒体库已添加";
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    async addDatasource(type) {
      this.prepareDatasource(type);
      await this.addLibrary();
    },
    async deleteLibrary(library) {
      if (!library) return;
      if (!confirm(`确认移除数据源？\n${library.name}`)) return;
      this.busy = true;
      this.status = "正在移除数据源";
      try {
        await this.api(`/api/libraries?id=${encodeURIComponent(library.id)}`, { method: "DELETE" });
        this.libraries = this.libraries.filter((item) => item.id !== library.id);
        if (this.selectedLibrary && this.selectedLibrary.id === library.id) {
          const next = this.filteredLibraries[0] || this.libraries[0] || null;
          if (next) {
            await this.selectLibrary(next);
          } else {
            this.selectedLibrary = null;
            this.items = [];
            this.selectedItem = null;
          }
        }
        this.status = "数据源已移除";
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    async selectLibrary(library) {
      if (!library) return;
      this.activeModule = library.type || "movie";
      this.selectedLibrary = library;
      this.items = [];
      this.selectedItem = null;
      this.selectedEntity = null;
      this.candidates = [];
      this.renamePreview = null;
      await this.loadTasks(library);
      await this.loadItems(library);
    },
    async selectLibraryById(id) {
      const library = this.libraries.find((item) => item.id === id);
      if (library) await this.selectLibrary(library);
    },
    startPolling() {
      if (this.poller) clearInterval(this.poller);
      this.poller = setInterval(async () => {
        if (!this.selectedLibrary) return;
        await this.loadTasks(this.selectedLibrary, true);
        if (this.selectedTask && (this.selectedTask.state === "running" || this.selectedTask.state === "completed")) {
          await this.loadItems(this.selectedLibrary, true);
        }
      }, 1500);
    },
    async loadTasks(library, quiet = false) {
      try {
        const result = await this.api(`/api/tasks?libraryId=${encodeURIComponent(library.id)}`);
        const tasks = result.tasks || [];
        if (!tasks.length) return;
        tasks.sort((a, b) => (a.startedAt < b.startedAt ? 1 : -1));
        this.tasks[library.id] = tasks[0];
        if (!quiet && this.tasks[library.id].state === "running") this.status = `${library.name} 正在扫描`;
      } catch (error) {
        if (!quiet) this.status = error.message;
      }
    },
    async loadItems(library, quiet = false) {
      if (!quiet) {
        this.busy = true;
        this.status = `正在加载 ${library.name}`;
      }
      try {
        const result = await this.api(`/api/items?libraryId=${encodeURIComponent(library.id)}`);
        this.items = result.items || [];
        if (!quiet) {
          this.status = this.items.length ? `已加载 ${this.items.length} 个缓存条目` : `已选择 ${library.name}，需要扫描`;
        }
      } catch (error) {
        if (!quiet) this.status = error.message;
      } finally {
        if (!quiet) this.busy = false;
      }
    },
    selectItem(item) {
      this.selectedItem = item;
      this.selectedEntity = { kind: item.kind === "tvshow" ? "episode" : "movie", payload: item };
      this.candidates = [];
      this.rename.pattern =
        item.kind === "tvshow"
          ? this.scraperSettings.tvShowRenamerFilename
          : this.scraperSettings.movieRenamerFilename;
      this.rename.title = item.titleGuess;
      this.rename.year = item.yearGuess || "";
      this.rename.tmdbId = item.matchedId || 0;
      this.scrapeSearch.query = item.titleGuess || item.showGuess || "";
      this.scrapeSearch.year = item.yearGuess || "";
      this.renamePreview = null;
    },
    handleKeydown(event) {
      if (event.key === "Escape") {
        this.closeContextMenu();
        if (this.chooser.open && !this.chooser.loading && !this.chooser.scraping) this.closeChooser();
      }
    },
    openContextMenu(event, scope, payload) {
      this.contextMenu = {
        open: true,
        x: event.clientX,
        y: event.clientY,
        scope,
        payload,
      };
      if (scope === "movie" || scope === "episode") this.selectItem(payload);
      if (scope === "show") this.selectTvGroup("show", payload);
      if (scope === "season") this.selectTvGroup("season", payload.season || payload);
    },
    closeContextMenu() {
      this.contextMenu.open = false;
    },
    contextMenuTitle() {
      if (this.contextMenu.scope === "show") return "搜索并刮削整剧...";
      if (this.contextMenu.scope === "season") return "搜索并刮削本季...";
      if (this.contextMenu.scope === "episode") return "搜索并刮削本集...";
      return "搜索并刮削...";
    },
    openChooserFromSelected() {
      if (!this.selectedItem) return;
      if (this.selectedEntity && this.selectedEntity.kind === "show") {
        const show = this.selectedEntity.payload;
        const first = this.firstTVItem(show);
        this.openChooser({
          scope: "show",
          mediaType: "tvshow",
          targetItem: first,
          targetShow: show,
          query: show.title || show.key || "",
          year: first ? first.yearGuess || "" : "",
          path: first ? this.showRootPath(first) : show.title,
        });
        return;
      }
      if (this.selectedEntity && this.selectedEntity.kind === "season") {
        const season = this.selectedEntity.payload;
        const first = season.items[0];
        this.openChooser({
          scope: "season",
          mediaType: "tvshow",
          targetItem: first,
          targetShow: { title: season.showTitle, key: season.showKey },
          targetSeason: season,
          query: season.showTitle || "",
          year: first ? first.yearGuess || "" : "",
          path: first ? this.showRootPath(first) : season.showTitle,
        });
        return;
      }
      this.openChooser({
        scope: this.selectedItem.kind === "tvshow" ? "episode" : "movie",
        mediaType: this.selectedItem.kind === "tvshow" ? "tvshow" : "movie",
        targetItem: this.selectedItem,
        query: this.selectedItem.kind === "tvshow" ? this.selectedItem.showGuess || this.selectedItem.titleGuess : this.selectedItem.titleGuess,
        year: this.selectedItem.yearGuess || "",
        path: this.selectedItem.path || this.selectedItem.dir || "",
      });
    },
    openChooserFromContext() {
      const scope = this.contextMenu.scope;
      const payload = this.contextMenu.payload;
      this.closeContextMenu();
      if (!scope || !payload) return;
      if (scope === "movie") {
        this.openChooser({
          scope: "movie",
          mediaType: "movie",
          targetItem: payload,
          query: payload.titleGuess || "",
          year: payload.yearGuess || "",
          path: payload.path || payload.dir || "",
        });
      } else if (scope === "show") {
        const first = this.firstTVItem(payload);
        this.openChooser({
          scope: "show",
          mediaType: "tvshow",
          targetItem: first,
          targetShow: payload,
          query: payload.title || payload.key || "",
          year: first ? first.yearGuess || "" : "",
          path: first ? this.showRootPath(first) : payload.title,
        });
      } else if (scope === "season") {
        const first = payload.season.items[0];
        this.openChooser({
          scope: "season",
          mediaType: "tvshow",
          targetItem: first,
          targetShow: payload.show,
          targetSeason: payload.season,
          query: payload.show.title || payload.show.key || "",
          year: first ? first.yearGuess || "" : "",
          path: first ? this.showRootPath(first) : payload.show.title,
        });
      } else if (scope === "episode") {
        this.openChooser({
          scope: "episode",
          mediaType: "tvshow",
          targetItem: payload,
          query: payload.showGuess || payload.titleGuess || "",
          year: payload.yearGuess || "",
          path: payload.path || payload.dir || "",
        });
      }
    },
    firstTVItem(show) {
      const firstSeason = show && show.seasons ? show.seasons[0] : null;
      return firstSeason && firstSeason.items.length ? firstSeason.items[0] : null;
    },
    showRootPath(item) {
      if (!item) return "";
      if (item.season > 0 || item.episode > 0) {
        const parts = String(item.dir || "").split("/");
        parts.pop();
        return parts.join("/") || item.dir || "";
      }
      return item.dir || item.path || "";
    },
    openChooser(config) {
      if (!config.targetItem) {
        this.status = "没有可刮削的条目";
        return;
      }
      const tv = config.mediaType === "tvshow";
      this.chooser = {
        open: true,
        scope: config.scope,
        mediaType: config.mediaType,
        targetItem: config.targetItem,
        targetShow: config.targetShow || null,
        targetSeason: config.targetSeason || null,
        path: config.path || "",
        query: config.query || "",
        year: config.year || "",
        candidates: [],
        selected: null,
        detail: null,
        loading: false,
        scraping: false,
        error: "",
        options: {
          metadata: tv ? this.scraperSettings.tvShowScrapeMetadata : this.scraperSettings.movieScrapeMetadata,
          nfo: tv ? this.scraperSettings.tvShowScrapeNfo : this.scraperSettings.movieScrapeNfo,
          artwork: tv ? this.scraperSettings.tvShowScrapeImages : this.scraperSettings.movieScrapeImages,
          overwrite: tv ? this.scraperSettings.tvShowScrapeOverwrite : this.scraperSettings.movieScrapeOverwrite,
          showMetadata: this.scraperSettings.tvShowScrapeMetadata,
          episodeMetadata: this.scraperSettings.tvShowEpisodeMetadata && (config.scope === "season" || config.scope === "show"),
        },
      };
      this.searchChooser();
    },
    closeChooser() {
      this.chooser.open = false;
      this.chooser.error = "";
    },
    chooserTitle() {
      if (this.chooser.scope === "show") return "刮削电视剧";
      if (this.chooser.scope === "season") return "刮削电视剧季";
      if (this.chooser.scope === "episode") return "刮削单集";
      return "刮削电影";
    },
    chooserScopeText() {
      if (this.chooser.scope === "show" && this.chooser.targetShow) {
        return `${this.chooser.targetShow.title} · 全剧 · ${this.chooser.targetShow.episodes} 集`;
      }
      if (this.chooser.scope === "season" && this.chooser.targetShow && this.chooser.targetSeason) {
        return `${this.chooser.targetShow.title} · ${this.chooser.targetSeason.title} · ${this.chooser.targetSeason.items.length} 集`;
      }
      if (this.chooser.targetItem) return this.chooser.targetItem.fileName || this.chooser.targetItem.titleGuess;
      return "";
    },
    imageURL(path, size = "w342") {
      if (!path) return "";
      return `https://image.tmdb.org/t/p/${size}${path}`;
    },
    candidateDate(candidate) {
      return candidate.releaseDate || candidate.firstAirDate || "";
    },
    candidateYear(candidate) {
      const date = this.candidateDate(candidate);
      return date ? date.slice(0, 4) : "";
    },
    selectedDetailTitle() {
      if (!this.chooser.detail) return "";
      return this.chooser.detail.title || this.chooser.detail.name || "";
    },
    async searchChooser() {
      const query = this.chooser.query.trim();
      if (!query) return;
      this.chooser.loading = true;
      this.chooser.error = "";
      this.chooser.candidates = [];
      this.chooser.selected = null;
      this.chooser.detail = null;
      this.status = `正在搜索 ${query}`;
      try {
        const params = new URLSearchParams();
        params.set("q", query);
        params.set("type", this.chooser.mediaType);
        if (this.chooser.year) params.set("year", this.chooser.year);
        this.chooser.candidates = await this.api(`/api/search?${params.toString()}`);
        this.status = `找到 ${this.chooser.candidates.length} 个候选`;
        if (this.chooser.candidates.length) await this.selectCandidate(this.chooser.candidates[0]);
      } catch (error) {
        this.chooser.error = error.message;
        this.status = error.message;
      } finally {
        this.chooser.loading = false;
      }
    },
    async selectCandidate(candidate) {
      this.chooser.selected = candidate;
      this.chooser.detail = null;
      this.chooser.error = "";
      this.chooser.loading = true;
      try {
        const params = new URLSearchParams();
        params.set("id", candidate.id);
        params.set("type", candidate.mediaType || this.chooser.mediaType);
        this.chooser.detail = await this.api(`/api/metadata?${params.toString()}`);
      } catch (error) {
        this.chooser.error = error.message;
      } finally {
        this.chooser.loading = false;
      }
    },
    async applyChooser() {
      if (!this.chooser.selected || !this.chooser.targetItem) return;
      this.chooser.scraping = true;
      this.chooser.error = "";
      this.status = "正在按选择项写入刮削结果";
      try {
        const body = {
          itemId: this.chooser.targetItem.id,
          scope: this.chooser.scope,
          libraryId: this.selectedLibrary ? this.selectedLibrary.id : this.chooser.targetItem.libraryId,
          tmdbId: this.chooser.selected.id,
          mediaType: this.chooser.selected.mediaType || this.chooser.mediaType,
          writeNfo: !!this.chooser.options.nfo,
          writeImages: !!this.chooser.options.artwork,
          writeMeta: !!this.chooser.options.metadata || !!this.chooser.options.showMetadata || !!this.chooser.options.episodeMetadata,
          overwrite: !!this.chooser.options.overwrite,
        };
        if (this.chooser.targetShow) body.showName = this.chooser.targetShow.key || this.chooser.targetShow.title;
        if (this.chooser.targetSeason) body.season = this.chooser.targetSeason.season;
        const result = await this.api("/api/scrape", {
          method: "POST",
          body: JSON.stringify(body),
        });
        if (result.items && result.items.length) {
          const byID = new Map(result.items.map((item) => [item.id, item]));
          this.items = this.items.map((item) => byID.get(item.id) || item);
          if (this.selectedItem && byID.has(this.selectedItem.id)) this.selectedItem = byID.get(this.selectedItem.id);
        } else if (result.item) {
          this.items = this.items.map((item) => (item.id === result.item.id ? result.item : item));
          this.selectedItem = result.item;
        }
        const scraped = result.movie || result.show || this.chooser.detail;
        if (scraped && this.selectedItem) {
          this.rename.title = scraped.title;
          this.rename.year = (scraped.releaseDate || scraped.firstAirDate || "").slice(0, 4);
          this.rename.tmdbId = scraped.id;
        }
        this.status = "刮削写入完成";
        this.closeChooser();
      } catch (error) {
        this.chooser.error = error.message;
        this.status = error.message;
      } finally {
        this.chooser.scraping = false;
      }
    },
    toggleShow(key) {
      this.expandedShows[key] = !this.isShowExpanded(key);
    },
    toggleSeason(key) {
      this.expandedSeasons[key] = !this.isSeasonExpanded(key);
    },
    isShowExpanded(key) {
      return this.expandedShows[key] === true;
    },
    isSeasonExpanded(key) {
      return this.expandedSeasons[key] === true;
    },
    selectTvGroup(kind, payload) {
      if (kind === "show") {
        this.selectedEntity = { kind: "show", payload };
        const firstSeason = payload.seasons[0];
        if (firstSeason && firstSeason.items[0]) {
          this.selectItem(firstSeason.items[0]);
          this.selectedEntity = { kind: "show", payload };
        }
      }
      if (kind === "season" && payload.items[0]) {
        this.selectItem(payload.items[0]);
        this.selectedEntity = { kind: "season", payload };
      }
    },
    itemSeasonText(item) {
      if (item.kind !== "tvshow") return item.yearGuess || "-";
      if (item.airDate) return item.airDate;
      if (item.season && item.episodes && item.episodes.length) {
        return `S${String(item.season).padStart(2, "0")}E${item.episodes.map((episode) => String(episode).padStart(2, "0")).join(",")}`;
      }
      if (item.season && item.episode) return `S${item.season}E${item.episode}`;
      return "-";
    },
    itemStatusText(item) {
      const values = [];
      if (item.hasNfo) values.push("NFO");
      if (item.hasPoster) values.push("海报");
      if (item.hasFanart) values.push("背景图");
      if (item.hasSubtitle) values.push("字幕");
      if (item.matchedName) values.push("已匹配");
      return values.length ? values.join(" / ") : "待完善";
    },
    taskDescription(task) {
      if (task.state === "running") return `已检查 ${task.visitedFiles || 0} 个文件，发现 ${task.foundItems || 0} 个视频`;
      if (task.state === "canceling") return `正在停止，已发现 ${task.foundItems || 0} 个视频`;
      if (task.state === "canceled") return `已停止，保留已发现的 ${task.resultCount || task.foundItems || 0} 个视频`;
      if (task.state === "completed") return `完成，导入 ${task.resultCount || task.foundItems || 0} 个视频`;
      if (task.state === "failed") return task.error || "任务失败";
      return task.state;
    },
    async scan() {
      if (!this.selectedLibrary) return;
      this.busy = true;
      this.status = this.selectedScanning ? "正在停止扫描任务" : "正在启动扫描任务";
      try {
        const result = this.selectedScanning
          ? await this.api("/api/scan/cancel", {
              method: "POST",
              body: JSON.stringify({ libraryId: this.selectedLibrary.id, taskId: this.selectedTask ? this.selectedTask.id : "" }),
            })
          : await this.api("/api/scan", {
              method: "POST",
              body: JSON.stringify({ libraryId: this.selectedLibrary.id }),
            });
        this.tasks[this.selectedLibrary.id] = result.task;
        if (this.selectedScanning) {
          this.status = "已请求停止扫描";
        } else {
          this.status = result.started ? "扫描任务已启动" : "该媒体库已有扫描任务在运行";
        }
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    async searchSelected() {
      if (!this.selectedItem) return;
      this.busy = true;
      this.status = "正在搜索 TMDb 候选";
      try {
        const params = new URLSearchParams();
        const query = this.scrapeSearch.query.trim();
        const year = String(this.scrapeSearch.year || "").trim();
        if (query) {
          params.set("q", query);
          if (year) params.set("year", year);
          params.set("type", this.selectedItem.kind || this.activeModule);
        } else {
          params.set("itemId", this.selectedItem.id);
        }
        this.candidates = await this.api(`/api/search?${params.toString()}`);
        this.status = `找到 ${this.candidates.length} 个候选`;
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    async scrape(candidate) {
      this.busy = true;
      this.status = "正在写入 NFO 和图片";
      const tv = (candidate.mediaType || this.selectedItem.kind || this.activeModule) === "tvshow";
      try {
        const result = await this.api("/api/scrape", {
          method: "POST",
          body: JSON.stringify({
            itemId: this.selectedItem.id,
            tmdbId: candidate.id,
            mediaType: candidate.mediaType || this.selectedItem.kind || this.activeModule,
            writeNfo: tv ? this.scraperSettings.tvShowScrapeNfo : this.scraperSettings.movieScrapeNfo,
            writeImages: tv ? this.scraperSettings.tvShowScrapeImages : this.scraperSettings.movieScrapeImages,
            writeMeta: tv ? this.scraperSettings.tvShowScrapeMetadata : this.scraperSettings.movieScrapeMetadata,
            overwrite: tv ? this.scraperSettings.tvShowScrapeOverwrite : this.scraperSettings.movieScrapeOverwrite,
          }),
        });
        this.selectedItem = result.item;
        const scraped = result.movie || result.show;
        this.rename.title = scraped.title;
        this.rename.year = (scraped.releaseDate || scraped.firstAirDate || "").slice(0, 4);
        this.rename.tmdbId = scraped.id;
        this.status = "刮削写入完成";
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    async previewRename() {
      if (!this.selectedItem) return;
      this.busy = true;
      this.status = "正在生成重命名预览";
      try {
        this.renamePreview = await this.api("/api/rename/preview", {
          method: "POST",
          body: JSON.stringify({
            itemId: this.selectedItem.id,
            ...this.rename,
            movieRenamerPathname: this.scraperSettings.movieRenamerPathname,
            movieRenamerFilename: this.selectedItem.kind === "tvshow" ? this.scraperSettings.movieRenamerFilename : this.rename.pattern,
            tvShowRenamerShowFolder: this.scraperSettings.tvShowRenamerShowFolder,
            tvShowRenamerSeason: this.scraperSettings.tvShowRenamerSeason,
            tvShowRenamerFilename: this.selectedItem.kind === "tvshow" ? this.rename.pattern : this.scraperSettings.tvShowRenamerFilename,
          }),
        });
        this.status = "重命名预览已生成";
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    async applyRename() {
      if (!this.renamePreview) return;
      if (!confirm("确认执行重命名？请确保 Plex/TMM 没有正在扫描该文件。")) return;
      this.busy = true;
      this.status = "正在执行重命名";
      try {
        await this.api("/api/rename/apply", {
          method: "POST",
          body: JSON.stringify(this.renamePreview),
        });
        this.status = "重命名完成，请重新扫描媒体库";
        this.renamePreview = null;
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
  },
}).mount("#app");
