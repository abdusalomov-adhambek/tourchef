<script setup lang="ts">
const discover = [
  { to: '/masterclasses', label: 'Masterclasses' },
  { to: '/reels', label: 'Reels' },
  { to: '/restaurants', label: 'Restaurants' },
  { to: '/dishes', label: 'Uzbek dishes' }
]
const company = [
  { to: '/about', label: 'About TourChef' },
  { to: '/register?role=family', label: 'Become a family host' },
  { to: '/contact', label: 'Contact' },
  { to: '/faq', label: 'FAQ' }
]
const legal = [
  { to: '/terms', label: 'Terms and Conditions' },
  { to: '/privacy', label: 'Privacy Policy' },
  { to: '/faq#cancellation', label: 'Cancellation policy' }
]
</script>

<template>
  <footer class="site-footer">
    <div class="container">
      <div class="footer-grid">
        <div>
          <NuxtLink to="/" class="brand">
            <img src="/logo.svg" alt="TourChef" class="logo-img" width="138" height="46" />
          </NuxtLink>
          <p style="margin-top: 14px; max-width: 34ch; font-size: 14px">
            Don't Just Taste Uzbekistan, Be Its Guest. Authentic kitchens, real stories,
            unforgettable experiences.
          </p>
        </div>
        <div>
          <h4>Discover</h4>
          <ul>
            <li v-for="l in discover" :key="l.to"><NuxtLink :to="l.to">{{ l.label }}</NuxtLink></li>
          </ul>
        </div>
        <div>
          <h4>Company</h4>
          <ul>
            <li v-for="l in company" :key="l.to"><NuxtLink :to="l.to">{{ l.label }}</NuxtLink></li>
          </ul>
        </div>
        <div>
          <h4>Legal</h4>
          <ul>
            <li v-for="l in legal" :key="l.to"><NuxtLink :to="l.to">{{ l.label }}</NuxtLink></li>
            <li>
              <a href="https://www.instagram.com/" target="_blank" rel="noopener">Instagram</a>
              ·
              <a href="https://t.me/" target="_blank" rel="noopener">Telegram</a>
            </li>
          </ul>
        </div>
      </div>
      <div class="footer-bottom">
        <span>© 2026 TourChef. All rights reserved.</span>
        <span>Made with care for Uzbek culinary culture.</span>
      </div>
    </div>
  </footer>
</template>
