import { chromium } from 'playwright'
// Two-user smoke test against a running v3 server (NEKO_LEGACY=false, multiuser provider):
//   NEKO_URL=https://host/ NEKO_ADMIN_PASSWORD=admin NEKO_USER_PASSWORD=neko npm run test:e2e
// Screenshots land in ./e2e-out. Needs a KDE/X desktop that opens KRunner on Alt+F2 for the typing step.
import fs from 'node:fs'
const out = 'e2e-out'
fs.mkdirSync(out, { recursive: true })
const URL = process.env.NEKO_URL || 'http://localhost:8080/'
const ADMIN = process.env.NEKO_ADMIN_PASSWORD || 'admin'
const USER = process.env.NEKO_USER_PASSWORD || 'neko'
// NEKO_CDP=http://127.0.0.1:9222 runs the same steps in a real Chromium-based browser (Brave, Chrome, Edge)
// started with --remote-debugging-port and the three flags below
const browser = process.env.NEKO_CDP
  ? await chromium.connectOverCDP(process.env.NEKO_CDP)
  : await chromium.launch({
      args: [
        '--autoplay-policy=no-user-gesture-required',
        '--use-fake-ui-for-media-stream',
        '--use-fake-device-for-media-stream',
      ],
    })
const errors = []
let A, B
let fails = 0
const log = (s) => console.log(new Date().toISOString().slice(11, 19), s)
async function step(name, fn) {
  try {
    await fn()
    log('PASS ' + name)
  } catch (e) {
    fails++
    log('FAIL ' + name + ' :: ' + e.message.split('\n')[0])
    // evidence for flaky failures: both users' screens and member lists
    const slug = name.replace(/\W+/g, '-')
    for (const [who, p] of [
      ['alice', A],
      ['bob', B],
    ]) {
      if (!p) continue
      await p.screenshot({ path: `${out}/fail-${slug}-${who}.png` }).catch(() => {})
      const members = await p
        .$$eval('.members-list .member', (els) => els.map((e) => e.getAttribute('aria-label') || 'self'))
        .catch(() => '?')
      log(`  ${who} members: ${JSON.stringify(members)}`)
    }
  }
}
async function user(name, password) {
  const ctx = await browser.newContext({
    ignoreHTTPSErrors: true,
    viewport: { width: 1400, height: 850 },
    permissions: ['clipboard-read', 'clipboard-write', 'microphone'],
  })
  const page = await ctx.newPage()
  page.on('pageerror', (e) => errors.push(`${name} pageerror: ${e.message}`))
  page.on(
    'console',
    (m) =>
      m.type() === 'error' &&
      // the outage steps make the socket fail on purpose; Brave reports that without the ERR_ text
      !/401|STUN|ERR_INTERNET_DISCONNECTED|WebSocket connection to .* failed/.test(m.text()) &&
      errors.push(`${name} console: ${m.text()}`),
  )
  await page.goto(URL)
  await page.fill('input[placeholder="Enter your display name"]', name)
  await page.fill('input[type=password]', password)
  await page.click('button[type=submit]')
  await page.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
  await page.waitForFunction(
    () => {
      const v = document.querySelector('video')
      return v && !v.paused && v.readyState >= 2
    },
    null,
    { timeout: 30000 },
  )
  return page
}
const api = (p, method, path, body) =>
  p.evaluate(
    async ([method, path, body]) => {
      const headers = {
        'Content-Type': 'application/json',
        Authorization: 'Bearer ' + localStorage.getItem('neko_session'),
      }
      const r = await fetch('api' + path, { method, headers, body: body && JSON.stringify(body) })
      if (!r.ok) throw new Error(r.status + ' ' + (await r.text()))
      return r.text()
    },
    [method, path, body],
  )
const openSide = async (p) => {
  if (!(await p.locator('aside.neko-menu').count())) await p.click('.header .fa-bars.toggle')
  await p.waitForSelector('aside.neko-menu', { timeout: 3000 })
}
const playing = (p) =>
  p.evaluate(() => {
    const v = document.querySelector('video')
    return `${v.videoWidth}x${v.videoHeight} paused=${v.paused} muted=${v.muted}`
  })

