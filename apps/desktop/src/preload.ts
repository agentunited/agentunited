import { contextBridge, ipcRenderer } from 'electron';

// App configuration interface
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

// Notification data interface
interface NotificationData {
  title: string;
  body: string;
  data?: {
    channelId?: string;
    messageId?: string;
    isMention?: boolean;
  };
}

// Expose protected methods that allow the renderer process to use
// the ipcRenderer without exposing the entire object
contextBridge.exposeInMainWorld('electronAPI', {
  // Deep link handling
  onDeepLink: (callback: (url: string) => void) => {
    ipcRenderer.on('deep-link', (_event, url) => callback(url));
  },
  
  // Menu action handling
  onMenuAction: (callback: (action: string, data?: any) => void) => {
    ipcRenderer.on('menu-action', (_event, action, data) => callback(action, data));
  },
  
  // Navigation handling
  onNavigateToChannel: (callback: (channelId: string) => void) => {
    ipcRenderer.on('navigate-to-channel', (_event, channelId) => callback(channelId));
  },
  
  // App configuration
  getAppConfig: (): Promise<AppConfig> => {
    return ipcRenderer.invoke('app-config-get');
  },
  
  setAppConfig: (config: Partial<AppConfig>): Promise<AppConfig> => {
    return ipcRenderer.invoke('app-config-set', config);
  },
  
  // Dock badge management
  setDockBadge: (count: number): Promise<void> => {
    return ipcRenderer.invoke('dock-badge-set', count);
  },
  
  clearDockBadge: (): Promise<void> => {
    return ipcRenderer.invoke('dock-badge-clear');
  },
  
  // Notifications
  showNotification: (data: NotificationData): Promise<void> => {
    return ipcRenderer.invoke('notification-show', data);
  },
  
  requestNotificationPermission: (): Promise<'granted' | 'denied' | 'default'> => {
    return ipcRenderer.invoke('notification-permission');
  },
  
  // App info and controls
  getAppVersion: (): Promise<string> => {
    return ipcRenderer.invoke('app-version');
  },
  
  quitApp: (): Promise<void> => {
    return ipcRenderer.invoke('app-quit');
  },
  
  // Window controls
  minimizeWindow: (): Promise<void> => {
    return ipcRenderer.invoke('window-minimize');
  },
  
  closeWindow: (): Promise<void> => {
    return ipcRenderer.invoke('window-close');
  },
  
  maximizeWindow: (): Promise<void> => {
    return ipcRenderer.invoke('window-maximize');
  },
  
  // Platform info
  platform: process.platform,
  
  // Version info
  versions: {
    node: process.versions.node,
    chrome: process.versions.chrome,
    electron: process.versions.electron,
  },
});

// Type definitions for renderer
export interface ElectronAPI {
  // Event listeners
  onDeepLink: (callback: (url: string) => void) => void;
  onMenuAction: (callback: (action: string, data?: any) => void) => void;
  onNavigateToChannel: (callback: (channelId: string) => void) => void;
  
  // App configuration
  getAppConfig: () => Promise<AppConfig>;
  setAppConfig: (config: Partial<AppConfig>) => Promise<AppConfig>;
  
  // Dock badge
  setDockBadge: (count: number) => Promise<void>;
  clearDockBadge: () => Promise<void>;
  
  // Notifications
  showNotification: (data: NotificationData) => Promise<void>;
  requestNotificationPermission: () => Promise<'granted' | 'denied' | 'default'>;
  
  // App controls
  getAppVersion: () => Promise<string>;
  quitApp: () => Promise<void>;
  
  // Window controls
  minimizeWindow: () => Promise<void>;
  closeWindow: () => Promise<void>;
  maximizeWindow: () => Promise<void>;
  
  // Platform info
  platform: string;
  versions: {
    node: string;
    chrome: string;
    electron: string;
  };
}

declare global {
  interface Window {
    electronAPI: ElectronAPI;
  }
}

export type { AppConfig, NotificationData };