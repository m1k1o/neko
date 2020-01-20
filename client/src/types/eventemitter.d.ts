declare module 'eventemitter3' {
  type Arguments<T> = [T] extends [(...args: infer U) => any] ? U : [T] extends [void] ? [] : [T]

  class TypedEventEmitter<Events> {
    addListener<E extends keyof Events>(event: E, listener: Events[E]): this
    on<E extends keyof Events>(event: E, listener: Events[E]): this
    once<E extends keyof Events>(event: E, listener: Events[E]): this
    prependListener<E extends keyof Events>(event: E, listener: Events[E]): this
    prependOnceListener<E extends keyof Events>(event: E, listener: Events[E]): this

    off<E extends keyof Events>(event: E, listener: Events[E]): this
    removeAllListeners<E extends keyof Events>(event?: E): this
    removeListener<E extends keyof Events>(event: E, listener: Events[E]): this

    emit<E extends keyof Events>(event: E, ...args: Arguments<Events[E]>): boolean
    eventNames(): (keyof Events)[]
    listeners<E extends keyof Events>(event: E): Function[]
    listenerCount<E extends keyof Events>(event: E): number

    getMaxListeners(): number
    setMaxListeners(maxListeners: number): this
  }

  export = TypedEventEmitter
}
