export function extractErrorMessage(error: unknown): string {
  if (!error) return ''
  if (typeof error === 'string') return error
  if (error instanceof Error) return error.message
  if (typeof error === 'object') {
    const data = (error as { data?: Record<string, unknown>, message?: string }).data
    if (data) {
      if (typeof data.Error === 'string') return data.Error
      if (typeof data.error === 'string') return data.error
      if (typeof data.message === 'string') return data.message
    }
    if (typeof (error as { message?: string }).message === 'string') {
      return (error as { message: string }).message
    }
  }
  return ''
}

export function formatDate(value?: string | null): string {
  if (!value) return 'Unknown'

  const normalized = value.includes('T') ? value : value.replace(' ', 'T')
  const parsed = new Date(normalized)

  if (Number.isNaN(parsed.getTime())) return value

  return parsed.toLocaleString()
}

export function toDataUrl(raw?: string | null, mime = 'image/png'): string | undefined {
  if (!raw) return undefined

  const trimmed = raw.trim()
  if (!trimmed) return undefined
  if (trimmed.startsWith('data:')) return trimmed

  return `data:${mime};base64,${trimmed}`
}

export function initialsFromName(name: string): string {
  if (!name) return '??'

  return name
    .split(/\s+/)
    .filter(Boolean)
    .slice(0, 2)
    .map(part => part[0]?.toUpperCase() ?? '')
    .join('') || '??'
}

export function buildFullName(first?: string | null, last?: string | null): string {
  return `${(first ?? '').trim()} ${(last ?? '').trim()}`.trim()
}

export function toSqlDateTime(localValue: string): string {
  if (!localValue) return ''

  if (localValue.includes('T')) {
    const [date, time] = localValue.split('T')
    const safeTime = time ?? ''
    if (!safeTime.includes(':')) {
      return `${date} ${safeTime}:00`
    }
    return `${date} ${safeTime.length === 5 ? `${safeTime}:00` : safeTime}`.replace('Z', '')
  }

  return localValue
}

export function fileToBase64(file: File): Promise<string> {
  return new Promise<string>((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      const result = typeof reader.result === 'string' ? reader.result : ''
      const payload = result.includes(',') ? result.split(',')[1] : result
      if (payload) {
        resolve(payload)
      } else {
        reject(new Error('empty-file'))
      }
    }
    reader.onerror = () => reject(reader.error ?? new Error('read-error'))
    reader.readAsDataURL(file)
  })
}
