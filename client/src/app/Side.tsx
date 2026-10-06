import { Fragment, useEffect, useLayoutEffect, useMemo, useRef, useState } from 'react'
import {
  useNeko,
  actions,
  t,
  isLocked,
  isMuted,
  setSetting,
  remember,
  client,
  ask,
  emoji,
  loadEmoji,
  pickedEmoji,
  a11y,
  closeOn,
  type FileItem,
} from './neko'
import { parseSafe, type Node as MdNode } from './markdown'
import { Avatar, openMenu } from './Room'
import type { MemberData } from '../core/types'
import './styles/side.scss'
import './styles/chat.scss'
import './styles/files.scss'
import './styles/settings.scss'
import './styles/emoji.scss'

const TABS = [
  ['chat', 'fa-comment-alt'],
  ['files', 'fa-file'],
  ['settings', 'fa-sliders-h'],
] as const

export function Side() {
  const { app } = useNeko()
  const admin = client.isAdmin
  const f = app.files
  const filesAllowed =
    !!f?.enabled && (admin || !isLocked('file_transfer')) && (admin || f.user_download || f.user_upload)
  const tab = app.tab === 'files' && !filesAllowed ? 'chat' : app.tab

  useEffect(() => {
    document.querySelector('aside')?.scrollIntoView({ behavior: 'smooth', block: 'start' })
  }, [])

  return (
    <aside className="neko-menu">
      <div className="tabs-container">
        <ul>
          {TABS.filter(([id]) => id !== 'files' || filesAllowed).map(([id, icon]) => (
            <li
              key={id}
              className={tab === id ? 'active' : ''}
              {...a11y(t(`side.${id}`), 'tab')}
              aria-selected={tab === id}
              onClick={() => ((app.tab = id), remember('tab', id))}
            >
              <i className={`fas ${icon}`} />
              <span>{t(`side.${id}`)}</span>
            </li>
          ))}
        </ul>
      </div>
      <div className="page-container">
        {tab === 'chat' && <Chat />}
        {tab === 'files' && <Files />}
        {tab === 'settings' && <Settings />}
      </div>
    </aside>
  )
}

const time = (d: Date) =>
  d.toDateString() === new Date().toDateString()
    ? `Today at ${d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })}`
    : d.toLocaleString([], { dateStyle: 'short', timeStyle: 'short' })

const MAX_MESSAGE = 512

function Chat() {
  const { app, state } = useNeko()
  const [text, setText] = useState('')
  const [picker, setPicker] = useState(false)
  const history = useRef<HTMLUListElement>(null)
  const input = useRef<HTMLTextAreaElement>(null)

  useEffect(loadEmoji, [])
  const last = app.chat[app.chat.length - 1]?.seq
  useLayoutEffect(() => {
    history.current!.scrollTop = history.current!.scrollHeight
  }, [last])

  const onKeyDown = (e: React.KeyboardEvent) => {
    if (e.key !== 'Enter' || e.shiftKey || e.nativeEvent.isComposing) return
    e.preventDefault()
    if (!text.trim()) return
    actions.sendChat(text)
    setText('')
  }

  // insert :name: at the cursor, like the legacy picker
  const onEmoji = (name: string) => {
    const el = input.current!
    const code = `:${name}:`
    const at = el.selectionStart ?? text.length
    const next = (text.slice(0, at) + code + text.slice(el.selectionEnd ?? at)).slice(0, MAX_MESSAGE)
    setText(next)
    setPicker(false)
    requestAnimationFrame(() => {
      el.focus()
      el.selectionStart = el.selectionEnd = at + code.length
    })
  }

  const lines = app.chat
  const nameOf = (m: (typeof lines)[number]) => state.sessions[m.id]?.profile.name ?? m.name
  return (
    <div className="chat">
      <ul className="chat-history" ref={history}>
        {lines.map((m, i) =>
          m.type === 'text' ? (
            <li
              key={m.seq}
              className={`message${i > 0 && lines[i - 1].id === m.id && lines[i - 1].type === 'text' ? ' bulk' : ''}`}
            >
              <div className="author" onContextMenu={(e) => openMenu(e, m.id)}>
                <Avatar seed={nameOf(m)} avatar={state.sessions[m.id]?.profile.avatar} size={40} />
              </div>
              <div className="content">
                <div className="content-head">
                  <span>{nameOf(m)}</span>
                  <span className="timestamp">{time(m.created)}</span>
                </div>
                <Markdown source={m.content} />
              </div>
            </li>
          ) : (
            <li key={m.seq} className="event">
              <div className="content" title={time(m.created)}>
                <strong>{m.name}</strong> {m.content}
              </div>
            </li>
          ),
        )}
      </ul>
      {app.chatEnabled && !isMuted() && (
        <div className="chat-send">
          <div className="accent" />
          <div className="text-container">
            <textarea
              ref={input}
              placeholder={t('send_a_message')}
              maxLength={MAX_MESSAGE}
              value={text}
              onChange={(e) => setText(e.target.value)}
              onKeyDown={onKeyDown}
            />
            {picker && <EmojiPicker onPick={onEmoji} onClose={() => setPicker(false)} />}
            <i
              className="emoji-menu fas fa-laugh"
              {...a11y('Emoji')}
              onClick={(e) => (e.stopPropagation(), setPicker(!picker))}
            />
          </div>
        </div>
      )}
    </div>
  )
}

