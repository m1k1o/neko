import { useTranslation } from 'react-i18next'
import { app } from '@/state/app'
import { Dialog, DialogPortal, DialogOverlay, DialogContent, DialogTitle } from '@/components/ui/dialog'
import logo from '@/assets/images/logo.svg'

const close = () => app.setState({ about: false })

export function About() {
  const { t } = useTranslation()
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
                <i className="fab fa-github" /> m1k1o/neko
              </a>
              <a
                className="mx-2.5 text-text-link"
                href="https://neko.m1k1o.net/"
                target="_blank"
                rel="noopener noreferrer"
              >
                <i className="fas fa-book" /> Documentation
              </a>
            </p>
            <p className="mb-[15px] text-[12px] text-text-muted">client {__APP_VERSION__}</p>
            <button
              className="cursor-pointer rounded-[5px] border-0 bg-style-primary px-6 py-2 text-white"
              data-testid="about-close"
              onClick={close}
            >
              {t('connection.button_confirm')}
            </button>
          </div>
        </DialogContent>
      </DialogPortal>
    </Dialog>
  )
}
