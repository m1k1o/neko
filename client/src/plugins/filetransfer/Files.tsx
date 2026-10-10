import { useRef, useState } from 'react'
import { useStore } from 'zustand'
import { useShallow } from 'zustand/react/shallow'
import { selectIsAdmin } from '@m1k1o/neko'
import { client } from '@/state/client'
import { ask } from '@/state/dialogs'
import { a11y } from '@/components/a11y'
import { t } from '@/i18n'
import { store, type FileItem } from './store'
import { fileUrl, fileDelete, refresh, upload } from './actions'
import './files.scss'

const size = (bytes?: number) => {
  if (bytes === undefined) return ''
  const u = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = bytes > 0 ? Math.min(Math.floor(Math.log(bytes) / Math.log(1024)), u.length - 1) : 0
  return `${+(bytes / 1024 ** i).toFixed(i ? 1 : 0)} ${u[i]}`
}

export function Files() {
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
    ask(t('files.delete_title', { name: n }), t('files.delete_confirm')).then((ok) => ok && fileDelete(n))
  const deleteSelected = async () => {
    if (!(await ask(t('files.delete_selected_title'), t('files.delete_selected_confirm', { count: selected.length }))))
      return
    const names = selected
    stopSelecting()
    for (const n of names) await fileDelete(n)
  }

  return (
    <div className="files">
      <div className="files-cwd">
        <p>{f.root_dir}</p>
        <i className="fas fa-rotate-right refresh" {...a11y('Refresh')} onClick={refresh} />
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
                <a href={fileUrl(item.name)} download={item.name}>
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
        {uploads.length > 0 && (
          <div className="transfers">
            <p className="transfers-list-header">
              <span>{t('files.uploads')}</span>
              <i
                className="fas fa-xmark remove-transfer"
                {...a11y('Clear finished uploads')}
                onClick={() => store.setState((s) => ({ uploads: s.uploads.filter((u) => u.status === 'inprogress') }))}
              />
            </p>
            {uploads.map((u) => (
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
            onDrop={(e) => (e.preventDefault(), setDrag(false), upload(e.dataTransfer.files))}
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
