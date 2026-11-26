export type ToastKind = 'ok' | 'err' | 'info'

export interface ToastItem {
  id: number
  kind: ToastKind
  message: string
}

let toastSeq = 0

export function useToast() {
  const toasts = useState<ToastItem[]>('tc-toasts', () => [])

  function dismiss(id: number) {
    toasts.value = toasts.value.filter((t) => t.id !== id)
  }

  function push(message: string, kind: ToastKind = 'ok') {
    const id = ++toastSeq
    toasts.value = [...toasts.value, { id, kind, message }].slice(-4)
    if (import.meta.client && typeof setTimeout !== 'undefined') {
      setTimeout(() => dismiss(id), 4200)
    }
    return id
  }

  return {
    toasts,
    dismiss,
    push,
    success: (message: string) => push(message, 'ok'),
    error: (message: string) => push(message, 'err'),
    info: (message: string) => push(message, 'info')
  }
}
