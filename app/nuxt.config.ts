// https://nuxt.com/docs/api/configuration/nuxt-config
export default defineNuxtConfig({
  compatibilityDate: '2025-01-01',
  devtools: { enabled: false },
  typescript: {
    strict: false
  },
  runtimeConfig: {
    // SSR va browser bir xil origin ustidan /api ga murojaat qiladi;
    // Nitro so'rovni backendga proxy qiladi (lokal: localhost, Docker: backend).
    apiProxyTarget: process.env.API_PROXY_TARGET || 'http://localhost:8080',
    public: {
      apiBase: process.env.API_BASE_URL || '/api'
    }
  },
  routeRules: {
    '/api/**': {
      proxy: `${process.env.API_PROXY_TARGET || 'http://localhost:8080'}/api/**`
    },
    '/uploads/**': {
      proxy: `${process.env.API_PROXY_TARGET || 'http://localhost:8080'}/uploads/**`
    }
  },
  css: ['~/assets/css/main.css'],
  app: {
    head: {
      title: 'TourChef — Don\'t Just Taste Uzbekistan, Be Its Guest.',
      htmlAttrs: { lang: 'en' },
      meta: [
        { charset: 'utf-8' },
        { name: 'viewport', content: 'width=device-width, initial-scale=1' },
        {
          name: 'description',
          content:
            'Authentic Uzbek culinary experiences, cooking masterclasses and restaurant stories. Discover, book and experience the flavors of Uzbekistan.'
        },
        { name: 'theme-color', content: '#F8F5EF' }
      ],
      script: [
        {
          innerHTML: "document.documentElement.classList.add('js-reveal')",
          tagPosition: 'head'
        }
      ],
      link: [
        { rel: 'icon', type: 'image/svg+xml', href: '/logo.svg' },
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' },
        {
          rel: 'stylesheet',
          href: 'https://fonts.googleapis.com/css2?family=DM+Sans:wght@400;500;600;700&family=Playfair+Display:ital,wght@0,500;0,600;0,700;1,600&display=swap'
        }
      ]
    }
  }
})
