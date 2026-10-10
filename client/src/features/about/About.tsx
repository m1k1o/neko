import { useTranslation } from 'react-i18next'
import { Book } from 'lucide-react'
import { useApp } from '@/state/provider'
import { Dialog, DialogPortal, DialogOverlay, DialogContent, DialogTitle } from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import logo from '@/assets/images/logo.svg'

// the GitHub mark (a brand icon, not in lucide): the octicon mark-github, MIT
const GitHub = () => (
  <svg className="inline size-3.5 fill-current align-[-0.125em]" viewBox="0 0 16 16" aria-hidden="true">
    <path d="M8 0C3.58 0 0 3.58 0 8c0 3.54 2.29 6.53 5.47 7.59.4.07.55-.17.55-.38 0-.19-.01-.82-.01-1.49-2.01.37-2.53-.49-2.69-.94-.09-.23-.48-.94-.82-1.13-.28-.15-.68-.52-.01-.53.63-.01 1.08.58 1.23.82.72 1.21 1.87.87 2.33.66.07-.52.28-.87.51-1.07-1.78-.2-3.64-.89-3.64-3.95 0-.87.31-1.59.82-2.15-.08-.2-.36-1.02.08-2.12 0 0 .67-.21 2.2.82.64-.18 1.32-.27 2-.27.68 0 1.36.09 2 .27 1.53-1.04 2.2-.82 2.2-.82.44 1.1.16 1.92.08 2.12.51.56.82 1.27.82 2.15 0 3.07-1.87 3.75-3.65 3.95.29.25.54.73.54 1.48 0 1.07-.01 1.93-.01 2.2 0 .21.15.46.55.38A8.013 8.013 0 0 0 16 8c0-4.42-3.58-8-8-8z" />
  </svg>
)

export function About() {
  const { t } = useTranslation()
  const app = useApp()
  const close = () => app.setState({ about: false })
  return (
    <Dialog open onOpenChange={(open) => !open && close()}>
      <DialogPortal>
        <DialogOverlay className="bg-background-floating/80" />
        <DialogContent
          className="max-h-[70vh] max-w-[70vw] overflow-x-hidden overflow-y-auto rounded-[5px] bg-background-secondary [scrollbar-color:var(--color-background-tertiary)_transparent] [scrollbar-width:thin]"
          aria-describedby={undefined}
          data-testid="about"
        >
          <DialogTitle className="sr-only">About n.eko</DialogTitle>
          <div className="px-10 py-7.5 text-center text-interactive-normal">
            <div className="mb-[15px] flex items-center justify-center">
              <img src={logo} alt="n.eko" className="mr-2.5 h-[70px]" />
              <span className="text-[30px] leading-14">
                <b>n</b>.eko
              </span>
            </div>
            <p className="mb-[15px]">A self hosted virtual browser that runs in docker and uses WebRTC.</p>
            <p className="mb-[15px]">
              <a
                className="mx-2.5 text-text-link"
                href="https://github.com/m1k1o/neko"
                target="_blank"
                rel="noopener noreferrer"
              >
                <GitHub /> m1k1o/neko
              </a>
              <a
                className="mx-2.5 text-text-link"
                href="https://neko.m1k1o.net/"
                target="_blank"
                rel="noopener noreferrer"
              >
                <Book className="inline size-3.5 align-[-0.125em]" /> Documentation
              </a>
            </p>
            <p className="mb-[15px] text-[12px] text-text-muted">client {__APP_VERSION__}</p>
            <Button
              className="px-6 py-2 leading-none font-normal text-white normal-case"
              data-testid="about-close"
              onClick={close}
            >
              {t('connection.button_confirm')}
            </Button>
          </div>
        </DialogContent>
      </DialogPortal>
    </Dialog>
  )
}
