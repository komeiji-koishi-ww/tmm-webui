export const mobileMixin = {
  computed: {
    mobileMovieRows() {
      return this.sortedMovieRows;
    },
  },
  methods: {
    updateViewportMode() {
      const next = window.matchMedia("(max-width: 760px)").matches;
      this.isMobile = next;
      if (!next) this.mobileDetailOpen = false;
    },
    openMobileItem(item, event = null) {
      if (!item) return;
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
