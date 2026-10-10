// per-viewer settings, stored like the legacy client did (same keys, '1'/'0' booleans)

export function get<T extends string | number | boolean>(key: string, def: T): T {
  try {
    const v = localStorage.getItem(key)
    if (!v) return def
    if (typeof def === 'boolean') return (v === '1') as T
    if (typeof def === 'number') return (isNaN(parseInt(v)) ? def : parseInt(v)) as T
    return v as T
  } catch {
    return def
  }
}

export function set(key: string, val: string | number | boolean) {
  try {
    localStorage.setItem(key, typeof val === 'boolean' ? (val ? '1' : '0') : String(val))
  } catch {}
}

export const remember = set
