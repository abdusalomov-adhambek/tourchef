/**
 * Thin wrapper around Nuxt's $fetch for the TourChef REST API.
 * - Reads the JWT from the `tc_token` cookie and sends it as a Bearer header.
 * - Normalises backend `{ error: "..." }` responses into readable JS errors.
 * - Exposes `status` on errors so pages can react to 401/403/404.
 */
export interface ApiError extends Error {
  status: number
  data?: unknown
}

export interface ApiRequestOptions {
  method?: 'GET' | 'POST' | 'PATCH' | 'PUT' | 'DELETE'
  body?: unknown
  query?: Record<string, string | number | boolean | null | undefined>
  headers?: Record<string, string>
  /** Do not send the Authorization header even when a token is present. */
  anonymous?: boolean
}

function normaliseError(error: any): ApiError {
  const status =
    error?.statusCode ||
    error?.status ||
    error?.response?.status ||
    (typeof error?.status === 'string' ? Number(error.status) : 500) ||
    500

  const data = error?.data ?? error?.response?._data ?? error?.response?.data
  let message =
    (typeof data === 'string' ? data : null) ||
    data?.error ||
    data?.message ||
    error?.statusMessage ||
    error?.message ||
    'Something went wrong. Please try again.'

  if (Array.isArray(message)) message = message.join(', ')
  if (typeof message !== 'string') message = 'Something went wrong. Please try again.'

  const err = new Error(message) as ApiError
  err.status = Number(status) || 500
  err.data = data
  return err
}

export function useApi() {
  const config = useRuntimeConfig()
  const tokenCookie = useCookie<string | null>('tc_token', { sameSite: 'lax' })
  const base = String(config.public.apiBase || 'http://localhost:8080/api').replace(/\/+$/, '')

  async function request<T = any>(path: string, options: ApiRequestOptions = {}): Promise<T> {
    const { method = 'GET', body, query, headers = {}, anonymous = false } = options
    const finalHeaders: Record<string, string> = { Accept: 'application/json', ...headers }

    if (body instanceof FormData) {
      delete finalHeaders['Content-Type'] // browser o'zi boundary bilan ko'rsatadi
    } else if (body !== undefined) {
      finalHeaders['Content-Type'] = 'application/json'
    }
    const token = tokenCookie.value
    if (!anonymous && token) finalHeaders.Authorization = `Bearer ${token}`

    const target = /^https?:\/\//i.test(path) ? path : `${base}${path.startsWith('/') ? '' : '/'}${path}`

    try {
      return await $fetch<T>(target, {
        method,
        headers: finalHeaders,
        query: query as any,
        body: body as any
      })
    } catch (error) {
      throw normaliseError(error)
    }
  }

  return {
    base,
    request,
    get: <T = any>(path: string, query?: ApiRequestOptions['query']) =>
      request<T>(path, { method: 'GET', query }),
    post: <T = any>(path: string, body?: unknown) =>
      request<T>(path, { method: 'POST', body }),
    patch: <T = any>(path: string, body?: unknown) =>
      request<T>(path, { method: 'PATCH', body }),
    put: <T = any>(path: string, body?: unknown) =>
      request<T>(path, { method: 'PUT', body }),
    del: <T = any>(path: string) => request<T>(path, { method: 'DELETE' }),
    upload: <T = any>(path: string, file: File, field = 'file') => {
      const fd = new FormData()
      fd.append(field, file)
      return request<T>(path, { method: 'POST', body: fd })
    },
    uploadMany: <T = any>(path: string, files: File[], field = 'multiimage') => {
      const fd = new FormData()
      files.forEach((f) => fd.append(field, f))
      return request<T>(path, { method: 'POST', body: fd })
    }
  }
}

/** Readable message for any caught API error. */
export function apiMessage(error: unknown): string {
  return error instanceof Error ? error.message : 'Something went wrong. Please try again.'
}
