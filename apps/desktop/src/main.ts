import { app, BrowserWindow, protocol, shell, Menu, Notification, ipcMain, nativeTheme, dialog } from 'electron';
import * as path from 'path';
import { autoUpdater } from 'electron-updater';

// Enable remote debugging in development
if (!app.isPackaged) {
  app.commandLine.appendSwitch('remote-debugging-port', '9222');
}

// Handle creating/removing shortcuts on Windows when installing/uninstalling
if (require('electron-squirrel-startup')) {
  app.quit();
}

let mainWindow: BrowserWindow | null = null;
let trayIcon: any = null;

// App configuration
const APP_CONFIG = {
  notifications: {
    enabled: true,
    sound: true,
    mentions: true,
    directMessages: true
  },
  dock: {
    badgeEnabled: true
  },
  startup: {
    launchAtLogin: false
  }
};

// Notification queue to prevent spam
const notificationQueue: Array<{ title: string; body: string; data?: any }> = [];
let isProcessingNotifications = false;

const createWindow = () => {
  // Create the browser window
  mainWindow = new BrowserWindow({
    width: 1200,
    height: 800,
    minWidth: 800,
    minHeight: 600,
    webPreferences: {
      preload: path.join(__dirname, 'preload.js'),
      nodeIntegration: false,
      contextIsolation: true,
    },
    titleBarStyle: 'hiddenInset', // macOS native title bar
    show: false, // Don't show until ready-to-show
    icon: path.join(__dirname, '../assets/icon.png'), // App icon
  });

  // Load the app
  // Use app.isPackaged to detect development mode (more reliable than NODE_ENV)
  if (!app.isPackaged) {
    // Development: load from Vite dev server
    mainWindow.loadURL('http://localhost:5180');
    mainWindow.webContents.openDevTools();
  } else {
    // Production: load from built files
    mainWindow.loadFile(path.join(__dirname, '../dist/index.html'));
  }

  // Show window when ready
  mainWindow.once('ready-to-show', () => {
    mainWindow?.show();
  });

  // Open external links in browser
  mainWindow.webContents.setWindowOpenHandler(({ url }) => {
    if (url.startsWith('http://') || url.startsWith('https://')) {
      shell.openExternal(url);
      return { action: 'deny' };
    }
    return { action: 'allow' };
  });

  // Handle window events
  mainWindow.on('closed', () => {
    mainWindow = null;
  });

  mainWindow.on('focus', () => {
    // Clear dock badge when app is focused
    if (process.platform === 'darwin' && APP_CONFIG.dock.badgeEnabled) {
      app.setBadgeCount(0);
    }
  });

  // Setup window menu
  createAppMenu();
};

