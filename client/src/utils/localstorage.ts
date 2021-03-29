export function set<T extends string | number | boolean>(key: string, val: T) {
  switch (typeof val) {
    case 'number':
      localStorage.setItem(key, val.toString())
      break
    case 'string':
      localStorage.setItem(key, val)
      break
    case 'boolean':
      localStorage.setItem(key, val ? '1' : '0')
      break
  }
}

export function get<T extends string | number | boolean>(key: string, def: T): T {
  const store = localStorage.getItem(key)
  if (store) {
    switch (typeof def) {
      case 'number':
        return parseInt(store) as T
      case 'string':
        return store as T
      case 'boolean':
        return (store === '1') as T
      default:
        return def
    }
  }

  return def
}
