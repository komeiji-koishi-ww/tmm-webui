export const mobileMixin = {
  computed: {
    mobileMovieRows() {
      return this.sortedMovieRows;
    },
  },
  methods: {
    initMobileHistory() {
      if (typeof window === "undefined" || !window.history) return;
      if (this.mobileHistory.initialized) return;
      const state = {
        ...(window.history.state || {}),
        tmmweb: true,
        tmmwebBase: true,
      };
      window.history.replaceState(state, "", window.location.href);
      this.mobileHistory.initialized = true;
    },
    updateViewportMode() {
      const next = window.matchMedia("(max-width: 760px)").matches;
      this.isMobile = next;
      if (!next) {
        this.mobileHistory.stack = [];
        this.mobileHistory.closingFromPop = false;
        this.mobileHistory.ignoreNextPop = false;
        this.mobileDetailOpen = false;
      } else {
        this.initMobileHistory();
      }
    },
    handleMobileSurfaceChange(kind, open, wasOpen) {
      if (!this.isMobile || open === wasOpen) return;
      if (kind === "detail" && !open && wasOpen) {
        this.restoreMobileLibraryScroll();
      }
      if (open) {
        this.pushMobileSurface(kind);
        return;
      }
      this.consumeMobileSurface(kind);
    },
    pushMobileSurface(kind) {
      if (typeof window === "undefined" || !window.history) return;
      this.initMobileHistory();
      const stack = this.mobileHistory.stack;
      if (stack[stack.length - 1] === kind) return;
      stack.push(kind);
      window.history.pushState(
        {
          tmmweb: true,
          tmmwebSurface: kind,
          tmmwebDepth: stack.length,
        },
        "",
        window.location.href,
      );
    },
    consumeMobileSurface(kind) {
      const historyState = this.mobileHistory;
      if (historyState.closingFromPop) return;
      const stack = historyState.stack;
      const index = stack.lastIndexOf(kind);
      if (index === -1) return;
      stack.splice(index, 1);
      if (typeof window === "undefined" || !window.history) return;
      historyState.ignoreNextPop = true;
      window.history.back();
    },
    handleMobilePopState() {
      if (!this.isMobile) return;
      const historyState = this.mobileHistory;
      if (historyState.ignoreNextPop) {
        historyState.ignoreNextPop = false;
        return;
      }
      const kind = historyState.stack.pop();
      if (!kind) return;
      historyState.closingFromPop = true;
      this.closeMobileSurfaceKind(kind);
      this.$nextTick(() => {
        historyState.closingFromPop = false;
      });
    },
    closeMobileSurfaceKind(kind) {
      if (kind === "rename") {
        this.localRename.open = false;
        this.localRename.error = "";
        return;
      }
      if (kind === "chooser") {
        this.chooser.open = false;
        this.chooser.error = "";
        return;
      }
      if (kind === "filters") {
        this.filterEditor.open = false;
        return;
      }
      if (kind === "browser") {
        this.browser.open = false;
        return;
      }
      if (kind === "settings") {
        this.settingsOpen = false;
        return;
      }
      if (kind === "detail") {
        this.mobileDetailOpen = false;
      }
    },
    mobileLibraryScreenElement() {
      if (this.isMobile && typeof document !== "undefined") {
        return document.scrollingElement || document.documentElement;
      }
      const ref = this.$refs.mobileLibraryScreen;
      return ref && (ref.$el || ref);
    },
    captureMobileLibraryScroll() {
      this.mobileLibraryScrollTop =
        this.captureMediaScrollAnchor() || this.mobileLibraryScrollTop || 0;
    },
    restoreMobileLibraryScroll() {
      const anchor = this.mobileLibraryScrollTop || 0;
      this.$nextTick(() => {
        window.requestAnimationFrame(() => {
          if (typeof anchor === "object") {
            this.restoreMediaScrollAnchor(anchor);
            return;
          }
          const screen = this.mobileLibraryScreenElement();
          if (screen) {
            screen.scrollTop = anchor;
            if (typeof window !== "undefined") window.scrollTo(0, anchor);
          }
        });
      });
    },
    openMobileItem(item, event = null) {
      if (!item) return;
      this.captureMobileLibraryScroll();
      this.selectItem(item, event);
      this.mobileDetailOpen = true;
    },
    closeMobileDetail() {
      this.mobileDetailOpen = false;
    },
    mobileTVRowTitle(row) {
      if (!row) return "";
      return String(row.title || "").replace(/^[▾▸]\s*/, "");
    },
    mobileTVSeasonKey(row) {
      if (!row || !row.payload) return "";
      if (row.level === "season") return row.payload.season.key;
      if (row.level !== "episode") return "";
      const item = row.payload;
      const showName = item.showGuess || item.titleGuess || "未知剧集";
      return `${showName}::${item.season || 0}`;
    },
    isMobileSeasonRenameActive(row) {
      return (
        !!this.mobileTVRenameSeasonKey &&
        this.mobileTVRenameSeasonKey === this.mobileTVSeasonKey(row)
      );
    },
    mobileEpisodeRenameSelected(item) {
      return !!item && this.mobileTVRenameSelectedIds.includes(item.id);
    },
    handleMobileTVRow(row, event = null) {
      if (!row) return;
      if (row.level === "episode") {
        if (this.isMobileSeasonRenameActive(row)) {
          this.toggleMobileEpisodeRenameSelection(row, event);
          return;
        }
        this.openMobileItem(row.payload, event);
        return;
      }
      this.activateTVRow(row, event);
    },
    openMobileTVChooser(row, event = null) {
      if (event) event.stopPropagation();
      if (!row || !row.payload) return;
      if (row.level === "show") this.selectTvGroup("show", row.payload);
      if (row.level === "season") this.selectTvGroup("season", row.payload.season);
      this.openChooserFromSelected();
    },
    startMobileSeasonRename(row, event = null) {
      if (event) event.stopPropagation();
      if (!row || row.level !== "season") return;
      const key = this.mobileTVSeasonKey(row);
      if (!this.isSeasonExpanded(key)) this.toggleSeason(key);
      this.mobileTVRenameSeasonKey = key;
      this.mobileTVRenameSelectedIds = [];
    },
    confirmMobileSeasonRename(row, event = null) {
      if (event) event.stopPropagation();
      if (!row || row.level !== "season") return;
      const selected = new Set(this.mobileTVRenameSelectedIds);
      const items = row.payload.season.items.filter((item) =>
        selected.has(item.id),
      );
      if (!items.length) {
        this.status = "请选择要重命名的集数";
        return;
      }
      this.openLocalRename(items, "tvshow", { singleManual: items.length === 1 });
      this.mobileTVRenameSeasonKey = "";
      this.mobileTVRenameSelectedIds = [];
    },
    cancelMobileSeasonRename(event = null) {
      if (event) event.stopPropagation();
      this.mobileTVRenameSeasonKey = "";
      this.mobileTVRenameSelectedIds = [];
    },
    toggleMobileEpisodeRenameSelection(row, event = null) {
      if (event) event.stopPropagation();
      if (!row || row.level !== "episode" || !row.payload) return;
      const id = row.payload.id;
      const selected = new Set(this.mobileTVRenameSelectedIds);
      if (selected.has(id)) selected.delete(id);
      else selected.add(id);
      this.mobileTVRenameSelectedIds = [...selected];
    },
  },
};
