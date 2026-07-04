import { GENRE_ZH, GENRE_ZH_LOOKUP } from "./config.js";

export const filtersMixin = {
  computed: {
    filterGroups() {
      const groups = [
        {
          id: "details",
          label: "详情",
          filters: [
            {
              id: "allInOne",
              label: "全部字段",
              type: "text",
              available: true,
            },
            { id: "title", label: "片名", type: "text", available: true },
            {
              id: "originalTitle",
              label: "原始片名",
              type: "text",
              available: true,
            },
            {
              id: "datasource",
              label: "数据源",
              type: "choice",
              available: true,
            },
            {
              id: "dateAdded",
              label: "添加日期",
              type: "date",
              available: true,
            },
            { id: "filename", label: "文件名", type: "text", available: true },
            { id: "path", label: "路径", type: "text", available: true },
            {
              id: "new",
              label: this.activeModule === "tvshow" ? "新剧集" : "新电影",
              type: "boolean",
              available: true,
            },
            {
              id: "duplicate",
              label: this.activeModule === "tvshow" ? "重复剧集" : "重复电影",
              type: "boolean",
              available: false,
            },
            {
              id: "watched",
              label: "已观看",
              type: "boolean",
              available: false,
            },
            {
              id: "locked",
              label: "已锁定",
              type: "boolean",
              available: false,
            },
          ],
        },
        {
          id: "metadata",
          label: "元数据",
          filters: [
            { id: "year", label: "年份", type: "number", available: true },
            { id: "decade", label: "年代", type: "choice", available: true },
            { id: "rating", label: "评分", type: "number", available: true },
            {
              id: "runtime",
              label: "片长",
              type: "number",
              available: this.activeModule === "movie",
            },
            { id: "genre", label: "类型", type: "choice", available: true },
            { id: "tmdbId", label: "TMDb ID", type: "number", available: true },
            { id: "imdbId", label: "IMDb ID", type: "text", available: true },
            {
              id: "missingMetadata",
              label: "缺少元数据",
              type: "boolean",
              available: true,
            },
            {
              id: "missingArtwork",
              label: "缺少图片",
              type: "boolean",
              available: true,
            },
            {
              id: "missingSubtitles",
              label: "缺少字幕",
              type: "boolean",
              available: true,
            },
            { id: "cast", label: "演员", type: "choice", available: false },
            { id: "country", label: "国家", type: "choice", available: false },
            {
              id: "certification",
              label: "分级",
              type: "choice",
              available: false,
            },
            { id: "tag", label: "标签", type: "choice", available: false },
            { id: "note", label: "备注", type: "text", available: false },
            {
              id: "episodeCount",
              label: "集数",
              type: "number",
              available: this.activeModule === "tvshow",
            },
            {
              id: "seasonCount",
              label: "季数",
              type: "number",
              available: this.activeModule === "tvshow",
            },
          ],
        },
        {
          id: "video",
          label: "视频",
          filters: [
            {
              id: "videoFormat",
              label: "视频格式",
              type: "choice",
              available: true,
            },
            {
              id: "videoCodec",
              label: "视频编码",
              type: "choice",
              available: false,
            },
            {
              id: "videoBitrate",
              label: "视频码率",
              type: "number",
              available: false,
            },
            {
              id: "videoBitdepth",
              label: "视频位深",
              type: "number",
              available: false,
            },
            {
              id: "videoContainer",
              label: "容器",
              type: "choice",
              available: false,
            },
            {
              id: "aspectRatio",
              label: "宽高比",
              type: "choice",
              available: false,
            },
            {
              id: "frameRate",
              label: "帧率",
              type: "number",
              available: false,
            },
            { id: "hdrFormat", label: "HDR", type: "choice", available: false },
            {
              id: "videoFilesize",
              label: "文件大小(GB)",
              type: "number",
              available: true,
            },
          ],
        },
        {
          id: "audio",
          label: "音频",
          filters: [
            {
              id: "audioCodec",
              label: "音频编码",
              type: "choice",
              available: true,
            },
            {
              id: "audioChannels",
              label: "声道",
              type: "choice",
              available: false,
            },
            {
              id: "audioLanguage",
              label: "音频语言",
              type: "choice",
              available: false,
            },
            {
              id: "audioTitle",
              label: "音轨标题",
              type: "text",
              available: false,
            },
            {
              id: "audioStreamCount",
              label: "音轨数量",
              type: "number",
              available: false,
            },
          ],
        },
        {
          id: "subtitles",
          label: "字幕",
          filters: [
            {
              id: "subtitleCount",
              label: "字幕数量",
              type: "number",
              available: false,
            },
            {
              id: "subtitleFormat",
              label: "字幕格式",
              type: "choice",
              available: false,
            },
            {
              id: "subtitleLanguage",
              label: "字幕语言",
              type: "choice",
              available: false,
            },
          ],
        },
        {
          id: "artwork",
          label: "图片",
          filters: [
            { id: "poster", label: "海报", type: "boolean", available: true },
            { id: "fanart", label: "同人画", type: "boolean", available: true },
            {
              id: "posterSize",
              label: "海报尺寸",
              type: "number",
              available: false,
            },
            {
              id: "fanartSize",
              label: "同人画尺寸",
              type: "number",
              available: false,
            },
            {
              id: "bannerSize",
              label: "横幅尺寸",
              type: "number",
              available: false,
            },
            {
              id: "clearLogoSize",
              label: "ClearLogo 尺寸",
              type: "number",
              available: false,
            },
            {
              id: "discArtSize",
              label: "DiscArt 尺寸",
              type: "number",
              available: false,
            },
          ],
        },
      ];
      if (this.activeModule === "movie") {
        groups[1].filters.push({
          id: "inMovieSet",
          label: "属于合集",
          type: "boolean",
          available: false,
        });
      } else {
        groups[1].filters.push({
          id: "status",
          label: "剧集状态",
          type: "choice",
          available: false,
        });
        groups[1].filters.push({
          id: "missingEpisodes",
          label: "缺少集",
          type: "boolean",
          available: false,
        });
        groups[1].filters.push({
          id: "uncategorizedEpisodes",
          label: "未分类集",
          type: "boolean",
          available: false,
        });
      }
      return groups;
    },
    filterDefinitions() {
      const map = new Map();
      for (const group of this.filterGroups) {
        for (const filter of group.filters) {
          map.set(filter.id, {
            ...filter,
            group: group.id,
            groupLabel: group.label,
          });
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
        base.push(
          { id: "season", label: "季" },
          { id: "episode", label: "集" },
          { id: "episodeCount", label: "集数" },
        );
      }
      return base;
    },
    datasourceOptions() {
      return (
        this.selectedLibrary
          ? this.selectedLibrary.paths || [this.selectedLibrary.path]
          : []
      ).filter(Boolean);
    },
    genreOptions() {
      const values = new Set();
      for (const item of this.items) {
        for (const genre of item.genres || []) {
          if (genre) values.add(this.localizeGenre(genre));
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
      const tvStats =
        this.activeModule === "tvshow" ? this.buildTVStats(this.items) : null;
      const filtered = this.items.filter((item) => {
        if (query && !(item.searchText || "").includes(query)) return false;
        return this.activeFilters.every((filter) =>
          this.filterAcceptsItem(filter, item, tvStats),
        );
      });
      return filtered;
    },
    sortedMovieRows() {
      return this.sortItems(this.visibleItems.slice());
    },
    movieRows() {
      return this.sortedMovieRows;
    },
    movieTotal() {
      return this.sortedMovieRows.length;
    },
    moviePageRangeText() {
      return `${this.movieTotal} / ${this.items.length} 条目`;
    },
    movieVirtualTotalHeight() {
      return this.movieTotal * this.movieVirtual.rowHeight;
    },
    virtualMovieRows() {
      const rowHeight = this.movieVirtual.rowHeight;
      const buffer = this.movieVirtual.buffer;
      const start = Math.max(
        0,
        Math.floor(this.movieVirtual.scrollTop / rowHeight) - buffer,
      );
      const count =
        Math.ceil(this.movieVirtual.viewportHeight / rowHeight) + buffer * 2;
      return this.movieRows.slice(start, start + count).map((item, offset) => {
        const index = start + offset;
        return { item, index, top: index * rowHeight };
      });
    },
    selectedCountText() {
      if (!this.selectedItem) return "未选择";
      if (this.activeModule === "tvshow" && this.selectedItemIds.length > 1)
        return `已选择 ${this.selectedItemIds.length} / ${this.visibleItems.length}`;
      return `已选择 1 / ${this.visibleItems.length}`;
    },
    activeFilterGroupFilters() {
      return (
        this.filterGroups.find((group) => group.id === this.filterEditor.tab) ||
        this.filterGroups[0] || { filters: [] }
      ).filters;
    },
  },
  methods: {
    filterDefinition(id) {
      return this.filterDefinitions.get(id) || null;
    },
    openFilterEditor() {
      this.filterEditor.open = true;
      this.filterEditor.tab = this.filterGroups[0]
        ? this.filterGroups[0].id
        : "details";
    },
    closeFilterEditor() {
      this.filterEditor.open = false;
    },
    isFilterActive(id) {
      return this.filters.some((filter) => filter.id === id);
    },
    addFilter(id) {
      const definition = this.filterDefinition(id);
      if (!definition || !definition.available || this.isFilterActive(id))
        return;
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
      if (definition.type === "date" && filter.operator !== "lastDays")
        return "date";
      return "text";
    },
    filterValueOptions(filter) {
      if (filter.id === "datasource") return this.datasourceOptions;
      if (filter.id === "genre") return this.genreOptions;
      if (filter.id === "videoFormat")
        return this.uniqueItemValues("videoFormat");
      if (filter.id === "audioCodec")
        return this.uniqueItemValues("audioCodec");
      if (filter.id === "decade") return this.decadeOptions;
      return [];
    },
    uniqueItemValues(field) {
      return [
        ...new Set(this.items.map((item) => item[field]).filter(Boolean)),
      ].sort((a, b) => this.compareText(a, b));
    },
    filterAcceptsItem(filter, item, tvStats = null) {
      const definition = this.filterDefinition(filter.id);
      if (!definition || !definition.available) return true;
      let accepted = true;
      if (definition.type === "number") {
        accepted = this.matchNumber(
          this.filterFieldValue(filter.id, item, tvStats),
          filter,
        );
      } else if (definition.type === "date") {
        accepted = this.matchDate(item.dateAdded, filter);
      } else if (definition.type === "boolean") {
        accepted = this.matchBoolean(
          this.filterFieldValue(filter.id, item, tvStats),
          filter,
        );
      } else if (definition.type === "choice") {
        accepted = this.matchChoice(
          this.filterFieldValue(filter.id, item, tvStats),
          filter,
        );
      } else {
        accepted = this.matchText(
          this.filterFieldValue(filter.id, item, tvStats),
          filter,
        );
      }
      return filter.invert ? !accepted : accepted;
    },
    filterFieldValue(id, item, tvStats = null) {
      switch (id) {
        case "allInOne":
          return [
            item.titleGuess,
            item.showGuess,
            item.originalTitle || item.original,
            item.yearGuess,
            item.matchedName,
            item.fileName,
            item.path,
            item.imdbId,
            item.matchedId,
            item.videoFormat,
            item.audioCodec,
            item.fileSize,
            ...(item.genres || []),
            ...this.localizedGenres(item.genres || []),
          ].join(" ");
        case "title":
          return item.kind === "tvshow"
            ? item.showGuess || item.titleGuess
            : item.titleGuess;
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
          return [
            ...(item.genres || []),
            ...this.localizedGenres(item.genres || []),
          ];
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
        if (!stats.has(key))
          stats.set(key, { episodes: 0, seasons: new Set() });
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
        return filter.operator === "contains"
          ? actual.includes(expected)
          : actual === expected;
      });
    },
    matchBoolean(value, filter) {
      const expected =
        filter.operator === "not"
          ? false
          : String(filter.value || "true") !== "false";
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
        return this.compareText(
          a.titleGuess || a.fileName,
          b.titleGuess || b.fileName,
        );
      });
    },
    sortEpisodes(items) {
      return items.sort(
        (a, b) =>
          (a.season || 0) - (b.season || 0) ||
          (a.episode || 0) - (b.episode || 0) ||
          this.compareEpisodeList(a.episodes, b.episodes) ||
          this.compareText(a.fileName, b.fileName) ||
          this.compareText(a.path, b.path),
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
        if (this.sortKey === "episodeCount")
          return (a.episodes - b.episodes) * direction;
        const result = this.compareSortValue(
          firstA || {},
          firstB || {},
          this.sortKey,
        );
        if (result !== 0) return result * direction;
        return this.compareText(a.title, b.title);
      });
    },
    compareSortValue(a, b, key) {
      if (
        [
          "year",
          "rating",
          "runtime",
          "season",
          "episode",
          "videoFilesize",
        ].includes(key)
      ) {
        return (
          Number(a[this.sortFieldName(key)] || 0) -
          Number(b[this.sortFieldName(key)] || 0)
        );
      }
      if (key === "dateAdded")
        return this.compareDate(a.dateAdded, b.dateAdded);
      if (key === "metadata") return Number(!!a.hasNfo) - Number(!!b.hasNfo);
      if (key === "artwork") return this.artworkScore(a) - this.artworkScore(b);
      if (key === "datasource")
        return this.compareText(a.sourcePath, b.sourcePath);
      if (key === "videoFormat")
        return this.compareText(a.videoFormat, b.videoFormat);
      if (key === "audioCodec")
        return this.compareText(a.audioCodec, b.audioCodec);
      if (key === "originalTitle")
        return this.compareText(
          a.originalTitle || a.original,
          b.originalTitle || b.original,
        );
      return this.compareText(
        a.kind === "tvshow" ? a.showGuess || a.titleGuess : a.titleGuess,
        b.kind === "tvshow" ? b.showGuess || b.titleGuess : b.titleGuess,
      );
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
      return String(a || "").localeCompare(String(b || ""), "zh-CN", {
        numeric: true,
        sensitivity: "base",
      });
    },
    compareDate(a, b) {
      return (
        (this.dateOnly(a)?.getTime() || 0) - (this.dateOnly(b)?.getTime() || 0)
      );
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
    localizeGenre(genre) {
      const text = String(genre || "").trim();
      if (!text) return "";
      return GENRE_ZH[text] || GENRE_ZH_LOOKUP[text.toLowerCase()] || text;
    },
    localizedGenres(genres) {
      const seen = new Set();
      const values = [];
      for (const genre of genres || []) {
        const localized = this.localizeGenre(genre);
        if (!localized || seen.has(localized)) continue;
        seen.add(localized);
        values.push(localized);
      }
      return values;
    },
    tvShowRating(show) {
      if (!show) return 0;
      const items = show.items || [];
      const withShowRating = items.find(
        (item) => Number(item.showRating || 0) > 0,
      );
      if (withShowRating) return Number(withShowRating.showRating || 0);
      const first = this.firstTVItem(show);
      return first ? Number(first.showRating || 0) : 0;
    },
  },
};
