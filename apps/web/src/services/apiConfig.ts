// Centralized API configuration with dynamic instance support

const STORAGE_KEY = 'agentunited_instance_url';
// Always default to same-origin API routing via nginx proxy (/api -> API_UPSTREAM).
// Do not use build-time API URL defaults here; invite links must resolve against current host.
const DEFAULT_INSTANCE_URL = '';

class ApiConfig {
  private instanceUrl: string;

  constructor() {
    // Initialize from localStorage or use default
    this.instanceUrl = this.loadInstanceUrl();
  }

  /**
   * Get the current API base URL
   */
  getBaseUrl(): string {
    return this.instanceUrl;
  }

  /**
   * Set the API base URL and persist to localStorage
   */
  setBaseUrl(url: string): void {
    // Validate URL format
    if (!this.isValidUrl(url)) {
      throw new Error(`Invalid instance URL: ${url}`);
    }

    // Normalize URL (remove trailing slash)
    const normalizedUrl = url.replace(/\/$/, '');
    
    this.instanceUrl = normalizedUrl;
    this.saveInstanceUrl(normalizedUrl);
  }

  /**
   * Reset to default instance URL
   */
  resetToDefault(): void {
    this.instanceUrl = DEFAULT_INSTANCE_URL;
    localStorage.removeItem(STORAGE_KEY);
  }

  /**
   * Check if current instance is the default (localhost)
   */
  isDefaultInstance(): boolean {
    return this.instanceUrl === DEFAULT_INSTANCE_URL;
  }

  /**
   * Parse instance URL from current page URL parameters
   */
  parseFromUrlParams(): { instanceUrl?: string; token?: string } {
    const params = new URLSearchParams(window.location.search);
    const instanceUrl = params.get('instance');
    const token = params.get('token');

    const result: { instanceUrl?: string; token?: string } = {};

    if (instanceUrl) {
      if (this.isValidUrl(instanceUrl)) {
        result.instanceUrl = instanceUrl;
      } else {
        console.warn(`Invalid instance URL in parameters: ${instanceUrl}`);
      }
    }

    if (token) {
      result.token = token;
    }

    return result;
  }

  /**
   * Apply URL parameters if present
   */
  applyUrlParams(): { instanceChanged: boolean; token?: string } {
    const { instanceUrl, token } = this.parseFromUrlParams();
    let instanceChanged = false;

    if (instanceUrl && instanceUrl !== this.instanceUrl) {
      this.setBaseUrl(instanceUrl);
      instanceChanged = true;
    }

    // Invite links should validate against the current host's API proxy by default.
    // If a token is present but no explicit instance override is provided, reset to default.
    if (token && !instanceUrl && this.instanceUrl !== DEFAULT_INSTANCE_URL) {
      this.resetToDefault();
      instanceChanged = true;
    }

    return { instanceChanged, token };
  }

  /**
   * Load instance URL from localStorage
   */
  private loadInstanceUrl(): string {
    try {
      const stored = localStorage.getItem(STORAGE_KEY);
      if (stored && this.isValidUrl(stored)) {
        return stored;
      }
    } catch (error) {
      console.warn('Failed to load instance URL from localStorage:', error);
    }
    return DEFAULT_INSTANCE_URL;
  }

  /**
   * Save instance URL to localStorage
   */
  private saveInstanceUrl(url: string): void {
    try {
      localStorage.setItem(STORAGE_KEY, url);
    } catch (error) {
      console.warn('Failed to save instance URL to localStorage:', error);
    }
  }

  /**
   * Validate URL format
   */
  private isValidUrl(url: string): boolean {
    try {
      const urlObj = new URL(url);
      return urlObj.protocol === 'http:' || urlObj.protocol === 'https:';
    } catch {
      return false;
    }
  }
}

// Singleton instance
export const apiConfig = new ApiConfig();

/**
 * Get the current API base URL
 */
export function getApiBaseUrl(): string {
  return apiConfig.getBaseUrl();
}

/**
 * Set the API base URL
 */
export function setApiBaseUrl(url: string): void {
  apiConfig.setBaseUrl(url);
}

/**
 * Initialize API configuration from URL parameters
 * Returns true if instance URL was changed
 */
export function initializeFromUrlParams(): { instanceChanged: boolean; token?: string } {
  return apiConfig.applyUrlParams();
}