await step('alice (admin) logs in, video plays', async () => {
  A = await user('alice', ADMIN)
  log('  alice video ' + (await playing(A)))
})
await step('bob (user) logs in, video plays', async () => {
  B = await user('bob', USER)
  log('  bob video ' + (await playing(B)))
})
await step('bob sees pre-existing alice in members', async () => {
  await B.waitForFunction(() => document.querySelectorAll('.members-list .member').length === 2, null, {
    timeout: 5000,
  })
})
await step('alice sees 2 members, admin shield', async () => {
  await A.waitForFunction(() => document.querySelectorAll('.members-list .member').length === 2, null, {
    timeout: 5000,
  })
  await A.waitForSelector('.room-settings .fa-shield-alt', { timeout: 5000 })
})
await step('side panel: bob joined event line', async () => {
  await openSide(A)
  await A.waitForSelector('aside.neko-menu .chat', { timeout: 3000 })
  await A.waitForFunction(
    () => [...document.querySelectorAll('.chat-history .event')].some((e) => /bob\s+connected/.test(e.textContent)),
    null,
    { timeout: 5000 },
  )
})
await step('chat: bob -> alice', async () => {
  await openSide(B)
  await B.fill('.chat-send textarea', 'hi from bob')
  await B.press('.chat-send textarea', 'Enter')
  await A.waitForFunction(
    () =>
      [...document.querySelectorAll('.chat-history .message .content-body')].some(
        (e) => e.textContent === 'hi from bob',
      ),
    null,
    { timeout: 5000 },
  )
})
await step('viewer typing in chat keeps the focus when the mouse crosses the video', async () => {
  await B.click('.chat-send textarea')
  await B.keyboard.type('hello ')
  const box = await B.locator('.player-container').boundingBox()
  await B.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
  await B.mouse.move(box.x + box.width / 2 + 5, box.y + box.height / 2 + 5)
  await B.keyboard.type('world')
  const v = await B.inputValue('.chat-send textarea')
  if (v !== 'hello world') throw new Error('chat box has ' + JSON.stringify(v))
  await B.fill('.chat-send textarea', '')
})
await step('bob takes control, alice sees host + event', async () => {
  await B.click('.neko-controls .fa-keyboard')
  await A.waitForFunction(() => document.querySelector('.members-list .member.host:not(.self)'), null, {
    timeout: 5000,
  })
  await A.waitForFunction(
    () =>
      [...document.querySelectorAll('.chat-history .event')].some((e) => /bob\s+took the controls/.test(e.textContent)),
    null,
    { timeout: 5000 },
  )
})
await step('alice: right-click bob -> Force Release', async () => {
  await A.click('.members-list .member:not(.self)', { button: 'right' })
  await A.click('.context >> text=Force Release Controls')
  await B.waitForFunction(() => !document.querySelector('.members-list .member.self.host'), null, { timeout: 5000 })
})
await step('alice sends emote, bob sees animation', async () => {
  await A.click('.emotes-bar .fa-grin-beam')
  await A.click('.context .emote.wave')
  await B.waitForSelector('.emotes .emote-container', { timeout: 3000 })
})
await step('files: alice uploads, the list shows the file', async () => {
  if (!(await A.locator('.tabs-container >> text=Files').count()))
    throw new Error('no Files tab: start the server with NEKO_FILETRANSFER_ENABLED=true')
  await A.click('.tabs-container >> text=Files')
  await A.setInputFiles('.upload-area input[type=file]', {
    name: 'react-e2e.txt',
    mimeType: 'text/plain',
    buffer: Buffer.from('hello'),
  })
  await A.waitForSelector('.transfers-list-item .fa-check', { timeout: 10000 })
  await A.waitForSelector('.files-list .file-name >> text=react-e2e.txt', { timeout: 5000 })
})
await step('files: download link works', async () => {
  const href = await A.getAttribute('.files-list-item:has-text("react-e2e.txt") a[download]', 'href')
  const r = await A.request.get(href)
  if ((await r.text()) !== 'hello') throw new Error('download body mismatch: ' + r.status())
  await A.click('.files-list-item:has-text("react-e2e.txt") .fa-trash')
  await A.click('.neko-dialog .confirm')
  await A.waitForSelector('.files-list .file-name >> text=react-e2e.txt', { state: 'detached', timeout: 5000 })
})
await step('markdown: bob -> alice renders safely', async () => {
  await B.click('.tabs-container >> text=Chat')
  await B.fill('.chat-send textarea', '**bold** ||secret|| :smile: https://example.com [x](javascript:alert(1))')
  await B.press('.chat-send textarea', 'Enter')
  await A.click('.tabs-container >> text=Chat')
  const last = A.locator('.chat-history .message .content-body').last()
  await last.locator('strong >> text=bold').waitFor({ timeout: 5000 })
  await last.locator('.spoiler:not(.active)').click()
  await last.locator('.spoiler.active').waitFor({ timeout: 2000 })
  await last.locator('.emoji[data-emoji="smile"]').waitFor({ timeout: 5000 })
  if ((await last.locator('a').count()) !== 1) throw new Error('expected exactly one (safe) link')
  if ((await last.locator('a').getAttribute('href')) !== 'https://example.com/') throw new Error('bad href')
})
await step('narrow screen: the member menu opens from a chat author', async () => {
  const ctx = await browser.newContext({ ignoreHTTPSErrors: true, viewport: { width: 700, height: 900 } })
  const N = await ctx.newPage()
  await N.goto(URL + '?usr=nina&pwd=' + USER + '&show_side=1')
  await N.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
  await B.fill('.chat-send textarea', 'hi nina')
  await B.press('.chat-send textarea', 'Enter')
  await N.waitForSelector('.chat-history .message .author', { timeout: 5000 })
  await N.click('.chat-history .message .author', { button: 'right' })
  await N.waitForSelector('.context[role=menu]', { state: 'visible', timeout: 3000 })
  await ctx.close()
})
await step('muting the sound keeps the desktop usable', async () => {
  await A.click('.neko-controls .volume i')
  await A.waitForFunction(() => document.querySelector('video').muted, null, { timeout: 3000 })
  if (await A.locator('.player-overlay').count()) throw new Error('unmute overlay covers the video')
  await A.click('.neko-controls .volume i')
  await A.waitForFunction(() => !document.querySelector('video').muted, null, { timeout: 3000 })
})
await step('emoji picker inserts :name:', async () => {
  await A.click('.chat-send .emoji-menu')
  await A.fill('.neko-emoji .search input', 'thumbsup')
  await A.click('.neko-emoji .list .emoji >> nth=0')
  const v = await A.inputValue('.chat-send textarea')
  if (!/^:[^:]+:$/.test(v)) throw new Error('textarea is ' + JSON.stringify(v))
  await A.fill('.chat-send textarea', '')
})
await step('language switch (de) and back', async () => {
  await A.selectOption('.room-settings select', 'de')
  await A.waitForSelector('.tabs-container >> text=Einstellungen', { timeout: 3000 })
  await A.selectOption('.room-settings select', 'en')
  await A.waitForSelector('.tabs-container >> text=Settings', { timeout: 3000 })
})
await step('admin locks controls, bob sees lock + event', async () => {
  await A.click('.header .fa-mouse')
  await B.waitForSelector('.header .fa-mouse.locked', { timeout: 5000 })
  await B.waitForFunction(
    () =>
      [...document.querySelectorAll('.chat-history .event')].some((e) => /alice\s+locked controls/.test(e.textContent)),
    null,
    { timeout: 5000 },
  )
  await A.click('.header .fa-mouse')
  await B.waitForSelector('.header .fa-mouse:not(.locked)', { timeout: 5000 })
})
await step('admin mutes bob: input gone, event line; then unmute', async () => {
  await A.click('.members-list .member:not(.self)', { button: 'right' })
  await A.click('.context >> text=Mute')
  await A.click('.neko-dialog .confirm')
  await B.waitForSelector('.chat-send textarea', { state: 'detached', timeout: 5000 })
  await A.waitForFunction(
    () => [...document.querySelectorAll('.chat-history .event')].some((e) => /You\s+muted bob/.test(e.textContent)),
    null,
    { timeout: 5000 },
  )
  await A.click('.members-list .member:not(.self)', { button: 'right' })
  await A.click('.context >> text=Unmute')
  await A.click('.neko-dialog .confirm')
  await B.waitForSelector('.chat-send textarea', { timeout: 5000 })
})
await step("a display name with # and ? cannot redirect the admin's mute", async () => {
  // multiuser session ids start with the login name; unencoded, this id would turn
  // POST /api/members/<id> into a request for a different path
  const E = await user('evil#?x', USER)
  await openSide(E)
  await E.waitForSelector('.chat-send textarea', { timeout: 5000 })
  await A.waitForFunction(() => document.querySelectorAll('.members-list .member').length === 3, null, {
    timeout: 5000,
  })
  await A.click('.members-list .member[aria-label="evil#?x"]', { button: 'right' })
  await A.click('.context >> text=Mute')
  await A.click('.neko-dialog .confirm')
  await E.waitForSelector('.chat-send textarea', { state: 'detached', timeout: 5000 })
  if (await A.locator('.connect').count()) throw new Error('admin lost the session')
  await E.context().close()
  await A.waitForFunction(() => document.querySelectorAll('.members-list .member').length === 2, null, {
    timeout: 5000,
  })
})
await step('ban hidden on shared-password (multiuser) provider', async () => {
  await A.click('.members-list .member:not(.self)', { button: 'right' })
  await A.waitForSelector('.context >> text=Kick', { timeout: 3000 })
  await A.waitForTimeout(500) // canBan() probe
  if (await A.locator('.context >> text=Ban').count()) throw new Error('Ban shown although accounts are not stored')
  await A.keyboard.press('Escape')
  await A.mouse.click(5, 5)
})
await step('keyboard: Enter on focused icon toggles side panel', async () => {
  await B.focus('.header .fa-bars.toggle')
  await B.keyboard.press('Enter')
  await B.waitForSelector('aside.neko-menu', { state: 'detached', timeout: 3000 })
  await B.keyboard.press('Enter')
  await B.waitForSelector('aside.neko-menu', { timeout: 3000 })
})
await step('a member who may not watch gets the room without video, not a spinner', async () => {
  const bobId = JSON.parse(await api(A, 'GET', '/sessions')).find(
    (s) => s.profile.name === 'bob' && s.state.is_connected,
  ).id
  await api(A, 'POST', '/members/' + encodeURIComponent(bobId), { can_watch: false })
  try {
    await B.reload()
    await B.waitForSelector('.connect', { state: 'detached', timeout: 20000 }) // the offer timeout (8 s) ends the spinner
    await openSide(B)
    await B.waitForSelector('.chat-send textarea', { timeout: 5000 })
  } finally {
    await api(A, 'POST', '/members/' + encodeURIComponent(bobId), { can_watch: true })
  }
  await B.reload()
  await B.waitForFunction(
    () => {
      const v = document.querySelector('video')
      return v && !v.paused && v.readyState >= 2
    },
    null,
    { timeout: 30000 },
  )
  await openSide(B)
})
await step('admin changes resolution, bob sees event; revert', async () => {
  await A.click('.video-menu .fa-desktop')
  const current = (await A.locator('.resolution li.active').textContent()) ?? ''
  const width = await A.evaluate(() => document.querySelector('video').videoWidth)
  await A.locator('.resolution li').filter({ hasNotText: current }).first().click()
  await B.waitForFunction(
    () =>
      [...document.querySelectorAll('.chat-history .event')].some((e) => /changed the resolution/.test(e.textContent)),
    null,
    { timeout: 8000 },
  )
  await A.click('.video-menu .fa-desktop')
  await A.locator('.resolution li', { hasText: current }).first().click()
  await A.mouse.click(5, 5)
  await A.waitForFunction((w) => document.querySelector('video').videoWidth === w, width, { timeout: 15000 })
})
await step('settings: keyboard layouts loaded, about opens', async () => {
  await A.click('.tabs-container >> text=Settings')
  const n = await A.locator('.side-settings select option').count()
  if (n < 20) throw new Error('only ' + n + ' keyboard layouts')
  await A.click('.room-settings .fa-question-circle')
  await A.waitForSelector('.about .about-content >> text=m1k1o/neko', { timeout: 3000 })
  await A.click('.about .about-content button')
  await A.waitForSelector('.about', { state: 'detached', timeout: 3000 })
  await A.click('.tabs-container >> text=Chat')
})
await step('implicit hosting: bob gets control by clicking', async () => {
  await api(A, 'POST', '/room/settings', { implicit_hosting: true })
  await B.waitForSelector('.neko-controls .fa-mouse-pointer', { timeout: 5000 })
  const box = await B.locator('.player-container').boundingBox()
  await B.mouse.click(box.x + box.width / 2, box.y + box.height / 2)
  await B.waitForSelector('.members-list .member.self.host', { timeout: 5000 })
  await api(A, 'POST', '/room/settings', { implicit_hosting: false })
  await A.waitForSelector('.neko-controls .fa-keyboard', { timeout: 5000 })
  await api(A, 'POST', '/room/control/reset')
  await B.waitForSelector('.members-list .member.self.host', { state: 'detached', timeout: 5000 })
})
await step('clipboard: local clipboard reaches the remote when hovering', async () => {
  await A.click('.neko-controls .fa-keyboard')
  await A.waitForSelector('.members-list .member.self.host', { timeout: 5000 })
  const text = 'neko-clip-' + Date.now()
  await A.evaluate((t) => navigator.clipboard.writeText(t), text)
  await A.mouse.move(5, 5)
  const box = await A.locator('.player-container').boundingBox()
  await A.mouse.move(box.x + 50, box.y + 50)
  await A.waitForFunction(
    async (t) => {
      const headers = { Authorization: 'Bearer ' + localStorage.getItem('neko_session') }
      const r = await fetch('api/room/clipboard', { headers })
      return r.ok && (await r.json()).text === t
    },
    text,
    { timeout: 8000, polling: 500 },
  )
})
await step('microphone: host enables and disables (fake device)', async () => {
  await A.click('.neko-controls .fa-microphone-slash')
  await A.waitForSelector('.neko-controls .fa-microphone:not(.fa-microphone-slash)', { timeout: 5000 })
  if (await A.locator('.neko-dialog[open]').count()) throw new Error(await A.locator('.neko-dialog').textContent())
  await A.click('.neko-controls .fa-microphone')
  await A.waitForSelector('.neko-controls .fa-microphone-slash', { timeout: 5000 })
  await api(A, 'POST', '/room/control/release')
})
// offline emulation only freezes an open websocket in Chromium (WebKit and Firefox keep delivering
// heartbeats to it, measured); elsewhere these two steps have nothing to notice and are skipped
const offlineWorks = browser.browserType().name() === 'chromium'
await step('network drop: bob notices the dead socket and reconnects', async () => {
  if (!offlineWorks) return log('  skipped: offline emulation does not reach the websocket here')
  // offline emulation freezes the websocket but not WebRTC: only the stale check (~25s) notices
  await B.context().setOffline(true)
  try {
    await B.waitForSelector('.connect', { timeout: 45000 })
  } finally {
    await B.context().setOffline(false)
  }
  await B.waitForSelector('.connect', { state: 'detached', timeout: 30000 })
  await B.waitForFunction(
    () => {
      const v = document.querySelector('video')
      return v && !v.paused && v.readyState >= 2
    },
    null,
    { timeout: 20000 },
  )
})
await step('long outage: session kept, Connect button instead of the login form', async () => {
  if (!offlineWorks) return log('  skipped: offline emulation does not reach the websocket here')
  // the core gives up after RECONNECT_MAX attempts (about 80 s); the session is still valid
  await B.context().setOffline(true)
  try {
    await B.waitForSelector('.connect form button[type=submit]', { timeout: 150000 })
  } finally {
    await B.context().setOffline(false)
  }
  if (await B.locator('.connect input[type=password]').count()) throw new Error('thrown back to the login form')
  await B.click('.neko-dialog .confirm') // "Disconnected: connection lost"
  await B.click('.connect form button[type=submit]')
  await B.waitForSelector('.connect', { state: 'detached', timeout: 30000 })
  await B.waitForFunction(
    () => {
      const v = document.querySelector('video')
      return v && !v.paused && v.readyState >= 2
    },
    null,
    { timeout: 20000 },
  )
})
await step('alice takes control and types into KDE', async () => {
  await A.click('.neko-controls .fa-keyboard')
  await A.waitForSelector('.members-list .member.self.host', { timeout: 5000 })
  const box = await A.locator('.player-container').boundingBox()
  await A.mouse.move(box.x + box.width / 2, box.y + box.height / 2)
  await A.mouse.move(box.x + box.width / 2 + 3, box.y + box.height / 2 + 3)
  await A.keyboard.press('Alt+F2')
  await A.waitForTimeout(1200)
  await A.keyboard.type('react gui works', { delay: 30 })
  await A.waitForTimeout(1200)
  await A.screenshot({ path: out + '/e2e-alice.png' })
  await B.screenshot({ path: out + '/e2e-bob.png' })
  // what arrived: select all and copy inside the remote, then read the remote clipboard
  const copyAll = async () => {
    await A.keyboard.press('Control+a')
    await A.keyboard.press('Control+c')
    await A.waitForTimeout(600)
    return JSON.parse(await api(A, 'GET', '/room/clipboard')).text
  }
  const typed = await copyAll()
  if (typed !== 'react gui works') throw new Error('remote received ' + JSON.stringify(typed))
  // Ctrl+V inside the remote pastes what was copied there, with the real Ctrl held. Home collapses
  // the selection to the start; End would not: with the caret already at the end, KRunner's results
  // list takes it (milou ResultsView.qml navigationKeyHandler) and swallows the following Ctrl+V
  await A.keyboard.press('Home')
  await A.keyboard.press('Control+v')
  await A.waitForTimeout(800)
  const pasted = await copyAll()
  if (pasted !== 'react gui worksreact gui works')
    throw new Error('after Ctrl+V the field has ' + JSON.stringify(pasted))
  await A.keyboard.press('Escape')
})
await step('alice kicks bob -> bob back at login', async () => {
  await A.click('.members-list .member:not(.self)', { button: 'right' })
  await A.click('.context >> text=Kick')
  await A.click('.neko-dialog .confirm')
  await B.waitForSelector('.connect input[type=password]', { timeout: 8000 })
  await A.click('.tabs-container >> text=Chat')
  await A.waitForFunction(
    () => [...document.querySelectorAll('.chat-history .event')].some((e) => /bob\s+disconnected/.test(e.textContent)),
    null,
    { timeout: 5000 },
  )
  await B.screenshot({ path: out + '/e2e-bob-kicked.png' })
})
await step('url params: ?usr&pwd&lang&show_side auto-login, url cleaned, logout', async () => {
  const ctx = await browser.newContext({ ignoreHTTPSErrors: true })
  const C = await ctx.newPage()
  await C.goto(URL + '?usr=carol&pwd=' + USER + '&lang=de&show_side=1')
  await C.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
  await C.waitForSelector('aside.neko-menu >> text=Einstellungen', { timeout: 5000 })
  if (/pwd=/.test(C.url())) throw new Error('password left in url: ' + C.url())
  // the same browser profile opens another invite link while its session is still valid:
  // the new login must win on both the websocket and REST (no split identity)
  await C.goto(URL + '?usr=carol2&pwd=' + USER + '&show_side=1')
  await C.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
  // the browser closes the first page's socket on navigation; Firefox does so without a close
  // code and the server then keeps that session for its 5 s reconnect grace
  let connected
  for (let i = 0; i < 16; i++) {
    await C.waitForTimeout(500)
    connected = JSON.parse(await api(A, 'GET', '/sessions'))
      .filter((s) => s.state.is_connected)
      .map((s) => s.profile.name)
    if (connected.includes('carol2') && !connected.includes('carol')) break
  }
  if (!connected.includes('carol2') || connected.includes('carol'))
    throw new Error('connected sessions after the second invite: ' + JSON.stringify(connected))
  await C.click('.tabs-container >> text=Einstellungen')
  await C.click('.side-settings button >> nth=-1')
  await C.waitForSelector('.connect input[type=password]', { timeout: 8000 })
  await ctx.close()
})
await step('url params: ?embed=1 hides header and room bar', async () => {
  const ctx = await browser.newContext({ ignoreHTTPSErrors: true })
  const C = await ctx.newPage()
  await C.goto(URL + '?embed=1&usr=dave&pwd=' + USER)
  await C.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
  if ((await C.locator('.header-container').count()) || (await C.locator('.room-container').count()))
    throw new Error('chrome visible in embed mode')
  await ctx.close()
})
// optional: ban/unban needs a provider that stores accounts, e.g.
// NEKO_FILE_URL=http://host:port/ with file provider members alice (admin, pw alice) and bob (pw bob)
if (process.env.NEKO_FILE_URL) {
  const FILE_URL = process.env.NEKO_FILE_URL
  const login = async (name) => {
    const ctx = await browser.newContext({ ignoreHTTPSErrors: true, viewport: { width: 1400, height: 850 } })
    const p = await ctx.newPage()
    await p.goto(FILE_URL)
    await p.fill('input[placeholder="Enter your display name"]', name)
    await p.fill('input[type=password]', name)
    await p.click('button[type=submit]')
    return p
  }
  let FA, FB
  await step('file provider: admin bans bob, bob cannot log in', async () => {
    FA = await login('alice')
    await FA.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
    FB = await login('bob')
    await FB.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
    await FA.waitForFunction(() => document.querySelectorAll('.members-list .member').length === 2, null, {
      timeout: 5000,
    })
    await FA.click('.members-list .member:not(.self)', { button: 'right' })
    await FA.click('.context >> text=Ban', { timeout: 5000 })
    await FA.click('.neko-dialog .confirm')
    await FB.waitForSelector('.connect input[type=password]', { timeout: 8000 })
    await FB.click('.neko-dialog .confirm').catch(() => {}) // "removed from this room"
    await FB.fill('input[placeholder="Enter your display name"]', 'bob')
    await FB.fill('input[type=password]', 'bob')
    await FB.click('button[type=submit]')
    await FB.waitForSelector('.neko-dialog[open]', { timeout: 5000 })
    const text = await FB.locator('.neko-dialog[open]').textContent()
    if (!/login/i.test(text ?? '') || !(await FB.locator('.connect input[type=password]').count()))
      throw new Error('expected a login error, got ' + JSON.stringify(text))
    await FB.click('.neko-dialog .confirm')
  })
  await step('file provider: unban from Settings, bob logs in again', async () => {
    await openSide(FA)
    await FA.click('.tabs-container >> text=Settings')
    await FA.click('.side-settings .banned button', { timeout: 5000 })
    await FA.waitForSelector('.side-settings .banned', { state: 'detached', timeout: 5000 })
    await FB.click('button[type=submit]')
    await FB.waitForSelector('.connect', { state: 'detached', timeout: 20000 })
  })
}
console.log(
  `\n${fails} failed. page errors:`,
  errors.length ? '\n  ' + [...new Set(errors)].slice(0, 12).join('\n  ') : 'none',
)
await browser.close()
process.exit(fails || errors.length ? 1 : 0)
