import { chromium } from 'playwright'
import { mkdir, writeFile } from 'node:fs/promises'
import process from 'node:process'

const baseURL = process.env.NEKO_E2E_BASE_URL || 'http://127.0.0.1:8080'
const username = process.env.NEKO_E2E_USERNAME || 'e2e-browser'
const password = process.env.NEKO_E2E_PASSWORD
const timeout = Number(process.env.NEKO_E2E_TIMEOUT_MS || 45_000)
const artifactDir = process.env.NEKO_E2E_ARTIFACT_DIR || '/tmp/neko-e2e-artifacts'
const outputPath = process.env.NEKO_E2E_OUTPUT || ''

if (!password) {
  throw new Error('NEKO_E2E_PASSWORD must be set')
}

const events = []
const startedAt = performance.now()
let browser
let page

function parseFrame(data) {
  if (typeof data !== 'string') {
    return
  }

  try {
    const message = JSON.parse(data)
    if (message && typeof message.event === 'string') {
      events.push(message.event)
    }
  } catch {
    // WebRTC signaling should be JSON, but a malformed frame is reported by
    // the final assertion instead of hiding the browser diagnostics.
  }
}

async function writeOutput(result) {
  const serialized = JSON.stringify(result, null, 2)
  process.stdout.write(`${serialized}\n`)
  if (outputPath) {
    await writeFile(outputPath, `${serialized}\n`, 'utf8')
  }
}

async function captureFailure() {
  await mkdir(artifactDir, { recursive: true })
  if (page) {
    await page.screenshot({ path: `${artifactDir}/failure.png`, fullPage: true }).catch(() => {})
  }
}

try {
  const target = new URL(baseURL)
  browser = await chromium.launch({
    headless: process.env.NEKO_E2E_HEADLESS !== '0',
    args: ['--no-sandbox', '--use-fake-ui-for-media-stream', '--autoplay-policy=no-user-gesture-required'],
  })
  const context = await browser.newContext({
    viewport: { width: 1280, height: 720 },
    permissions: ['microphone'],
  })
  page = await context.newPage()
  page.setDefaultTimeout(timeout)
  page.on('websocket', (socket) => {
    socket.on('framereceived', parseFrame)
  })

  const loginResponse = page.waitForResponse(
    (response) => response.url().endsWith('/api/login') && response.request().method() === 'POST',
  )
  await page.goto(target.toString(), { waitUntil: 'domcontentloaded' })
  await page.getByTestId('displayname-input').fill(username)
  await page.getByTestId('password-input').fill(password)
  await page.getByTestId('connect-submit').click()

  const login = await loginResponse
  if (!login.ok()) {
    throw new Error(`login failed with HTTP ${login.status()}`)
  }

  await page.getByTestId('connection-indicator').waitFor({ state: 'attached' })
  await page.waitForFunction(
    () => document.querySelector('[data-testid="connection-indicator"]')?.classList.contains('connected'),
    undefined,
    { timeout },
  )
  const connectionMs = Math.round(performance.now() - startedAt)

  await page.waitForFunction(
    () => {
      const video = document.querySelector('[data-testid="remote-video"]')
      return video && video.readyState >= 2 && video.videoWidth > 0 && video.videoHeight > 0
    },
    undefined,
    { timeout },
  )

  const frame = await page.evaluate(async () => {
    const video = document.querySelector('[data-testid="remote-video"]')
    if (!(video instanceof HTMLVideoElement)) {
      throw new Error('remote video element is missing')
    }
    await video.play().catch(() => {})
    if (typeof video.requestVideoFrameCallback === 'function') {
      await new Promise((resolve) => video.requestVideoFrameCallback(() => resolve()))
    } else {
      await new Promise((resolve) => requestAnimationFrame(() => resolve()))
    }
    return { width: video.videoWidth, height: video.videoHeight, readyState: video.readyState }
  })
  const firstFrameMs = Math.round(performance.now() - startedAt)

  if (!events.includes('system/init')) {
    throw new Error('signaling did not deliver system/init')
  }
  if (events.includes('screen_sizes_list')) {
    throw new Error('deprecated screen_sizes_list event was observed')
  }

  const result = {
    baseURL: target.origin,
    username,
    connectionMs,
    firstFrameMs,
    video: frame,
    signalingEvents: [...new Set(events)],
    browser: (await browser.version()).trim(),
    profile: process.env.NEKO_E2E_PROFILE || 'unspecified',
  }
  await writeOutput(result)
} catch (error) {
  await captureFailure()
  const message = error instanceof Error ? error.message : String(error)
  await writeOutput({
    baseURL,
    username,
    error: message,
    signalingEvents: [...new Set(events)],
  })
  process.exitCode = 1
} finally {
  await browser?.close()
}