function Spoiler({ children }: { children: React.ReactNode }) {
  const [shown, setShown] = useState(false)
  return (
    <span
      className={`spoiler${shown ? ' active' : ''}`}
      {...(shown ? {} : a11y('Spoiler'))}
      onClick={() => setShown(true)}
    >
      <span>{children}</span>
    </span>
  )
}

function Markdown({ source }: { source: string }) {
  const { app } = useNeko() // re-render once emoji names are known
  const nodes = useMemo(() => parseSafe(source), [source])
  // open-in-app needs the plugin and control of the desktop
  const inApp = app.openInApp && client.controlling
  const open = (href: string) => (e: React.MouseEvent) => {
    if (!inApp || !app.settings.links_in_app) return
    e.preventDefault()
    actions.openInApp(href)
  }

  const render = (list: MdNode[]): React.ReactNode[] =>
    list.map((n, i) => {
      switch (n.t) {
        case 'text':
          return n.v
        case 'br':
          return <br key={i} />
        case 'code':
          return <code key={i}>{n.v}</code>
        case 'pre':
          return (
            <pre key={i}>
              <code>{n.v}</code>
            </pre>
          )
        case 'emoji':
          return app.emojiReady && emoji.names.has(n.v) ? (
            <span key={i} className="emoji" data-emoji={n.v} title={`:${n.v}:`} />
          ) : (
            `:${n.v}:`
          )
        case 'link':
          return (
            <Fragment key={i}>
              <a href={n.href} target="_blank" rel="noopener noreferrer" onClick={open(n.href)}>
                {render(n.c)}
              </a>
              {inApp && (
                <i
                  className="open-in-app fas fa-arrow-up-right-from-square"
                  {...a11y('Open in app')}
                  onClick={() => actions.openInApp(n.href)}
                />
              )}
            </Fragment>
          )
        case 'spoiler':
          return <Spoiler key={i}>{render(n.c)}</Spoiler>
        case 'quote':
          return <blockquote key={i}>{render(n.c)}</blockquote>
        default: {
          const Tag = n.t // strong, em, u, s
          return <Tag key={i}>{render(n.c)}</Tag>
        }
      }
    })

  return <div className="content-body">{render(nodes)}</div>
}

