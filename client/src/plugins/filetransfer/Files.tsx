import { useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { client } from '@/state/client'
import { ask } from '@/state/dialogs'
import { cn } from '@/lib/utils'
import { a11y } from '@/components/a11y'
import { store, type FileItem } from './store'
import { fileUrl, fileDelete, refresh, upload } from './actions'

const box = 'rounded-[5px] bg-white/5'
const action = 'cursor-pointer font-semibold text-white/60 transition-colors duration-200 hover:text-white'
const name = 'overflow-hidden text-ellipsis whitespace-nowrap'
// the file and transfer lists scroll with a thin bar
const scroll = '[scrollbar-color:var(--color-background-tertiary)_transparent] [scrollbar-width:thin]'

const size = (bytes?: number) => {
  if (bytes === undefined) return ''
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = bytes > 0 ? Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), u.length - 1) : 0
  return `${+(bytes / 1024 ** i).toFixed(i ? 1 : 0)} ${u[i]}`
}

export function Files() {
  const { t } = useTranslation()
  const { files: f, uploads } = useStore(
    store,
    useShallow((s) => ({ files: s.files!, uploads: s.uploads })),
  )
  const admin = useStore(client.store, selectIsAdmin)
  const [drag, setDrag] = useState(false)
  const [selecting, setSelecting] = useState(false)
  const [selected, setSelected] = useState<string[]>([])
  const input = useRef<HTMLInputElement>(null)

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
    ask(t('files:delete_title', { name: n }), t('files:delete_confirm')).then((ok) => ok && fileDelete(n))
  const deleteSelected = async () => {
    if (!(await ask(t('files:delete_selected_title'), t('files:delete_selected_confirm', { count: selected.length }))))
      return
    const names = selected
    stopSelecting()
    for (const n of names) await fileDelete(n)
  }

  return (
    <div className="flex max-w-full flex-1 flex-col">
      <div className={cn(box, 'mx-2.5 mt-2.5 flex flex-row p-2 font-semibold')}>
        <p>{f.root_dir}</p>
        <i className="fas fa-rotate-right ml-auto cursor-pointer" {...a11y('Refresh')} onClick={refresh} />
      </div>
      {plain.length > 2 && canDelete && (
        <div className={cn(box, 'mx-2.5 mt-2.5 flex flex-row items-center justify-between p-2 text-[0.9em]')}>
          <div className="flex items-center">
            {!selecting ? (
              <span className={action} role="button" tabIndex={0} onClick={() => setSelecting(true)}>
                {t('files:select')}
              </span>
            ) : (
              <span
                className={action}
                role="button"
                tabIndex={0}
                onClick={() => setSelected(allSelected ? [] : plain.map((x) => x.name))}
              >
                {t(allSelected ? 'files:unselect_all' : 'files:select_all')}
              </span>
            )}
          </div>
          {selecting && (
            <div className="flex items-center gap-[1.2em]">
              {selected.length > 0 && (
                <span
                  className={cn(action, 'text-style-error hover:text-[#dc5959]')}
                  role="button"
                  tabIndex={0}
                  onClick={deleteSelected}
                >
                  <i className="fas fa-trash"></i> {t('files:delete')} ({selected.length})
                </span>
              )}
              <span className={action} role="button" tabIndex={0} onClick={stopSelecting}>
                {t('files:cancel')}
              </span>
            </div>
          )}
        </div>
      )}
      <div className={cn(box, 'm-2.5 overflow-y-scroll', scroll)}>
        {files.map((item) => {
          const sel = selecting && item.type !== 'dir'
          return (
            <div
              key={item.name}
              className={cn(
                'flex flex-row border-b-2 border-white/10 p-2 leading-[1.2] last:border-b-0',
                sel && 'cursor-pointer transition-colors duration-200 hover:bg-white/8',
                sel && selected.includes(item.name) && 'bg-white/12',
              )}
              data-testid="file"
              onClick={(e) => onItemClick(item, e)}
            >
              {sel && (
                <input
                  type="checkbox"
                  className="mr-[0.8em] cursor-pointer accent-style-primary"
                  checked={selected.includes(item.name)}
                  onClick={(e) => e.stopPropagation()}
                  onChange={() => toggle(item.name)}
                />
              )}
              <i className={`fas ${item.type === 'dir' ? 'fa-folder' : 'fa-file'} mr-2 w-3.5`} />
              <p className={name} data-testid="file-name" title={item.name}>
                {item.name}
              </p>
              <p className="mr-2 ml-auto whitespace-nowrap text-white/40">{size(item.size)}</p>
              {/* native download: the browser's download manager shows progress */}
              {!selecting && item.type !== 'dir' && canDownload && (
                <a href={fileUrl(item.name)} download={item.name}>
                  <i className="fas fa-download cursor-pointer" aria-label={`Download ${item.name}`} />
                </a>
              )}
              {!selecting && item.type !== 'dir' && canDelete && (
                <i
                  className="fas fa-trash ml-2 cursor-pointer"
                  data-testid="file-delete"
                  {...a11y(`${t('files:delete')} ${item.name}`)}
                  onClick={() => deleteOne(item.name)}
                />
              )}
            </div>
          )
        })}
      </div>
      <div className="mt-auto">
        {uploads.length > 0 && (
          <div className={cn(box, 'm-2.5 max-h-[50vh] overflow-x-hidden overflow-y-scroll', scroll)}>
            <p className="flex justify-between border-b-2 border-white/10 p-2.5 font-semibold">
              <span>{t('files:uploads')}</span>
              <i
                className="fas fa-xmark cursor-pointer"
                {...a11y('Clear finished uploads')}
                onClick={() => store.setState((s) => ({ uploads: s.uploads.filter((u) => u.status === 'inprogress') }))}
              />
            </p>
            {uploads.map((u) => (
              <div key={u.id} data-testid="transfer" data-status={u.status}>
                <div className="flex max-w-full flex-row p-2.5">
                  <i
                    className={`fas ${u.status === 'inprogress' ? 'fa-arrows-rotate' : u.status === 'completed' ? 'fa-check' : 'fa-warning'} mr-2 w-3.5`}
                  />
                  <p className={name} title={u.name}>
                    {u.name}
                  </p>
                  <p className="mr-2 ml-auto whitespace-nowrap text-white/40">
                    {Math.min(100, Math.round((u.progress / (u.size || 1)) * 100))}%
                  </p>
                </div>
                {u.status === 'failed' ? (
                  <div className="rounded-[5px] border border-style-error p-2.5">{u.error}</div>
                ) : (
                  <progress
                    className="mx-2.5 mb-2.5 h-3.5 w-[95%] appearance-none overflow-hidden rounded-full [&::-moz-progress-bar]:bg-style-primary [&::-webkit-progress-bar]:bg-background-tertiary [&::-webkit-progress-value]:bg-style-primary"
                    value={u.progress}
                    max={u.size}
                  />
                )}
              </div>
            ))}
          </div>
        )}
        {canUpload && (
          <div
            className={cn(
              box,
              'm-2.5 flex cursor-pointer flex-col justify-center text-center hover:bg-white/10',
              drag && 'bg-white/10',
            )}
            data-testid="upload"
            onDragOver={(e) => (e.preventDefault(), setDrag(true))}
            onDragLeave={(e) => (e.preventDefault(), setDrag(false))}
            onDrop={(e) => (e.preventDefault(), setDrag(false), upload(e.dataTransfer.files))}
            {...a11y(t('files:upload_here'))}
            onClick={() => input.current!.click()}
          >
            <i className="fas fa-file-arrow-up m-2.5 text-[4em]" />
            <p className="mx-2.5 mb-2.5">{t('files:upload_here')}</p>
            <input
              ref={input}
              type="file"
              multiple
              hidden
              onChange={(e) => {
                if (e.target.files) upload(e.target.files)
                e.target.value = '' // so the same file can be chosen again
              }}
            />
          </div>
        )}
      </div>
    </div>
  )
}
