export const tvMixin = {
  computed: {
    tvTree() {
      const shows = new Map();
      for (const item of this.visibleItems) {
        const showName = item.showGuess || item.titleGuess || "未知剧集";
        if (!shows.has(showName)) {
          shows.set(showName, {
            key: showName,
            title: showName,
            episodes: 0,
            seasons: new Map(),
            items: [],
          });
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
            title: seasonNumber
              ? `Season ${String(seasonNumber).padStart(2, "0")}`
              : "未识别季",
            items: [],
          });
        }
        show.seasons.get(seasonKey).items.push(item);
        show.episodes += 1;
      }
      return this.sortShows(Array.from(shows.values())).map((show) => ({
        ...show,
        seasons: Array.from(show.seasons.values())
          .map((season) => ({
            ...season,
            items: this.sortEpisodes(season.items.slice()),
          }))
          .sort((a, b) => a.season - b.season),
      }));
    },
    tvTreeRows() {
      return this.tvTree.flatMap((show) => {
        const firstShowItem = this.firstTVItem(show) || {};
        const rows = [
          {
            key: `show:${show.key}`,
            level: "show",
            title: `${this.isShowExpanded(show.key) ? "▾" : "▸"} ${show.title}`,
            year: firstShowItem.yearGuess || "-",
            rating: this.tvShowRating(show),
            dateAdded: firstShowItem.dateAdded || "",
            status: `${show.seasons.length} 季 / ${show.episodes} 集 · ${firstShowItem.hasNfo ? "show.nfo" : "未匹配"}`,
            payload: show,
          },
        ];
        if (!this.isShowExpanded(show.key)) return rows;
        for (const season of show.seasons) {
          const firstSeasonItem = season.items[0] || {};
          rows.push({
            key: `season:${season.key}`,
            level: "season",
            title: `${this.isSeasonExpanded(season.key) ? "▾" : "▸"} ${season.title}`,
            year: firstSeasonItem.yearGuess || "-",
            rating: 0,
            dateAdded: firstSeasonItem.dateAdded || "",
            status: `${season.items.length} 集 · ${firstSeasonItem.hasNfo ? "season.nfo" : "待写入"}`,
            payload: { show, season },
          });
          if (!this.isSeasonExpanded(season.key)) continue;
          for (const item of season.items) {
            rows.push({
              key: `episode:${item.id}`,
              level: "episode",
              title: this.itemSeasonText(item),
              year: item.yearGuess || "-",
              rating: this.tvEpisodeRating(item),
              dateAdded: item.dateAdded || "",
              status: this.itemStatusText(item),
              payload: item,
            });
          }
        }
        return rows;
      });
    },
  },
  methods: {
    selectItem(item, event = null) {
      const isTV = item.kind === "tvshow";
      const rangeSelect = isTV && this.isShiftSelection(event);
      const multiToggle = isTV && this.isToggleSelection(event);
      if (isTV && event) this.suppressBrowserTextSelection(event);
      if (rangeSelect) {
        const range = this.tvSelectionRange(item.id);
        this.selectedItemIds = range.length ? range : [item.id];
      } else if (multiToggle) {
        const selected = new Set(this.selectedItemIds);
        if (selected.has(item.id)) selected.delete(item.id);
        else selected.add(item.id);
        this.selectedItemIds = [...selected];
        this.lastSelectedTVItemId = item.id;
        this.tvRangeAnchorItemId = item.id;
      } else {
        this.selectedItemIds = [item.id];
        this.lastSelectedTVItemId = isTV ? item.id : "";
        this.tvRangeAnchorItemId = isTV ? item.id : "";
      }
      this.selectedItem = item;
      this.selectedEntity = {
        kind: item.kind === "tvshow" ? "episode" : "movie",
        payload: item,
      };
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
    suppressBrowserTextSelection(event = null) {
      if (event && event.cancelable) event.preventDefault();
      const selection = window.getSelection ? window.getSelection() : null;
      if (selection && selection.removeAllRanges) selection.removeAllRanges();
    },
    isShiftSelection(event = null) {
      return !!((event && event.shiftKey) || this.keyboardModifiers.shift);
    },
    isToggleSelection(event = null) {
      return !!(
        (event && (event.metaKey || event.ctrlKey)) ||
        this.keyboardModifiers.meta ||
        this.keyboardModifiers.ctrl
      );
    },
    updateKeyboardModifiers(event = null) {
      if (!event) return;
      this.keyboardModifiers.shift = !!event.shiftKey;
      this.keyboardModifiers.ctrl = !!event.ctrlKey;
      this.keyboardModifiers.meta = !!event.metaKey;
    },
    clearKeyboardModifiers() {
      this.keyboardModifiers.shift = false;
      this.keyboardModifiers.ctrl = false;
      this.keyboardModifiers.meta = false;
    },
    rememberTVPointerEvent(event) {
      this.lastTVPointerEvent = event || null;
      this.updateKeyboardModifiers(event);
      if (event && event.shiftKey) this.suppressBrowserTextSelection(event);
    },
    prepareTVCellPointer(event) {
      this.rememberTVPointerEvent(event);
      if (event) this.suppressBrowserTextSelection(event);
    },
    tvClickEvent(event = null) {
      return event || this.lastTVPointerEvent || window.event || null;
    },
    shouldSkipTVClick(row) {
      if (!row) return true;
      const now = Date.now();
      if (this.lastTVClickKey === row.key && now - this.lastTVClickAt < 40)
        return true;
      this.lastTVClickKey = row.key;
      this.lastTVClickAt = now;
      return false;
    },
    handleTVTableRowClick(row, _column, event) {
      this.activateTVRow(row, event);
    },
    handleTVTitleClick(row, event) {
      this.activateTVRow(row, event);
    },
    activateTVRow(row, event = null) {
      if (!row) return;
      if (event) this.suppressBrowserTextSelection(event);
      if (this.shouldSkipTVClick(row)) return;
      const clickEvent = this.tvClickEvent(event);
      this.updateKeyboardModifiers(clickEvent);
      if (row.level === "show") {
        this.toggleShow(row.payload.key);
        this.selectTvGroup("show", row.payload);
        return;
      }
      if (row.level === "season") {
        this.toggleSeason(row.payload.season.key);
        this.selectTvGroup("season", row.payload.season);
        return;
      }
      this.selectItem(row.payload, clickEvent);
    },
    handleTVTableContextMenu(row, _column, event) {
      if (!row || !event) return;
      this.updateKeyboardModifiers(event);
      event.preventDefault();
      if (row.level === "show") {
        this.openContextMenu(event, "show", row.payload);
        return;
      }
      if (row.level === "season") {
        this.openContextMenu(event, "season", row.payload);
        return;
      }
      this.openContextMenu(event, "episode", row.payload);
    },
    visibleTVEpisodeItems() {
      return this.tvTreeRows
        .filter((row) => row.level === "episode" && row.payload)
        .map((row) => row.payload);
    },
    tvSelectionRange(targetId) {
      const items = this.visibleTVEpisodeItems();
      if (!items.length) return [];
      const anchorId =
        this.tvRangeAnchorItemId ||
        this.lastSelectedTVItemId ||
        (this.selectedItem && this.selectedItem.kind === "tvshow"
          ? this.selectedItem.id
          : "");
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
      if (this.activeModule === "tvshow")
        return this.selectedItemIds.includes(item.id);
      return !!this.selectedItem && this.selectedItem.id === item.id;
    },
    tvRowClassName({ row }) {
      if (!row) return "";
      const classes = [`tv-row-${row.level}`];
      if (
        row.level === "episode" &&
        row.payload &&
        this.selectedItemIds.includes(row.payload.id)
      ) {
        classes.push("is-tv-selected");
      }
      if (
        row.level === "show" &&
        this.selectedEntity &&
        this.selectedEntity.kind === "show" &&
        this.selectedEntity.payload &&
        this.selectedEntity.payload.key === row.payload.key
      ) {
        classes.push("is-tv-selected");
      }
      if (
        row.level === "season" &&
        this.selectedEntity &&
        this.selectedEntity.kind === "season" &&
        this.selectedEntity.payload &&
        this.selectedEntity.payload.key === row.payload.season.key
      ) {
        classes.push("is-tv-selected");
      }
      return classes.join(" ");
    },
    selectedTVRenameItems(payload = null) {
      if (
        payload &&
        payload.kind === "tvshow" &&
        this.selectedItemIds.includes(payload.id)
      ) {
        const selected = new Set(this.selectedItemIds);
        return this.visibleTVEpisodeItems().filter((item) =>
          selected.has(item.id),
        );
      }
      if (payload && payload.kind === "tvshow") return [payload];
      return this.selectedItem && this.selectedItem.kind === "tvshow"
        ? [this.selectedItem]
        : [];
    },
    handleKeydown(event) {
      this.updateKeyboardModifiers(event);
      if (event.key === "Escape") {
        this.closeContextMenu();
        if (
          this.chooser.open &&
          !this.chooser.loading &&
          !this.chooser.scraping
        )
          this.closeChooser();
        if (this.localRename.open && !this.localRename.saving)
          this.closeLocalRename();
      }
    },
    handleKeyup(event) {
      this.updateKeyboardModifiers(event);
    },
    openContextMenu(event, scope, payload) {
      const keepTVSelection =
        scope === "episode" &&
        payload &&
        this.selectedItemIds.includes(payload.id) &&
        this.selectedItemIds.length > 1;
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
      if (scope === "season")
        this.selectTvGroup("season", payload.season || payload);
    },
    handleContextCommand(command) {
      if (command === "scrape") this.openChooserFromContext();
      if (command === "rename") this.openLocalRenameFromContext();
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
      return (
        this.contextMenu.scope === "movie" ||
        this.contextMenu.scope === "episode"
      );
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
    openRenameFromToolbar() {
      if (!this.selectedItem) return;
      if (this.activeModule === "tvshow") {
        this.openLocalRename(
          this.selectedTVRenameItems(this.selectedItem),
          "tvshow",
        );
        return;
      }
      this.openLocalRename([this.selectedItem], "movie");
    },
    firstTVItem(show) {
      const firstSeason = show && show.seasons ? show.seasons[0] : null;
      return firstSeason && firstSeason.items.length
        ? firstSeason.items[0]
        : null;
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
        this.tvRangeAnchorItemId = "";
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
        this.tvRangeAnchorItemId = "";
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
      if (item.season && item.episode)
        return `S${String(item.season).padStart(2, "0")}E${String(item.episode).padStart(2, "0")}`;
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
  },
};
