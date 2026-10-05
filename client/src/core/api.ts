// Thin fetch wrapper over the v3 REST API (server/openapi.yaml).
// Types come from `npm run gen:api`; no runtime client is generated.
import type { components } from './api.gen.ts'

export type Schemas = components['schemas']

export class ApiError extends Error {
  readonly status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

export class NekoApi {
  url = location.href.replace(/[?#].*$/, '').replace(/\/+$/, '')
  token = ''

  private headers(json: boolean): Record<string, string> {
    const h: Record<string, string> = {}
    if (json) h['Content-Type'] = 'application/json'
    if (this.token) h['Authorization'] = 'Bearer ' + this.token
    return h
  }

  async req<T = void>(method: string, path: string, body?: unknown): Promise<T> {
    const res = await fetch(this.url + '/api' + path, {
      method,
      credentials: 'include',
      headers: this.headers(body !== undefined),
      body: body === undefined ? undefined : JSON.stringify(body),
    })
    if (!res.ok) {
      const err = await res.json().catch(() => null)
      throw new ApiError(res.status, err?.message || res.statusText)
    }
    const text = await res.text()
    return (text ? JSON.parse(text) : undefined) as T
  }

  async blob(path: string): Promise<Blob> {
    const res = await fetch(this.url + '/api' + path, { credentials: 'include', headers: this.headers(false) })
    if (!res.ok) throw new ApiError(res.status, res.statusText)
    return res.blob()
  }

  // XHR because fetch cannot report upload progress
  upload(path: string, form: FormData, onProgress?: (p: { loaded: number; total: number }) => void): Promise<void> {
    return new Promise((resolve, reject) => {
      const xhr = new XMLHttpRequest()
      xhr.open('POST', this.url + '/api' + path)
      xhr.withCredentials = true
      for (const [k, v] of Object.entries(this.headers(false))) xhr.setRequestHeader(k, v)
      if (onProgress) xhr.upload.onprogress = (e) => onProgress({ loaded: e.loaded, total: e.total })
      xhr.onload = () => (xhr.status < 300 ? resolve() : reject(new ApiError(xhr.status, xhr.statusText)))
      xhr.onerror = () => reject(new ApiError(0, 'network error'))
      xhr.send(form)
    })
  }
}
