// Authentication service for handling tokens and login state

const AUTH_TOKEN_KEY = 'auth-token';

export interface LoginResult {
  success: boolean;
  error?: string;
}

export class AuthService {
  /**
   * Store authentication token
   */
  static setToken(token: string): void {
    try {
      localStorage.setItem(AUTH_TOKEN_KEY, token);
    } catch (error) {
      console.error('Failed to store auth token:', error);
    }
  }

  /**
   * Get stored authentication token
   */
  static getToken(): string | null {
    try {
      return localStorage.getItem(AUTH_TOKEN_KEY);
    } catch (error) {
      console.error('Failed to retrieve auth token:', error);
      return null;
    }
  }

  /**
   * Remove authentication token
   */
  static removeToken(): void {
    try {
      localStorage.removeItem(AUTH_TOKEN_KEY);
    } catch (error) {
      console.error('Failed to remove auth token:', error);
    }
  }

  /**
   * Check if user is authenticated
   */
  static isAuthenticated(): boolean {
    const token = this.getToken();
    return !!token && !this.isTokenExpired(token);
  }

  /**
   * Auto-login with provided token
   */
  static async autoLogin(token: string): Promise<LoginResult> {
    try {
      // Validate token format (basic JWT structure check)
      if (!this.isValidJwtFormat(token)) {
        return {
          success: false,
          error: 'Invalid token format'
        };
      }

      // Store the token
      this.setToken(token);

      // In a real app, you might want to verify the token with the server
      // For now, we'll assume the token is valid if it has proper JWT format

      return {
        success: true
      };
    } catch (error) {
      console.error('Auto-login failed:', error);
      return {
        success: false,
        error: error instanceof Error ? error.message : 'Login failed'
      };
    }
  }

  /**
   * Logout user
   */
  static logout(): void {
    this.removeToken();
  }

  /**
   * Check if token is expired (basic check)
   */
  private static isTokenExpired(token: string): boolean {
    try {
      const payload = this.parseJwtPayload(token);
      if (payload?.exp) {
        const expirationTime = payload.exp * 1000; // Convert to milliseconds
        return Date.now() >= expirationTime;
      }
      return false; // If no expiration, assume not expired
    } catch {
      return true; // If can't parse, assume expired
    }
  }

  /**
   * Basic JWT format validation
   */
  static isValidJwtFormat(token: string): boolean {
    const parts = token.split('.');
    return parts.length === 3;
  }

  /**
   * Parse JWT payload (without verification)
   */
  private static parseJwtPayload(token: string): any {
    try {
      const parts = token.split('.');
      if (parts.length !== 3) return null;
      
      const payload = parts[1];
      const decoded = atob(payload);
      return JSON.parse(decoded);
    } catch {
      return null;
    }
  }
}

// Convenience functions
export const isAuthenticated = () => AuthService.isAuthenticated();
export const getAuthToken = () => AuthService.getToken();
export const setAuthToken = (token: string) => AuthService.setToken(token);
export const logout = () => AuthService.logout();
export const autoLogin = (token: string) => AuthService.autoLogin(token);