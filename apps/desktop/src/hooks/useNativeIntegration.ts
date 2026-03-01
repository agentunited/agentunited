import { useState, useEffect, useCallback, useRef } from 'react';

interface AppConfig {
  notifications: {
    enabled: boolean;
    sound: boolean;
    mentions: boolean;
    directMessages: boolean;
  };
  dock: {
    badgeEnabled: boolean;
  };
  startup: {
    launchAtLogin: boolean;
  };
}

interface NotificationOptions {
  title: string;
  body: string;
  channelId?: string;
  messageId?: string;
  isMention?: boolean;
}

export function useNativeIntegration() {
  const [appConfig, setAppConfig] = useState<AppConfig | null>(null);
  const [isElectron, setIsElectron] = useState(false);
  const [dockBadgeCount, setDockBadgeCount] = useState(0);
  const menuActionCallbacks = useRef<{ [key: string]: (data?: any) => void }>({});

  // Check if running in Electron
  useEffect(() => {
    setIsElectron(typeof window !== 'undefined' && !!window.electronAPI);
  }, []);

  // Load app configuration
  useEffect(() => {
    if (isElectron) {
      window.electronAPI.getAppConfig().then(config => {
        setAppConfig(config);
      });
    }
  }, [isElectron]);

  // Setup menu action handlers
  useEffect(() => {
    if (!isElectron) return;

    const handleMenuAction = (action: string, data?: any) => {
      const callback = menuActionCallbacks.current[action];
      if (callback) {
        callback(data);
      } else {
        console.warn('Unhandled menu action:', action, data);
      }
    };

    window.electronAPI.onMenuAction(handleMenuAction);
  }, [isElectron]);

  // Register menu action callback
  const registerMenuAction = useCallback((action: string, callback: (data?: any) => void) => {
    menuActionCallbacks.current[action] = callback;
  }, []);

  // Unregister menu action callback
  const unregisterMenuAction = useCallback((action: string) => {
    delete menuActionCallbacks.current[action];
  }, []);

  // Update app configuration
  const updateAppConfig = useCallback(async (updates: Partial<AppConfig>): Promise<boolean> => {
    if (!isElectron || !appConfig) return false;

    try {
      const newConfig = await window.electronAPI.setAppConfig(updates);
      setAppConfig(newConfig);
      return true;
    } catch (error) {
      console.error('Failed to update app config:', error);
      return false;
    }
  }, [isElectron, appConfig]);

  // Show native notification
  const showNotification = useCallback(async (options: NotificationOptions): Promise<boolean> => {
    if (!isElectron || !appConfig?.notifications.enabled) return false;

    // Check specific notification types
    if (options.isMention && !appConfig.notifications.mentions) return false;
    if (options.channelId?.startsWith('dm_') && !appConfig.notifications.directMessages) return false;

    try {
      await window.electronAPI.showNotification({
        title: options.title,
        body: options.body,
        data: {
          channelId: options.channelId,
          messageId: options.messageId,
          isMention: options.isMention
        }
      });
      return true;
    } catch (error) {
      console.error('Failed to show notification:', error);
      return false;
    }
  }, [isElectron, appConfig]);

  // Update dock badge
  const updateDockBadge = useCallback(async (count: number) => {
    if (!isElectron || !appConfig?.dock.badgeEnabled) return;

    setDockBadgeCount(count);
    
    try {
      if (count > 0) {
        await window.electronAPI.setDockBadge(count);
      } else {
        await window.electronAPI.clearDockBadge();
      }
    } catch (error) {
      console.error('Failed to update dock badge:', error);
    }
  }, [isElectron, appConfig]);

  // Clear dock badge
  const clearDockBadge = useCallback(async () => {
    await updateDockBadge(0);
  }, [updateDockBadge]);

  // Request notification permission (mostly for consistency)
  const requestNotificationPermission = useCallback(async (): Promise<boolean> => {
    if (!isElectron) return false;

    try {
      const permission = await window.electronAPI.requestNotificationPermission();
      return permission === 'granted';
    } catch (error) {
      console.error('Failed to request notification permission:', error);
      return false;
    }
  }, [isElectron]);

  // Get app version
  const getAppVersion = useCallback(async (): Promise<string | null> => {
    if (!isElectron) return null;

    try {
      return await window.electronAPI.getAppVersion();
    } catch (error) {
      console.error('Failed to get app version:', error);
      return null;
    }
  }, [isElectron]);

  // Window controls
  const minimizeWindow = useCallback(async () => {
    if (isElectron) {
      await window.electronAPI.minimizeWindow();
    }
  }, [isElectron]);

  const closeWindow = useCallback(async () => {
    if (isElectron) {
      await window.electronAPI.closeWindow();
    }
  }, [isElectron]);

  const maximizeWindow = useCallback(async () => {
    if (isElectron) {
      await window.electronAPI.maximizeWindow();
    }
  }, [isElectron]);

  const quitApp = useCallback(async () => {
    if (isElectron) {
      await window.electronAPI.quitApp();
    }
  }, [isElectron]);

  // Platform info
  const platform = isElectron ? window.electronAPI.platform : 'web';
  const isMacOS = platform === 'darwin';
  const versions = isElectron ? window.electronAPI.versions : null;

  return {
    // State
    isElectron,
    isMacOS,
    platform,
    versions,
    appConfig,
    dockBadgeCount,

    // Menu actions
    registerMenuAction,
    unregisterMenuAction,

    // App configuration
    updateAppConfig,

    // Notifications
    showNotification,
    requestNotificationPermission,

    // Dock badge
    updateDockBadge,
    clearDockBadge,

    // App controls
    getAppVersion,
    quitApp,

    // Window controls
    minimizeWindow,
    closeWindow,
    maximizeWindow
  };
}

// Helper hook for handling menu actions
export function useMenuAction(action: string, callback: (data?: any) => void) {
  const { registerMenuAction, unregisterMenuAction } = useNativeIntegration();

  useEffect(() => {
    registerMenuAction(action, callback);
    return () => unregisterMenuAction(action);
  }, [action, callback, registerMenuAction, unregisterMenuAction]);
}