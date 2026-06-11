import { resolve } from "node:path"
import { remarkDocsLiveDemo } from "./app/lib/remarkDocsLiveDemo"

const docsLiveDemoRemarkPlugin = resolve("./app/lib/remarkDocsLiveDemo.ts").replace(/\\/g, "/")

const adminIconClientBundle = [
  "lucide:activity",
  "lucide:align-center",
  "lucide:align-left",
  "lucide:align-right",
  "lucide:arrow-down",
  "lucide:arrow-left",
  "lucide:arrow-right",
  "lucide:arrow-up",
  "lucide:badge",
  "lucide:badge-help",
  "lucide:ban",
  "lucide:blocks",
  "lucide:bold",
  "lucide:book-open",
  "lucide:bookmark",
  "lucide:building",
  "lucide:building-2",
  "lucide:calendar",
  "lucide:check",
  "lucide:chevron-down",
  "lucide:chevron-left",
  "lucide:chevron-right",
  "lucide:chevron-up",
  "lucide:chevrons-down",
  "lucide:circle-alert",
  "lucide:clock-3",
  "lucide:code",
  "lucide:code-2",
  "lucide:compass",
  "lucide:copy",
  "lucide:database",
  "lucide:download",
  "lucide:edit-3",
  "lucide:external-link",
  "lucide:file",
  "lucide:file-archive",
  "lucide:file-audio",
  "lucide:file-code",
  "lucide:file-image",
  "lucide:file-text",
  "lucide:file-video",
  "lucide:files",
  "lucide:flip-horizontal-2",
  "lucide:flip-vertical-2",
  "lucide:folder",
  "lucide:folder-open",
  "lucide:fullscreen",
  "lucide:hash",
  "lucide:hard-drive",
  "lucide:heart",
  "lucide:history",
  "lucide:id-card",
  "lucide:image-up",
  "lucide:info",
  "lucide:italic",
  "lucide:key-round",
  "lucide:layout-dashboard",
  "lucide:link",
  "lucide:list-checks",
  "lucide:list-filter",
  "lucide:list-video",
  "lucide:loader-circle",
  "lucide:lock-keyhole",
  "lucide:log-in",
  "lucide:log-out",
  "lucide:mail",
  "lucide:maximize",
  "lucide:message-circle",
  "lucide:message-square-off",
  "lucide:message-square-text",
  "lucide:minimize",
  "lucide:minimize-2",
  "lucide:monitor",
  "lucide:monitor-check",
  "lucide:moon",
  "lucide:more-horizontal",
  "lucide:palette",
  "lucide:panel-left",
  "lucide:panel-right",
  "lucide:panel-right-close",
  "lucide:panel-right-open",
  "lucide:pause",
  "lucide:play",
  "lucide:play-square",
  "lucide:plus",
  "lucide:refresh-cw",
  "lucide:repeat",
  "lucide:rocket",
  "lucide:route",
  "lucide:rotate-ccw",
  "lucide:rotate-cw",
  "lucide:save",
  "lucide:scroll-text",
  "lucide:search",
  "lucide:send",
  "lucide:server",
  "lucide:settings",
  "lucide:share-2",
  "lucide:shield-check",
  "lucide:shield-plus",
  "lucide:sliders-horizontal",
  "lucide:sparkles",
  "lucide:strikethrough",
  "lucide:sun",
  "lucide:ticket",
  "lucide:trash-2",
  "lucide:type",
  "lucide:underline",
  "lucide:upload",
  "lucide:user",
  "lucide:user-check",
  "lucide:user-plus",
  "lucide:users",
  "lucide:video",
  "lucide:video-off",
  "lucide:volume-2",
  "lucide:volume-x",
  "lucide:x"
]

export default defineNuxtConfig({
  compatibilityDate: "2026-06-03",
  devtools: { enabled: false },
  ssr: false,
  srcDir: "app",

  app: {
    baseURL: process.env.NUXT_APP_BASE_URL || "/admin/"
  },

  modules: [
    "@nuxt/content",
    "@nuxt/icon",
    "@pinia/nuxt",
    "@nuxtjs/i18n"
  ],

  css: [
    "~/assets/css/tokens.css",
    "~/assets/css/main.css"
  ],

  components: [
    {
      path: "~/components",
      pathPrefix: false
    }
  ],

  runtimeConfig: {
    public: {
      adminBaseURL: process.env.NUXT_APP_BASE_URL || "/admin/",
      apiBaseURL: process.env.NUXT_PUBLIC_API_BASE_URL || "",
      apiMock: process.env.NUXT_PUBLIC_API_MOCK === "true",
      showDemoTodo: process.env.NUXT_PUBLIC_SHOW_DEMO_TODO === "true"
    }
  },

  routeRules: {
    "/docs": { prerender: true },
    "/docs/**": { prerender: true }
  },

  content: {
    build: {
      markdown: {
        remarkPlugins: {
          "remark-docs-live-demo": {
            instance: remarkDocsLiveDemo,
            src: docsLiveDemoRemarkPlugin
          }
        }
      }
    },
    experimental: {
      nativeSqlite: true
    }
  },

  icon: {
    provider: "server",
    fallbackToApi: false,
    serverBundle: {
      collections: ["lucide"]
    },
    clientBundle: {
      icons: adminIconClientBundle,
      scan: true
    }
  },

  i18n: {
    defaultLocale: "zh-CN",
    strategy: "no_prefix",
    detectBrowserLanguage: false,
    locales: [
      { code: "zh-CN", language: "zh-CN", name: "简体中文", file: "zh-CN.json" },
      { code: "en", language: "en-US", name: "English", file: "en.json" },
      { code: "ja", language: "ja-JP", name: "日本語", file: "ja.json" }
    ]
  },

  vue: {
    compilerOptions: {
      isCustomElement: (tag) => tag.startsWith("md-")
    }
  }
})
