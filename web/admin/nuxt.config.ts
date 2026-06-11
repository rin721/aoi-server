import { resolve } from "node:path"
import { remarkDocsLiveDemo } from "./app/lib/remarkDocsLiveDemo"

const docsLiveDemoRemarkPlugin = resolve("./app/lib/remarkDocsLiveDemo.ts").replace(/\\/g, "/")

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
