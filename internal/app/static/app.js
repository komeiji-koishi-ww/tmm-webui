import { chooserMixin } from "./js/chooser.js";
import { filtersMixin } from "./js/filters.js";
import { layoutMixin } from "./js/layout.js";
import { libraryMixin } from "./js/library.js";
import { mobileMixin } from "./js/mobile.js";
import { renamerMixin } from "./js/renamer.js";
import { createInitialState } from "./js/state.js";
import { settingsMixin } from "./js/settings.js";
import { summaryMixin } from "./js/summary.js";
import { tvMixin } from "./js/tv.js";

const { createApp } = Vue;
const ElementIcons = window.ElementPlusIconsVue || {};
const {
  ArrowUp,
  EditPen,
  Filter,
  Folder,
  MagicStick,
  Monitor,
  Refresh,
  Search,
  Setting,
  SortDown,
  SortUp,
} = ElementIcons;

const app = createApp({
  mixins: [
    layoutMixin,
    libraryMixin,
    filtersMixin,
    tvMixin,
    summaryMixin,
    renamerMixin,
    chooserMixin,
    settingsMixin,
    mobileMixin,
  ],
  data() {
    return createInitialState({
      ArrowUp,
      EditPen,
      Filter,
      Folder,
      MagicStick,
      Monitor,
      Refresh,
      Search,
      Setting,
      SortDown,
      SortUp,
    });
  },
  computed: {
    tmdbStatusText() {
      if (!this.scraperSettings.tmdbConfigured) return "未配置";
      if (this.scraperSettings.tmdbKeySource === "environment")
        return "已通过环境变量配置";
      if (this.scraperSettings.tmdbKeySource === "settings") return "已配置";
      return "已启用";
    },
    proxyStatusText() {
      if (!this.scraperSettings.proxyEnabled) return "未启用代理";
      const host = this.scraperSettings.proxyHost || "未设置主机";
      const port = this.scraperSettings.proxyPort || 80;
      return `HTTP 代理 ${host}:${port}`;
    },
  },
  watch: {
    mobileDetailOpen(open, wasOpen) {
      this.handleMobileSurfaceChange("detail", open, wasOpen);
    },
    settingsOpen(open, wasOpen) {
      this.handleMobileSurfaceChange("settings", open, wasOpen);
    },
    "browser.open"(open, wasOpen) {
      this.handleMobileSurfaceChange("browser", open, wasOpen);
    },
    "localRename.open"(open, wasOpen) {
      this.handleMobileSurfaceChange("rename", open, wasOpen);
    },
    "filterEditor.open"(open, wasOpen) {
      this.handleMobileSurfaceChange("filters", open, wasOpen);
    },
    "chooser.open"(open, wasOpen) {
      this.handleMobileSurfaceChange("chooser", open, wasOpen);
    },
    query() {
      this.resetMoviePage();
    },
    filters: {
      deep: true,
      handler() {
        this.resetMoviePage();
      },
    },
    sortKey() {
      this.resetMoviePage();
    },
    sortDirection() {
      this.resetMoviePage();
    },
    activeModule() {
      this.resetMoviePage();
    },
  },
  async mounted() {
    await this.loadSettings();
    this.loadLayoutSettings();
    await this.loadLibraries();
    this.startPolling();
    window.addEventListener("click", this.closeContextMenu);
    window.addEventListener("keydown", this.handleKeydown);
    window.addEventListener("keyup", this.handleKeyup);
    window.addEventListener("blur", this.clearKeyboardModifiers);
    window.addEventListener("pointermove", this.handleResizeMove);
    window.addEventListener("pointerup", this.stopResize);
    window.addEventListener("mousemove", this.handleResizeMove);
    window.addEventListener("mouseup", this.stopResize);
    window.addEventListener("resize", this.updateViewportMode);
    window.addEventListener("popstate", this.handleMobilePopState);
    this.updateViewportMode();
    this.initMobileHistory();
    this.status = "就绪";
  },
  beforeUnmount() {
    if (this.poller) clearInterval(this.poller);
    window.removeEventListener("click", this.closeContextMenu);
    window.removeEventListener("keydown", this.handleKeydown);
    window.removeEventListener("keyup", this.handleKeyup);
    window.removeEventListener("blur", this.clearKeyboardModifiers);
    window.removeEventListener("pointermove", this.handleResizeMove);
    window.removeEventListener("pointerup", this.stopResize);
    window.removeEventListener("mousemove", this.handleResizeMove);
    window.removeEventListener("mouseup", this.stopResize);
    window.removeEventListener("resize", this.updateViewportMode);
    window.removeEventListener("popstate", this.handleMobilePopState);
  },
  methods: {
    async api(path, options = {}) {
      const response = await fetch(path, {
        headers: { "Content-Type": "application/json" },
        ...options,
      });
      if (!response.ok) throw new Error(await response.text());
      return response.json();
    },
  },
});

if (window.ElementPlus) {
  app.use(window.ElementPlus);
}
for (const [name, component] of Object.entries(ElementIcons)) {
  app.component(name, component);
}
app.mount("#app");
