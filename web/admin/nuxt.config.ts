export default defineNuxtConfig({
  compatibilityDate: "2026-06-03",
  devtools: { enabled: false },
  ssr: false,
  srcDir: "app",

  app: {
    baseURL: process.env.NUXT_APP_BASE_URL || "/admin/"
  },

  modules: [
    "@nuxt/icon",
    "@pinia/nuxt"
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
      apiBaseURL: process.env.NUXT_PUBLIC_API_BASE_URL || "",
      showDemoTodo: process.env.NUXT_PUBLIC_SHOW_DEMO_TODO === "true"
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

  vue: {
    compilerOptions: {
      isCustomElement: (tag) => tag.startsWith("md-")
    }
  }
})