const createAppMenu = () => {
  if (process.platform !== 'darwin') {
    return; // macOS only for now
  }

  const template: Electron.MenuItemConstructorOptions[] = [
    {
      label: 'Agent United',
      submenu: [
        {
          label: 'About Agent United',
          role: 'about'
        },
        { type: 'separator' },
        {
          label: 'Preferences...',
          accelerator: 'Cmd+,',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'preferences');
          }
        },
        { type: 'separator' },
        {
          label: 'Hide Agent United',
          accelerator: 'Cmd+H',
          role: 'hide'
        },
        {
          label: 'Hide Others',
          accelerator: 'Cmd+Shift+H',
          role: 'hideOthers'
        },
        {
          label: 'Show All',
          role: 'unhide'
        },
        { type: 'separator' },
        {
          label: 'Quit Agent United',
          accelerator: 'Cmd+Q',
          click: () => {
            app.quit();
          }
        }
      ]
    },
    {
      label: 'File',
      submenu: [
        {
          label: 'New Channel',
          accelerator: 'Cmd+Shift+N',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'new-channel');
          }
        },
        {
          label: 'New Direct Message',
          accelerator: 'Cmd+Shift+D',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'new-dm');
          }
        },
        { type: 'separator' },
        {
          label: 'Close Window',
          accelerator: 'Cmd+W',
          role: 'close'
        }
      ]
    },
    {
      label: 'Edit',
      submenu: [
        {
          label: 'Undo',
          accelerator: 'Cmd+Z',
          role: 'undo'
        },
        {
          label: 'Redo',
          accelerator: 'Cmd+Shift+Z',
          role: 'redo'
        },
        { type: 'separator' },
        {
          label: 'Cut',
          accelerator: 'Cmd+X',
          role: 'cut'
        },
        {
          label: 'Copy',
          accelerator: 'Cmd+C',
          role: 'copy'
        },
        {
          label: 'Paste',
          accelerator: 'Cmd+V',
          role: 'paste'
        },
        {
          label: 'Select All',
          accelerator: 'Cmd+A',
          role: 'selectAll'
        }
      ]
    },
    {
      label: 'View',
      submenu: [
        {
          label: 'Command Palette',
          accelerator: 'Cmd+K',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'command-palette');
          }
        },
        { type: 'separator' },
        {
          label: 'Toggle Sidebar',
          accelerator: 'Cmd+Shift+S',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'toggle-sidebar');
          }
        },
        { type: 'separator' },
        {
          label: 'Zoom In',
          accelerator: 'Cmd+Plus',
          role: 'zoomIn'
        },
        {
          label: 'Zoom Out',
          accelerator: 'Cmd+-',
          role: 'zoomOut'
        },
        {
          label: 'Reset Zoom',
          accelerator: 'Cmd+0',
          role: 'resetZoom'
        },
        { type: 'separator' },
        {
          label: 'Toggle Full Screen',
          accelerator: 'Cmd+Ctrl+F',
          role: 'togglefullscreen'
        },
        {
          label: 'Toggle Developer Tools',
          accelerator: 'Cmd+Option+I',
          click: () => {
            mainWindow?.webContents.toggleDevTools();
          }
        }
      ]
    },
    {
      label: 'Navigate',
      submenu: [
        {
          label: 'Previous Channel',
          accelerator: 'Cmd+[',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'previous-channel');
          }
        },
        {
          label: 'Next Channel',
          accelerator: 'Cmd+]',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'next-channel');
          }
        },
        { type: 'separator' },
        {
          label: 'Jump to Channel 1',
          accelerator: 'Cmd+1',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'jump-to-channel', 0);
          }
        },
        {
          label: 'Jump to Channel 2',
          accelerator: 'Cmd+2',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'jump-to-channel', 1);
          }
        },
        {
          label: 'Jump to Channel 3',
          accelerator: 'Cmd+3',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'jump-to-channel', 2);
          }
        },
        {
          label: 'Jump to Channel 4',
          accelerator: 'Cmd+4',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'jump-to-channel', 3);
          }
        },
        {
          label: 'Jump to Channel 5',
          accelerator: 'Cmd+5',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'jump-to-channel', 4);
          }
        }
      ]
    },
    {
      label: 'Window',
      submenu: [
        {
          label: 'Minimize',
          accelerator: 'Cmd+M',
          role: 'minimize'
        },
        {
          label: 'Bring All to Front',
          role: 'front'
        }
      ]
    },
    {
      label: 'Help',
      submenu: [
        {
          label: 'Agent United Help',
          click: () => {
            shell.openExternal('https://docs.agentunited.com');
          }
        },
        {
          label: 'Keyboard Shortcuts',
          accelerator: 'Cmd+/',
          click: () => {
            mainWindow?.webContents.send('menu-action', 'keyboard-shortcuts');
          }
        },
        { type: 'separator' },
        {
          label: 'Report Issue',
          click: () => {
            shell.openExternal('https://github.com/agentunited/agentunited/issues');
          }
        },
        {
          label: 'View License',
          click: () => {
            shell.openExternal('https://github.com/agentunited/agentunited/blob/main/LICENSE');
          }
        }
      ]
    }
  ];

  const menu = Menu.buildFromTemplate(template);
  Menu.setApplicationMenu(menu);
};

// Dock context menu (macOS)
const createDockMenu = () => {
  if (process.platform !== 'darwin') {
    return null;
  }

  const template: Electron.MenuItemConstructorOptions[] = [
    {
      label: 'New Channel',
      click: () => {
        if (mainWindow) {
          mainWindow.show();
          mainWindow.webContents.send('menu-action', 'new-channel');
        }
      }
    },
    {
      label: 'New Direct Message',
      click: () => {
        if (mainWindow) {
          mainWindow.show();
          mainWindow.webContents.send('menu-action', 'new-dm');
        }
      }
    },
    { type: 'separator' },
    {
      label: 'Preferences...',
      click: () => {
        if (mainWindow) {
          mainWindow.show();
          mainWindow.webContents.send('menu-action', 'preferences');
        }
      }
    }
  ];

  return Menu.buildFromTemplate(template);
};

// Show native notification
const showNotification = async (title: string, body: string, data?: any) => {
  if (!APP_CONFIG.notifications.enabled || !Notification.isSupported()) {
    return;
  }

  // Check if app is focused (don't show notifications when app is active)
  if (mainWindow && mainWindow.isFocused()) {
    return;
  }

  // Add to queue
  notificationQueue.push({ title, body, data });
  
  if (!isProcessingNotifications) {
    processNotificationQueue();
  }
};

