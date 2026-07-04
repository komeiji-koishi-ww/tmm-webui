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
    handleMobileTVRow(row, event = null) {
      if (!row) return;
      if (row.level === "episode") {
        this.openMobileItem(row.payload, event);
        return;
      }
      this.activateTVRow(row, event);
    },
  },
};
