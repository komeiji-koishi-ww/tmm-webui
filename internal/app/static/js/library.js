const { markRaw } = Vue;

export const libraryMixin = {
  computed: {
    filteredLibraries() {
      return this.libraries.filter(
        (library) => library.type === this.activeModule,
      );
    },
    moduleTitle() {
      return this.activeModule === "tvshow" ? "电视剧" : "电影";
    },
    selectedTypeText() {
      if (!this.selectedLibrary) return "";
      return this.selectedLibrary.type === "tvshow" ? "电视剧" : "电影";
    },
    selectedTask() {
      if (!this.selectedLibrary) return null;
      return this.tasks[this.selectedLibrary.id] || null;
    },
    selectedScanning() {
      return this.selectedTask && this.selectedTask.state === "running";
    },
    selectedTaskActive() {
      return this.scanFlowVisible(this.selectedTask);
    },
    allTasks() {
      return Object.values(this.tasks)
        .filter(Boolean)
        .sort((a, b) => (a.startedAt < b.startedAt ? 1 : -1));
    },
    scanProgressText() {
      const task = this.selectedTask;
      if (!task) return "";
      if (task.state === "running")
        return `已检查 ${task.visitedFiles || 0} 个文件，发现 ${task.foundItems || 0} 个视频`;
      if (task.state === "completed")
        return `扫描完成，共 ${task.resultCount || task.foundItems || 0} 个视频`;
      if (task.state === "failed")
        return `扫描失败：${task.error || "未知错误"}`;
      return "";
    },
  },
  methods: {
    scanFlowVisible(task) {
      return !!(
        task &&
        (task.state === "running" || task.state === "canceling")
      );
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
    async switchModule(module) {
      this.activeModule = module;
      this.mobileDetailOpen = false;
      this.newLibrary.type = module;
      this.newLibrary.name = module === "tvshow" ? "电视剧" : "电影";
      this.query = "";
      this.filters = [];
      this.sortKey = "dateAdded";
      this.sortDirection = "desc";
      this.selectedItemIds = [];
      this.lastSelectedTVItemId = "";
      this.tvRangeAnchorItemId = "";
      const first = this.filteredLibraries[0];
      if (first) {
        await this.selectLibrary(first);
      } else {
        this.selectedLibrary = null;
        this.items = [];
        this.selectedItem = null;
        this.selectedItemIds = [];
        this.lastSelectedTVItemId = "";
        this.tvRangeAnchorItemId = "";
        this.selectedEntity = null;
        this.status = `未配置${this.moduleTitle}数据源`;
      }
    },
    addPendingPath() {
      const path = this.pendingPath.trim();
      if (!path || this.newLibrary.paths.includes(path)) return;
      this.newLibrary.paths.push(path);
      this.pendingPath = "";
      if (
        !this.newLibrary.name ||
        this.newLibrary.name === "电影" ||
        this.newLibrary.name === "电视剧"
      ) {
        this.newLibrary.name =
          this.newLibrary.type === "tvshow" ? "电视剧" : "电影";
      }
    },
    removePath(path) {
      this.newLibrary.paths = this.newLibrary.paths.filter(
        (item) => item !== path,
      );
    },
    prepareDatasource(type) {
      const previousType = this.newLibrary.type;
      const defaultNames = new Set(["电影", "电视剧"]);
      this.newLibrary.type = type;
      if (
        !this.newLibrary.name ||
        (previousType !== type && defaultNames.has(this.newLibrary.name))
      ) {
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
        const result = await this.api(
          `/api/browse?path=${encodeURIComponent(path)}`,
        );
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
        await this.api(`/api/libraries?id=${encodeURIComponent(library.id)}`, {
          method: "DELETE",
        });
        this.libraries = this.libraries.filter(
          (item) => item.id !== library.id,
        );
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
            this.tvRangeAnchorItemId = "";
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
      this.mobileDetailOpen = false;
      this.resetMoviePage();
      this.items = [];
      this.selectedItem = null;
      this.selectedItemIds = [];
      this.lastSelectedTVItemId = "";
      this.tvRangeAnchorItemId = "";
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
        const task = this.selectedTask;
        if (!task) return;
        if (task.state === "running") {
          await this.loadItems(this.selectedLibrary, true);
          return;
        }
        if (task.state === "completed" && !this.completedTaskReloads[task.id]) {
          this.completedTaskReloads[task.id] = true;
          await this.loadItems(this.selectedLibrary, true);
        }
      }, 1500);
    },
    resetMoviePage() {
      this.movieVirtual.scrollTop = 0;
      this.$nextTick(() => {
        const scroller = this.movieScrollerElement();
        if (scroller) {
          scroller.scrollTop = 0;
          this.movieVirtual.viewportHeight =
            scroller.clientHeight || this.movieVirtual.viewportHeight;
        }
      });
    },
    ensureMoviePageInRange() {
      this.handleMovieVirtualScroll();
    },
    movieScrollerElement() {
      return (
        this.$refs.movieScroller &&
        (this.$refs.movieScroller.$el || this.$refs.movieScroller)
      );
    },
    handleMovieVirtualScroll() {
      const scroller = this.movieScrollerElement();
      if (!scroller) return;
      this.movieVirtual.scrollTop = scroller.scrollTop || 0;
      this.movieVirtual.viewportHeight =
        scroller.clientHeight || this.movieVirtual.viewportHeight;
    },
    async loadTasks(library, quiet = false) {
      try {
        const result = await this.api(
          `/api/tasks?libraryId=${encodeURIComponent(library.id)}`,
        );
        const tasks = result.tasks || [];
        if (!tasks.length) return;
        tasks.sort((a, b) => (a.startedAt < b.startedAt ? 1 : -1));
        this.tasks[library.id] = tasks[0];
        if (!quiet && this.tasks[library.id].state === "running")
          this.status = `${library.name} 正在扫描`;
      } catch (error) {
        if (!quiet) this.status = error.message;
      }
    },
    normalizeItems(items) {
      return (items || []).map((item) => this.normalizeItem(item));
    },
    normalizeItem(item) {
      if (!item) return item;
      const normalized = {
        id: item.id,
        libraryId: item.libraryId,
        sourcePath: item.sourcePath,
        kind: item.kind,
        mediaType: item.mediaType,
        path: item.path,
        dir: item.dir,
        fileName: item.fileName,
        titleGuess: item.titleGuess,
        yearGuess: item.yearGuess,
        originalTitle: item.originalTitle,
        original: item.original,
        overview: item.overview,
        runtime: item.runtime,
        rating: item.rating,
        showRating: item.showRating,
        genres: Array.isArray(item.genres) ? item.genres.slice(0, 24) : [],
        actors: Array.isArray(item.actors) ? item.actors.slice(0, 40) : [],
        premiered: item.premiered,
        dateAdded: item.dateAdded,
        modTimeUnix: item.modTimeUnix,
        nfoModTimeUnix: item.nfoModTimeUnix,
        fileSize: item.fileSize,
        fileSizeBytes: item.fileSizeBytes,
        videoFormat: item.videoFormat,
        audioCodec: item.audioCodec,
        mediaDurationSeconds: item.mediaDurationSeconds,
        imdbId: item.imdbId,
        showGuess: item.showGuess,
        season: item.season,
        episode: item.episode,
        episodes: Array.isArray(item.episodes)
          ? item.episodes.slice(0, 12)
          : [],
        hasNfo: item.hasNfo,
        hasPoster: item.hasPoster,
        hasFanart: item.hasFanart,
        hasSubtitle: item.hasSubtitle,
        matchedId: item.matchedId,
        matchedName: item.matchedName,
      };
      normalized.searchText = [
        normalized.titleGuess,
        normalized.showGuess,
        normalized.fileName,
        normalized.path,
        normalized.yearGuess,
        normalized.matchedName,
        normalized.imdbId,
        normalized.matchedId,
        ...(normalized.genres || []),
        ...this.localizedGenres(normalized.genres || []),
      ]
        .filter(Boolean)
        .join(" ")
        .toLowerCase();
      return markRaw(normalized);
    },
    async loadItems(library, quiet = false) {
      if (!quiet) {
        this.busy = true;
        this.status = `正在加载 ${library.name}`;
      }
      const scrollAnchor = quiet ? this.captureMediaScrollAnchor() : null;
      try {
        const result = await this.api(
          `/api/items?libraryId=${encodeURIComponent(library.id)}`,
        );
        const loadedItems = this.normalizeItems(result.items || []);
        this.items =
          quiet && this.selectedTask && this.selectedTask.state === "running"
            ? this.mergeLoadedItems(this.items, loadedItems)
            : loadedItems;
        this.ensureMoviePageInRange();
        if (!quiet) {
          await this.$nextTick();
          this.selectInitialDesktopItem();
        }
        if (scrollAnchor) {
          await this.$nextTick();
          this.restoreMediaScrollAnchor(scrollAnchor);
        }
        if (!quiet) {
          this.status = this.items.length
            ? `已加载 ${this.items.length} 个缓存条目`
            : `已选择 ${library.name}，需要扫描`;
        }
      } catch (error) {
        if (!quiet) this.status = error.message;
      } finally {
        if (!quiet) this.busy = false;
      }
    },
    selectInitialDesktopItem() {
      if (this.isMobile || this.selectedItem || this.selectedEntity) return;
      if (this.activeModule === "movie") {
        const firstMovie = this.sortedMovieRows[0];
        if (firstMovie) this.selectItem(firstMovie);
        return;
      }
      const firstTVRow = this.tvTreeRows[0];
      if (!firstTVRow) return;
      if (firstTVRow.level === "show") {
        this.selectTvGroup("show", firstTVRow.payload);
        return;
      }
      if (firstTVRow.level === "season") {
        this.selectTvGroup("season", firstTVRow.payload.season);
        return;
      }
      if (firstTVRow.level === "episode") this.selectItem(firstTVRow.payload);
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
      if (this.isMobile) return this.mobileLibraryScreenElement();
      if (this.activeModule === "movie") return this.movieScrollerElement();
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
        const escapedKey =
          window.CSS && CSS.escape
            ? CSS.escape(anchor.key)
            : anchor.key.replace(/["\\]/g, "\\$&");
        const row = scroller.querySelector(`[data-row-key="${escapedKey}"]`);
        if (row) {
          scroller.scrollTop = Math.max(0, row.offsetTop - anchor.offset);
          return;
        }
      }
      scroller.scrollTop = anchor.scrollTop || 0;
    },
    taskDescription(task) {
      if (task.state === "running")
        return `已检查 ${task.visitedFiles || 0} 个文件，发现 ${task.foundItems || 0} 个视频`;
      if (task.state === "canceling")
        return `正在停止，已发现 ${task.foundItems || 0} 个视频`;
      if (task.state === "canceled")
        return `已停止，保留已发现的 ${task.resultCount || task.foundItems || 0} 个视频`;
      if (task.state === "completed")
        return `完成，导入 ${task.resultCount || task.foundItems || 0} 个视频`;
      if (task.state === "failed") return task.error || "任务失败";
      return task.state;
    },
    async scan() {
      if (!this.selectedLibrary) return;
      this.busy = true;
      this.status = this.selectedScanning
        ? "正在停止扫描任务"
        : "正在启动扫描任务";
      try {
        const result = this.selectedScanning
          ? await this.api("/api/scan/cancel", {
              method: "POST",
              body: JSON.stringify({
                libraryId: this.selectedLibrary.id,
                taskId: this.selectedTask ? this.selectedTask.id : "",
              }),
            })
          : await this.api("/api/scan", {
              method: "POST",
              body: JSON.stringify({ libraryId: this.selectedLibrary.id }),
            });
        this.tasks[this.selectedLibrary.id] = result.task;
        if (this.selectedScanning) {
          this.status = "已请求停止扫描";
        } else {
          this.status = result.started
            ? "扫描任务已启动"
            : "该媒体库已有扫描任务在运行";
        }
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
  },
};
