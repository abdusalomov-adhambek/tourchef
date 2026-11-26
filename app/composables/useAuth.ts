import { useApi, apiMessage } from './useApi'
import { useToast } from './useToast'

export type UserRole = 'client' | 'chef' | 'admin'
export type UserStatus = 'active' | 'pending' | 'suspended'

export interface AuthUser {
  id: string | number
  email: string
  full_name: string
  role: UserRole
  status: UserStatus
  phone?: string
  created_at?: string
}

export interface AuthResult {
  token: string
  user: AuthUser
}

export function homeForRole(role?: UserRole | null): string {
  if (role === 'admin') return '/admin'
  if (role === 'chef') return '/family'
  return '/client'
}

export function useAuth() {
  const api = useApi()
  const toast = useToast()
  const token = useCookie<string | null>('tc_token', { sameSite: 'lax' })
  const user = useState<AuthUser | null>('tc-user', () => null)
  const ready = useState<boolean>('tc-auth-ready', () => false)
  const pending = useState<boolean>('tc-auth-pending', () => false)

  async function fetchUser(force = false): Promise<AuthUser | null> {
    if (!token.value) {
      user.value = null
      ready.value = true
      return null
    }
    if (!force && (ready.value || pending.value)) return user.value
    pending.value = true
    try {
      const me = await api.get<AuthUser>('/auth/me')
      user.value = me && typeof me === 'object' && me.id ? me : null
      if (!user.value) token.value = null
      ready.value = true
      return user.value
    } catch (error: any) {
      if (error?.status === 401 || error?.status === 403) token.value = null
      user.value = null
      ready.value = true
      return null
    } finally {
      pending.value = false
    }
  }

  async function login(email: string, password: string): Promise<AuthUser> {
    const result = await api.post<AuthResult>('/auth/login', { email, password }) as AuthResult
    token.value = result.token
    user.value = result.user
    ready.value = true
    return result.user
  }

  async function register(payload: {
    email: string
    password: string
    full_name: string
    role: 'client' | 'chef'
    phone?: string
  }): Promise<AuthUser> {
    const result = await api.post<AuthResult>('/auth/register', payload) as AuthResult
    token.value = result.token
    user.value = result.user
    ready.value = true
    return result.user
  }

  function logout() {
    token.value = null
    user.value = null
    ready.value = true
  }

  function is(role: UserRole | UserRole[]): boolean {
    if (!user.value) return false
    return Array.isArray(role) ? role.includes(user.value.role) : user.value.role === role
  }

  function requireRole(role: UserRole): boolean {
    if (!user.value) return false
    return user.value.role === role
  }

  async function signOutAndGo(router?: any) {
    logout()
    toast.info('You have been signed out.')
    if (router) await router.push('/')
  }

  return {
    token,
    user,
    ready,
    pending,
    fetchUser,
    login,
    register,
    logout,
    signOutAndGo,
    is,
    requireRole,
    api
  }
}