const processNotificationQueue = async () => {
  if (notificationQueue.length === 0) {
    isProcessingNotifications = false;
    return;
  }

  isProcessingNotifications = true;
  const { title, body, data } = notificationQueue.shift()!;

  const notification = new Notification({
    title,
    body,
    icon: path.join(__dirname, '../assets/icon.png'),
    silent: !APP_CONFIG.notifications.sound
  });

  notification.on('click', () => {
    if (mainWindow) {
      if (mainWindow.isMinimized()) {
        mainWindow.restore();
      }
      mainWindow.show();
      mainWindow.focus();

      // Navigate to channel if data provided
      if (data?.channelId) {
        mainWindow.webContents.send('navigate-to-channel', data.channelId);
      }
    }
  });

  notification.show();

  // Bounce dock icon (macOS)
  if (process.platform === 'darwin' && data?.isMention) {
    app.dock.bounce('informational');
  }

  // Process next notification after delay
  setTimeout(() => {
    processNotificationQueue();
  }, 1000); // 1 second between notifications
};

// Handle deep links (agentunited://)
const handleDeepLink = (url: string) => {
  console.log('Deep link:', url);
  
  if (mainWindow) {
    if (mainWindow.isMinimized()) {
      mainWindow.restore();
    }
    mainWindow.focus();
    
    // Send deep link to renderer
    mainWindow.webContents.send('deep-link', url);
  }
};

// IPC Handlers
ipcMain.handle('app-config-get', () => {
  return APP_CONFIG;
});

ipcMain.handle('app-config-set', (event, config) => {
  Object.assign(APP_CONFIG, config);
  
  // Apply launch at login setting
  if (process.platform === 'darwin') {
    app.setLoginItemSettings({
      openAtLogin: APP_CONFIG.startup.launchAtLogin,
      openAsHidden: APP_CONFIG.startup.launchAtLogin
    });
  }
  
  return APP_CONFIG;
});

ipcMain.handle('dock-badge-set', (event, count: number) => {
  if (process.platform === 'darwin' && APP_CONFIG.dock.badgeEnabled) {
    app.setBadgeCount(count);
  }
});

ipcMain.handle('dock-badge-clear', () => {
  if (process.platform === 'darwin') {
    app.setBadgeCount(0);
  }
});

ipcMain.handle('notification-show', (event, { title, body, data }) => {
  showNotification(title, body, data);
});

ipcMain.handle('notification-permission', async () => {
  // On macOS, notifications are allowed by default
  // This is mainly for consistency with web APIs
  return 'granted';
});

ipcMain.handle('app-version', () => {
  return app.getVersion();
});

ipcMain.handle('app-quit', () => {
  app.quit();
});

ipcMain.handle('window-minimize', () => {
  mainWindow?.minimize();
});

ipcMain.handle('window-close', () => {
  mainWindow?.close();
});

ipcMain.handle('window-maximize', () => {
  if (mainWindow?.isMaximized()) {
    mainWindow.unmaximize();
  } else {
    mainWindow?.maximize();
  }
});

// macOS: Register protocol handler
if (process.platform === 'darwin') {
  app.setAsDefaultProtocolClient('agentunited');
}

// Windows/Linux: Register protocol handler
if (process.platform === 'win32' || process.platform === 'linux') {
  app.setAsDefaultProtocolClient('agentunited');
}

// Handle deep link on macOS
app.on('open-url', (event, url) => {
  event.preventDefault();
  handleDeepLink(url);
});

// Handle deep link on Windows/Linux
const gotTheLock = app.requestSingleInstanceLock();
if (!gotTheLock) {
  app.quit();
} else {
  app.on('second-instance', (event, commandLine) => {
    // Windows/Linux: handle deep link from second instance
    const url = commandLine.find((arg) => arg.startsWith('agentunited://'));
    if (url) {
      handleDeepLink(url);
    }
    
    // Focus window
    if (mainWindow) {
      if (mainWindow.isMinimized()) {
        mainWindow.restore();
      }
      mainWindow.focus();
    }
  });
}

// App ready
app.whenReady().then(() => {
  createWindow();
  
  // Set dock menu (macOS)
  if (process.platform === 'darwin') {
    const dockMenu = createDockMenu();
    if (dockMenu) {
      app.dock.setMenu(dockMenu);
    }
  }
  
  // Check for updates (production only)
  if (process.env.NODE_ENV !== 'development') {
    autoUpdater.checkForUpdatesAndNotify();
  }
  
  // macOS: re-create window when dock icon clicked
  app.on('activate', () => {
    if (BrowserWindow.getAllWindows().length === 0) {
      createWindow();
    } else if (mainWindow) {
      mainWindow.show();
    }
  });
});

// Quit when all windows closed (except macOS)
app.on('window-all-closed', () => {
  if (process.platform !== 'darwin') {
    app.quit();
  }
});

// Auto-updater events
autoUpdater.on('update-available', () => {
  console.log('Update available');
  showNotification('Update Available', 'A new version of Agent United is ready to download.');
});

autoUpdater.on('update-downloaded', () => {
  console.log('Update downloaded');
  showNotification('Update Ready', 'Agent United will restart to apply the update.');
});