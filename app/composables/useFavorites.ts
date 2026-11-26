import { useApi } from './useApi'
import { useAuth } from './useAuth'
import { useToast } from './useToast'

/**
 * Favorites are masterclass IDs backed by GET/POST/DELETE /client/favorites.
 * The array is normalised because the backend may return either raw masterclasses
 * or favorite records with a `masterclass_id`.
 */
export function useFavorites() {
  const api = useApi()
  const auth = useAuth()
  const toast = useToast()
  const ids = useState<number[]>('tc-favorite-ids', () => [])
  const records = useState<any[]>('tc-favorite-records', () => [])
  const loaded = useState<boolean>('tc-favorites-loaded', () => false)

  function normalize(list: any[]): { ids: number[]; records: any[] } {
    const out: number[] = []
    for (const item of list || []) {
      const id = Number(
        item?.masterclass_id ?? item?.masterclass?.id ?? item?.id ?? item?.mc_id ?? 0
      )
      if (id && !out.includes(id)) out.push(id)
    }
    return { ids: out, records: list || [] }
  }

  async function load(force = false): Promise<number[]> {
    if (!auth.token.value) {
      ids.value = []
      records.value = []
      loaded.value = false
      return ids.value
    }
    if (loaded.value && !force) return ids.value
    try {
      const list = await api.get<any[]>('/client/favorites')
      const n = normalize(Array.isArray(list) ? list : [])
      ids.value = n.ids
      records.value = n.records
      loaded.value = true
    } catch {
      ids.value = []
      records.value = []
      loaded.value = true
    }
    return ids.value
  }

  function has(masterclassId: unknown): boolean {
    return ids.value.includes(Number(masterclassId))
  }

  async function toggle(masterclassId: unknown, redirectPath?: string): Promise<boolean> {
    const id = Number(masterclassId)
    if (!id) return false
    if (!auth.token.value) {
      const redirect = redirectPath || useRoute().fullPath
      await useRouter().push(`/login?redirect=${encodeURIComponent(redirect)}`)
      toast.info('Sign in to save favorites.')
      return false
    }
    await load()
    const record = records.value.find(
      (r) =>
        Number(r?.masterclass_id ?? r?.masterclass?.id ?? r?.id) === id
    )
    try {
      if (has(id)) {
        const favId = record?.id ?? id
        await api.del(`/client/favorites/${favId}`)
        ids.value = ids.value.filter((x) => x !== id)
        records.value = records.value.filter(
          (r) => Number(r?.masterclass_id ?? r?.masterclass?.id ?? r?.id) !== id
        )
        toast.success('Removed from favorites.')
      } else {
        const created = await api.post<any>('/client/favorites', { masterclass_id: id })
        ids.value = [...ids.value, id]
        records.value = [...records.value, created ?? { id, masterclass_id: id }]
        toast.success('Saved to favorites.')
      }
      return true
    } catch (error: any) {
      toast.error(error?.message || 'Could not update favorites.')
      return false
    }
  }

  function reset() {
    ids.value = []
    records.value = []
    loaded.value = false
  }

  return { ids, records, loaded, load, has, toggle, reset }
}
