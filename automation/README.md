# Neko Automation Service

Neko's built-in browser automation platform. It turns the Neko virtual
browser into a **reusable, API-driven scraping/automation service** that
knows how to discover the browser, enumerate its tabs, detect and reuse the
user's logged-in sessions, drive headless Chromium through Playwright, and
run resumable scraping jobs — **for any website, not just Pinduoduo**.

This service exists so that a future AI agent (or human) never has to
rebuild the ad-hoc machinery that was needed before: Firefox profile
discovery, SQLite cookie extraction, cookie conversion/injection,
Playwright wiring, network interception, session-loss detection, or a
command-transfer workaround. It is all behind one authenticated HTTP API.

```
┌──────────────────────────── neko container ────────────────────────────┐
│  supervisord                                                            │
│   ├─ Xorg / openbox / pulseaudio                                        │
│   ├─ firefox ──(WebDriver BiDi, 127.0.0.1:9222)──┐                      │
│   ├─ neko server (port 8080, web UI + WebRTC)   │                      │
│   └─ automation service (port 9100) ◄────────────┘                     │
│        │  FastAPI + persistent BiDi client                               │
│        │  cookie reader (safe SQLite copies)                            │
│        │  Playwright Chromium (headless, session reuse)                 │
│        │  job manager (persistent, resumable)                           │
│        ▼                                                                 │
│   /var/lib/neko-automation  (jobs, site registry, audit log)            │
│   /home/neko/.mozilla       (Firefox profile - MOUNT AS A VOLUME)       │
└──────────────────────────────────────────────────────────────────────────┘
```

## What each piece solves

| Capability | How |
|---|---|
| Is the browser up? Which profile? | `GET /api/automation/status` |
| What tabs are open right now? | WebDriver BiDi (live) with sessionstore fallback |
| Is `mobile.yangkeduo.com` logged in? | `GET /api/automation/session/status?domain=...` |
| Scrape a JS-rendered site with the user's session | `POST /api/automation/context` (cookies auto-injected) |
| Capture the site's XHR/JSON APIs | page network endpoints |
| Long scrapes that survive restarts | jobs API (persistent state, resume) |
| Login expired / CAPTCHA appeared | job pauses as `needs_attention`, user fixes it in the Neko UI, job resumes |
| A website the service doesn't know | register a site config at runtime - no code |

## Authentication

Every `/api/*` request needs the bearer token configured at deployment:

```bash
TOKEN="$NEKO_AUTOMATION_TOKEN"          # e.g. openssl rand -hex 24
BASE="https://neko.tingsrepo.com"       # or http://<host>:9100
curl -H "Authorization: Bearer $TOKEN" "$BASE/api/automation/status"
```

`/health` is unauthenticated (container healthcheck). If no token is
configured the API answers `503` for everything - it never runs open.

## The 60-second tutorial (for an AI agent)

```bash
# 1. Is everything alive? Browser? Automation? Profile?
curl -H "Authorization: Bearer $TOKEN" "$BASE/api/automation/status"

# 2. What tabs are open in the user's browser?
curl -H "Authorization: Bearer $TOKEN" "$BASE/api/automation/tabs"

# 3. Find the Pinduoduo tab (any site: use your own pattern)
curl -H "Authorization: Bearer $TOKEN" "$BASE/api/automation/tabs/find?pattern=yangkeduo"

# 4. Is that site logged in? (cookie signals; add &probe=true for a live check)
curl -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/automation/session/status?domain=mobile.yangkeduo.com"

# 5. Get an authenticated automation context (cookies auto-injected)
CTX=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"domain":"mobile.yangkeduo.com","is_mobile":true}' \
  "$BASE/api/automation/context" | python3 -c 'import json,sys;print(json.load(sys.stdin)["context_id"])')

# 6. Open a page and extract rendered content
PAGE=$(curl -s -X POST -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"url":"https://mobile.yangkeduo.com/goods.html?goods_id=799108744880"}' \
  "$BASE/api/automation/context/$CTX/pages" | python3 -c 'import json,sys;print(json.load(sys.stdin)["page_id"])')
curl -H "Authorization: Bearer $TOKEN" "$BASE/api/automation/pages/$PAGE/text"
curl -X POST -H "Authorization: Bearer $TOKEN" -H "Content-Type: application/json" \
  -d '{"expression":"document.title"}' "$BASE/api/automation/pages/$PAGE/evaluate"

# 7. See which XHR the page fired (JSON APIs!)
curl -H "Authorization: Bearer $TOKEN" \
  "$BASE/api/automation/pages/$PAGE/network?pattern=api&decode_json=true"
```

