export const summaryMixin = {
  computed: {
    detailTitle() {
      if (!this.selectedSummary) return "";
      return this.selectedSummary.title;
    },
    selectedSummary() {
      if (this.selectedEntity && this.selectedEntity.kind === "show") {
        const show = this.selectedEntity.payload;
        const first = this.firstTVItem(show);
        return this.buildSummary(first, {
          entityType: "show",
          title: show.title,
          subtitle: `${show.seasons.length} 季 / ${show.episodes} 集`,
          year: first ? first.yearGuess : "",
          rating: this.tvShowRating(show),
          showMediaFields: false,
          showRatingField: true,
          itemCount: show.episodes,
          folderPath: first ? this.showRootPath(first) : "",
        });
      }
      if (this.selectedEntity && this.selectedEntity.kind === "season") {
        const season = this.selectedEntity.payload;
        const first = season.items[0];
        return this.buildSummary(first, {
          entityType: "season",
          title:
            `${season.showTitle || (first && first.showGuess) || ""} ${season.title}`.trim(),
          subtitle: `${season.items.length} 集`,
          year: first ? first.yearGuess : "",
          rating: 0,
          showMediaFields: false,
          showRatingField: false,
          season: season.season,
          itemCount: season.items.length,
          folderPath: first ? first.dir || this.showRootPath(first) : "",
        });
      }
      if (!this.selectedItem) return null;
      return this.buildSummary(this.selectedItem, {
        entityType: this.selectedItem.kind === "tvshow" ? "episode" : "movie",
        title:
          this.selectedItem.kind === "tvshow"
            ? this.selectedItem.showGuess || this.selectedItem.titleGuess
            : this.selectedItem.titleGuess,
        subtitle:
          this.selectedItem.kind === "tvshow"
            ? this.itemSeasonText(this.selectedItem)
            : this.selectedItem.originalTitle ||
              this.selectedItem.original ||
              "",
        year: this.selectedItem.yearGuess,
        rating:
          this.selectedItem.kind === "tvshow"
            ? this.tvEpisodeRating(this.selectedItem)
            : this.selectedItem.rating || 0,
        showMediaFields: true,
        showRatingField: true,
        folderPath: this.selectedItem.dir || this.selectedItem.sourcePath || "",
      });
    },
  },
  methods: {
    artworkURL(item, type, entityType = "") {
      if (!item) return "";
      if (
        item.kind !== "tvshow" &&
        ((type === "poster" && !item.hasPoster) ||
          (type === "fanart" && !item.hasFanart))
      )
        return "";
      const scope = item.kind === "tvshow" ? entityType || "episode" : "movie";
      return `/api/artwork?id=${encodeURIComponent(item.id)}&type=${encodeURIComponent(type)}&scope=${encodeURIComponent(scope)}&v=${encodeURIComponent(item.dateAdded || item.matchedId || "")}`;
    },
    hideBrokenImage(event) {
      if (event && event.target) event.target.style.display = "none";
    },
    showLoadedImage(event) {
      if (event && event.target) event.target.style.display = "";
    },
    buildSummary(item, overrides = {}) {
      const fallback = item || {};
      const entityType = overrides.entityType || fallback.kind || "movie";
      return {
        item: fallback,
        entityType,
        title:
          overrides.title ||
          fallback.titleGuess ||
          fallback.showGuess ||
          "未命名",
        subtitle:
          overrides.subtitle ||
          fallback.originalTitle ||
          fallback.original ||
          "",
        year: overrides.year || fallback.yearGuess || "",
        season: overrides.season || fallback.season || 0,
        itemCount: overrides.itemCount || 1,
        poster: this.artworkURL(fallback, "poster", entityType),
        fanart: this.artworkURL(fallback, "fanart", entityType),
        rating: Object.prototype.hasOwnProperty.call(overrides, "rating")
          ? overrides.rating
          : fallback.rating || 0,
        showMediaFields: Object.prototype.hasOwnProperty.call(
          overrides,
          "showMediaFields",
        )
          ? overrides.showMediaFields
          : true,
        showRatingField: Object.prototype.hasOwnProperty.call(
          overrides,
          "showRatingField",
        )
          ? overrides.showRatingField
          : true,
        genres: fallback.genres || [],
        actors: fallback.actors || [],
        overview: fallback.overview || "",
        dateAdded: fallback.dateAdded || "",
        matchedName: fallback.matchedName || "",
        matchedId: fallback.matchedId || 0,
        imdbId: fallback.imdbId || "",
        fileSize: fallback.fileSize || "",
        fileSizeBytes: fallback.fileSizeBytes || 0,
        videoFormat: fallback.videoFormat || "",
        audioCodec: fallback.audioCodec || "",
        hasNfo: !!fallback.hasNfo,
        hasPoster: !!fallback.hasPoster,
        hasFanart: !!fallback.hasFanart,
        hasSubtitle: !!fallback.hasSubtitle,
        sourcePath: fallback.sourcePath || "",
        folderPath:
          overrides.folderPath || fallback.dir || fallback.sourcePath || "",
      };
    },
  },
};
