export const layoutMixin = {
  computed: {
    workbenchStyle() {
      const width = this.clampedInspectorWidth(this.layout.inspectorWidth);
      return {
        "--inspector-width": `${width}px`,
      };
    },
    filterEditorStyle() {
      return {
        gridTemplateColumns: `${this.layout.filterNavWidth}px 6px minmax(0, 1fr)`,
      };
    },
    movieGridStyle() {
      return {
        gridTemplateColumns: this.movieGridTemplate,
      };
    },
    movieGridTemplate() {
      const columns = this.layout.movieColumns;
      return `${columns[0]}px ${columns[1]}px ${columns[2]}px ${columns[3]}px minmax(${columns[4]}px, 1fr)`;
    },
    tvGridStyle() {
      return {
        gridTemplateColumns: this.layout.tvColumns
          .map((width) => `${width}px`)
          .join(" "),
      };
    },
  },
  methods: {
    loadLayoutSettings() {
      const browserWidth = Number(
        localStorage.getItem("tmmweb.browserWidth") || 0,
      );
      const inspectorWidth = Number(
        localStorage.getItem("tmmweb.inspectorWidth") || 0,
      );
      const filterNavWidth = Number(
        localStorage.getItem("tmmweb.filterNavWidth") || 0,
      );
      const movieColumns = this.loadColumnWidths(
        "tmmweb.movieColumns",
        this.layout.movieColumns,
      );
      const tvColumns = this.loadColumnWidths(
        "tmmweb.tvColumns",
        this.layout.tvColumns,
      );
      if (inspectorWidth >= 320)
        this.layout.inspectorWidth = this.clampedInspectorWidth(inspectorWidth);
      else if (browserWidth >= 520) {
        const migrated = window.innerWidth
          ? window.innerWidth - browserWidth - 6
          : 440;
        this.layout.inspectorWidth = this.clampedInspectorWidth(migrated);
      }
      if (filterNavWidth >= 140) this.layout.filterNavWidth = filterNavWidth;
      this.layout.movieColumns = movieColumns;
      this.layout.tvColumns = tvColumns;
    },
    clampedInspectorWidth(value) {
      const viewport = window.innerWidth || 1440;
      const max = Math.max(360, Math.min(680, viewport - 640));
      return Math.min(max, Math.max(360, Number(value) || 440));
    },
    loadColumnWidths(key, fallback) {
      try {
        const values = JSON.parse(localStorage.getItem(key) || "[]");
        if (!Array.isArray(values) || values.length !== fallback.length)
          return fallback.slice();
        return values.map((value, index) =>
          Math.max(
            this.minColumnWidth(index),
            Number(value) || fallback[index],
          ),
        );
      } catch (_) {
        return fallback.slice();
      }
    },
    startWorkbenchResize(event) {
      const workbench =
        this.$refs.workbench &&
        (this.$refs.workbench.$el || this.$refs.workbench);
      const rect =
        workbench && workbench.getBoundingClientRect
          ? workbench.getBoundingClientRect()
          : null;
      if (!rect) return;
      this.layout.resizing = {
        type: "workbench",
        startX: event.clientX,
        startWidth: this.clampedInspectorWidth(this.layout.inspectorWidth),
        containerWidth: rect.width,
      };
      event.preventDefault();
    },
    tableColumnIndex(column) {
      const key =
        column &&
        (column.columnKey ||
          column.property ||
          column.rawColumnKey ||
          column.label);
      return {
        title: 0,
        titleGuess: 0,
        标题: 0,
        "标题 / 季 / 集": 0,
        year: 1,
        yearGuess: 1,
        年份: 1,
        rating: 2,
        评分: 2,
        dateAdded: 3,
        添加日期: 3,
        status: 4,
        "NFO / 图片 / 媒体信息": 4,
      }[key];
    },
    handleTableHeaderDragend(kind, newWidth, _oldWidth, column) {
      const index = this.tableColumnIndex(column);
      if (index === undefined) return;
      const key = kind === "movie" ? "movieColumns" : "tvColumns";
      const widths = this.layout[key].slice();
      widths[index] = Math.max(
        this.minColumnWidth(index),
        Number(newWidth) || widths[index],
      );
      this.layout[key] = widths;
      localStorage.setItem(
        `tmmweb.${key}`,
        JSON.stringify(widths.map((value) => Math.round(value))),
      );
    },
    saveColumnWidths(kind) {
      const key = kind === "movie" ? "movieColumns" : "tvColumns";
      localStorage.setItem(
        `tmmweb.${key}`,
        JSON.stringify(this.layout[key].map((value) => Math.round(value))),
      );
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
      const columns =
        kind === "movie" ? this.layout.movieColumns : this.layout.tvColumns;
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
        const max = Math.max(360, Math.min(680, resizing.containerWidth - 640));
        this.layout.inspectorWidth = Math.min(
          max,
          Math.max(360, resizing.startWidth - delta),
        );
        return;
      }
      if (resizing.type === "filterNav") {
        this.layout.filterNavWidth = Math.min(
          320,
          Math.max(140, resizing.startWidth + delta),
        );
      }
      if (resizing.type === "column") {
        const key = resizing.kind === "movie" ? "movieColumns" : "tvColumns";
        const columns = this.layout[key].slice();
        columns[resizing.index] = Math.max(
          this.minColumnWidth(resizing.index),
          resizing.startWidth + delta,
        );
        this.layout[key] = columns;
      }
    },
    stopResize() {
      if (!this.layout.resizing) return;
      const resizing = this.layout.resizing;
      localStorage.setItem(
        "tmmweb.inspectorWidth",
        String(
          Math.round(this.clampedInspectorWidth(this.layout.inspectorWidth)),
        ),
      );
      localStorage.removeItem("tmmweb.browserWidth");
      localStorage.setItem(
        "tmmweb.filterNavWidth",
        String(Math.round(this.layout.filterNavWidth)),
      );
      localStorage.setItem(
        "tmmweb.movieColumns",
        JSON.stringify(
          this.layout.movieColumns.map((value) => Math.round(value)),
        ),
      );
      localStorage.setItem(
        "tmmweb.tvColumns",
        JSON.stringify(this.layout.tvColumns.map((value) => Math.round(value))),
      );
      this.layout.resizing = null;
      if (resizing.type === "column") this.saveColumnWidths(resizing.kind);
    },
    minColumnWidth(index) {
      if (index === 0) return 180;
      if (index === 4) return 220;
      return 70;
    },
  },
};
