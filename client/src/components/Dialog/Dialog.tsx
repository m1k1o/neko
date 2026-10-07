import { useEffect, useRef } from 'react'
import { useNeko } from '@/state/hooks'
import { t } from '@/i18n'
import './dialog.scss'

const DIALOG_ICON = { warning: 'fa-exclamation-triangle', error: 'fa-times-circle', info: 'fa-info-circle' }

// the one modal dialog, shown for ask() and tell() (state/dialogs.ts)
export function Dialog() {
  const { app } = useNeko()
  const ref = useRef<HTMLDialogElement>(null)
  const d = app.dialog

  useEffect(() => {
    if (d && !ref.current!.open) ref.current!.showModal()
    if (!d && ref.current!.open) ref.current!.close()
  }, [d])

  const done = (ok: boolean) => {
    const resolve = app.dialog?.resolve
    app.dialog = null
    resolve?.(ok)
  }

  return (
    <dialog ref={ref} className="neko-dialog" onCancel={(e) => (e.preventDefault(), done(false))}>
      {d && (
        <>
          <i className={`icon ${d.icon} fas ${DIALOG_ICON[d.icon]}`} aria-hidden="true" />
          <h2>{d.title}</h2>
          {d.text && <p>{d.text}</p>}
          <div className="actions">
            <button className="confirm" autoFocus onClick={() => done(true)}>
              {d.cancel ? t('context.confirm.button_yes') : t('connection.button_confirm')}
            </button>
            {d.cancel && (
              <button className="cancel" onClick={() => done(false)}>
                {t('context.confirm.button_cancel')}
              </button>
            )}
          </div>
        </>
      )}
    </dialog>
  )
}
