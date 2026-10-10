import { useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { useTranslation } from 'react-i18next'
import { client } from '@/state/client'
import { ask } from '@/state/dialogs'
import { Check, Download, File, FileUp, Folder, RefreshCw, RotateCw, Trash2, TriangleAlert, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { IconButton } from '@/components/IconButton'
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
        <IconButton label="Refresh" className="ml-auto" onClick={refresh}>
          <RotateCw className="size-3.5" />
        </IconButton>
      </div>
      {plain.length > 2 && canDelete && (
        <div className={cn(box, 'mx-2.5 mt-2.5 flex flex-row items-center justify-between p-2 text-[0.9em]')}>
          <div className="flex items-center">
            {!selecting ? (
              <button type="button" className={action} onClick={() => setSelecting(true)}>
                {t('files:select')}
              </button>
            ) : (
              <button
                type="button"
                className={action}
                onClick={() => setSelected(allSelected ? [] : plain.map((x) => x.name))}
              >
                {t(allSelected ? 'files:unselect_all' : 'files:select_all')}
              </button>
            )}
          </div>
          {selecting && (
            <div className="flex items-center gap-[1.2em]">
              {selected.length > 0 && (
                <button
                  type="button"
                  className={cn(action, 'text-style-error hover:text-[#dc5959]')}
                  onClick={deleteSelected}
                >
                  <Trash2 className="inline size-3.5 align-[-0.125em]" /> {t('files:delete')} ({selected.length})
                </button>
              )}
              <button type="button" className={action} onClick={stopSelecting}>
                {t('files:cancel')}
              </button>
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
              {item.type === 'dir' ? (
                <Folder className="mr-2 size-3.5 shrink-0" />
              ) : (
                <File className="mr-2 size-3.5 shrink-0" />
              )}
              <p className={name} data-testid="file-name" title={item.name}>
                {item.name}
              </p>
              <p className="mr-2 ml-auto whitespace-nowrap text-white/40">{size(item.size)}</p>
              {/* native download: the browser's download manager shows progress */}
              {!selecting && item.type !== 'dir' && canDownload && (
                <a
                  href={fileUrl(item.name)}
                  download={item.name}
                  aria-label={`Download ${item.name}`}
                  title={`Download ${item.name}`}
                >
                  <Download className="size-3.5" />
                </a>
              )}
              {!selecting && item.type !== 'dir' && canDelete && (
                <IconButton
                  label={`${t('files:delete')} ${item.name}`}
                  className="ml-2"
                  data-testid="file-delete"
                  onClick={() => deleteOne(item.name)}
                >
                  <Trash2 className="size-3.5" />
                </IconButton>
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
              <IconButton
                label="Clear finished uploads"
                onClick={() => store.setState((s) => ({ uploads: s.uploads.filter((u) => u.status === 'inprogress') }))}
              >
                <X className="size-3.5" />
              </IconButton>
            </p>
            {uploads.map((u) => (
              <div key={u.id} data-testid="transfer" data-status={u.status}>
                <div className="flex max-w-full flex-row p-2.5">
                  {u.status === 'inprogress' ? (
                    <RefreshCw className="mr-2 size-3.5 shrink-0" />
                  ) : u.status === 'completed' ? (
                    <Check className="mr-2 size-3.5 shrink-0" />
                  ) : (
                    <TriangleAlert className="mr-2 size-3.5 shrink-0" />
                  )}
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
          <>
            <button
              type="button"
              className={cn(
                box,
                'm-2.5 flex w-[calc(100%-20px)] cursor-pointer flex-col justify-center text-center hover:bg-white/10 focus-visible:outline-2 focus-visible:outline-style-primary',
                drag && 'bg-white/10',
              )}
              data-testid="upload"
              onDragOver={(e) => (e.preventDefault(), setDrag(true))}
              onDragLeave={(e) => (e.preventDefault(), setDrag(false))}
              onDrop={(e) => (e.preventDefault(), setDrag(false), upload(e.dataTransfer.files))}
              onClick={() => input.current!.click()}
            >
              <FileUp className="m-2.5 size-14 self-center" />
              <span className="mx-2.5 mb-2.5">{t('files:upload_here')}</span>
            </button>
            <input
              ref={input}
              type="file"
              multiple
              hidden
              data-testid="upload-input"
              aria-label={t('files:upload_here')}
              onChange={(e) => {
                if (e.target.files) upload(e.target.files)
                e.target.value = '' // so the same file can be chosen again
              }}
            />
          </>
        )}
      </div>
    </div>
  )
}
