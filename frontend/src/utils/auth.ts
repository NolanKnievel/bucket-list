import { supabase } from "../lib/supabase";

/**
 * Get the current user's JWT token for API requests
 */
export const getAuthToken = async (): Promise<string | null> => {
  const {
    data: { session },
  } = await supabase.auth.getSession();
  return session?.access_token ?? null;
};

/**
 * Check if user is currently authenticated
 */
export const isAuthenticated = async (): Promise<boolean> => {
  const {
    data: { session },
  } = await supabase.auth.getSession();
  return !!session;
};

/**
 * Get current user information
 */
export const getCurrentUser = async () => {
  const {
    data: { user },
  } = await supabase.auth.getUser();
  return user;
};

/**
 * Create authorization headers for API requests
 */
export const createAuthHeaders = async (): Promise<Record<string, string>> => {
  const token = await getAuthToken();
  if (!token) {
    return {};
  }

  return {
    Authorization: `Bearer ${token}`,
    "Content-Type": "application/json",
  };
};

/**
 * Refresh the current session
 */
export const refreshSession = async () => {
  const { data, error } = await supabase.auth.refreshSession();
  if (error) {
    console.error("Error refreshing session:", error);
    return null;
  }
  return data.session;
};

/**
 * Check if the current session is expired
 */
export const isSessionExpired = async (): Promise<boolean> => {
  const {
    data: { session },
  } = await supabase.auth.getSession();
  if (!session) return true;

  const now = Math.floor(Date.now() / 1000);
  return session.expires_at ? session.expires_at < now : true;
};