If step 4 says `LOGIN_REQUIRED`/`VERIFICATION_REQUIRED`, ask the user to
open the site in the Neko browser UI and log in / pass the check; then
re-run step 4 and continue. Never try to defeat a CAPTCHA.

## API reference

### Status
- `GET /health` - liveness (no auth)
- `GET /api/automation/status` - neko/browser/bidi/profile/playwright/jobs

### Tabs (the LIVE Firefox the user sees)
- `GET  /api/automation/tabs` - all tabs: id, url, title, domain, readyState,
  loading, per-tab auth state guess
- `GET  /api/automation/tabs/active`
- `GET  /api/automation/tabs/find?pattern=<regex or substring>`
- `GET  /api/automation/tabs/{id}`
- `POST /api/automation/tabs?url=<url>&background=false` - open new tab
- `POST /api/automation/tabs/{id}/activate`
- `POST /api/automation/tabs/{id}/navigate?url=<url>`
- `POST /api/automation/tabs/{id}/reload`
- `POST /api/automation/tabs/{id}/close`
- `POST /api/automation/tabs/{id}/screenshot` - png (base64) of that tab
- `POST /api/automation/tabs/{id}/evaluate` - JS in the live tab
  (`{"expression": "...", "await_promise": false}`)

Tab ops need Firefox's remote debugging (BiDi) - enabled by default in
this image. Tab listing falls back to sessionstore parsing if BiDi is down.

### Sessions
- `GET /api/automation/session/status?domain=<d>&probe=false[&probe_url=...]`

States: `AUTHENTICATED`, `PROBABLE_AUTHENTICATED`, `LOGIN_REQUIRED`,
`VERIFICATION_REQUIRED`, `SESSION_EXPIRED`, `NO_SESSION`, `UNKNOWN`.
Each response includes a `hints` field explaining what to do next.
`probe=true` performs a live check (renders a page with the session
cookies and classifies the landing state).

### Site registry (make ANY website first-class)
- `GET    /api/automation/sites`
- `POST   /api/automation/sites` - `{"name":"my-shop","domains":["myshop.com"],
  "auth_cookies":["sessionid"],"login_path_patterns":["*/login*"],
  "verification_path_patterns":["*captcha*"],"probe_url":"https://myshop.com/account",
  "probe_ua":"Mozilla/5.0 ..."}`
- `DELETE /api/automation/sites/{name}`

Registered sites persist in `/var/lib/neko-automation/sites.json`.
Built-ins included: pinduoduo, taobao, 1688, aliexpress, jd.

### Contexts & pages (Playwright with session reuse)
- `POST   /api/automation/context` - `{"domain":"...","reuse_existing_session":true,
  "is_mobile":true,"user_agent":"...","viewport":{...},"locale":"zh-CN"}` → `context_id`
- `GET    /api/automation/contexts`
- `DELETE /api/automation/contexts/{id}`
- `POST   /api/automation/contexts/{id}/pages?url=...` → `page_id`
- `GET    /api/automation/contexts/{id}/pages`

Page operations:
- `POST /api/automation/pages/{id}/navigate` `{"url","wait_until","timeout_ms"}`
- `GET  /api/automation/pages/{id}` - url + title
- `GET  /api/automation/pages/{id}/content` - rendered HTML
- `GET  /api/automation/pages/{id}/text` - visible text
- `POST /api/automation/pages/{id}/evaluate` `{"expression","await_promise"}`
- `POST /api/automation/pages/{id}/click` `{"selector"}`
- `POST /api/automation/pages/{id}/fill` `{"selector","value"}`
- `POST /api/automation/pages/{id}/scroll` `{"dy":1000}`
- `POST /api/automation/pages/{id}/wait` `{"selector"|"text"|"ms","timeout_ms","network_idle"}`
- `GET  /api/automation/pages/{id}/query?selector=...` - element info list
- `POST /api/automation/pages/{id}/screenshot?full_page=false`
- `POST /api/automation/pages/{id}/wait-for-response` `{"pattern","timeout_ms"}`
- `GET  /api/automation/pages/{id}/network?pattern=...&decode_json=true`
- `DELETE /api/automation/pages/{id}`

Network capture stores XHR/fetch responses (bodies up to 2 MB each,
last 500 per page) with `set-cookie`/`authorization` headers stripped.

### Jobs (long-running, resumable)
- `POST /api/automation/jobs` - `{"type":"scrape.urls","params":{...}}`
- `GET  /api/automation/jobs`
- `GET  /api/automation/jobs/{id}[?with_result=false]`
- `POST /api/automation/jobs/{id}/cancel`
- `POST /api/automation/jobs/{id}/resume`

