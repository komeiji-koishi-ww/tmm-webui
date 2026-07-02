const { createApp } = Vue;

const TMM_SCRAPER_FIELDS = {
  movie: [
    { title: "元数据", items: [["ID", "ID / TMDb / IMDb"], ["TITLE", "标题"], ["ORIGINAL_TITLE", "原始标题"], ["TAGLINE", "标语"], ["PLOT", "简介"], ["YEAR", "年份"], ["RELEASE_DATE", "上映日期"], ["RATING", "评分"], ["TOP250", "Top 250"], ["RUNTIME", "时长"], ["CERTIFICATION", "分级"], ["GENRES", "类型"], ["SPOKEN_LANGUAGES", "语言"], ["COUNTRY", "国家"], ["PRODUCTION_COMPANY", "制作公司"], ["TAGS", "标签"], ["COLLECTION", "电影集合"], ["TRAILER", "预告片"]] },
    { title: "演职员", items: [["ACTORS", "演员"], ["PRODUCERS", "制片"], ["DIRECTORS", "导演"], ["WRITERS", "编剧"]] },
    { title: "图片", items: [["POSTER", "海报"], ["FANART", "同人图"], ["BANNER", "横幅"], ["CLEARART", "ClearArt"], ["THUMB", "缩略图"], ["CLEARLOGO", "ClearLogo"], ["DISCART", "碟图"], ["KEYART", "KeyArt"], ["EXTRAFANART", "Extra Fanart"], ["EXTRATHUMB", "Extra Thumb"]] },
  ],
  tvshow: [
    { title: "剧集元数据", items: [["ID", "ID / TMDb"], ["TITLE", "标题"], ["ORIGINAL_TITLE", "原始标题"], ["PLOT", "简介"], ["YEAR", "年份"], ["AIRED", "首播"], ["STATUS", "状态"], ["RATING", "评分"], ["TOP250", "Top 250"], ["RUNTIME", "时长"], ["CERTIFICATION", "分级"], ["GENRES", "类型"], ["COUNTRY", "国家"], ["STUDIO", "电视台/工作室"], ["TAGS", "标签"], ["TRAILER", "预告片"], ["SEASON_NAMES", "季名称"], ["SEASON_OVERVIEW", "季简介"]] },
    { title: "演职员", items: [["ACTORS", "演员"]] },
    { title: "剧集图片", items: [["POSTER", "海报"], ["FANART", "同人图"], ["BANNER", "横幅"], ["CLEARART", "ClearArt"], ["THUMB", "缩略图"], ["CLEARLOGO", "ClearLogo"], ["DISCART", "碟图"], ["KEYART", "KeyArt"], ["CHARACTERART", "角色图"], ["EXTRAFANART", "Extra Fanart"], ["SEASON_POSTER", "季海报"], ["SEASON_FANART", "季同人图"], ["SEASON_BANNER", "季横幅"], ["SEASON_THUMB", "季缩略图"], ["THEME", "主题曲"]] },
  ],
  episode: [
    { title: "单集元数据", items: [["TITLE", "标题"], ["ORIGINAL_TITLE", "原始标题"], ["PLOT", "简介"], ["SEASON_EPISODE", "季/集编号"], ["AIRED", "首播"], ["RATING", "评分"], ["TAGS", "标签"]] },
    { title: "演职员", items: [["ACTORS", "演员"], ["DIRECTORS", "导演"], ["WRITERS", "编剧"]] },
    { title: "图片", items: [["THUMB", "缩略图"]] },
  ],
};

const TMM_RENAMER_TOKENS = [
  "${title}", "${originalTitle}", "${edition}", "${year}", "${releaseDate}", "${rating}", "${imdb}", "${tmdb}",
  "${videoFormat}", "${audioCodec}", "${fileSize}",
  "${showTitle}", "${showYear}", "${seasonNr}", "${seasonNr2}", "${episodeNr}", "${episodeNr2}", "${aired}",
];

const MOVIE_RENAMER_TOKEN_ROWS = [
  ["${title}", "标题", "银翼杀手"],
  ["${originalTitle}", "原始标题", "Blade Runner"],
  ["${originalFilename}", "原始文件名", "Blade.Runner.1982.1080p.DTS.mkv"],
  ["${originalBasename}", "不带扩展名的原始文件名", "Blade.Runner.1982.1080p.DTS"],
  ["${title[0]}", "标题第一个字符", "银"],
  ["${title;first}", "首字母或非字母替换字符", "银"],
  ["${title[0,2]}", "标题前两个字符", "银翼"],
  ["${titleSortable}", "排序标题", "银翼杀手"],
  ["${year}", "年份", "1982"],
  ["${releaseDate}", "上映日期", "1982-06-25"],
  ["${movieSet.title}", "电影集合标题", "银翼杀手系列"],
  ["${movieSet.titleSortable}", "电影集合排序标题", "银翼杀手系列"],
  ["${movieSetIndex}", "集合中的序号", "1"],
  ["${rating}", "评分", "8.1"],
  ["${imdb}", "IMDb ID", "tt0083658"],
  ["${tmdb}", "TMDb ID", "78"],
  ["${certification}", "分级", "R"],
  ["${directors[0].name}", "第一位导演", "Ridley Scott"],
  ["${actors[0].name}", "第一位演员", "Harrison Ford"],
  ["${genres[0]}", "第一个类型", "科幻"],
  ["${genresAsString}", "全部类型", "科幻, 惊悚"],
  ["${genres[0].name}", "第一个英文类型", "Science Fiction"],
  ["${tags[0]}", "第一个标签", "Cyberpunk"],
  ["${productionCompany}", "制作公司", "Warner Bros."],
  ["${productionCompanyAsArray[0]}", "第一制作公司", "Warner Bros."],
  ["${language}", "语言", "zh-CN"],
  ["${videoResolution}", "视频分辨率", "1920x1080"],
  ["${aspectRatio}", "画面比例", "178"],
  ["${aspectRatio2}", "第二画面比例", ""],
  ["${videoCodec}", "视频编码", "H.264"],
  ["${videoFormat}", "视频格式", "1080p"],
  ["${videoBitDepth}", "视频位深", "10"],
  ["${videoBitRate}", "视频码率", "10.5 Mbps"],
  ["${framerate}", "帧率", "23.976"],
  ["${audioCodec}", "默认音频编码", "DTS"],
  ["${audioCodecList}", "全部音频编码数组", "[DTS, AC3]"],
  ["${audioCodecsAsString}", "全部音频编码", "DTS, AC3"],
  ["${audioChannels}", "默认音轨声道", "6ch"],
  ["${audioChannelList}", "全部声道数组", "[6ch, 2ch]"],
  ["${audioChannelsAsString}", "全部声道", "6ch, 2ch"],
  ["${audioChannelsDot}", "默认声道点号格式", "5.1"],
  ["${audioChannelDotList}", "全部点号声道数组", "[5.1, 2.0]"],
  ["${audioChannelsDotAsString}", "全部点号声道", "5.1, 2.0"],
  ["${audioLanguage}", "默认音轨语言", "ZH"],
  ["${audioLanguageList}", "全部音轨语言数组", "[ZH, EN]"],
  ["${audioLanguagesAsString}", "全部音轨语言", "ZH, EN"],
  ["${subtitleLanguageList}", "字幕语言数组", "[ZH, EN]"],
  ["${subtitleLanguagesAsString}", "字幕语言", "ZH, EN"],
  ["${mediaSource}", "媒体来源", "BluRay"],
  ["${3Dformat}", "3D 标记", "3D SBS"],
  ["${edition}", "版本", "Final Cut"],
  ["${hdr}", "HDR", "HDR"],
  ["${hdrformat}", "HDR 格式", "HDR10"],
  ["${filesize}", "视频文件大小", "12.4 GB"],
  ["${fileSize}", "视频文件大小", "12.4GB"],
  ["${parent}", "数据源和父目录之间的路径", "科幻"],
  ["${note}", "备注", "Director's Cut"],
  ["${decadeLong}", "年代", "1980s"],
  ["${decadeShort}", "年代短格式", "80s"],
  ["${crc32}", "CRC32", "A1B2C3D4"],
];