function EmojiPicker({ onPick, onClose }: { onPick: (name: string) => void; onClose: () => void }) {
  const { app } = useNeko()
  const [search, setSearch] = useState('')
  const [hovered, setHovered] = useState('')
  const [active, setActive] = useState(0)
  const ref = useRef<HTMLDivElement>(null)
  const scroll = useRef<HTMLDivElement>(null)
  const groupEls = useRef<(HTMLLIElement | null)[]>([])

  useEffect(() => closeOn(onClose), [onClose])

  const groups = [{ id: 'recent', name: 'Recent', list: app.emojiRecent }, ...emoji.groups]
  const q = search.trim().toLowerCase()
  const filtered = q
    ? [...emoji.names].filter((n) => n.includes(q) || emoji.keywords[n]?.some((k) => k.includes(q)))
    : []

  const pick = (name: string) => {
    pickedEmoji(name)
    onPick(name)
  }
  const item = (name: string, key: string) => (
    <li key={key} className={`emoji-container${hovered === name ? ' active' : ''}`}>
      <span
        className="emoji"
        data-emoji={name}
        {...a11y(`:${name}:`)}
        tabIndex={-1}
        onMouseEnter={() => setHovered(name)}
        onFocus={() => setHovered(name)}
        onClick={() => pick(name)}
      />
    </li>
  )
  const onScroll = () => {
    const top = scroll.current!.scrollTop
    let i = 0
    groupEls.current.forEach((el, idx) => el && el.offsetTop <= top && (i = idx))
    setActive(i)
  }

  return (
    <div className="neko-emoji" ref={ref} onClick={(e) => e.stopPropagation()}>
      <div className="search">
        <div className="search-contianer">
          <input
            type="text"
            autoFocus
            value={search}
            placeholder={hovered ? `:${hovered}:` : ''}
            onChange={(e) => setSearch(e.target.value)}
            onKeyDown={(e) => e.key === 'Enter' && filtered[0] && pick(filtered[0])}
          />
        </div>
      </div>
      <div className="list" ref={scroll} onScroll={onScroll}>
        {q ? (
          <ul className="emoji-container" style={{ display: 'flex' }}>
            {filtered.map((n) => item(n, n))}
          </ul>
        ) : (
          <ul className="group-list">
            {groups.map((g, gi) => (
              <li key={g.id} className="group" ref={(el) => void (groupEls.current[gi] = el)}>
                <span className="label">{g.name}</span>
                <ul className="emoji-list">{g.list.map((n) => item(n, `${g.id}-${n}`))}</ul>
              </li>
            ))}
          </ul>
        )}
      </div>
      <div className="details">
        {hovered && (
          <div className="details-container">
            <span className="emoji" data-emoji={hovered} />
            <span className="emoji-id">:{hovered}:</span>
          </div>
        )}
      </div>
      <div className="groups">
        <ul>
          {groups.map((g, gi) => (
            <li
              key={g.id}
              className={`${g.id}${active === gi && !q ? ' active' : ''}`}
              {...a11y(g.name, 'tab')}
              onClick={() => (scroll.current!.scrollTop = gi === 0 ? 0 : (groupEls.current[gi]?.offsetTop ?? 0))}
            >
              <span className={`group-${g.id} fas`} />
            </li>
          ))}
        </ul>
      </div>
    </div>
  )
}

const size = (bytes?: number) => {
  if (bytes === undefined) return ''
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = bytes > 0 ? Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), u.length - 1) : 0
  return `${+(bytes / 1024 ** i).toFixed(i ? 1 : 0)} ${u[i]}`
}

