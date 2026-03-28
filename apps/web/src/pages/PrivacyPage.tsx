import { Link } from 'react-router-dom'
import ReactMarkdown from 'react-markdown'
import privacyMarkdown from '../legal/privacy.md?raw'

export function PrivacyPage() {
  return (
    <div className="min-h-screen bg-background text-foreground">
      <main className="mx-auto w-full max-w-3xl px-6 py-10">
        <div className="mb-8 flex items-center justify-between gap-4">
          <h1 className="text-3xl font-semibold tracking-tight">Privacy Policy</h1>
          <Link to="/" className="text-sm text-muted-foreground hover:text-foreground">← Back</Link>
        </div>


        <article className="prose prose-slate max-w-none dark:prose-invert">
          <ReactMarkdown>{privacyMarkdown}</ReactMarkdown>
        </article>
      </main>
    </div>
  )
}
