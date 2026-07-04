import {
  DEFAULT_MOVIE_RENAMER_FILE,
  DEFAULT_MOVIE_RENAMER_PATH,
  DEFAULT_TVSHOW_RENAMER_FILE,
  DEFAULT_TVSHOW_RENAMER_PATH,
  DEFAULT_TVSHOW_RENAMER_SEASON,
  MOVIE_RENAMER_TOKEN_ROWS,
  TMM_RENAMER_TOKENS,
  TV_RENAMER_TOKEN_ROWS,
} from "./config.js";

export const renamerMixin = {
  computed: {
    renamePreviewText() {
      if (!this.renamePreview) return "";
      if (
        this.renamePreview.operations &&
        this.renamePreview.operations.length
      ) {
        return this.renamePreview.operations
          .map((op) => `${op.kind || "file"}:\n${op.source}\n=>\n${op.target}`)
          .join("\n\n");
      }
      return `文件:\n${this.renamePreview.sourceFile}\n=>\n${this.renamePreview.targetFile}\n\n目录:\n${this.renamePreview.sourceDir}\n=>\n${this.renamePreview.targetDir}`;
    },
    localRenamePreviewRows() {
      return this.localRename.rows.map((row) => ({
        ...row,
        previewName: this.localRenamePreviewName(row),
      }));
    },
    localRenameCanApply() {
      if (
        !this.localRename.open ||
        this.localRename.saving ||
        !this.localRename.rows.length
      )
        return false;
      const rows = this.localRenamePreviewRows;
      return (
        rows.every((row) => row.previewName) &&
        rows.some((row) => row.previewName !== row.fileName)
      );
    },
  },
  methods: {
    renamerTokens() {
      return TMM_RENAMER_TOKENS;
    },
    renamerTokenRows(kind) {
      return kind === "tvshow"
        ? TV_RENAMER_TOKEN_ROWS
        : MOVIE_RENAMER_TOKEN_ROWS;
    },
    resetRenamerPattern(kind, field) {
      const defaults = {
        moviePath: DEFAULT_MOVIE_RENAMER_PATH,
        movieFile: DEFAULT_MOVIE_RENAMER_FILE,
        tvShow: DEFAULT_TVSHOW_RENAMER_PATH,
        tvSeason: DEFAULT_TVSHOW_RENAMER_SEASON,
        tvFile: DEFAULT_TVSHOW_RENAMER_FILE,
      };
      const setting = {
        moviePath: "movieRenamerPathname",
        movieFile: "movieRenamerFilename",
        tvShow: "tvShowRenamerShowFolder",
        tvSeason: "tvShowRenamerSeason",
        tvFile: "tvShowRenamerFilename",
      }[`${kind}${field}`];
      const value = defaults[`${kind}${field}`];
      if (setting && value) this.scraperSettings[setting] = value;
    },
    renameOptions(kind, segment) {
      if (kind === "movie" && segment === "folder") {
        return {
          spaceSubstitution:
            !!this.scraperSettings.movieRenamerPathSpaceSubstitution,
          spaceReplacement:
            this.scraperSettings.movieRenamerPathSpaceReplacement || "_",
          colonReplacement:
            this.scraperSettings.movieRenamerColonReplacement || "-",
          asciiReplacement: !!this.scraperSettings.movieRenamerAsciiReplacement,
        };
      }
      if (kind === "movie") {
        return {
          spaceSubstitution:
            !!this.scraperSettings.movieRenamerFilenameSpaceSubstitution,
          spaceReplacement:
            this.scraperSettings.movieRenamerFilenameSpaceReplacement || "_",
          colonReplacement:
            this.scraperSettings.movieRenamerColonReplacement || "-",
          asciiReplacement: !!this.scraperSettings.movieRenamerAsciiReplacement,
        };
      }
      if (segment === "show") {
        return {
          spaceSubstitution:
            !!this.scraperSettings.tvShowRenamerShowFolderSpaceSubstitution,
          spaceReplacement:
            this.scraperSettings.tvShowRenamerShowFolderSpaceReplacement || "_",
          colonReplacement:
            this.scraperSettings.tvShowRenamerColonReplacement || " ",
          asciiReplacement:
            !!this.scraperSettings.tvShowRenamerAsciiReplacement,
        };
      }
      if (segment === "season") {
        return {
          spaceSubstitution:
            !!this.scraperSettings.tvShowRenamerSeasonFolderSpaceSubstitution,
          spaceReplacement:
            this.scraperSettings.tvShowRenamerSeasonFolderSpaceReplacement ||
            "_",
          colonReplacement:
            this.scraperSettings.tvShowRenamerColonReplacement || " ",
          asciiReplacement:
            !!this.scraperSettings.tvShowRenamerAsciiReplacement,
        };
      }
      return {
        spaceSubstitution:
          !!this.scraperSettings.tvShowRenamerFilenameSpaceSubstitution,
        spaceReplacement:
          this.scraperSettings.tvShowRenamerFilenameSpaceReplacement || "_",
        colonReplacement:
          this.scraperSettings.tvShowRenamerColonReplacement || " ",
        asciiReplacement: !!this.scraperSettings.tvShowRenamerAsciiReplacement,
      };
    },
    previewPattern(pattern, kind, segment = "file") {
      const replacements = {
        "{title}": "银翼杀手",
        "${title}": "银翼杀手",
        "{title[0]}": "银",
        "${title[0]}": "银",
        "{title;first}": this.firstCharacterToken(
          "银翼杀手",
          this.scraperSettings.movieRenamerFirstCharacterReplacement,
        ),
        "${title;first}": this.firstCharacterToken(
          "银翼杀手",
          this.scraperSettings.movieRenamerFirstCharacterReplacement,
        ),
        "{title[0,2]}": "银翼",
        "${title[0,2]}": "银翼",
        "{originalTitle}": "Blade Runner",
        "${originalTitle}": "Blade Runner",
        "{originalFilename}": "Blade.Runner.1982.1080p.DTS.mkv",
        "${originalFilename}": "Blade.Runner.1982.1080p.DTS.mkv",
        "{originalBasename}": "Blade.Runner.1982.1080p.DTS",
        "${originalBasename}": "Blade.Runner.1982.1080p.DTS",
        "{edition}": "Final Cut",
        "${edition}": "Final Cut",
        "${- ,edition,}": "- Final Cut",
        "{year}": "1982",
        "${year}": "1982",
        "{releaseDate}": "1982-06-25",
        "${releaseDate}": "1982-06-25",
        "{rating}": "8.1",
        "${rating}": "8.1",
        "{imdb}": "tt0083658",
        "${imdb}": "tt0083658",
        "{tmdb}": "78",
        "${tmdb}": "78",
        "{tmdbid}": "78",
        "${tmdbid}": "78",
        "{videoFormat}": "1080p",
        "${videoFormat}": "1080p",
        "{audioCodec}": "DTS",
        "${audioCodec}": "DTS",
        "{fileSize}": "12.4GB",
        "${fileSize}": "12.4GB",
        "{filesize}": "12.4GB",
        "${filesize}": "12.4GB",
        "{showTitle}": "绝命毒师",
        "${showTitle}": "绝命毒师",
        "{showOriginalTitle}": "Breaking Bad",
        "${showOriginalTitle}": "Breaking Bad",
        "{showYear}": "2008",
        "${showYear}": "2008",
        "{seasonNr}": "1",
        "${seasonNr}": "1",
        "{seasonNr2}": "01",
        "${seasonNr2}": "01",
        "{episodeNr}": "1",
        "${episodeNr}": "1",
        "{episodeNr2}": "01",
        "${episodeNr2}": "01",
        "{aired}": "2008-01-20",
        "${aired}": "2008-01-20",
        "{airedDate}": "2008-01-20",
        "${airedDate}": "2008-01-20",
      };
      if (kind === "tvshow") {
        Object.assign(replacements, {
          "{title}": "试播集",
          "${title}": "试播集",
          "{title[0]}": "试",
          "${title[0]}": "试",
          "{title;first}": this.firstCharacterToken(
            "试播集",
            this.scraperSettings.tvShowRenamerFirstCharacterReplacement,
          ),
          "${title;first}": this.firstCharacterToken(
            "试播集",
            this.scraperSettings.tvShowRenamerFirstCharacterReplacement,
          ),
          "{title[0,2]}": "试播",
          "${title[0,2]}": "试播",
          "{originalTitle}": "Pilot",
          "${originalTitle}": "Pilot",
          "{originalFilename}": "Breaking.Bad.S01E01.1080p.EAC3.mkv",
          "${originalFilename}": "Breaking.Bad.S01E01.1080p.EAC3.mkv",
          "{originalBasename}": "Breaking.Bad.S01E01.1080p.EAC3",
          "${originalBasename}": "Breaking.Bad.S01E01.1080p.EAC3",
          "{audioCodec}": "EAC3",
          "${audioCodec}": "EAC3",
          "{fileSize}": "2.1GB",
          "${fileSize}": "2.1GB",
          "{filesize}": "2.1GB",
          "${filesize}": "2.1GB",
        });
      }
      let value =
        pattern ||
        (kind === "movie"
          ? DEFAULT_MOVIE_RENAMER_FILE
          : DEFAULT_TVSHOW_RENAMER_FILE);
      Object.keys(replacements)
        .sort((a, b) => b.length - a.length)
        .forEach((key) => {
          value = value.split(key).join(replacements[key]);
        });
      return this.cleanPreviewName(value, this.renameOptions(kind, segment));
    },
    cleanPreviewName(value, options = {}) {
      let result = String(value || "").replace(/\$\{[^}]+\}/g, "");
      const colonReplacement =
        options.colonReplacement === undefined ||
        options.colonReplacement === null
          ? "-"
          : options.colonReplacement;
      result = result.replace(/[:：]/g, colonReplacement);
      result = result
        .replace(/[\\/]/g, " ")
        .replace(/[*?"]/g, "")
        .replace(/[<>]/g, "")
        .replace(/\|/g, " ");
      if (options.asciiReplacement) {
        result = result
          .normalize("NFKD")
          .replace(/[\u0300-\u036f]/g, "")
          .replace(/[^\x00-\x7F]/g, "");
      }
      result = result
        .replace(/\s+/g, " ")
        .replace(/\( \)/g, "")
        .replace(/\[ \]/g, "")
        .replace(/[ ._-]+$/g, "")
        .trim();
      if (options.spaceSubstitution) {
        const replacement = options.spaceReplacement || "_";
        const escaped = replacement.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");
        result = result
          .replace(/ /g, replacement)
          .replace(new RegExp(`${escaped}+`, "g"), replacement)
          .replace(/[ ._-]+$/g, "")
          .trim();
      }
      return result;
    },
    firstCharacterToken(value, replacement) {
      const text = String(value || "").trim();
      if (!text) return "";
      const first = [...text][0];
      if (/^\p{L}$/u.test(first)) return first;
      return String(replacement || "#").trim() || "#";
    },
    movieRenamerExample() {
      return {
        movie: "银翼杀手",
        datasource: "/media/movies",
        folder: this.previewPattern(
          this.scraperSettings.movieRenamerPathname,
          "movie",
          "folder",
        ),
        filename: `${this.previewPattern(this.scraperSettings.movieRenamerFilename, "movie", "file")}.mkv`,
      };
    },
    tvRenamerExample() {
      const show = this.previewPattern(
        this.scraperSettings.tvShowRenamerShowFolder,
        "tvshow",
        "show",
      );
      const season = this.previewPattern(
        this.scraperSettings.tvShowRenamerSeason,
        "tvshow",
        "season",
      );
      return {
        show: "绝命毒师",
        episode: "1.1 试播集",
        datasource: "/media/tv",
        folder: `${show}/${season}`,
        filename: `${this.previewPattern(this.scraperSettings.tvShowRenamerFilename, "tvshow", "file")}.mkv`,
      };
    },
    openLocalRename(items, mode, options = {}) {
      const rows = items.filter(Boolean).map((item) => ({
        itemId: item.id,
        fileName: item.fileName || this.basename(item.path),
        newFileName: item.fileName || this.basename(item.path),
      }));
      if (!rows.length) return;
      const singleManual =
        options.singleManual || mode === "movie" || rows.length === 1;
      this.localRename = {
        open: true,
        mode,
        tab: singleManual ? "manual" : "replace",
        saving: false,
        error: "",
        replaceText: "",
        replaceWith: "",
        addPosition: "prefix",
        addText: "",
        rows,
      };
    },
    closeLocalRename() {
      this.localRename.open = false;
      this.localRename.error = "";
    },
    basename(path) {
      return (
        String(path || "")
          .split("/")
          .pop() || ""
      );
    },
    splitFilename(fileName) {
      const index = fileName.lastIndexOf(".");
      if (index <= 0) return { base: fileName, ext: "" };
      return { base: fileName.slice(0, index), ext: fileName.slice(index) };
    },
    localRenamePreviewName(row) {
      if (
        this.localRename.mode === "movie" ||
        this.localRename.tab === "manual"
      )
        return row.newFileName.trim();
      if (this.localRename.tab === "replace") {
        const source = this.localRename.replaceText;
        if (!source) return row.fileName;
        return row.fileName.split(source).join(this.localRename.replaceWith);
      }
      if (this.localRename.tab === "add") {
        const addition = this.localRename.addText;
        if (!addition) return row.fileName;
        const parts = this.splitFilename(row.fileName);
        if (this.localRename.addPosition === "suffix")
          return `${parts.base}${addition}${parts.ext}`;
        return `${addition}${parts.base}${parts.ext}`;
      }
      return row.fileName;
    },
    async applyLocalRename() {
      if (!this.localRenameCanApply) return;
      this.localRename.saving = true;
      this.localRename.error = "";
      this.status = "正在重命名本地文件";
      try {
        const requests = this.localRenamePreviewRows.map((row) => ({
          itemId: row.itemId,
          newFileName: row.previewName,
        }));
        const result = await this.api("/api/local-rename", {
          method: "POST",
          body: JSON.stringify({ items: requests }),
        });
        const updatedItems = this.normalizeItems(result.items || []);
        const byOldID = new Map(
          this.localRename.rows.map((row, index) => [
            row.itemId,
            updatedItems[index],
          ]),
        );
        const byNewID = new Map(updatedItems.map((item) => [item.id, item]));
        this.items = this.items.map(
          (item) => byOldID.get(item.id) || byNewID.get(item.id) || item,
        );
        this.selectedItemIds = updatedItems.map((item) => item.id);
        this.lastSelectedTVItemId = updatedItems[0] ? updatedItems[0].id : "";
        this.tvRangeAnchorItemId = this.lastSelectedTVItemId;
        if (this.selectedItem) {
          this.selectedItem =
            byOldID.get(this.selectedItem.id) ||
            byNewID.get(this.selectedItem.id) ||
            updatedItems[0] ||
            this.selectedItem;
        } else {
          this.selectedItem = updatedItems[0] || null;
        }
        if (this.selectedItem)
          this.selectedEntity = {
            kind: this.selectedItem.kind === "tvshow" ? "episode" : "movie",
            payload: this.selectedItem,
          };
        this.status = `已重命名 ${updatedItems.length} 个文件`;
        this.closeLocalRename();
      } catch (error) {
        this.localRename.error = error.message;
        this.status = error.message;
      } finally {
        this.localRename.saving = false;
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
            movieRenamerPathname: this.scraperSettings.movieRenamerPathname,
            movieRenamerFilename:
              this.selectedItem.kind === "tvshow"
                ? this.scraperSettings.movieRenamerFilename
                : this.rename.pattern,
            tvShowRenamerShowFolder:
              this.scraperSettings.tvShowRenamerShowFolder,
            tvShowRenamerSeason: this.scraperSettings.tvShowRenamerSeason,
            tvShowRenamerFilename:
              this.selectedItem.kind === "tvshow"
                ? this.rename.pattern
                : this.scraperSettings.tvShowRenamerFilename,
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
      if (!confirm("确认执行重命名？请确保 Plex/TMM 没有正在扫描该文件。"))
        return;
      this.busy = true;
      this.status = "正在执行重命名";
      try {
        const result = await this.api("/api/rename/apply", {
          method: "POST",
          body: JSON.stringify(this.renamePreview),
        });
        if (result.item) {
          const updatedItem = this.normalizeItem(result.item);
          const oldPath = this.renamePreview.sourceFile;
          this.items = this.items.map((item) =>
            item.id === updatedItem.id || item.path === oldPath
              ? updatedItem
              : item,
          );
          this.selectedItem = updatedItem;
        }
        this.status = "重命名完成";
        this.renamePreview = null;
      } catch (error) {
        this.status = error.message;
      } finally {
        this.busy = false;
      }
    },
  },
};
