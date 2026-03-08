import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import { useEffect, useState } from 'react';
import { HomePage } from './pages/HomePage';
import { LoginPage } from './pages/LoginPage';
import { InviteAcceptPage } from './pages/InviteAcceptPage';
import { ChatPage } from './pages/ChatPage';
import { AgentFoundationSettingsPage } from './pages/AgentFoundationSettingsPage';
import { DashboardHomePage } from './pages/DashboardHomePage';
import { TunnelDashboardPage } from './pages/TunnelDashboardPage';
import { PairInstancePage } from './pages/PairInstancePage';
import { UserSettingsPage } from './pages/UserSettingsPage';
import { IntegrationSettingsPage } from './pages/IntegrationSettingsPage';
import { PricingPage } from './pages/PricingPage';
import { initializeFromUrlParams } from './services/apiConfig';
import { AuthService, autoLogin } from './services/authService';
import { ErrorBoundary } from './components/ui/ErrorBoundary';

// Protected Route component
function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const isAuthenticated = AuthService.isAuthenticated();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }
  
  return <>{children}</>;
}

function App() {
  const [isInitializing, setIsInitializing] = useState(true);
  const [autoLoginRedirect, setAutoLoginRedirect] = useState<string | null>(null);

  useEffect(() => {
    const initializeApp = async () => {
      try {
        // Parse URL parameters and update API configuration
        const { instanceChanged, token } = initializeFromUrlParams();

        // Handle auto-login if token is provided
        if (token) {
          const result = await autoLogin(token);
          if (result.success) {
            // Clear URL parameters after successful login
            const url = new URL(window.location.href);
            url.searchParams.delete('token');
            // Keep instance parameter for visibility, but could remove if desired
            // url.searchParams.delete('instance');
            window.history.replaceState({}, '', url.toString());
            
            // Redirect to chat after initialization
            setAutoLoginRedirect('/chat');
          } else {
            console.error('Auto-login failed:', result.error);
            // Could show an error message to user
          }
        }

        if (instanceChanged) {
          console.log('Connected to instance:', window.location.search);
        }
      } catch (error) {
        console.error('Failed to initialize app:', error);
      } finally {
        setIsInitializing(false);
      }
    };

    initializeApp();
  }, []);

  // Show loading screen while initializing
  if (isInitializing) {
    return (
      <div className="flex min-h-screen flex-col items-center justify-center gap-4 bg-white dark:bg-background">
        <div className="flex h-12 w-12 items-center justify-center rounded-full border border-emerald-500/50 shadow-[0_0_20px_rgba(16,185,129,0.3)] animate-pulse">
          <div className="h-3 w-3 rounded-full bg-emerald-500" />
        </div>
        <p className="text-sm font-medium text-muted-foreground">Loading Agent United…</p>
      </div>
    );
  }

  return (
    <ErrorBoundary title="App error" message="The app shell hit an unexpected issue. Reload to recover.">
      <BrowserRouter>
        {autoLoginRedirect && (
          <Navigate to={autoLoginRedirect} replace />
        )}
        <Routes>
          <Route path="/" element={<HomePage />} />
          <Route path="/login" element={<LoginPage />} />
          <Route path="/invite" element={<InviteAcceptPage />} />
          <Route path="/pricing" element={<PricingPage />} />
          <Route
            path="/chat"
            element={
              <ProtectedRoute>
                <ChatPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/agents/:agentId/settings"
            element={
              <ProtectedRoute>
                <AgentFoundationSettingsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/dashboard"
            element={
              <ProtectedRoute>
                <DashboardHomePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/dashboard/tunnel"
            element={
              <ProtectedRoute>
                <TunnelDashboardPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/pair-instance"
            element={
              <ProtectedRoute>
                <PairInstancePage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/settings/profile"
            element={
              <ProtectedRoute>
                <UserSettingsPage />
              </ProtectedRoute>
            }
          />
          <Route
            path="/settings/integrations"
            element={
              <ProtectedRoute>
                <IntegrationSettingsPage />
              </ProtectedRoute>
            }
          />
        </Routes>
      </BrowserRouter>
    </ErrorBoundary>
  );
}

export default App;
