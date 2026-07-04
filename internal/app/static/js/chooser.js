export const chooserMixin = {
  methods: {
    openChooserFromSelected() {
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
      if (!this.selectedItem) return;
      this.openChooser({
        scope: this.selectedItem.kind === "tvshow" ? "episode" : "movie",
        mediaType: this.selectedItem.kind === "tvshow" ? "tvshow" : "movie",
        targetItem: this.selectedItem,
        query:
          this.selectedItem.kind === "tvshow"
            ? this.selectedItem.showGuess || this.selectedItem.titleGuess
            : this.selectedItem.titleGuess,
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
          metadata: tv
            ? this.scraperSettings.tvShowScrapeMetadata
            : this.scraperSettings.movieScrapeMetadata,
          nfo: tv
            ? this.scraperSettings.tvShowScrapeNfo
            : this.scraperSettings.movieScrapeNfo,
          artwork: tv
            ? this.scraperSettings.tvShowScrapeImages
            : this.scraperSettings.movieScrapeImages,
          overwrite: true,
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
      if (
        this.chooser.scope === "season" &&
        this.chooser.targetShow &&
        this.chooser.targetSeason
      ) {
        return `${this.chooser.targetShow.title} · ${this.chooser.targetSeason.title} · ${this.chooser.targetSeason.items.length} 集`;
      }
      if (this.chooser.targetItem)
        return (
          this.chooser.targetItem.fileName || this.chooser.targetItem.titleGuess
        );
      return "";
    },
    imageURL(path, size = "w342") {
      if (!path) return "";
      const params = new URLSearchParams();
      params.set("path", path);
      params.set("size", size);
      return `/api/tmdb-image?${params.toString()}`;
    },
    chooserFanartURL() {
      const path =
        (this.chooser.detail && this.chooser.detail.backdropPath) ||
        (this.chooser.selected && this.chooser.selected.backdropPath) ||
        "";
      return this.imageURL(path, "w780");
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
        this.chooser.candidates = await this.api(
          `/api/search?${params.toString()}`,
        );
        this.status = `找到 ${this.chooser.candidates.length} 个候选`;
        if (this.chooser.candidates.length)
          await this.selectCandidate(this.chooser.candidates[0]);
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
        this.chooser.detail = await this.api(
          `/api/metadata?${params.toString()}`,
        );
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
        const writeShowMeta = tv
          ? !!this.chooser.options.showMetadata
          : !!this.chooser.options.metadata;
        const writeEpisodeMeta = tv && !!this.chooser.options.episodeMetadata;
        const body = {
          itemId: this.chooser.targetItem.id,
          scope: this.chooser.scope,
          libraryId: this.selectedLibrary
            ? this.selectedLibrary.id
            : this.chooser.targetItem.libraryId,
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
        if (this.chooser.targetShow)
          body.showName =
            this.chooser.targetShow.key || this.chooser.targetShow.title;
        if (this.chooser.targetSeason)
          body.season = this.chooser.targetSeason.season;
        const result = await this.api("/api/scrape", {
          method: "POST",
          body: JSON.stringify(body),
        });
        if (result.items && result.items.length) {
          const updatedItems = this.normalizeItems(result.items);
          const byID = new Map(updatedItems.map((item) => [item.id, item]));
          const byOldPath = new Map(
            (result.renamePreviews || []).map((preview, index) => [
              preview.sourceFile,
              updatedItems[index],
            ]),
          );
          this.items = this.items.map(
            (item) => byID.get(item.id) || byOldPath.get(item.path) || item,
          );
          if (this.selectedItem && byID.has(this.selectedItem.id))
            this.selectedItem = byID.get(this.selectedItem.id);
          if (this.selectedItem && byOldPath.has(this.selectedItem.path))
            this.selectedItem = byOldPath.get(this.selectedItem.path);
        } else if (result.item) {
          const updatedItem = this.normalizeItem(result.item);
          const oldPath = result.renamePreview
            ? result.renamePreview.sourceFile
            : "";
          this.items = this.items.map((item) =>
            item.path === oldPath ||
            item.path === updatedItem.path ||
            item.id === updatedItem.id
              ? updatedItem
              : item,
          );
          this.selectedItem = updatedItem;
        }
        const scraped = result.movie || result.show || this.chooser.detail;
        if (scraped && this.selectedItem) {
          this.rename.title = scraped.title;
          this.rename.year = (
            scraped.releaseDate ||
            scraped.firstAirDate ||
            ""
          ).slice(0, 4);
          this.rename.tmdbId = scraped.id;
        }
        this.status = result.renamed
          ? "刮削写入完成，已自动重命名"
          : result.renameWarnings
            ? `刮削完成，重命名未完成：${result.renameWarnings[0]}`
            : "刮削写入完成";
        this.closeChooser();
      } catch (error) {
        this.chooser.error = error.message;
        this.status = error.message;
      } finally {
        this.chooser.scraping = false;
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
      const tv =
        (candidate.mediaType || this.selectedItem.kind || this.activeModule) ===
        "tvshow";
      try {
        const result = await this.api("/api/scrape", {
          method: "POST",
          body: JSON.stringify({
            itemId: this.selectedItem.id,
            tmdbId: candidate.id,
            mediaType:
              candidate.mediaType ||
              this.selectedItem.kind ||
              this.activeModule,
            writeNfo: tv
              ? this.scraperSettings.tvShowScrapeNfo
              : this.scraperSettings.movieScrapeNfo,
            writeImages: tv
              ? this.scraperSettings.tvShowScrapeImages
              : this.scraperSettings.movieScrapeImages,
            writeMeta: tv
              ? this.scraperSettings.tvShowScrapeMetadata ||
                this.scraperSettings.tvShowEpisodeMetadata
              : this.scraperSettings.movieScrapeMetadata,
            writeShowMeta: tv
              ? this.scraperSettings.tvShowScrapeMetadata
              : this.scraperSettings.movieScrapeMetadata,
            writeEpisodeMeta: tv
              ? this.scraperSettings.tvShowEpisodeMetadata
              : false,
            overwrite: tv
              ? this.scraperSettings.tvShowScrapeOverwrite
              : this.scraperSettings.movieScrapeOverwrite,
            metadataFields: this.scraperFields(tv ? "tvshow" : "movie"),
            episodeFields: tv ? this.scraperFields("episode") : [],
          }),
        });
        if (result.item) {
          const updatedItem = this.normalizeItem(result.item);
          this.selectedItem = updatedItem;
          const oldPath = result.renamePreview
            ? result.renamePreview.sourceFile
            : "";
          this.items = this.items.map((item) =>
            item.id === updatedItem.id || item.path === oldPath
              ? updatedItem
              : item,
          );
        }
        const scraped = result.movie || result.show;
        this.rename.title = scraped.title;
        this.rename.year = (
          scraped.releaseDate ||
          scraped.firstAirDate ||
          ""
        ).slice(0, 4);
        this.rename.tmdbId = scraped.id;
        this.status = result.renamed
          ? "刮削写入完成，已自动重命名"
          : result.renameWarnings
            ? `刮削完成，重命名未完成：${result.renameWarnings[0]}`
            : "刮削写入完成";
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
  },
};