function Files() {
  const { app } = useNeko()
  const [drag, setDrag] = useState(false)
  const [selecting, setSelecting] = useState(false)
  const [selected, setSelected] = useState<string[]>([])
  const input = useRef<HTMLInputElement>(null)

  const f = app.files!
  const admin = client.isAdmin
  const canDownload = admin || f.user_download
  const canUpload = admin || f.user_upload
  const canDelete = admin || f.user_delete
  const files = f.files ?? []
  const plain = files.filter((x) => x.type !== 'dir')
  const allSelected = plain.length > 0 && plain.every((x) => selected.includes(x.name))

  const toggle = (n: string) => setSelected((s) => (s.includes(n) ? s.filter((x) => x !== n) : [...s, n]))
  const stopSelecting = () => (setSelecting(false), setSelected([]))
  const onItemClick = (item: FileItem, e: React.MouseEvent) => {
    if (item.type === 'dir') return
    if (e.ctrlKey || e.metaKey) {
      setSelecting(true)
      toggle(item.name)
    } else if (selecting) toggle(item.name)
  }
  const deleteOne = (n: string) =>
    ask(t('files.delete_title', { name: n }), t('files.delete_confirm')).then((ok) => ok && actions.fileDelete(n))
  const deleteSelected = async () => {
    if (!(await ask(t('files.delete_selected_title'), t('files.delete_selected_confirm', { count: selected.length }))))
      return
    const names = selected
    stopSelecting()
    for (const n of names) await actions.fileDelete(n)
  }

  return (
    <div className="files">
      <div className="files-cwd">
        <p>{f.root_dir}</p>
        <i className="fas fa-rotate-right refresh" {...a11y('Refresh')} onClick={actions.filesRefresh} />
      </div>
      {plain.length > 2 && canDelete && (
        <div className="files-actions">
          <div className="left-controls">
            {!selecting ? (
              <span className="action-btn select-toggle" role="button" tabIndex={0} onClick={() => setSelecting(true)}>
                {t('files.select')}
              </span>
            ) : (
              <span
                className="action-btn select-all"
                role="button"
                tabIndex={0}
                onClick={() => setSelected(allSelected ? [] : plain.map((x) => x.name))}
              >
                {t(allSelected ? 'files.unselect_all' : 'files.select_all')}
              </span>
            )}
          </div>
          {selecting && (
            <div className="right-controls">
              {selected.length > 0 && (
                <span className="action-btn delete-selected" role="button" tabIndex={0} onClick={deleteSelected}>
                  <i className="fas fa-trash"></i> {t('files.delete')} ({selected.length})
                </span>
              )}
              <span className="action-btn select-toggle cancel-btn" role="button" tabIndex={0} onClick={stopSelecting}>
                {t('files.cancel')}
              </span>
            </div>
          )}
        </div>
      )}
      <div className="files-list">
        {files.map((item) => {
          const sel = selecting && item.type !== 'dir'
          return (
            <div
              key={item.name}
              className={`files-list-item${sel ? ' selectable-item' : ''}${sel && selected.includes(item.name) ? ' selected-item' : ''}`}
              onClick={(e) => onItemClick(item, e)}
            >
              {sel && (
                <input
                  type="checkbox"
                  className="file-checkbox"
                  checked={selected.includes(item.name)}
                  onClick={(e) => e.stopPropagation()}
                  onChange={() => toggle(item.name)}
                />
              )}
              <i className={`file-icon fas ${item.type === 'dir' ? 'fa-folder' : 'fa-file'}`} />
              <p className="file-name" title={item.name}>
                {item.name}
              </p>
              <p className="file-size">{size(item.size)}</p>
              {/* native download: the browser's download manager shows progress */}
              {!selecting && item.type !== 'dir' && canDownload && (
                <a href={actions.fileUrl(item.name)} download={item.name}>
                  <i className="fas fa-download download" aria-label={`Download ${item.name}`} />
                </a>
              )}
              {!selecting && item.type !== 'dir' && canDelete && (
                <i
                  className="fas fa-trash delete"
                  {...a11y(`${t('files.delete')} ${item.name}`)}
                  onClick={() => deleteOne(item.name)}
                />
              )}
            </div>
          )
        })}
      </div>
      <div className="transfer-area">
        {app.uploads.length > 0 && (
          <div className="transfers">
            <p className="transfers-list-header">
              <span>{t('files.uploads')}</span>
              <i
                className="fas fa-xmark remove-transfer"
                {...a11y('Clear finished uploads')}
                onClick={() => (app.uploads = app.uploads.filter((u) => u.status === 'inprogress'))}
              />
            </p>
            {app.uploads.map((u) => (
              <div key={u.id} className="transfers-list-item">
                <div className="transfer-info">
                  <i
                    className={`fas transfer-status ${u.status === 'inprogress' ? 'fa-arrows-rotate' : u.status === 'completed' ? 'fa-check' : 'fa-warning'}`}
                  />
                  <p className="file-name" title={u.name}>
                    {u.name}
                  </p>
                  <p className="file-size">{Math.min(100, Math.round((u.progress / (u.size || 1)) * 100))}%</p>
                </div>
                {u.status === 'failed' ? (
                  <div className="transfer-error">{u.error}</div>
                ) : (
                  <progress className="transfer-progress" value={u.progress} max={u.size} />
                )}
              </div>
            ))}
          </div>
        )}
        {canUpload && (
          <div
            className={`upload-area${drag ? ' upload-area-drag' : ''}`}
            onDragOver={(e) => (e.preventDefault(), setDrag(true))}
            onDragLeave={(e) => (e.preventDefault(), setDrag(false))}
            onDrop={(e) => (e.preventDefault(), setDrag(false), actions.upload(e.dataTransfer.files))}
            {...a11y(t('files.upload_here'))}
            onClick={() => input.current!.click()}
          >
            <i className="fas fa-file-arrow-up" />
            <p>{t('files.upload_here')}</p>
            <input
              ref={input}
              type="file"
              multiple
              hidden
              onChange={(e) => {
                if (e.target.files) actions.upload(e.target.files)
                e.target.value = '' // so the same file can be chosen again
              }}
            />
          </div>
        )}
      </div>
    </div>
  )
}