Job statuses: `queued`, `running`, `completed`, `failed`, `cancelled`,
`interrupted` (service restarted mid-run), `needs_attention` (login
expired / verification required - resumable after the user fixes the
session in the Neko UI).

Generic job types (any website):

**`scrape.urls`**
```json
{
  "type": "scrape.urls",
  "params": {
    "urls": ["https://example.com/a", "https://example.com/b"],
    "extract": "({title: document.title, h1: document.querySelector('h1')?.innerText})",
    "wait_ms": 4000,
    "network_patterns": ["/api/"],
    "context": {"domain": "example.com", "is_mobile": false},
    "retries": 1,
    "rate_ms": 1500,
    "stop_on_auth_loss": true
  }
}
```

**`scrape.search`** (search page + XHR harvesting + scroll pagination)
```json
{
  "type": "scrape.search",
  "params": {
    "search_url_template": "https://example.com/search?q={query}",
    "queries": ["widget", "gadget"],
    "scroll_times": 8,
    "response_pattern": "/api/search",
    "extract_response": "body.items",
    "dedup_key": "r.id",
    "context": {"domain": "example.com"}
  }
}
```

Resumability: processed URLs / queries / harvested records are persisted
after every page; a resumed or restarted job continues where it stopped and
never reprocesses successful items.

### Adapters & recipes
- `GET  /api/automation/adapters` - adapters, site configs, recipes
- `POST /api/automation/adapters/recipes/run` - `{"adapter":"pinduoduo",
  "recipe":"search","query":{"queries":["generator"]}}`

The Pinduoduo adapter ships two recipes (both battle-tested):

- `pinduoduo/search` - keywords → goods records via
  `search_result.html` + `/proxy/api/search` interception
- `pinduoduo/enrich` - goods_ids → decrypted goods object
  (name, specs, galleries, SKUs) from the page heap

Pinduoduo specifics encoded in the adapter: mobile iPhone UA required for
goods pages; `mobile.yangkeduo.com` (marketplace) and `mms.pinduoduo.com`
(merchant backend) have separate sessions; login redirects mean session
loss; `psnl_verification` means human verification needed.

### Cookie export (disabled by default)
- `POST /api/automation/admin/cookies` - `{"domain":"...","names":[...]}`

Returns Playwright-format cookies. Requires BOTH the bearer token AND
`NEKO_AUTOMATION_ALLOW_COOKIE_EXPORT=true` at deployment. Every use is
audit-logged. Prefer contexts (automatic injection) over exporting.

## Login expiry / CAPTCHA flow

1. A job or probe detects a login/verification redirect.
2. The job pauses: status `needs_attention`, `attention.reason` is
   `LOGIN_REQUIRED` or `VERIFICATION_REQUIRED`.
3. Ask the user: "please open <site> in the Neko browser and log in /
   complete the check".
4. Poll `GET /api/automation/session/status?domain=...` until
   `AUTHENTICATED`.
5. `POST /api/automation/jobs/{id}/resume` - the job continues with the
   fresh cookies (contexts created afterwards pick them up automatically).

The service never attempts to defeat CAPTCHAs.

## Deployment

The image is built by `.github/workflows/automation-image.yml` →
`ghcr.io/<owner>/<repo>-automation:latest`. Deploy as a compose service
(see `docker-compose.automation.yaml`):

```yaml
services:
  neko:
    image: ghcr.io/stevenlimo20123/neko-automation:latest
    restart: unless-stopped
    shm_size: "2gb"
    ports: ["8080:8080", "52000-52100:52000-52100/udp"]
    environment:
      NEKO_MEMBER_MULTIUSER_ADMIN_PASSWORD: ${NEKO_ADMIN_PASSWORD}
      NEKO_WEBRTC_EPR: 52000-52100
      NEKO_WEBRTC_ICELITE: "1"
      NEKO_AUTOMATION_TOKEN: ${NEKO_AUTOMATION_TOKEN}
    volumes:
      - neko-profile:/home/neko/.mozilla
      - neko-automation-state:/var/lib/neko-automation
```

Route the automation API through your reverse proxy:
- separate domain → container port 9100, or
- same domain with a path prefix (e.g. `/automation` → port 9100).
  Set `NEKO_AUTOMATION_PREFIX=/automation` - the router answers on both
  `/api/...` and `/automation/api/...`, so it works whether or not the
  proxy strips the prefix.

### Environment variables

