const { createApp } = Vue;

createApp({
  data() {
    return {
      libraries: [],
      selectedLibrary: null,
      items: [],
      selectedItem: null,
      candidates: [],
      rename: {
        pattern: "{title} ({year}) {tmdb-{tmdbid}}",
        title: "",
        year: "",
        tmdbId: 0,
      },
      renamePreview: null,
      newLibrary: {
        name: "电影",
        path: "",
        type: "movie",
      },
      status: "正在初始化",
      busy: false,
    };
  },
  computed: {
    renamePreviewText() {
      if (!this.renamePreview) return "";
      return `文件:\n${this.renamePreview.sourceFile}\n=>\n${this.renamePreview.targetFile}\n\n目录:\n${this.renamePreview.sourceDir}\n=>\n${this.renamePreview.targetDir}`;
    },
  },
  async mounted() {
    await this.loadLibraries();
    this.status = "就绪";
  },
  methods: {
    async api(path, options = {}) {
      const response = await fetch(path, {
        headers: { "Content-Type": "application/json" },
        ...options,
      });
      if (!response.ok) {
        throw new Error(await response.text());
      }
      return response.json();
    },
    async loadLibraries() {
      this.libraries = await this.api("/api/libraries");
      if (this.libraries.length && !this.selectedLibrary) {
        this.selectLibrary(this.libraries[0]);
      }
    },
    async addLibrary() {
      this.busy = true;
      this.status = "正在添加媒体库";
      try {
        const library = await this.api("/api/libraries", {
          method: "POST",
          body: JSON.stringify(this.newLibrary),
        });
        this.libraries.push(library);
        this.selectLibrary(library);
        this.newLibrary.path = "";
        this.status = "媒体库已添加";
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
    selectLibrary(library) {
      this.selectedLibrary = library;
      this.items = [];
      this.selectedItem = null;
      this.candidates = [];
      this.status = `已选择 ${library.name}`;
    },
    selectItem(item) {
      this.selectedItem = item;
      this.candidates = [];
      this.rename.title = item.titleGuess;
      this.rename.year = item.yearGuess || "";
      this.rename.tmdbId = item.matchedId || 0;
      this.renamePreview = null;
    },
    async scan() {
      if (!this.selectedLibrary) return;
      this.busy = true;
      this.status = "正在扫描媒体库";
      try {
        const result = await this.api("/api/scan", {
          method: "POST",
          body: JSON.stringify({ libraryId: this.selectedLibrary.id }),
        });
        this.items = result.items;
        this.status = `扫描完成，共 ${result.count} 个视频文件`;
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
        this.candidates = await this.api(`/api/search?itemId=${encodeURIComponent(this.selectedItem.id)}`);
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
      try {
        const result = await this.api("/api/scrape", {
          method: "POST",
          body: JSON.stringify({
            itemId: this.selectedItem.id,
            tmdbId: candidate.id,
            writeNfo: true,
            writeImages: true,
          }),
        });
        this.selectedItem = result.item;
        this.rename.title = result.movie.title;
        this.rename.year = (result.movie.releaseDate || "").slice(0, 4);
        this.rename.tmdbId = result.movie.id;
        this.status = "刮削写入完成";
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
        await this.api("/api/rename/apply", {
          method: "POST",
          body: JSON.stringify(this.renamePreview),
        });
        this.status = "重命名完成，请重新扫描媒体库";
        this.renamePreview = null;
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
  },
}).mount("#app");