function Settings() {
  const { app } = useNeko()
  const s = app.settings
  const [url, setUrl] = useState(app.broadcast.url)
  useEffect(() => setUrl(app.broadcast.url), [app.broadcast.url])

  const toggle = (
    key: 'scroll_invert' | 'autoplay' | 'ignore_emotes' | 'chat_sound' | 'links_in_app',
    label: string,
  ) => (
    <li>
      <span>{t(`setting.${label}`)}</span>
      <label className="switch">
        <input
          type="checkbox"
          aria-label={t(`setting.${label}`)}
          checked={s[key]}
          onChange={(e) => setSetting(key, e.target.checked)}
        />
        <span />
      </label>
    </li>
  )

  return (
    <div className="side-settings">
      <ul>
        <li>
          <span>{t('setting.scroll')}</span>
          <label className="slider">
            <input
              type="range"
              min="-5"
              max="5"
              step="1"
              value={s.scroll_sensitivity}
              onChange={(e) => setSetting('scroll_sensitivity', Number(e.target.value))}
            />
          </label>
        </li>
        {toggle('scroll_invert', 'scroll_invert')}
        {toggle('autoplay', 'autoplay')}
        {toggle('ignore_emotes', 'ignore_emotes')}
        {toggle('chat_sound', 'chat_sound')}
        {app.openInApp && toggle('links_in_app', 'links_in_app')}
        <li>
          <span>{t('setting.keyboard_layout')}</span>
          <label className="select">
            <select value={s.keyboard_layout} onChange={(e) => setSetting('keyboard_layout', e.target.value)}>
              {Object.entries(app.keyboardLayouts).map(([code, name]) => (
                <option key={code} value={code}>
                  {name}
                </option>
              ))}
            </select>
            <span />
          </label>
        </li>
        {client.isAdmin && (
          <li className="broadcast">
            <div>
              <span>{t('setting.broadcast_title')}</span>
              {!app.broadcast.active ? (
                <button aria-label={t('setting.broadcast_title')} onClick={() => actions.broadcastStart(url)}>
                  <i className="fas fa-play"></i>
                </button>
              ) : (
                <button aria-label={t('setting.broadcast_title')} onClick={actions.broadcastStop} className="btn-red">
                  <i className="fas fa-stop"></i>
                </button>
              )}
            </div>
            <input
              value={url}
              onChange={(e) => setUrl(e.target.value)}
              disabled={app.broadcast.active}
              className="input"
              placeholder="rtmp://a.rtmp.youtube.com/live2/<stream-key>"
            />
          </li>
        )}
        {client.isAdmin && <Banned />}
        <li>
          <button onClick={actions.logout}>{t('logout')}</button>
        </li>
      </ul>
    </div>
  )
}

// accounts with can_login=false (see actions.ban); empty unless the provider stores accounts
function Banned() {
  const [banned, setBanned] = useState<MemberData[] | null>(null)
  const { app } = useNeko()
  const load = () => actions.members().then((all) => setBanned(all.filter((m) => m.profile?.can_login === false)))
  useEffect(() => void load(), [app.bans]) // reloads after a ban or unban from anywhere

  if (!banned?.length) return null
  return (
    <li className="banned">
      <span>{t('setting.banned')}</span>
      <ul>
        {banned.map((m) => (
          <li key={m.id}>
            <span>{m.profile?.name || m.id}</span>
            <button onClick={() => actions.unban(m.id!)}>{t('context.unban')}</button>
          </li>
        ))}
      </ul>
    </li>
  )
}