| Variable | Default | Purpose |
|---|---|---|
| `NEKO_AUTOMATION_TOKEN` | (empty → API disabled) | bearer token |
| `NEKO_AUTOMATION_ALLOW_COOKIE_EXPORT` | false | enable the export endpoint |
| `NEKO_AUTOMATION_PREFIX` | (empty) | extra route prefix |
| `NEKO_AUTOMATION_STATE_DIR` | /var/lib/neko-automation | jobs/sites/audit |
| `NEKO_AUTOMATION_PROFILE_DIR` | /home/neko/.mozilla/firefox | Firefox profile root |
| `NEKO_AUTOMATION_BIDI_URL` | ws://127.0.0.1:9222/session | Firefox BiDi endpoint |
| `NEKO_AUTOMATION_CONTEXT_IDLE_TTL` | 900 | context idle close (s) |
| `NEKO_AUTOMATION_JOB_CONCURRENCY` | 1 | parallel job workers |

### Volumes (IMPORTANT)

`/home/neko/.mozilla` (profile) and `/var/lib/neko-automation` (state)
must be volumes, otherwise a redeploy loses the user's logins and job
state. When migrating an existing container to volumes, seed the profile
volume first (copy the profile into the volume mounted at a temporary
path, then remount at the real path) - do not just mount an empty volume
over the profile.

### What the image fixes vs upstream neko

The upstream `m1k1o/neko/firefox` image ships a `policies.json` with
`SanitizeOnShutdown: {Cookies: true, Sessions: true, ...}` - every clean
Firefox shutdown DELETED ALL COOKIES AND SESSIONS, silently destroying the
user's logins on any container restart/redeploy. This image removes that
policy and sets `Homepage.StartPage: "previous-session"` so tabs are
restored. These two lines are what make "persistent sessions" real.

## Security model

- All `/api/*` routes: bearer token (constant-time compare). No token
  configured → API answers 503, never runs open.
- Cookie values are never returned by normal endpoints (names/counts
  only); raw export is opt-in via env + token + domain filter + audit.
- No arbitrary shell execution anywhere in the API.
- `evaluate` endpoints run JS inside the browser page sandbox (page
  context only), gated by the token and audit-logged.
- Network capture strips `set-cookie`/`authorization` headers.
- Audit log (JSONL) records every sensitive action with redaction.
- Firefox remote debugging binds to container-loopback only; port 9222 is
  not published.
- Do not commit tokens to git; pass them as environment secrets.

## Troubleshooting

| Symptom | Cause / fix |
|---|---|
| `bidi.connected: false` | Firefox starting or crashed; the client retries automatically. Check `/api/automation/status` → `supervisor` |
| `Maximum number of active sessions` in bidi.last_error | leaked BiDi sessions (e.g. after a hard service kill); the client restarts Firefox automatically if `NEKO_AUTOMATION_ALLOW_BROWSER_RESTART` is on - this is safe now (cookies persist) |
| Tabs list is `source: sessionstore` | BiDi down; read-only fallback, ~15 s stale |
| `scrape` job `needs_attention` | login/CAPTCHA - see flow above |
| Playwright `chromium launch failed` | run `python3 -m playwright install --with-deps chromium` inside the container |
| Goods page redirects to login on PDD | use `is_mobile: true` context (iPhone UA) - desktop UA gets login-walled |

## Local development

```bash
cd automation
pip install -r requirements.txt
python3 -m playwright install --with-deps chromium
NEKO_AUTOMATION_STATE_DIR=/tmp/na NEKO_AUTOMATION_TOKEN=t \
  python3 -m uvicorn neko_automation.main:app --port 9100
```

The service degrades gracefully without Firefox/BiDi (tabs fall back to
sessionstore, contexts still work).

## Profile seeding (migration & disaster recovery)

When moving an existing Neko container to volume-backed persistence (or
rebuilding after a volume loss), set these one-time variables on the
service:

```
NEKO_AUTOMATION_PROFILE_SEED_URL=<https url of a .tar.gz of firefox/ profile>
NEKO_AUTOMATION_PROFILE_SEED_TOKEN=<bearer token if the URL needs auth>
```

On startup, before Firefox launches, the seeder extracts the archive into
the profile root **only if** no cookies database exists yet (idempotent,
marker-file guarded, path-traversal filtered). Remove the variables after
the first successful boot.

Migration checklist for an existing container whose profile lives in the
container layer:

1. While the old container is running, snapshot the profile:
   `tar czf profile.tar.gz --exclude=cache2 -C /home/neko/.mozilla firefox`
   and host it somewhere the new container can fetch it (auth-protected).
2. Switch the compose to the automation image with volumes:
   `neko-profile:/home/neko/.mozilla` and
   `neko-automation-state:/var/lib/neko-automation`.
3. Set the two SEED variables above, deploy once, verify cookies arrived
   (`/api/automation/session/status?domain=...`), then remove the SEED
   variables and redeploy.
