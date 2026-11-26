// Narxlar bazada so'mda saqlanadi, interfeysda esa dollarda ko'rsatiladi.
const UZS_PER_USD = 12_000

export function money(value: unknown): string {
  const n = Number(value || 0) / UZS_PER_USD
  if (!Number.isFinite(n)) return '$0'
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 0,
    maximumFractionDigits: 2
  }).format(n)
}

export function compactNumber(value: unknown): string {
  const n = Number(value || 0)
  if (!Number.isFinite(n)) return '0'
  if (n >= 1_000_000) return `${(n / 1_000_000).toFixed(1)}M`
  if (n >= 1_000) return `${(n / 1_000).toFixed(n >= 10_000 ? 0 : 1)}K`
  return String(Math.round(n))
}

export function duration(minutes: unknown): string {
  const n = Number(minutes || 0)
  if (!n) return '—'
  const h = Math.floor(n / 60)
  const m = n % 60
  if (!h) return `${m} min`
  return m ? `${h}h ${m}m` : `${h} hours`
}

export function dateLabel(value: unknown, withTime = true): string {
  if (!value) return '—'
  const d = new Date(String(value))
  if (Number.isNaN(d.getTime())) return String(value)
  const date = d.toLocaleDateString('en-GB', { day: 'numeric', month: 'short', year: 'numeric' })
  if (!withTime) return date
  const time = d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
  return `${date} · ${time}`
}

export function timeLabel(value: unknown): string {
  if (!value) return '—'
  const d = new Date(String(value))
  if (Number.isNaN(d.getTime())) return String(value)
  return d.toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit' })
}

export function initials(name: unknown): string {
  return String(name || 'TC')
    .trim()
    .split(/\s+/)
    .slice(0, 2)
    .map((part) => part.charAt(0).toUpperCase())
    .join('')
}

export function pillClass(status: unknown): string {
  const s = String(status || '').toLowerCase()
  if (['active', 'approved', 'confirmed', 'published', 'paid', 'paid_dev', 'completed', 'attended'].includes(s)) return 'pill ok'
  if (['pending', 'pending_moderation', 'scheduled', 'processing', 'draft', 'pending_payment'].includes(s)) return 'pill warn'
  if (['suspended', 'cancelled', 'rejected', 'failed', 'hidden', 'no_show', 'banned'].includes(s)) return 'pill bad'
  return 'pill neutral'
}

export function apPillClass(status: unknown): string {
  const s = String(status || '').toLowerCase()
  if (['active', 'approved', 'confirmed', 'published', 'paid', 'paid_dev', 'completed'].includes(s)) return 'apill g'
  if (['pending', 'pending_moderation', 'scheduled', 'processing', 'draft'].includes(s)) return 'apill w'
  if (['suspended', 'cancelled', 'rejected', 'failed', 'hidden'].includes(s)) return 'apill b'
  return 'apill n'
}

export function humanStatus(status: unknown): string {
  return String(status || '—').replace(/_/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
}

/** Booking records can nest their related session/masterclass in several ways. */
export function bookingMc(b: any): any {
  return b?.masterclass || b?.mc || b?.masterclass_data || {}
}
export function bookingSession(b: any): any {
  return b?.session || b?.masterclass_session || {}
}
export function bookingStart(b: any): unknown {
  return (
    bookingSession(b)?.starts_at ||
    b?.starts_at ||
    b?.session_date ||
    bookingSession(b)?.start_time ||
    b?.start_time
  )
}
export function bookingTotal(b: any): number {
  return Number(b?.total_price ?? b?.total ?? b?.amount ?? b?.price ?? 0)
}

/** Placeholder image used when the API record has no image URL. */
export const FALLBACK_IMAGE =
  'https://thumb.wikimedia.org/wikipedia/commons/thumb/0/04/Uzbek_national_cuisine.jpg/960px-Uzbek_national_cuisine.jpg'

export function imageOr(value: unknown, fallback = FALLBACK_IMAGE): string {
  return typeof value === 'string' && value.trim() ? value : fallback
}
