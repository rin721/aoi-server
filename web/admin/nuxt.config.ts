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
      apiBaseURL: process.env.NUXT_PUBLIC_API_BASE_URL || ""
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
