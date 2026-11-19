// Unit tests of the provider and its hooks: `npm test` (vitest, react-dom/server).
import { expect, test, vi } from 'vitest'
import { renderToString } from 'react-dom/server'
import { fakeBrowser, fakeTransport } from '@/test/browser'

vi.mock('sonner', () => ({ toast: {} }))
fakeBrowser()
const { createNekoApp } = await import('./neko.ts')
const { NekoProvider, useNeko, useClient, useApp } = await import('./provider.tsx')

const neko = createNekoApp({ transport: fakeTransport() })
const other = createNekoApp({ transport: fakeTransport() })
function Probe() {
  const n = useNeko()
  const client = useClient()
  const app = useApp()
  return <i>{[n === neko, client === neko.client, app === neko.app, client === other.client].join()}</i>
}

test('useClient() and the other hooks return the instance of the <NekoProvider> above', () => {
  expect(
    renderToString(
      <NekoProvider neko={neko}>
        <Probe />
      </NekoProvider>,
    ),
  ).toBe('<i>true,true,true,false</i>')
})

test('outside a <NekoProvider> the hooks throw, naming it', () => {
  expect(() => renderToString(<Probe />)).toThrow('useNeko: no <NekoProvider> above this component')
})
