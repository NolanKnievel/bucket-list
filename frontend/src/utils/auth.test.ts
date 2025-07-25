import { vi, describe, it, expect, beforeEach } from "vitest";

// Mock Supabase
vi.mock("../lib/supabase", () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
      getUser: vi.fn(),
      refreshSession: vi.fn(),
    },
  },
}));

import {
  getAuthToken,
  isAuthenticated,
  getCurrentUser,
  createAuthHeaders,
} from "./auth";
import { supabase } from "../lib/supabase";

const mockSupabase = supabase as any;

describe("Auth Utilities", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe("getAuthToken", () => {
    it("should return token when session exists", async () => {
      const mockToken = "mock-jwt-token";
      mockSupabase.auth.getSession.mockResolvedValue({
        data: {
          session: {
            access_token: mockToken,
            refresh_token: "refresh-token",
            expires_at: Date.now() / 1000 + 3600,
            token_type: "bearer",
            user: {
              id: "user-id",
              email: "test@example.com",
              aud: "authenticated",
              role: "authenticated",
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
              app_metadata: {},
              user_metadata: {},
            },
          },
        },
        error: null,
      });

      const token = await getAuthToken();
      expect(token).toBe(mockToken);
    });

    it("should return null when no session exists", async () => {
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: null },
        error: null,
      });

      const token = await getAuthToken();
      expect(token).toBeNull();
    });
  });

  describe("isAuthenticated", () => {
    it("should return true when session exists", async () => {
      mockSupabase.auth.getSession.mockResolvedValue({
        data: {
          session: {
            access_token: "token",
            refresh_token: "refresh-token",
            expires_at: Date.now() / 1000 + 3600,
            token_type: "bearer",
            user: {
              id: "user-id",
              email: "test@example.com",
              aud: "authenticated",
              role: "authenticated",
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
              app_metadata: {},
              user_metadata: {},
            },
          },
        },
        error: null,
      });

      const authenticated = await isAuthenticated();
      expect(authenticated).toBe(true);
    });

    it("should return false when no session exists", async () => {
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: null },
        error: null,
      });

      const authenticated = await isAuthenticated();
      expect(authenticated).toBe(false);
    });
  });

  describe("createAuthHeaders", () => {
    it("should create headers with token when authenticated", async () => {
      const mockToken = "mock-jwt-token";
      mockSupabase.auth.getSession.mockResolvedValue({
        data: {
          session: {
            access_token: mockToken,
            refresh_token: "refresh-token",
            expires_at: Date.now() / 1000 + 3600,
            token_type: "bearer",
            user: {
              id: "user-id",
              email: "test@example.com",
              aud: "authenticated",
              role: "authenticated",
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
              app_metadata: {},
              user_metadata: {},
            },
          },
        },
        error: null,
      });

      const headers = await createAuthHeaders();
      expect(headers).toEqual({
        Authorization: `Bearer ${mockToken}`,
        "Content-Type": "application/json",
      });
    });

    it("should return empty headers when not authenticated", async () => {
      mockSupabase.auth.getSession.mockResolvedValue({
        data: { session: null },
        error: null,
      });

      const headers = await createAuthHeaders();
      expect(headers).toEqual({});
    });
  });
});
