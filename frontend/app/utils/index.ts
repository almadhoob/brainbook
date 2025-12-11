export function randomInt(min: number, max: number): number {
  return Math.floor(Math.random() * (max - min + 1)) + min
}

export function randomFrom<T>(array: T[]): T {
  return array[Math.floor(Math.random() * array.length)]!
}

export function normalizeAvatar(raw?: string | null): string | undefined {
  if (!raw) return undefined
  if (raw.startsWith('data:')) return raw
  return `data:image/png;base64,${raw}`
}
