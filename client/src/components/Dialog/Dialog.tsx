import { useEffect, useRef } from 'react'
import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { app } from '@/state/app'

const DIALOG_ICON = { warning: 'fa-exclamation-triangle', error: 'fa-times-circle', info: 'fa-info-circle' }
const ICON_COLOR = { warning: 'text-[#f8bb86]', error: 'text-[#f27474]', info: 'text-[#3fc3ee]' }
const button =
  'cursor-pointer rounded-[0.25em] border-0 px-[2em] py-[0.625em] text-[1.0625em] text-white focus-visible:outline-2 focus-visible:outline-style-primary'

// the one modal dialog, shown for ask() and tell() (state/dialogs.ts); the look of the legacy SweetAlert theme
export function Dialog() {
  const { t } = useTranslation()
  const d = useStore(app, (s) => s.dialog)
  const ref = useRef<HTMLDialogElement>(null)

  useEffect(() => {
    if (d && !ref.current!.open) ref.current!.showModal()
    if (!d && ref.current!.open) ref.current!.close()
  }, [d])

  const done = (ok: boolean) => {
    const resolve = app.getState().dialog?.resolve
    app.setState({ dialog: null })
    resolve?.(ok)
  }

  return (
    <dialog
      ref={ref}
      className="m-auto w-[34.5em] max-w-[calc(100%-2em)] rounded-[0.3125em] border-0 bg-background-secondary p-[1.25em] text-center text-interactive-hover backdrop:bg-black/40"
      data-testid="dialog"
      onCancel={(e) => (e.preventDefault(), done(false))}
    >
      {d && (
        <>
          <i
            className={`fas ${DIALOG_ICON[d.icon]} ${ICON_COLOR[d.icon]} mx-auto mt-[0.5em] mb-[0.8em] block text-[64px]`}
            aria-hidden="true"
          />
          <h2 className="mb-[0.4em] text-[1.875em] font-semibold">{d.title}</h2>
          {d.text && <p className="text-[1.125em] leading-[1.4]">{d.text}</p>}
          <div className="mt-[1.25em] flex justify-center gap-[0.6em]">
            <button
              className={`${button} bg-background-tertiary`}
              data-testid="dialog-confirm"
              autoFocus
              onClick={() => done(true)}
            >
              {d.cancel ? t('context.confirm.button_yes') : t('connection.button_confirm')}
            </button>
            {d.cancel && (
              <button className={`${button} bg-background-floating`} onClick={() => done(false)}>
                {t('context.confirm.button_cancel')}
              </button>
            )}
          </div>
        </>
      )}
    </dialog>
  )
}
