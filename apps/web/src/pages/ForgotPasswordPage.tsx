import { useState } from 'react';
import { Link } from 'react-router-dom';
import { AlertCircle, ArrowLeft, CheckCircle2, Loader2 } from 'lucide-react';
import { forgotPassword } from '../services/authApi';

type PageState =
  | { status: 'idle' }
  | { status: 'loading' }
  | { status: 'success' }
  | { status: 'error'; message: string };

export function ForgotPasswordPage() {
  const [email, setEmail] = useState('');
  const [state, setState] = useState<PageState>({ status: 'idle' });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!email) return;
    setState({ status: 'loading' });
    try {
      await forgotPassword(email);
      setState({ status: 'success' });
    } catch {
      setState({ status: 'error', message: 'Something went wrong. Please try again.' });
    }
  };

  return (
    <div
      className="relative flex min-h-screen flex-col items-center justify-center px-4"
      style={{
        fontFamily: 'Manrope, system-ui, -apple-system, sans-serif',
        background:
          'radial-gradient(circle at top left, rgba(16, 185, 129, 0.12), transparent 45%), radial-gradient(circle at 80% 20%, rgba(14, 165, 233, 0.08), transparent 40%), #f8fafc',
      }}
    >
      <div className="mb-8 flex flex-col items-center gap-3">
        <div className="flex h-12 w-12 items-center justify-center rounded-full border border-emerald-500/50 shadow-[0_0_20px_rgba(16,185,129,0.35)]">
          <div className="h-3 w-3 rounded-full bg-emerald-500" />
        </div>
        <h1 className="text-xl font-semibold tracking-tight text-slate-900">Agent United</h1>
      </div>

      <div className="w-full max-w-sm rounded-2xl border border-white/70 bg-white/90 p-8 shadow-[0_24px_80px_-48px_rgba(15,23,42,0.3)] backdrop-blur">
        {state.status === 'success' ? (
          <SuccessState email={email} />
        ) : (
          <RequestForm
            email={email}
            setEmail={setEmail}
            state={state}
            onSubmit={handleSubmit}
          />
        )}
      </div>

      <Link
        to="/login"
        className="mt-6 flex items-center gap-1.5 text-xs text-slate-400 transition-colors hover:text-slate-600"
      >
        <ArrowLeft className="h-3.5 w-3.5" />
        Back to sign in
      </Link>
    </div>
  );
}

interface RequestFormProps {
  email: string;
  setEmail: (v: string) => void;
  state: PageState;
  onSubmit: (e: React.FormEvent) => void;
}

function RequestForm({ email, setEmail, state, onSubmit }: RequestFormProps) {
  const isLoading = state.status === 'loading';
  return (
    <>
      <h2 className="mb-1 text-lg font-semibold text-slate-900">Forgot your password?</h2>
      <p className="mb-6 text-sm text-slate-500">
        Enter your email and we'll send you a reset link.
      </p>

      {state.status === 'error' && (
        <div
          className="mb-4 flex items-center gap-2 rounded-xl border border-red-200 bg-red-50 px-4 py-3"
          role="alert"
        >
          <AlertCircle className="h-4 w-4 shrink-0 text-red-500" aria-hidden="true" />
          <p className="text-sm text-red-700">{state.message}</p>
        </div>
      )}

      <form onSubmit={onSubmit} className="space-y-4">
        <div className="flex flex-col gap-1.5">
          <label
            htmlFor="fp-email"
            className="text-xs font-semibold tracking-[0.1em] text-slate-600 uppercase"
          >
            Email address
          </label>
          <input
            id="fp-email"
            type="email"
            autoComplete="email"
            placeholder="you@example.com"
            disabled={isLoading}
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
            className="rounded-xl border border-slate-200 bg-white px-4 py-2.5 text-sm text-slate-900 outline-none transition-all placeholder:text-slate-400 focus:border-emerald-400/60 focus:ring-4 focus:ring-emerald-500/10 disabled:opacity-70"
          />
        </div>

        <button
          type="submit"
          disabled={isLoading || !email}
          className="mt-2 w-full rounded-xl bg-emerald-500 py-2.5 text-sm font-semibold text-white shadow-sm transition-all hover:-translate-y-0.5 hover:bg-emerald-600 disabled:translate-y-0 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {isLoading ? (
            <span className="flex items-center justify-center gap-2">
              <Loader2 className="h-4 w-4 animate-spin" aria-hidden="true" /> Sending…
            </span>
          ) : (
            'Send reset link'
          )}
        </button>
      </form>
    </>
  );
}

function SuccessState({ email }: { email: string }) {
  return (
    <div className="flex flex-col items-center gap-4 py-2 text-center">
      <CheckCircle2 className="h-10 w-10 text-emerald-500" aria-hidden="true" />
      <div>
        <h2 className="text-lg font-semibold text-slate-900">Check your email</h2>
        <p className="mt-2 text-sm text-slate-500">
          If <span className="font-medium text-slate-700">{email}</span> has an account, a
          password reset link is on its way. Check your spam folder if you don't see it.
        </p>
      </div>
      <Link
        to="/login"
        className="mt-2 w-full rounded-xl bg-emerald-500 py-2.5 text-center text-sm font-semibold text-white shadow-sm transition-all hover:-translate-y-0.5 hover:bg-emerald-600"
      >
        Back to sign in
      </Link>
    </div>
  );
}
