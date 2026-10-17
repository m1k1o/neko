// Minimal typed event emitter (on/off/once/emit).
export class Emitter<E extends { [K in keyof E]: (...args: any[]) => void }> {
  private handlers = new Map<keyof E, Set<Function>>()

  on<K extends keyof E>(event: K, fn: E[K]) {
    if (!this.handlers.has(event)) this.handlers.set(event, new Set())
    this.handlers.get(event)!.add(fn)
    return () => this.off(event, fn)
  }

  off<K extends keyof E>(event: K, fn: E[K]) {
    this.handlers.get(event)?.delete(fn)
  }

  once<K extends keyof E>(event: K, fn: E[K]) {
    const wrap = ((...args: any[]) => (this.off(event, wrap), fn(...args))) as E[K]
    return this.on(event, wrap)
  }

  emit<K extends keyof E>(event: K, ...args: Parameters<E[K]>) {
    for (const fn of [...(this.handlers.get(event) ?? [])]) fn(...args)
  }
}
