import {
  DEFAULT_MOVIE_RENAMER_FILE,
  DEFAULT_MOVIE_RENAMER_PATH,
  DEFAULT_TVSHOW_RENAMER_FILE,
  DEFAULT_TVSHOW_RENAMER_PATH,
  DEFAULT_TVSHOW_RENAMER_SEASON,
  TMM_SCRAPER_FIELDS,
} from "./config.js";

export const settingsMixin = {
  computed: {
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
    filteredSettingsSections() {
      const query = this.settingsFilter.trim().toLowerCase();
      if (!query) return this.settingsSections;
      return this.settingsSections
        .map((section) => {
          const children = this.filterSettingsChildren(
            section.children,
            `${section.title} `,
            query,
          );
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
  },
  methods: {
    fieldList(kind) {
      return (TMM_SCRAPER_FIELDS[kind] || []).flatMap((group) =>
        group.items.map((item) => item[0]),
      );
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
      this.scraperSettings[setting] = [...selected].filter((value) =>
        this.fieldList(kind).includes(value),
      );
    },
    selectAllScraperFields(kind) {
      this.scraperSettings[this.scraperFieldSetting(kind)] =
        this.fieldList(kind);
    },
    clearScraperFields(kind) {
      this.scraperSettings[this.scraperFieldSetting(kind)] = [];
    },
    applySettings(settings) {
      this.scraperSettings.tmdbConfigured = !!settings.tmdbConfigured;
      this.scraperSettings.tmdbKeySource = settings.tmdbKeySource || "none";
      this.scraperSettings.proxyEnabled = !!settings.proxyEnabled;
      this.scraperSettings.proxyHost = settings.proxyHost || "";
      this.scraperSettings.proxyPort = settings.proxyPort || 7890;
      this.scraperSettings.proxyUsername = settings.proxyUsername || "";
      this.scraperSettings.proxyPasswordConfigured = !!settings.proxyPassword;
      this.scraperSettings.movieScrapeMetadata =
        settings.movieScrapeMetadata !== false;
      this.scraperSettings.movieScrapeNfo = settings.movieScrapeNfo !== false;
      this.scraperSettings.movieScrapeImages =
        settings.movieScrapeImages !== false;
      this.scraperSettings.movieScrapeOverwrite =
        !!settings.movieScrapeOverwrite;
      this.scraperSettings.tvShowScrapeMetadata =
        settings.tvShowScrapeMetadata !== false;
      this.scraperSettings.tvShowEpisodeMetadata =
        settings.tvShowEpisodeMetadata !== false;
      this.scraperSettings.tvShowScrapeNfo = settings.tvShowScrapeNfo !== false;
      this.scraperSettings.tvShowScrapeImages =
        settings.tvShowScrapeImages !== false;
      this.scraperSettings.tvShowScrapeOverwrite =
        !!settings.tvShowScrapeOverwrite;
      this.scraperSettings.movieRenameAfterScrape =
        !!settings.movieRenameAfterScrape;
      this.scraperSettings.tvShowRenameAfterScrape =
        !!settings.tvShowRenameAfterScrape;
      this.scraperSettings.movieScraperFields = this.normalizeFieldList(
        settings.movieScraperFields,
        "movie",
      );
      this.scraperSettings.tvShowScraperFields = this.normalizeFieldList(
        settings.tvShowScraperFields,
        "tvshow",
      );
      this.scraperSettings.tvEpisodeScraperFields = this.normalizeFieldList(
        settings.tvEpisodeScraperFields,
        "episode",
      );
      this.scraperSettings.movieRenamerPathname =
        settings.movieRenamerPathname || DEFAULT_MOVIE_RENAMER_PATH;
      this.scraperSettings.movieRenamerFilename =
        settings.movieRenamerFilename || DEFAULT_MOVIE_RENAMER_FILE;
      this.scraperSettings.movieRenamerPathSpaceSubstitution =
        !!settings.movieRenamerPathSpaceSubstitution;
      this.scraperSettings.movieRenamerPathSpaceReplacement =
        settings.movieRenamerPathSpaceReplacement || "_";
      this.scraperSettings.movieRenamerFilenameSpaceSubstitution =
        !!settings.movieRenamerFilenameSpaceSubstitution;
      this.scraperSettings.movieRenamerFilenameSpaceReplacement =
        settings.movieRenamerFilenameSpaceReplacement || "_";
      this.scraperSettings.movieRenamerColonReplacement =
        settings.movieRenamerColonReplacement || "-";
      this.scraperSettings.movieRenamerAsciiReplacement =
        !!settings.movieRenamerAsciiReplacement;
      this.scraperSettings.movieRenamerFirstCharacterReplacement =
        settings.movieRenamerFirstCharacterReplacement || "#";
      this.scraperSettings.movieRenamerCreateSingleMovieSet =
        !!settings.movieRenamerCreateSingleMovieSet;
      this.scraperSettings.movieRenamerNfoCleanup =
        !!settings.movieRenamerNfoCleanup;
      this.scraperSettings.movieRenamerCleanupUnwanted =
        !!settings.movieRenamerCleanupUnwanted;
      this.scraperSettings.movieRenamerAllowMerge =
        !!settings.movieRenamerAllowMerge;
      this.scraperSettings.tvShowRenamerShowFolder =
        settings.tvShowRenamerShowFolder || DEFAULT_TVSHOW_RENAMER_PATH;
      this.scraperSettings.tvShowRenamerSeason =
        settings.tvShowRenamerSeason || DEFAULT_TVSHOW_RENAMER_SEASON;
      this.scraperSettings.tvShowRenamerFilename =
        settings.tvShowRenamerFilename || DEFAULT_TVSHOW_RENAMER_FILE;
      this.scraperSettings.tvShowRenamerShowFolderSpaceSubstitution =
        !!settings.tvShowRenamerShowFolderSpaceSubstitution;
      this.scraperSettings.tvShowRenamerShowFolderSpaceReplacement =
        settings.tvShowRenamerShowFolderSpaceReplacement || "_";
      this.scraperSettings.tvShowRenamerSeasonFolderSpaceSubstitution =
        !!settings.tvShowRenamerSeasonFolderSpaceSubstitution;
      this.scraperSettings.tvShowRenamerSeasonFolderSpaceReplacement =
        settings.tvShowRenamerSeasonFolderSpaceReplacement || "_";
      this.scraperSettings.tvShowRenamerFilenameSpaceSubstitution =
        !!settings.tvShowRenamerFilenameSpaceSubstitution;
      this.scraperSettings.tvShowRenamerFilenameSpaceReplacement =
        settings.tvShowRenamerFilenameSpaceReplacement || "_";
      this.scraperSettings.tvShowRenamerColonReplacement =
        settings.tvShowRenamerColonReplacement || " ";
      this.scraperSettings.tvShowRenamerAsciiReplacement =
        !!settings.tvShowRenamerAsciiReplacement;
      this.scraperSettings.tvShowRenamerFirstCharacterReplacement =
        settings.tvShowRenamerFirstCharacterReplacement || "#";
      this.scraperSettings.tvShowRenamerCleanupUnwanted =
        !!settings.tvShowRenamerCleanupUnwanted;
      this.scraperSettings.moviePosterName =
        settings.moviePosterName || "poster.jpg";
      this.scraperSettings.movieFanartName =
        settings.movieFanartName || "fanart.jpg";
      this.scraperSettings.moviePosterNames =
        settings.moviePosterNames ||
        "poster.jpg\nfolder.jpg\n{filename}-poster.jpg";
      this.scraperSettings.movieFanartNames =
        settings.movieFanartNames || "fanart.jpg\n{filename}-fanart.jpg";
      this.scraperSettings.tvShowPosterName =
        settings.tvShowPosterName || "poster.jpg";
      this.scraperSettings.tvShowFanartName =
        settings.tvShowFanartName || "fanart.jpg";
      this.scraperSettings.tvShowPosterNames =
        settings.tvShowPosterNames || "poster.jpg\nfolder.jpg";
      this.scraperSettings.tvShowFanartNames =
        settings.tvShowFanartNames || "fanart.jpg\nbackdrop.jpg";
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
        movieRenamerPathSpaceSubstitution:
          !!this.scraperSettings.movieRenamerPathSpaceSubstitution,
        movieRenamerPathSpaceReplacement:
          this.scraperSettings.movieRenamerPathSpaceReplacement,
        movieRenamerFilenameSpaceSubstitution:
          !!this.scraperSettings.movieRenamerFilenameSpaceSubstitution,
        movieRenamerFilenameSpaceReplacement:
          this.scraperSettings.movieRenamerFilenameSpaceReplacement,
        movieRenamerColonReplacement:
          this.scraperSettings.movieRenamerColonReplacement,
        movieRenamerAsciiReplacement:
          !!this.scraperSettings.movieRenamerAsciiReplacement,
        movieRenamerFirstCharacterReplacement:
          this.scraperSettings.movieRenamerFirstCharacterReplacement,
        movieRenamerCreateSingleMovieSet:
          !!this.scraperSettings.movieRenamerCreateSingleMovieSet,
        movieRenamerNfoCleanup: !!this.scraperSettings.movieRenamerNfoCleanup,
        movieRenamerCleanupUnwanted:
          !!this.scraperSettings.movieRenamerCleanupUnwanted,
        movieRenamerAllowMerge: !!this.scraperSettings.movieRenamerAllowMerge,
        tvShowRenamerShowFolder: this.scraperSettings.tvShowRenamerShowFolder,
        tvShowRenamerSeason: this.scraperSettings.tvShowRenamerSeason,
        tvShowRenamerFilename: this.scraperSettings.tvShowRenamerFilename,
        tvShowRenamerShowFolderSpaceSubstitution:
          !!this.scraperSettings.tvShowRenamerShowFolderSpaceSubstitution,
        tvShowRenamerShowFolderSpaceReplacement:
          this.scraperSettings.tvShowRenamerShowFolderSpaceReplacement,
        tvShowRenamerSeasonFolderSpaceSubstitution:
          !!this.scraperSettings.tvShowRenamerSeasonFolderSpaceSubstitution,
        tvShowRenamerSeasonFolderSpaceReplacement:
          this.scraperSettings.tvShowRenamerSeasonFolderSpaceReplacement,
        tvShowRenamerFilenameSpaceSubstitution:
          !!this.scraperSettings.tvShowRenamerFilenameSpaceSubstitution,
        tvShowRenamerFilenameSpaceReplacement:
          this.scraperSettings.tvShowRenamerFilenameSpaceReplacement,
        tvShowRenamerColonReplacement:
          this.scraperSettings.tvShowRenamerColonReplacement,
        tvShowRenamerAsciiReplacement:
          !!this.scraperSettings.tvShowRenamerAsciiReplacement,
        tvShowRenamerFirstCharacterReplacement:
          this.scraperSettings.tvShowRenamerFirstCharacterReplacement,
        tvShowRenamerCleanupUnwanted:
          !!this.scraperSettings.tvShowRenamerCleanupUnwanted,
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
          const nested = child.children
            ? this.filterSettingsChildren(
                child.children,
                `${prefix}${child.title} `,
                query,
              )
            : [];
          const text = `${prefix}${child.title}`.toLowerCase();
          if (text.includes(query) || nested.length)
            return { ...child, children: nested };
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
        if (this.scraperSettings.tmdbApiKey)
          body.tmdbApiKey = this.scraperSettings.tmdbApiKey;
        if (this.scraperSettings.proxyPassword)
          body.proxyPassword = this.scraperSettings.proxyPassword;
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
          body: JSON.stringify(
            this.settingsPayload({
              clearTmdbKey: true,
              clearProxyPassword: false,
            }),
          ),
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
  },
};
