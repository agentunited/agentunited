import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { getApiBaseUrl } from '../services/apiConfig';

export function HomePage() {
  const [workspaceName, setWorkspaceName] = useState<string | null>(null);
  const navigate = useNavigate();

  useEffect(() => {
    const timeout = setTimeout(() => {
      setWorkspaceName((prev) => prev ?? 'Agent United');
    }, 1500);

    fetch(`${getApiBaseUrl()}/api/v1/workspace/info`)
      .then((r) => (r.ok ? r.json() : null))
      .then((d) => setWorkspaceName(d?.name || 'Agent United'))
      .catch(() => setWorkspaceName('Agent United'))
      .finally(() => clearTimeout(timeout));

    return () => clearTimeout(timeout);
  }, []);

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
        {workspaceName === null ? (
          <div className="h-7 w-36 animate-pulse rounded-lg bg-slate-200" />
        ) : (
          <h1 className="text-2xl font-semibold tracking-tight text-slate-900">{workspaceName}</h1>
        )}
      </div>

      <div className="w-full max-w-sm rounded-2xl border border-white/70 bg-white/90 p-8 shadow-[0_24px_80px_-48px_rgba(15,23,42,0.3)] backdrop-blur">
        <h2 className="text-lg font-semibold text-slate-900">Sign in to your workspace</h2>
        <p className="mt-1 text-sm text-slate-500">Enter your credentials to continue.</p>

        <button
          onClick={() => navigate('/login')}
          className="mt-6 w-full rounded-xl bg-emerald-500 px-6 py-2.5 text-sm font-semibold text-white transition-colors hover:bg-emerald-600"
        >
          Sign In
        </button>

        <div className="mt-5 border-t border-slate-100 pt-5">
          <p className="text-center text-sm text-slate-500">
            Have an invite? Open your invite link to join.
          </p>
        </div>
      </div>

      <p className="absolute bottom-6 text-xs text-slate-400">
        Powered by{' '}
        <a
          href="https://agentunited.ai"
          target="_blank"
          rel="noopener noreferrer"
          className="transition-colors hover:text-slate-600"
        >
          Agent United
        </a>
      </p>
    </div>
  );
}
