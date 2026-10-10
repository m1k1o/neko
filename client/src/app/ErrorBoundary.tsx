import { Component, type ReactNode } from 'react'
import logo from '@/assets/images/logo.svg'

// a render bug should not leave a blank page
export class ErrorBoundary extends Component<{ children: ReactNode }, { error?: Error }> {
  state: { error?: Error } = {}

  static getDerivedStateFromError(error: Error) {
    return { error }
  }

  render() {
    if (!this.state.error) return this.props.children
    return (
      <div className="absolute inset-0 flex items-center justify-center">
        <div className="w-[320px] rounded-[5px] bg-background-secondary p-2.5">
          <div className="flex w-full items-center justify-center">
            <img src={logo} alt="n.eko" className="mr-2.5 h-[90px]" />
            <span className="text-[30px] leading-14">
              <b className="font-black">n</b>.eko
            </span>
          </div>
          <div className="flex flex-col">
            <span className="block text-center leading-7.5 uppercase">
              Something went wrong: {this.state.error.message}
            </span>
            <button
              className="my-[5px] cursor-pointer rounded-[5px] bg-style-primary p-1 text-center leading-7.5 font-bold text-text-normal uppercase"
              onClick={() => location.reload()}
            >
              Reload
            </button>
          </div>
        </div>
      </div>
    )
  }
}
