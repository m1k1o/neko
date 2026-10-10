import type { NekoApp } from './app'
import { toast } from './dialogs'

// REST call for user actions: failures become a toast, the result says whether it worked.
// Session ids are built from the login name on some providers, so they go through encodeURIComponent
// wherever they are part of a path: otherwise a name like `x/../../logout?` redirects the request.
export const api = ({ client }: NekoApp, method: string, path: string, body?: unknown) =>
  client.api.req(method, path, body).then(
    () => true,
    (err) => (toast(err.message), false),
  )
