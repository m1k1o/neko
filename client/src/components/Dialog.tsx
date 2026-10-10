import { useStore } from 'zustand'
import { useTranslation } from 'react-i18next'
import { CircleX, Info, TriangleAlert } from 'lucide-react'
import { useApp } from '@/state/provider'
import { Button } from '@/components/ui/button'
import {
  Dialog as Root,
  DialogPortal,
  DialogOverlay,
  DialogContent,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog'

const ICON = { warning: TriangleAlert, error: CircleX, info: Info }
const ICON_COLOR = { warning: 'text-[#f8bb86]', error: 'text-[#f27474]', info: 'text-[#3fc3ee]' }

// the one modal dialog, shown for ask() and tell() (state/dialogs.ts); the look of the legacy
// SweetAlert theme. Escape and a click outside dismiss it (false for ask)
export function Dialog() {
  const { t } = useTranslation()
  const app = useApp()
  const d = useStore(app, (s) => s.dialog)

  const done = (ok: boolean) => {
    const resolve = app.getState().dialog?.resolve
    app.setState({ dialog: null })
    resolve?.(ok)
  }

  const Icon = d ? ICON[d.icon] : Info
  return (
    <Root open={d !== null} onOpenChange={(open) => !open && done(false)}>
      {d && (
        <DialogPortal>
          <DialogOverlay />
          <DialogContent
            className="w-[34.5em] max-w-[calc(100%-2em)] rounded-[0.3125em] bg-background-secondary p-[1.25em] text-center text-interactive-hover"
            data-testid="dialog"
            {...(d.text ? {} : { 'aria-describedby': undefined })}
          >
            <Icon className={`${ICON_COLOR[d.icon]} mx-auto mt-[0.5em] mb-[0.8em] block size-16`} aria-hidden="true" />
            <DialogTitle className="mb-[0.4em] text-[1.875em] font-semibold">{d.title}</DialogTitle>
            {d.text && <DialogDescription className="text-[1.125em] leading-[1.4]">{d.text}</DialogDescription>}
            <div className="mt-[1.25em] flex justify-center gap-[0.6em]">
              <Button variant="confirm" data-testid="dialog-confirm" autoFocus onClick={() => done(true)}>
                {d.cancel ? t('context.confirm.button_yes') : t('connection.button_confirm')}
              </Button>
              {d.cancel && (
                <Button variant="cancel" onClick={() => done(false)}>
                  {t('context.confirm.button_cancel')}
                </Button>
              )}
            </div>
          </DialogContent>
        </DialogPortal>
      )}
    </Root>
  )
}