const TV_RENAMER_TOKEN_ROWS = [
  ["${showTitle}", "电视剧标题", "绝命毒师"],
  ["${showOriginalTitle}", "电视剧原始标题", "Breaking Bad"],
  ["${showTitleSortable}", "电视剧排序标题", "Breaking Bad"],
  ["${showImdb}", "电视剧 IMDb ID", "tt0903747"],
  ["${showTmdb}", "电视剧 TMDb ID", "1396"],
  ["${showTvdb}", "电视剧 TVDB ID", "81189"],
  ["${showYear}", "电视剧年份", "2008"],
  ["${showStatus}", "电视剧状态", "Ended"],
  ["${showTags[0]}", "剧集第一个标签", "Crime"],
  ["${title}", "单集标题", "试播集"],
  ["${originalTitle}", "单集原始标题", "Pilot"],
  ["${originalFilename}", "单集原始文件名", "Breaking.Bad.S01E01.1080p.EAC3.mkv"],
  ["${originalBasename}", "不带扩展名的原始文件名", "Breaking.Bad.S01E01.1080p.EAC3"],
  ["${titleSortable}", "单集排序标题", "Pilot"],
  ["${year}", "单集年份", "2008"],
  ["${airedDate}", "单集首播日期", "2008-01-20"],
  ["${aired}", "单集首播日期", "2008-01-20"],
  ["${seasonNr}", "季编号", "1"],
  ["${seasonNr2}", "两位季编号", "01"],
  ["${seasonNrAired}", "播出季编号", "1"],
  ["${seasonNrAired2}", "两位播出季编号", "01"],
  ["${seasonName}", "季名称", "Season 1"],
  ["${seasonNrDvd}", "DVD 季编号", "1"],
  ["${seasonNrDvd2}", "两位 DVD 季编号", "01"],
  ["${episodeNr}", "集编号", "1"],
  ["${episodeNr2}", "两位集编号", "01"],
  ["${episodeNrAired}", "播出集编号", "1"],
  ["${episodeNrAired2}", "两位播出集编号", "01"],
  ["${episodeNrDvd}", "DVD 集编号", "1"],
  ["${episodeNrDvd2}", "两位 DVD 集编号", "01"],
  ["${absoluteNr}", "绝对集编号", "1"],
  ["${absoluteNr2}", "两位绝对集编号", "01"],
  ["${episodeImdb}", "单集 IMDb ID", "tt0959621"],
  ["${episodeTmdb}", "单集 TMDb ID", "62085"],
  ["${episodeTvdb}", "单集 TVDB ID", "349232"],
  ["${episodeTags[0]}", "单集第一个标签", "Pilot"],
  ["${videoResolution}", "视频分辨率", "1920x1080"],
  ["${aspectRatio}", "画面比例", "178"],
  ["${aspectRatio2}", "第二画面比例", ""],
  ["${videoCodec}", "视频编码", "H.264"],
  ["${videoFormat}", "视频格式", "1080p"],
  ["${videoBitDepth}", "视频位深", "10"],
  ["${videoBitRate}", "视频码率", "10.5 Mbps"],
  ["${framerate}", "帧率", "23.976"],
  ["${audioCodec}", "第一音轨编码", "EAC3"],
  ["${audioCodecList}", "全部音频编码数组", "[EAC3, AC3]"],
  ["${audioCodecsAsString}", "全部音频编码", "EAC3, AC3"],
  ["${audioChannels}", "默认音轨声道", "6ch"],
  ["${audioChannelList}", "全部声道数组", "[6ch, 2ch]"],
  ["${audioChannelsAsString}", "全部声道", "6ch, 2ch"],
  ["${audioChannelsDot}", "默认声道点号格式", "5.1"],
  ["${audioChannelDotList}", "全部点号声道数组", "[5.1, 2.0]"],
  ["${audioChannelsDotAsString}", "全部点号声道", "5.1, 2.0"],
  ["${audioLanguage}", "第一音轨语言", "ZH"],
  ["${audioLanguageList}", "全部音轨语言数组", "[ZH, EN]"],
  ["${audioLanguagesAsString}", "全部音轨语言", "ZH, EN"],
  ["${subtitleLanguageList}", "字幕语言数组", "[ZH, EN]"],
  ["${subtitleLanguagesAsString}", "字幕语言", "ZH, EN"],
  ["${mediaSource}", "媒体来源", "WEB-DL"],
  ["${3Dformat}", "3D 标记", ""],
  ["${hdr}", "HDR", "HDR"],
  ["${hdrformat}", "HDR 格式", "HDR10"],
  ["${filesize}", "单集文件大小", "2.1 GB"],
  ["${fileSize}", "单集文件大小", "2.1GB"],
  ["${parent}", "数据源和剧集父目录之间的路径", "美剧"],
  ["${showNote}", "电视剧备注", ""],
  ["${note}", "单集备注", ""],
  ["${showGenres[0]}", "电视剧第一个类型", "剧情"],
  ["${showGenresAsString}", "电视剧全部类型", "剧情, 犯罪"],
  ["${showGenres[0].name}", "电视剧第一个英文类型", "Drama"],
  ["${showRating}", "电视剧评分", "8.9"],
  ["${episodeRating}", "单集评分", "8.2"],
  ["${showProductionCompany}", "电视剧制作公司", "AMC"],
  ["${showProductionCompanyAsArray[0]}", "第一制作公司", "AMC"],
  ["${showCertification}", "电视剧分级", "TV-MA"],
  ["${crc32}", "CRC32", "A1B2C3D4"],
];

const DEFAULT_MOVIE_RENAMER_PATH = "${title} ${- ,edition,} (${year}) ${videoFormat} - ${fileSize}";
const DEFAULT_MOVIE_RENAMER_FILE = "${title} ${- ,edition,} (${year}) ${videoFormat} ${audioCodec} ${fileSize}";
const DEFAULT_TVSHOW_RENAMER_PATH = "${showTitle} (${showYear})";
const DEFAULT_TVSHOW_RENAMER_SEASON = "Season ${seasonNr2}";
const DEFAULT_TVSHOW_RENAMER_FILE = "${showTitle}.S${seasonNr2}E${episodeNr2}.${title}";

