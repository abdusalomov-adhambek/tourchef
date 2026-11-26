/**
 * Scroll reveal: `.reveal` elementlar ko'rinishga kirganda `.is-visible` qo'shadi.
 * JS o'chirilgan bo'lsa, `js-reveal` class qo'shilmaydi va kontent ko'rinib turadi.
 */
export default defineNuxtPlugin(() => {
  if (typeof window === 'undefined' || !('IntersectionObserver' in window)) return

  const observer = new IntersectionObserver(
    (entries) => {
      entries.forEach((entry) => {
        if (!entry.isIntersecting) return
        entry.target.classList.add('is-visible')
        observer.unobserve(entry.target)
      })
    },
    { threshold: 0.08, rootMargin: '0px 0px -40px 0px' }
  )

  function observeAll() {
    document.querySelectorAll('.reveal:not(.is-visible)').forEach((el) => observer.observe(el))
  }

  requestAnimationFrame(observeAll)
  window.addEventListener('load', observeAll)

  // Ma'lumotlar API dan kelganda yangi kartalar paydo bo'ladi — ularni ham kuzatay.
  const mutations = new MutationObserver(() => observeAll())
  mutations.observe(document.body, { childList: true, subtree: true })
})
