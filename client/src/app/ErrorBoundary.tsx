import { Component, type ReactNode } from 'react'
import { Logo } from '@/components/Logo'
import './unsupported.scss'

// a render bug should not leave a blank page
export class ErrorBoundary extends Component<{ children: ReactNode }, { error?: Error }> {
  state: { error?: Error } = {}

  static getDerivedStateFromError(error: Error) {
    return { error }
  }

  render() {
    if (!this.state.error) return this.props.children
    return (
      <div className="unsupported">
        <div className="window">
          <div className="logo">
            <Logo />
          </div>
          <div className="message">
            <span>Something went wrong: {this.state.error.message}</span>
            <button onClick={() => location.reload()}>Reload</button>
          </div>
        </div>
      </div>
    )
  }
}
