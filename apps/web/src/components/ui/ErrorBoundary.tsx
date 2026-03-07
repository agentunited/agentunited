import React, { type ErrorInfo, type ReactNode } from 'react'

interface ErrorBoundaryProps {
  children: ReactNode
  title?: string
  message?: string
  className?: string
}

interface ErrorBoundaryState {
  hasError: boolean
}

export class ErrorBoundary extends React.Component<ErrorBoundaryProps, ErrorBoundaryState> {
  public state: ErrorBoundaryState = { hasError: false }

  public static getDerivedStateFromError(): ErrorBoundaryState {
    return { hasError: true }
  }

  public componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    console.error('ErrorBoundary caught an error:', error, errorInfo)
  }

  private handleReload = () => {
    window.location.reload()
  }

  public render() {
    if (this.state.hasError) {
      return (
        <div className={`flex min-h-[220px] w-full items-center justify-center px-4 py-8 ${this.props.className ?? ''}`}>
          <div className="w-full max-w-md rounded-2xl border border-emerald-500/30 bg-white/90 p-6 text-center shadow-[0_20px_60px_-40px_rgba(16,185,129,0.45)] backdrop-blur dark:bg-card/90">
            <div className="mx-auto mb-4 flex h-10 w-10 items-center justify-center rounded-full border border-emerald-500/45 shadow-[0_0_16px_rgba(16,185,129,0.35)]">
              <div className="h-2.5 w-2.5 rounded-full bg-emerald-500" />
            </div>
            <h2 className="text-lg font-semibold text-foreground">
              {this.props.title ?? 'Something went wrong'}
            </h2>
            <p className="mt-1 text-sm text-muted-foreground">
              {this.props.message ?? 'Refresh to continue.'}
            </p>
            <button
              onClick={this.handleReload}
              className="mt-5 rounded-xl bg-emerald-500 px-4 py-2 text-sm font-semibold text-white transition-colors hover:bg-emerald-600"
            >
              Reload
            </button>
          </div>
        </div>
      )
    }

    return this.props.children
  }
}