createApp({
  data() {
    return {
      libraries: [],
      activeModule: "movie",
      selectedLibrary: null,
      items: [],
      selectedItem: null,
      selectedItemIds: [],
      lastSelectedTVItemId: "",
      candidates: [],
      scrapeSearch: {
        query: "",
        year: "",
      },
      rename: {
        pattern: DEFAULT_MOVIE_RENAMER_FILE,
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
        movieRenameAfterScrape: true,
        tvShowRenameAfterScrape: true,
        movieScraperFields: [],
        tvShowScraperFields: [],
        tvEpisodeScraperFields: [],
        movieRenamerPathname: DEFAULT_MOVIE_RENAMER_PATH,
        movieRenamerFilename: DEFAULT_MOVIE_RENAMER_FILE,
        movieRenamerPathSpaceSubstitution: false,
        movieRenamerPathSpaceReplacement: "_",
        movieRenamerFilenameSpaceSubstitution: false,
        movieRenamerFilenameSpaceReplacement: "_",
        movieRenamerColonReplacement: "-",
        movieRenamerAsciiReplacement: false,
        movieRenamerFirstCharacterReplacement: "#",
        movieRenamerCreateSingleMovieSet: false,
        movieRenamerNfoCleanup: false,
        movieRenamerCleanupUnwanted: false,
        movieRenamerAllowMerge: false,
        tvShowRenamerShowFolder: DEFAULT_TVSHOW_RENAMER_PATH,
        tvShowRenamerSeason: DEFAULT_TVSHOW_RENAMER_SEASON,
        tvShowRenamerFilename: DEFAULT_TVSHOW_RENAMER_FILE,
        tvShowRenamerShowFolderSpaceSubstitution: false,
        tvShowRenamerShowFolderSpaceReplacement: "_",
        tvShowRenamerSeasonFolderSpaceSubstitution: false,
        tvShowRenamerSeasonFolderSpaceReplacement: "_",
        tvShowRenamerFilenameSpaceSubstitution: false,
        tvShowRenamerFilenameSpaceReplacement: "_",
        tvShowRenamerColonReplacement: " ",
        tvShowRenamerAsciiReplacement: false,
        tvShowRenamerFirstCharacterReplacement: "#",
        tvShowRenamerCleanupUnwanted: false,
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
        movieColumns: [210, 72, 72, 112, 64, 82],
        tvColumns: [230, 72, 72, 112, 150],
        resizing: null,
      },
      sortKey: "dateAdded",
      sortDirection: "desc",
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
      localRename: {
        open: false,
        mode: "movie",
        tab: "replace",
        saving: false,
        error: "",
        replaceText: "",
        replaceWith: "",
        addPosition: "prefix",
        addText: "",
        rows: [],
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
      DEFAULT_MOVIE_RENAMER_PATH,
      DEFAULT_MOVIE_RENAMER_FILE,
      DEFAULT_TVSHOW_RENAMER_PATH,
      DEFAULT_TVSHOW_RENAMER_SEASON,
      DEFAULT_TVSHOW_RENAMER_FILE,
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
      if (this.renamePreview.operations && this.renamePreview.operations.length) {
        return this.renamePreview.operations
          .map((op) => `${op.kind || "file"}:\n${op.source}\n=>\n${op.target}`)
          .join("\n\n");
      }
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
    movieGridStyle() {
      return {
        gridTemplateColumns: this.layout.movieColumns.map((width) => `${width}px`).join(" "),
      };
    },
    tvGridStyle() {
      return {
        gridTemplateColumns: this.layout.tvColumns.map((width) => `${width}px`).join(" "),
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
            { id: "videoFormat", label: "视频格式", type: "choice", available: true },
            { id: "videoCodec", label: "视频编码", type: "choice", available: false },
            { id: "videoBitrate", label: "视频码率", type: "number", available: false },
            { id: "videoBitdepth", label: "视频位深", type: "number", available: false },
            { id: "videoContainer", label: "容器", type: "choice", available: false },
            { id: "aspectRatio", label: "宽高比", type: "choice", available: false },
            { id: "frameRate", label: "帧率", type: "number", available: false },
            { id: "hdrFormat", label: "HDR", type: "choice", available: false },
            { id: "videoFilesize", label: "文件大小(GB)", type: "number", available: true },
          ],
        },
        {
          id: "audio",
          label: "音频",
          filters: [
            { id: "audioCodec", label: "音频编码", type: "choice", available: true },
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
        { id: "videoFilesize", label: "文件大小" },
        { id: "videoFormat", label: "视频格式" },
        { id: "audioCodec", label: "音频编码" },
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
      if (this.activeModule === "tvshow" && this.selectedItemIds.length > 1) return `已选择 ${this.selectedItemIds.length} / ${this.visibleItems.length}`;
      return `已选择 1 / ${this.visibleItems.length}`;
    },
    localRenamePreviewRows() {
      return this.localRename.rows.map((row) => ({
        ...row,
        previewName: this.localRenamePreviewName(row),
      }));
    },
    localRenameCanApply() {
      if (!this.localRename.open || this.localRename.saving || !this.localRename.rows.length) return false;
      const rows = this.localRenamePreviewRows;
      return rows.every((row) => row.previewName) && rows.some((row) => row.previewName !== row.fileName);
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
          rating: this.tvShowRating(show),
          showMediaFields: false,
          showRatingField: true,
          itemCount: show.episodes,
          folderPath: first ? this.showRootPath(first) : "",
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
          rating: 0,
          showMediaFields: false,
          showRatingField: false,
          season: season.season,
          itemCount: season.items.length,
          folderPath: first ? first.dir || this.showRootPath(first) : "",
        });
      }
      if (!this.selectedItem) return null;
      return this.buildSummary(this.selectedItem, {
        entityType: this.selectedItem.kind === "tvshow" ? "episode" : "movie",
        title: this.selectedItem.kind === "tvshow" ? this.selectedItem.showGuess || this.selectedItem.titleGuess : this.selectedItem.titleGuess,
        subtitle: this.selectedItem.kind === "tvshow" ? this.itemSeasonText(this.selectedItem) : this.selectedItem.originalTitle || this.selectedItem.original || "",
        year: this.selectedItem.yearGuess,
        rating: this.selectedItem.kind === "tvshow" ? this.tvEpisodeRating(this.selectedItem) : this.selectedItem.rating || 0,
        showMediaFields: true,
        showRatingField: true,
        folderPath: this.selectedItem.dir || this.selectedItem.sourcePath || "",
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
      const movieColumns = this.loadColumnWidths("tmmweb.movieColumns", this.layout.movieColumns);
      const tvColumns = this.loadColumnWidths("tmmweb.tvColumns", this.layout.tvColumns);
      if (browserWidth >= 420) this.layout.browserWidth = browserWidth;
      if (filterNavWidth >= 140) this.layout.filterNavWidth = filterNavWidth;
      this.layout.movieColumns = movieColumns;
      this.layout.tvColumns = tvColumns;
    },
    loadColumnWidths(key, fallback) {
      try {
        const values = JSON.parse(localStorage.getItem(key) || "[]");
        if (!Array.isArray(values) || values.length !== fallback.length) return fallback.slice();
        return values.map((value, index) => Math.max(this.minColumnWidth(index), Number(value) || fallback[index]));
      } catch (_) {
        return fallback.slice();
      }
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
    startColumnResize(kind, index, event) {
      const columns = kind === "movie" ? this.layout.movieColumns : this.layout.tvColumns;
      this.layout.resizing = {
        type: "column",
        kind,
        index,
        startX: event.clientX,
        startWidth: columns[index],
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
      if (resizing.type === "column") {
        const key = resizing.kind === "movie" ? "movieColumns" : "tvColumns";
        const columns = this.layout[key].slice();
        columns[resizing.index] = Math.max(this.minColumnWidth(resizing.index), resizing.startWidth + delta);
        this.layout[key] = columns;
      }
    },
    stopResize() {
      if (!this.layout.resizing) return;
      if (this.layout.browserWidth) localStorage.setItem("tmmweb.browserWidth", String(Math.round(this.layout.browserWidth)));
      localStorage.setItem("tmmweb.filterNavWidth", String(Math.round(this.layout.filterNavWidth)));
      localStorage.setItem("tmmweb.movieColumns", JSON.stringify(this.layout.movieColumns.map((value) => Math.round(value))));
      localStorage.setItem("tmmweb.tvColumns", JSON.stringify(this.layout.tvColumns.map((value) => Math.round(value))));
      this.layout.resizing = null;
    },
    minColumnWidth(index) {
      return index === 0 ? 140 : 56;
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
      if (filter.id === "videoFormat") return this.uniqueItemValues("videoFormat");
      if (filter.id === "audioCodec") return this.uniqueItemValues("audioCodec");
      if (filter.id === "decade") return this.decadeOptions;
      return [];
    },
    uniqueItemValues(field) {
      return [...new Set(this.items.map((item) => item[field]).filter(Boolean))]
        .sort((a, b) => this.compareText(a, b));
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
          return [item.titleGuess, item.showGuess, item.originalTitle || item.original, item.yearGuess, item.matchedName, item.fileName, item.path, item.imdbId, item.matchedId, item.videoFormat, item.audioCodec, item.fileSize, ...(item.genres || [])].join(" ");
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
        case "videoFormat":
          return item.videoFormat || "";
        case "audioCodec":
          return item.audioCodec || "";
        case "videoFilesize":
          return Number(item.fileSizeBytes || 0) / (1024 * 1024 * 1024);
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
      return items.sort((a, b) =>
        (a.season || 0) - (b.season || 0) ||
        (a.episode || 0) - (b.episode || 0) ||
        this.compareEpisodeList(a.episodes, b.episodes) ||
        this.compareText(a.fileName, b.fileName) ||
        this.compareText(a.path, b.path)
      );
    },
    compareEpisodeList(a = [], b = []) {
      const left = Array.isArray(a) ? a : [];
      const right = Array.isArray(b) ? b : [];
      const length = Math.max(left.length, right.length);
      for (let index = 0; index < length; index += 1) {
        const result = Number(left[index] || 0) - Number(right[index] || 0);
        if (result !== 0) return result;
      }
      return 0;
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
      if (["year", "rating", "runtime", "season", "episode", "videoFilesize"].includes(key)) {
        return Number(a[this.sortFieldName(key)] || 0) - Number(b[this.sortFieldName(key)] || 0);
      }
      if (key === "dateAdded") return this.compareDate(a.dateAdded, b.dateAdded);
      if (key === "metadata") return Number(!!a.hasNfo) - Number(!!b.hasNfo);
      if (key === "artwork") return this.artworkScore(a) - this.artworkScore(b);
      if (key === "datasource") return this.compareText(a.sourcePath, b.sourcePath);
      if (key === "videoFormat") return this.compareText(a.videoFormat, b.videoFormat);
      if (key === "audioCodec") return this.compareText(a.audioCodec, b.audioCodec);
      if (key === "originalTitle") return this.compareText(a.originalTitle || a.original, b.originalTitle || b.original);
      return this.compareText(a.kind === "tvshow" ? a.showGuess || a.titleGuess : a.titleGuess, b.kind === "tvshow" ? b.showGuess || b.titleGuess : b.titleGuess);
    },
    sortFieldName(key) {
      if (key === "year") return "yearGuess";
      if (key === "rating") return "rating";
      if (key === "runtime") return "runtime";
      if (key === "season") return "season";
      if (key === "episode") return "episode";
      if (key === "videoFilesize") return "fileSizeBytes";
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
    tvEpisodeRating(item) {
      if (!item || item.kind !== "tvshow") return Number(item?.rating || 0);
      const rating = Number(item.rating || 0);
      if (rating <= 0) return 0;
      const showRating = Number(item.showRating || 0);
      if (showRating > 0 && Math.abs(rating - showRating) < 0.05) return 0;
      return rating;
    },
    tvShowRating(show) {
      if (!show) return 0;
      const items = show.items || [];
      const withShowRating = items.find((item) => Number(item.showRating || 0) > 0);
      if (withShowRating) return Number(withShowRating.showRating || 0);
      const first = this.firstTVItem(show);
      return first ? Number(first.showRating || 0) : 0;
    },
    artworkURL(item, type, entityType = "") {
      if (!item) return "";
      if (item.kind !== "tvshow" && ((type === "poster" && !item.hasPoster) || (type === "fanart" && !item.hasFanart))) return "";
      const scope = item.kind === "tvshow" ? entityType || "episode" : "movie";
      return `/api/artwork?id=${encodeURIComponent(item.id)}&type=${encodeURIComponent(type)}&scope=${encodeURIComponent(scope)}&v=${encodeURIComponent(item.dateAdded || item.matchedId || "")}`;
    },
    hideBrokenImage(event) {
      if (event && event.target) event.target.style.display = "none";
    },
    showLoadedImage(event) {
      if (event && event.target) event.target.style.display = "";
    },
    buildSummary(item, overrides = {}) {
      const fallback = item || {};
      const entityType = overrides.entityType || fallback.kind || "movie";
      return {
        item: fallback,
        entityType,
        title: overrides.title || fallback.titleGuess || fallback.showGuess || "未命名",
        subtitle: overrides.subtitle || fallback.originalTitle || fallback.original || "",
        year: overrides.year || fallback.yearGuess || "",
        season: overrides.season || fallback.season || 0,
        itemCount: overrides.itemCount || 1,
        poster: this.artworkURL(fallback, "poster", entityType),
        fanart: this.artworkURL(fallback, "fanart", entityType),
        rating: Object.prototype.hasOwnProperty.call(overrides, "rating") ? overrides.rating : fallback.rating || 0,
        showMediaFields: Object.prototype.hasOwnProperty.call(overrides, "showMediaFields") ? overrides.showMediaFields : true,
        showRatingField: Object.prototype.hasOwnProperty.call(overrides, "showRatingField") ? overrides.showRatingField : true,
        genres: fallback.genres || [],
        actors: fallback.actors || [],
        overview: fallback.overview || "",
        dateAdded: fallback.dateAdded || "",
        matchedName: fallback.matchedName || "",
        matchedId: fallback.matchedId || 0,
        imdbId: fallback.imdbId || "",
        fileSize: fallback.fileSize || "",
        fileSizeBytes: fallback.fileSizeBytes || 0,
        videoFormat: fallback.videoFormat || "",
        audioCodec: fallback.audioCodec || "",
        hasNfo: !!fallback.hasNfo,
        hasPoster: !!fallback.hasPoster,
        hasFanart: !!fallback.hasFanart,
        hasSubtitle: !!fallback.hasSubtitle,
        sourcePath: fallback.sourcePath || "",
        folderPath: overrides.folderPath || fallback.dir || fallback.sourcePath || "",
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
    fieldList(kind) {
      return (TMM_SCRAPER_FIELDS[kind] || []).flatMap((group) => group.items.map((item) => item[0]));
    },
    fieldGroups(kind) {
      return TMM_SCRAPER_FIELDS[kind] || [];
    },
    normalizeFieldList(values, kind) {
      const allowed = this.fieldList(kind);
      if (!Array.isArray(values)) return allowed.slice();
      if (!values.length) return [];
      const selected = values.filter((value) => allowed.includes(value));
      return selected.length ? [...new Set(selected)] : [];
    },
    scraperFieldSetting(kind) {
      if (kind === "movie") return "movieScraperFields";
      if (kind === "episode") return "tvEpisodeScraperFields";
      return "tvShowScraperFields";
    },
    scraperFields(kind) {
      return this.scraperSettings[this.scraperFieldSetting(kind)] || [];
    },
    scraperFieldSelected(kind, key) {
      return this.scraperFields(kind).includes(key);
    },
    toggleScraperField(kind, key, event) {
      const setting = this.scraperFieldSetting(kind);
      const selected = new Set(this.scraperSettings[setting] || []);
      if (event.target.checked) selected.add(key);
      else selected.delete(key);
      this.scraperSettings[setting] = [...selected].filter((value) => this.fieldList(kind).includes(value));
    },
    selectAllScraperFields(kind) {
      this.scraperSettings[this.scraperFieldSetting(kind)] = this.fieldList(kind);
    },
    clearScraperFields(kind) {
      this.scraperSettings[this.scraperFieldSetting(kind)] = [];
    },
    renamerTokens() {
      return TMM_RENAMER_TOKENS;
    },
    renamerTokenRows(kind) {
      return kind === "tvshow" ? TV_RENAMER_TOKEN_ROWS : MOVIE_RENAMER_TOKEN_ROWS;
    },
    resetRenamerPattern(kind, field) {
      const defaults = {
        moviePath: DEFAULT_MOVIE_RENAMER_PATH,
        movieFile: DEFAULT_MOVIE_RENAMER_FILE,
        tvShow: DEFAULT_TVSHOW_RENAMER_PATH,
        tvSeason: DEFAULT_TVSHOW_RENAMER_SEASON,
        tvFile: DEFAULT_TVSHOW_RENAMER_FILE,
      };
      const setting = {
        moviePath: "movieRenamerPathname",
        movieFile: "movieRenamerFilename",
        tvShow: "tvShowRenamerShowFolder",
        tvSeason: "tvShowRenamerSeason",
        tvFile: "tvShowRenamerFilename",
      }[`${kind}${field}`];
      const value = defaults[`${kind}${field}`];
      if (setting && value) this.scraperSettings[setting] = value;
    },
    renameOptions(kind, segment) {
      if (kind === "movie" && segment === "folder") {
        return {
          spaceSubstitution: !!this.scraperSettings.movieRenamerPathSpaceSubstitution,
          spaceReplacement: this.scraperSettings.movieRenamerPathSpaceReplacement || "_",
          colonReplacement: this.scraperSettings.movieRenamerColonReplacement || "-",
          asciiReplacement: !!this.scraperSettings.movieRenamerAsciiReplacement,
        };
      }
      if (kind === "movie") {
        return {
          spaceSubstitution: !!this.scraperSettings.movieRenamerFilenameSpaceSubstitution,
          spaceReplacement: this.scraperSettings.movieRenamerFilenameSpaceReplacement || "_",
          colonReplacement: this.scraperSettings.movieRenamerColonReplacement || "-",
          asciiReplacement: !!this.scraperSettings.movieRenamerAsciiReplacement,
        };
      }
      if (segment === "show") {
        return {
          spaceSubstitution: !!this.scraperSettings.tvShowRenamerShowFolderSpaceSubstitution,
          spaceReplacement: this.scraperSettings.tvShowRenamerShowFolderSpaceReplacement || "_",
          colonReplacement: this.scraperSettings.tvShowRenamerColonReplacement || " ",
          asciiReplacement: !!this.scraperSettings.tvShowRenamerAsciiReplacement,
        };
      }
      if (segment === "season") {
        return {
          spaceSubstitution: !!this.scraperSettings.tvShowRenamerSeasonFolderSpaceSubstitution,
          spaceReplacement: this.scraperSettings.tvShowRenamerSeasonFolderSpaceReplacement || "_",
          colonReplacement: this.scraperSettings.tvShowRenamerColonReplacement || " ",
          asciiReplacement: !!this.scraperSettings.tvShowRenamerAsciiReplacement,
        };
      }
      return {
        spaceSubstitution: !!this.scraperSettings.tvShowRenamerFilenameSpaceSubstitution,
        spaceReplacement: this.scraperSettings.tvShowRenamerFilenameSpaceReplacement || "_",
        colonReplacement: this.scraperSettings.tvShowRenamerColonReplacement || " ",
        asciiReplacement: !!this.scraperSettings.tvShowRenamerAsciiReplacement,
      };
    },
    previewPattern(pattern, kind, segment = "file") {
      const replacements = {
        "{title}": "银翼杀手",
        "${title}": "银翼杀手",
        "{title[0]}": "银",
        "${title[0]}": "银",
        "{title;first}": this.firstCharacterToken("银翼杀手", this.scraperSettings.movieRenamerFirstCharacterReplacement),
        "${title;first}": this.firstCharacterToken("银翼杀手", this.scraperSettings.movieRenamerFirstCharacterReplacement),
        "{title[0,2]}": "银翼",
        "${title[0,2]}": "银翼",
        "{originalTitle}": "Blade Runner",
        "${originalTitle}": "Blade Runner",
        "{originalFilename}": "Blade.Runner.1982.1080p.DTS.mkv",
        "${originalFilename}": "Blade.Runner.1982.1080p.DTS.mkv",
        "{originalBasename}": "Blade.Runner.1982.1080p.DTS",
        "${originalBasename}": "Blade.Runner.1982.1080p.DTS",
        "{edition}": "Final Cut",
        "${edition}": "Final Cut",
        "${- ,edition,}": "- Final Cut",
        "{year}": "1982",
        "${year}": "1982",
        "{releaseDate}": "1982-06-25",
        "${releaseDate}": "1982-06-25",
        "{rating}": "8.1",
        "${rating}": "8.1",
        "{imdb}": "tt0083658",
        "${imdb}": "tt0083658",
        "{tmdb}": "78",
        "${tmdb}": "78",
        "{tmdbid}": "78",
        "${tmdbid}": "78",
        "{videoFormat}": "1080p",
        "${videoFormat}": "1080p",
        "{audioCodec}": "DTS",
        "${audioCodec}": "DTS",
        "{fileSize}": "12.4GB",
        "${fileSize}": "12.4GB",
        "{filesize}": "12.4GB",
        "${filesize}": "12.4GB",
        "{showTitle}": "绝命毒师",
        "${showTitle}": "绝命毒师",
        "{showOriginalTitle}": "Breaking Bad",
        "${showOriginalTitle}": "Breaking Bad",
        "{showYear}": "2008",
        "${showYear}": "2008",
        "{seasonNr}": "1",
        "${seasonNr}": "1",
        "{seasonNr2}": "01",
        "${seasonNr2}": "01",
        "{episodeNr}": "1",
        "${episodeNr}": "1",
        "{episodeNr2}": "01",
        "${episodeNr2}": "01",
        "{aired}": "2008-01-20",
        "${aired}": "2008-01-20",
        "{airedDate}": "2008-01-20",
        "${airedDate}": "2008-01-20",
      };
      if (kind === "tvshow") {
        Object.assign(replacements, {
          "{title}": "试播集",
          "${title}": "试播集",
          "{title[0]}": "试",
          "${title[0]}": "试",
          "{title;first}": this.firstCharacterToken("试播集", this.scraperSettings.tvShowRenamerFirstCharacterReplacement),
          "${title;first}": this.firstCharacterToken("试播集", this.scraperSettings.tvShowRenamerFirstCharacterReplacement),
          "{title[0,2]}": "试播",
          "${title[0,2]}": "试播",
          "{originalTitle}": "Pilot",
          "${originalTitle}": "Pilot",
          "{originalFilename}": "Breaking.Bad.S01E01.1080p.EAC3.mkv",
          "${originalFilename}": "Breaking.Bad.S01E01.1080p.EAC3.mkv",
          "{originalBasename}": "Breaking.Bad.S01E01.1080p.EAC3",
          "${originalBasename}": "Breaking.Bad.S01E01.1080p.EAC3",
          "{audioCodec}": "EAC3",
          "${audioCodec}": "EAC3",
          "{fileSize}": "2.1GB",
          "${fileSize}": "2.1GB",
          "{filesize}": "2.1GB",
          "${filesize}": "2.1GB",
        });
      }
      let value = pattern || (kind === "movie" ? DEFAULT_MOVIE_RENAMER_FILE : DEFAULT_TVSHOW_RENAMER_FILE);
      Object.keys(replacements)
        .sort((a, b) => b.length - a.length)
        .forEach((key) => {
          value = value.split(key).join(replacements[key]);
        });
      return this.cleanPreviewName(value, this.renameOptions(kind, segment));
    },
    cleanPreviewName(value, options = {}) {
      let result = String(value || "").replace(/\$\{[^}]+\}/g, "");
      const colonReplacement = options.colonReplacement === undefined || options.colonReplacement === null ? "-" : options.colonReplacement;
      result = result.replace(/[:：]/g, colonReplacement);
      result = result.replace(/[\\/]/g, " ").replace(/[*?"]/g, "").replace(/[<>]/g, "").replace(/\|/g, " ");
      if (options.asciiReplacement) {
        result = result.normalize("NFKD").replace(/[\u0300-\u036f]/g, "").replace(/[^\x00-\x7F]/g, "");
      }
      result = result.replace(/\s+/g, " ").replace(/\( \)/g, "").replace(/\[ \]/g, "").replace(/[ ._-]+$/g, "").trim();
      if (options.spaceSubstitution) {
        const replacement = options.spaceReplacement || "_";
        const escaped = replacement.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        result = result.replace(/ /g, replacement).replace(new RegExp(`${escaped}+`, "g"), replacement).replace(/[ ._-]+$/g, "").trim();
      }
      return result;
    },
    firstCharacterToken(value, replacement) {
      const text = String(value || "").trim();
      if (!text) return "";
      const first = [...text][0];
      if (/^\p{L}$/u.test(first)) return first;
      return String(replacement || "#").trim() || "#";
    },
    movieRenamerExample() {
      return {
        movie: "银翼杀手",
        datasource: "/media/movies",
        folder: this.previewPattern(this.scraperSettings.movieRenamerPathname, "movie", "folder"),
        filename: `${this.previewPattern(this.scraperSettings.movieRenamerFilename, "movie", "file")}.mkv`,
      };
    },
    tvRenamerExample() {
      const show = this.previewPattern(this.scraperSettings.tvShowRenamerShowFolder, "tvshow", "show");
      const season = this.previewPattern(this.scraperSettings.tvShowRenamerSeason, "tvshow", "season");
      return {
        show: "绝命毒师",
        episode: "1.1 试播集",
        datasource: "/media/tv",
        folder: `${show}/${season}`,
        filename: `${this.previewPattern(this.scraperSettings.tvShowRenamerFilename, "tvshow", "file")}.mkv`,
      };
    },
    applySettings(settings) {
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
      this.scraperSettings.movieRenameAfterScrape = !!settings.movieRenameAfterScrape;
      this.scraperSettings.tvShowRenameAfterScrape = !!settings.tvShowRenameAfterScrape;
      this.scraperSettings.movieScraperFields = this.normalizeFieldList(settings.movieScraperFields, "movie");
      this.scraperSettings.tvShowScraperFields = this.normalizeFieldList(settings.tvShowScraperFields, "tvshow");
      this.scraperSettings.tvEpisodeScraperFields = this.normalizeFieldList(settings.tvEpisodeScraperFields, "episode");
      this.scraperSettings.movieRenamerPathname = settings.movieRenamerPathname || DEFAULT_MOVIE_RENAMER_PATH;
      this.scraperSettings.movieRenamerFilename = settings.movieRenamerFilename || DEFAULT_MOVIE_RENAMER_FILE;
      this.scraperSettings.movieRenamerPathSpaceSubstitution = !!settings.movieRenamerPathSpaceSubstitution;
      this.scraperSettings.movieRenamerPathSpaceReplacement = settings.movieRenamerPathSpaceReplacement || "_";
      this.scraperSettings.movieRenamerFilenameSpaceSubstitution = !!settings.movieRenamerFilenameSpaceSubstitution;
      this.scraperSettings.movieRenamerFilenameSpaceReplacement = settings.movieRenamerFilenameSpaceReplacement || "_";
      this.scraperSettings.movieRenamerColonReplacement = settings.movieRenamerColonReplacement || "-";
      this.scraperSettings.movieRenamerAsciiReplacement = !!settings.movieRenamerAsciiReplacement;
      this.scraperSettings.movieRenamerFirstCharacterReplacement = settings.movieRenamerFirstCharacterReplacement || "#";
      this.scraperSettings.movieRenamerCreateSingleMovieSet = !!settings.movieRenamerCreateSingleMovieSet;
      this.scraperSettings.movieRenamerNfoCleanup = !!settings.movieRenamerNfoCleanup;
      this.scraperSettings.movieRenamerCleanupUnwanted = !!settings.movieRenamerCleanupUnwanted;
      this.scraperSettings.movieRenamerAllowMerge = !!settings.movieRenamerAllowMerge;
      this.scraperSettings.tvShowRenamerShowFolder = settings.tvShowRenamerShowFolder || DEFAULT_TVSHOW_RENAMER_PATH;
      this.scraperSettings.tvShowRenamerSeason = settings.tvShowRenamerSeason || DEFAULT_TVSHOW_RENAMER_SEASON;
      this.scraperSettings.tvShowRenamerFilename = settings.tvShowRenamerFilename || DEFAULT_TVSHOW_RENAMER_FILE;
      this.scraperSettings.tvShowRenamerShowFolderSpaceSubstitution = !!settings.tvShowRenamerShowFolderSpaceSubstitution;
      this.scraperSettings.tvShowRenamerShowFolderSpaceReplacement = settings.tvShowRenamerShowFolderSpaceReplacement || "_";
      this.scraperSettings.tvShowRenamerSeasonFolderSpaceSubstitution = !!settings.tvShowRenamerSeasonFolderSpaceSubstitution;
      this.scraperSettings.tvShowRenamerSeasonFolderSpaceReplacement = settings.tvShowRenamerSeasonFolderSpaceReplacement || "_";
      this.scraperSettings.tvShowRenamerFilenameSpaceSubstitution = !!settings.tvShowRenamerFilenameSpaceSubstitution;
      this.scraperSettings.tvShowRenamerFilenameSpaceReplacement = settings.tvShowRenamerFilenameSpaceReplacement || "_";
      this.scraperSettings.tvShowRenamerColonReplacement = settings.tvShowRenamerColonReplacement || " ";
      this.scraperSettings.tvShowRenamerAsciiReplacement = !!settings.tvShowRenamerAsciiReplacement;
      this.scraperSettings.tvShowRenamerFirstCharacterReplacement = settings.tvShowRenamerFirstCharacterReplacement || "#";
      this.scraperSettings.tvShowRenamerCleanupUnwanted = !!settings.tvShowRenamerCleanupUnwanted;
      this.scraperSettings.moviePosterName = settings.moviePosterName || "poster.jpg";
      this.scraperSettings.movieFanartName = settings.movieFanartName || "fanart.jpg";
      this.scraperSettings.moviePosterNames = settings.moviePosterNames || "poster.jpg\nfolder.jpg\n{filename}-poster.jpg";
      this.scraperSettings.movieFanartNames = settings.movieFanartNames || "fanart.jpg\n{filename}-fanart.jpg";
      this.scraperSettings.tvShowPosterName = settings.tvShowPosterName || "poster.jpg";
      this.scraperSettings.tvShowFanartName = settings.tvShowFanartName || "fanart.jpg";
      this.scraperSettings.tvShowPosterNames = settings.tvShowPosterNames || "poster.jpg\nfolder.jpg";
      this.scraperSettings.tvShowFanartNames = settings.tvShowFanartNames || "fanart.jpg\nbackdrop.jpg";
    },
    settingsPayload(extra = {}) {
      return {
        ...extra,
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
        movieRenameAfterScrape: !!this.scraperSettings.movieRenameAfterScrape,
        tvShowRenameAfterScrape: !!this.scraperSettings.tvShowRenameAfterScrape,
        movieScraperFields: this.scraperSettings.movieScraperFields,
        tvShowScraperFields: this.scraperSettings.tvShowScraperFields,
        tvEpisodeScraperFields: this.scraperSettings.tvEpisodeScraperFields,
        movieRenamerPathname: this.scraperSettings.movieRenamerPathname,
        movieRenamerFilename: this.scraperSettings.movieRenamerFilename,
        movieRenamerPathSpaceSubstitution: !!this.scraperSettings.movieRenamerPathSpaceSubstitution,
        movieRenamerPathSpaceReplacement: this.scraperSettings.movieRenamerPathSpaceReplacement,
        movieRenamerFilenameSpaceSubstitution: !!this.scraperSettings.movieRenamerFilenameSpaceSubstitution,
        movieRenamerFilenameSpaceReplacement: this.scraperSettings.movieRenamerFilenameSpaceReplacement,
        movieRenamerColonReplacement: this.scraperSettings.movieRenamerColonReplacement,
        movieRenamerAsciiReplacement: !!this.scraperSettings.movieRenamerAsciiReplacement,
        movieRenamerFirstCharacterReplacement: this.scraperSettings.movieRenamerFirstCharacterReplacement,
        movieRenamerCreateSingleMovieSet: !!this.scraperSettings.movieRenamerCreateSingleMovieSet,
        movieRenamerNfoCleanup: !!this.scraperSettings.movieRenamerNfoCleanup,
        movieRenamerCleanupUnwanted: !!this.scraperSettings.movieRenamerCleanupUnwanted,
        movieRenamerAllowMerge: !!this.scraperSettings.movieRenamerAllowMerge,
        tvShowRenamerShowFolder: this.scraperSettings.tvShowRenamerShowFolder,
        tvShowRenamerSeason: this.scraperSettings.tvShowRenamerSeason,
        tvShowRenamerFilename: this.scraperSettings.tvShowRenamerFilename,
        tvShowRenamerShowFolderSpaceSubstitution: !!this.scraperSettings.tvShowRenamerShowFolderSpaceSubstitution,
        tvShowRenamerShowFolderSpaceReplacement: this.scraperSettings.tvShowRenamerShowFolderSpaceReplacement,
        tvShowRenamerSeasonFolderSpaceSubstitution: !!this.scraperSettings.tvShowRenamerSeasonFolderSpaceSubstitution,
        tvShowRenamerSeasonFolderSpaceReplacement: this.scraperSettings.tvShowRenamerSeasonFolderSpaceReplacement,
        tvShowRenamerFilenameSpaceSubstitution: !!this.scraperSettings.tvShowRenamerFilenameSpaceSubstitution,
        tvShowRenamerFilenameSpaceReplacement: this.scraperSettings.tvShowRenamerFilenameSpaceReplacement,
        tvShowRenamerColonReplacement: this.scraperSettings.tvShowRenamerColonReplacement,
        tvShowRenamerAsciiReplacement: !!this.scraperSettings.tvShowRenamerAsciiReplacement,
        tvShowRenamerFirstCharacterReplacement: this.scraperSettings.tvShowRenamerFirstCharacterReplacement,
        tvShowRenamerCleanupUnwanted: !!this.scraperSettings.tvShowRenamerCleanupUnwanted,
        moviePosterName: this.scraperSettings.moviePosterName,
        movieFanartName: this.scraperSettings.movieFanartName,
        moviePosterNames: this.scraperSettings.moviePosterNames,
        movieFanartNames: this.scraperSettings.movieFanartNames,
        tvShowPosterName: this.scraperSettings.tvShowPosterName,
        tvShowFanartName: this.scraperSettings.tvShowFanartName,
        tvShowPosterNames: this.scraperSettings.tvShowPosterNames,
        tvShowFanartNames: this.scraperSettings.tvShowFanartNames,
      };
    },
    async loadLibraries() {
      const libraries = await this.api("/api/libraries");
      this.libraries = Array.isArray(libraries) ? libraries : [];
      if (this.libraries.length && !this.selectedLibrary) {
        const first = this.filteredLibraries[0] || this.libraries[0];
        this.activeModule = first.type || "movie";
        await this.selectLibrary(first);
      }
    },
    async loadSettings() {
      try {
        const settings = await this.api("/api/settings");
        this.applySettings(settings);
        this.scraperSettings.proxyPassword = "";
        this.scraperSettings.clearProxyPassword = false;
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
        const body = this.settingsPayload();
        if (this.scraperSettings.tmdbApiKey) body.tmdbApiKey = this.scraperSettings.tmdbApiKey;
        if (this.scraperSettings.proxyPassword) body.proxyPassword = this.scraperSettings.proxyPassword;
        const settings = await this.api("/api/settings", {
          method: "PUT",
          body: JSON.stringify(body),
        });
        this.applySettings(settings);
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
          body: JSON.stringify(this.settingsPayload({
            clearTmdbKey: true,
            clearProxyPassword: false,
          })),
        });
        this.applySettings(settings);
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
      this.sortKey = "dateAdded";
      this.sortDirection = "desc";
      this.selectedItemIds = [];
      this.lastSelectedTVItemId = "";
      const first = this.filteredLibraries[0];
      if (first) {
        await this.selectLibrary(first);
      } else {
        this.selectedLibrary = null;
        this.items = [];
        this.selectedItem = null;
        this.selectedItemIds = [];
        this.lastSelectedTVItemId = "";
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
      const previousType = this.newLibrary.type;
      const defaultNames = new Set(["电影", "电视剧"]);
      this.newLibrary.type = type;
      if (!this.newLibrary.name || (previousType !== type && defaultNames.has(this.newLibrary.name))) {
        this.newLibrary.name = type === "tvshow" ? "电视剧" : "电影";
      }
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
            this.selectedItemIds = [];
            this.lastSelectedTVItemId = "";
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
      this.selectedItemIds = [];
      this.lastSelectedTVItemId = "";
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
      const scrollAnchor = quiet ? this.captureMediaScrollAnchor() : null;
      try {
        const result = await this.api(`/api/items?libraryId=${encodeURIComponent(library.id)}`);
        const loadedItems = result.items || [];
        this.items = quiet && this.selectedTask && this.selectedTask.state === "running"
          ? this.mergeLoadedItems(this.items, loadedItems)
          : loadedItems;
        if (scrollAnchor) {
          await this.$nextTick();
          this.restoreMediaScrollAnchor(scrollAnchor);
        }
        if (!quiet) {
          this.status = this.items.length ? `已加载 ${this.items.length} 个缓存条目` : `已选择 ${library.name}，需要扫描`;
        }
      } catch (error) {
        if (!quiet) this.status = error.message;
      } finally {
        if (!quiet) this.busy = false;
      }
    },
    mergeLoadedItems(currentItems, loadedItems) {
      const loadedByID = new Map(loadedItems.map((item) => [item.id, item]));
      const merged = currentItems
        .filter((item) => loadedByID.has(item.id))
        .map((item) => loadedByID.get(item.id));
      const existingIDs = new Set(merged.map((item) => item.id));
      for (const item of loadedItems) {
        if (!existingIDs.has(item.id)) merged.push(item);
      }
      return merged;
    },
    mediaScrollElement() {
      const ref = this.$refs.mediaScroller;
      return Array.isArray(ref) ? ref.find(Boolean) : ref;
    },
    captureMediaScrollAnchor() {
      const scroller = this.mediaScrollElement();
      if (!scroller) return null;
      const scrollTop = scroller.scrollTop;
      const rows = scroller.querySelectorAll("[data-row-key]");
      for (const row of rows) {
        if (row.offsetTop + row.offsetHeight >= scrollTop) {
          return {
            key: row.dataset.rowKey,
            offset: row.offsetTop - scrollTop,
            scrollTop,
          };
        }
      }
      return { scrollTop };
    },
    restoreMediaScrollAnchor(anchor) {
      const scroller = this.mediaScrollElement();
      if (!scroller || !anchor) return;
      if (anchor.key) {
        const escapedKey = window.CSS && CSS.escape ? CSS.escape(anchor.key) : anchor.key.replace(/["\\]/g, "\\$&");
        const row = scroller.querySelector(`[data-row-key="${escapedKey}"]`);
        if (row) {
          scroller.scrollTop = Math.max(0, row.offsetTop - anchor.offset);
          return;
        }
      }
      scroller.scrollTop = anchor.scrollTop || 0;
    },
    selectItem(item, event = null) {
      const isTV = item.kind === "tvshow";
      const rangeSelect = isTV && event && event.shiftKey;
      const multiToggle = isTV && event && (event.metaKey || event.ctrlKey);
      if (rangeSelect) {
        const range = this.tvSelectionRange(item.id);
        this.selectedItemIds = range.length ? range : [item.id];
      } else if (multiToggle) {
        const selected = new Set(this.selectedItemIds);
        if (selected.has(item.id)) selected.delete(item.id);
        else selected.add(item.id);
        this.selectedItemIds = [...selected];
        this.lastSelectedTVItemId = item.id;
      } else {
        this.selectedItemIds = [item.id];
        this.lastSelectedTVItemId = isTV ? item.id : "";
      }
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
    visibleTVEpisodeItems() {
      return this.tvTree.flatMap((show) => show.seasons.flatMap((season) => season.items));
    },
    tvSelectionRange(targetId) {
      const items = this.visibleTVEpisodeItems();
      if (!items.length) return [];
      const anchorId = this.lastSelectedTVItemId || (this.selectedItem && this.selectedItem.kind === "tvshow" ? this.selectedItem.id : "");
      const targetIndex = items.findIndex((item) => item.id === targetId);
      const anchorIndex = items.findIndex((item) => item.id === anchorId);
      if (targetIndex < 0) return [targetId];
      if (anchorIndex < 0) return [targetId];
      const start = Math.min(anchorIndex, targetIndex);
      const end = Math.max(anchorIndex, targetIndex);
      return items.slice(start, end + 1).map((item) => item.id);
    },
    isItemSelected(item) {
      if (!item) return false;
      if (this.activeModule === "tvshow") return this.selectedItemIds.includes(item.id);
      return !!this.selectedItem && this.selectedItem.id === item.id;
    },
    selectedTVRenameItems(payload = null) {
      if (payload && payload.kind === "tvshow" && this.selectedItemIds.includes(payload.id)) {
        const selected = new Set(this.selectedItemIds);
        return this.visibleTVEpisodeItems().filter((item) => selected.has(item.id));
      }
      if (payload && payload.kind === "tvshow") return [payload];
      return this.selectedItem && this.selectedItem.kind === "tvshow" ? [this.selectedItem] : [];
    },
    handleKeydown(event) {
      if (event.key === "Escape") {
        this.closeContextMenu();
        if (this.chooser.open && !this.chooser.loading && !this.chooser.scraping) this.closeChooser();
        if (this.localRename.open && !this.localRename.saving) this.closeLocalRename();
      }
    },
    openContextMenu(event, scope, payload) {
      const keepTVSelection = scope === "episode" && payload && this.selectedItemIds.includes(payload.id) && this.selectedItemIds.length > 1;
      this.contextMenu = {
        open: true,
        x: event.clientX,
        y: event.clientY,
        scope,
        payload,
      };
      if (scope === "movie" || scope === "episode") {
        if (!keepTVSelection) this.selectItem(payload);
        else {
          this.selectedItem = payload;
          this.selectedEntity = { kind: "episode", payload };
        }
      }
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
    canRenameFromContext() {
      return this.contextMenu.scope === "movie" || this.contextMenu.scope === "episode";
    },
    openLocalRenameFromContext() {
      const scope = this.contextMenu.scope;
      const payload = this.contextMenu.payload;
      this.closeContextMenu();
      if (!payload) return;
      if (scope === "movie") {
        this.openLocalRename([payload], "movie");
        return;
      }
      if (scope === "episode") {
        this.openLocalRename(this.selectedTVRenameItems(payload), "tvshow");
      }
    },
    openLocalRename(items, mode) {
      const rows = items
        .filter(Boolean)
        .map((item) => ({
          itemId: item.id,
          fileName: item.fileName || this.basename(item.path),
          newFileName: item.fileName || this.basename(item.path),
        }));
      if (!rows.length) return;
      this.localRename = {
        open: true,
        mode,
        tab: mode === "movie" ? "manual" : "replace",
        saving: false,
        error: "",
        replaceText: "",
        replaceWith: "",
        addPosition: "prefix",
        addText: "",
        rows,
      };
    },
    closeLocalRename() {
      this.localRename.open = false;
      this.localRename.error = "";
    },
    basename(path) {
      return String(path || "").split("/").pop() || "";
    },
    splitFilename(fileName) {
      const index = fileName.lastIndexOf(".");
      if (index <= 0) return { base: fileName, ext: "" };
      return { base: fileName.slice(0, index), ext: fileName.slice(index) };
    },
    localRenamePreviewName(row) {
      if (this.localRename.mode === "movie" || this.localRename.tab === "manual") return row.newFileName.trim();
      if (this.localRename.tab === "replace") {
        const source = this.localRename.replaceText;
        if (!source) return row.fileName;
        return row.fileName.split(source).join(this.localRename.replaceWith);
      }
      if (this.localRename.tab === "add") {
        const addition = this.localRename.addText;
        if (!addition) return row.fileName;
        const parts = this.splitFilename(row.fileName);
        if (this.localRename.addPosition === "suffix") return `${parts.base}${addition}${parts.ext}`;
        return `${addition}${parts.base}${parts.ext}`;
      }
      return row.fileName;
    },
    async applyLocalRename() {
      if (!this.localRenameCanApply) return;
      this.localRename.saving = true;
      this.localRename.error = "";
      this.status = "正在重命名本地文件";
      try {
        const requests = this.localRenamePreviewRows.map((row) => ({
          itemId: row.itemId,
          newFileName: row.previewName,
        }));
        const result = await this.api("/api/local-rename", {
          method: "POST",
          body: JSON.stringify({ items: requests }),
        });
        const updatedItems = result.items || [];
        const byOldID = new Map(this.localRename.rows.map((row, index) => [row.itemId, updatedItems[index]]));
        const byNewID = new Map(updatedItems.map((item) => [item.id, item]));
        this.items = this.items.map((item) => byOldID.get(item.id) || byNewID.get(item.id) || item);
        this.selectedItemIds = updatedItems.map((item) => item.id);
        this.lastSelectedTVItemId = updatedItems[0] ? updatedItems[0].id : "";
        if (this.selectedItem) {
          this.selectedItem = byOldID.get(this.selectedItem.id) || byNewID.get(this.selectedItem.id) || updatedItems[0] || this.selectedItem;
        } else {
          this.selectedItem = updatedItems[0] || null;
        }
        if (this.selectedItem) this.selectedEntity = { kind: this.selectedItem.kind === "tvshow" ? "episode" : "movie", payload: this.selectedItem };
        this.status = `已重命名 ${updatedItems.length} 个文件`;
        this.closeLocalRename();
      } catch (error) {
        this.localRename.error = error.message;
        this.status = error.message;
      } finally {
        this.localRename.saving = false;
      }
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
	          renameAfterScrape: true,
	          showMetadata: this.scraperSettings.tvShowScrapeMetadata,
	          episodeMetadata: this.scraperSettings.tvShowEpisodeMetadata && tv,
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
        const tv = this.chooser.mediaType === "tvshow";
        const writeShowMeta = tv ? !!this.chooser.options.showMetadata : !!this.chooser.options.metadata;
        const writeEpisodeMeta = tv && !!this.chooser.options.episodeMetadata;
        const body = {
          itemId: this.chooser.targetItem.id,
          scope: this.chooser.scope,
          libraryId: this.selectedLibrary ? this.selectedLibrary.id : this.chooser.targetItem.libraryId,
          tmdbId: this.chooser.selected.id,
          mediaType: this.chooser.selected.mediaType || this.chooser.mediaType,
          writeNfo: !!this.chooser.options.nfo,
          writeImages: !!this.chooser.options.artwork,
          writeMeta: tv ? writeShowMeta || writeEpisodeMeta : writeShowMeta,
          writeShowMeta,
          writeEpisodeMeta,
          overwrite: !!this.chooser.options.overwrite,
          renameAfterScrape: !!this.chooser.options.renameAfterScrape,
          metadataFields: this.scraperFields(tv ? "tvshow" : "movie"),
          episodeFields: tv ? this.scraperFields("episode") : [],
        };
        if (this.chooser.targetShow) body.showName = this.chooser.targetShow.key || this.chooser.targetShow.title;
        if (this.chooser.targetSeason) body.season = this.chooser.targetSeason.season;
        const result = await this.api("/api/scrape", {
          method: "POST",
          body: JSON.stringify(body),
        });
        if (result.items && result.items.length) {
          const byID = new Map(result.items.map((item) => [item.id, item]));
          const byOldPath = new Map((result.renamePreviews || []).map((preview, index) => [preview.sourceFile, result.items[index]]));
          this.items = this.items.map((item) => byID.get(item.id) || byOldPath.get(item.path) || item);
          if (this.selectedItem && byID.has(this.selectedItem.id)) this.selectedItem = byID.get(this.selectedItem.id);
          if (this.selectedItem && byOldPath.has(this.selectedItem.path)) this.selectedItem = byOldPath.get(this.selectedItem.path);
        } else if (result.item) {
          const oldPath = result.renamePreview ? result.renamePreview.sourceFile : "";
          this.items = this.items.map((item) => (item.path === oldPath || item.path === result.item.path || item.id === result.item.id ? result.item : item));
          this.selectedItem = result.item;
        }
        const scraped = result.movie || result.show || this.chooser.detail;
        if (scraped && this.selectedItem) {
          this.rename.title = scraped.title;
          this.rename.year = (scraped.releaseDate || scraped.firstAirDate || "").slice(0, 4);
          this.rename.tmdbId = scraped.id;
        }
        this.status = result.renamed ? "刮削写入完成，已自动重命名" : result.renameWarnings ? `刮削完成，重命名未完成：${result.renameWarnings[0]}` : "刮削写入完成";
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
        this.selectedItemIds = [];
        this.lastSelectedTVItemId = "";
        this.selectedEntity = { kind: "show", payload };
        const firstSeason = payload.seasons[0];
        if (firstSeason && firstSeason.items[0]) {
          this.selectItem(firstSeason.items[0]);
          this.selectedEntity = { kind: "show", payload };
        }
      }
      if (kind === "season" && payload.items[0]) {
        this.selectedItemIds = [];
        this.lastSelectedTVItemId = "";
        this.selectItem(payload.items[0]);
        this.selectedEntity = { kind: "season", payload };
      }
    },
    itemSeasonText(item) {
      if (item.kind !== "tvshow") return item.yearGuess || "-";
      const matchedName = (item.matchedName || "").trim();
      const showName = (item.showGuess || "").trim();
      if (matchedName && matchedName !== showName) return matchedName;
      if (item.season && item.episodes && item.episodes.length) {
        return `S${String(item.season).padStart(2, "0")}E${item.episodes.map((episode) => String(episode).padStart(2, "0")).join(",")}`;
      }
      if (item.season && item.episode) return `S${String(item.season).padStart(2, "0")}E${String(item.episode).padStart(2, "0")}`;
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
            writeMeta: tv ? this.scraperSettings.tvShowScrapeMetadata || this.scraperSettings.tvShowEpisodeMetadata : this.scraperSettings.movieScrapeMetadata,
            writeShowMeta: tv ? this.scraperSettings.tvShowScrapeMetadata : this.scraperSettings.movieScrapeMetadata,
            writeEpisodeMeta: tv ? this.scraperSettings.tvShowEpisodeMetadata : false,
            overwrite: tv ? this.scraperSettings.tvShowScrapeOverwrite : this.scraperSettings.movieScrapeOverwrite,
            metadataFields: this.scraperFields(tv ? "tvshow" : "movie"),
            episodeFields: tv ? this.scraperFields("episode") : [],
          }),
        });
        this.selectedItem = result.item;
        if (result.item) {
          const oldPath = result.renamePreview ? result.renamePreview.sourceFile : "";
          this.items = this.items.map((item) => (item.id === result.item.id || item.path === oldPath ? result.item : item));
        }
        const scraped = result.movie || result.show;
        this.rename.title = scraped.title;
        this.rename.year = (scraped.releaseDate || scraped.firstAirDate || "").slice(0, 4);
        this.rename.tmdbId = scraped.id;
        this.status = result.renamed ? "刮削写入完成，已自动重命名" : result.renameWarnings ? `刮削完成，重命名未完成：${result.renameWarnings[0]}` : "刮削写入完成";
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
        const result = await this.api("/api/rename/apply", {
          method: "POST",
          body: JSON.stringify(this.renamePreview),
        });
        if (result.item) {
          const oldPath = this.renamePreview.sourceFile;
          this.items = this.items.map((item) => (item.id === result.item.id || item.path === oldPath ? result.item : item));
          this.selectedItem = result.item;
        }
        this.status = "重命名完成";
        this.renamePreview = null;
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
  },
}).mount("#app");
