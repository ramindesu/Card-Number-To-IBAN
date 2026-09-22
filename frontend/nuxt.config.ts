export default defineNuxtConfig({
  compatibilityDate: '2024-11-01',
  devtools: { enabled: true },
  modules: ['@nuxtjs/tailwindcss'],
  css: ['~/assets/css/main.css'],
  runtimeConfig: {
    public: {
      apiBaseUrl: process.env.NUXT_PUBLIC_API_BASE_URL || '/api/v1',
    },
  },
  nitro: {
    devProxy: {
      '/api/': {
        target: process.env.NUXT_API_PROXY_TARGET || 'http://localhost:8080',
        changeOrigin: true,
      },
    },
  },
  app: {
    head: {
      htmlAttrs: { lang: 'fa', dir: 'rtl' },
      title: 'شبایار | تبدیل شماره کارت به شبا',
      meta: [
        {
          name: 'description',
          content: 'تبدیل سریع و امن شماره کارت به شماره شبا',
        },
      ],
    },
  },
})
