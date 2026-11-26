import { homeForRole, type UserRole } from '~/composables/useAuth'

const RULES: { prefix: string; role: UserRole }[] = [
  { prefix: '/client', role: 'client' },
  { prefix: '/family', role: 'chef' },
  { prefix: '/admin', role: 'admin' }
]

export default defineNuxtRouteMiddleware(async (to) => {
  const rule = RULES.find((r) => to.path === r.prefix || to.path.startsWith(`${r.prefix}/`))
  if (!rule) return

  const auth = useAuth()
  const current = await auth.fetchUser(true)

  if (!auth.token.value || !current) {
    return navigateTo(`/login?redirect=${encodeURIComponent(to.fullPath)}`)
  }

  if (current.role !== rule.role) {
    return navigateTo(homeForRole(current.role))
  }
})
