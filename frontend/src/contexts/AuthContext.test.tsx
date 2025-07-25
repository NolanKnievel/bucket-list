import { vi, describe, it, expect, beforeEach } from "vitest";
import { render, screen, waitFor } from "@testing-library/react";

// Mock Supabase
vi.mock("../lib/supabase", () => ({
  supabase: {
    auth: {
      getSession: vi.fn(),
      onAuthStateChange: vi.fn(),
      signInWithPassword: vi.fn(),
      signUp: vi.fn(),
      signOut: vi.fn(),
      resetPasswordForEmail: vi.fn(),
    },
  },
}));

import { AuthProvider, useAuth } from "./AuthContext";
import { supabase } from "../lib/supabase";

const mockSupabase = supabase as any;

// Test component that uses the auth context
const TestComponent = () => {
  const { user, loading, signIn } = useAuth();

  if (loading) return <div>Loading...</div>;

  return (
    <div>
      <div data-testid="user-status">
        {user ? `Logged in as ${user.email}` : "Not logged in"}
      </div>
      <button onClick={() => signIn("test@example.com", "password")}>
        Sign In
      </button>
    </div>
  );
};

describe("AuthContext", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should provide authentication context", async () => {
    // Mock no initial session
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null,
    });

    // Mock auth state change listener
    mockSupabase.auth.onAuthStateChange.mockReturnValue({
      data: { subscription: { unsubscribe: vi.fn() } },
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    // Should show loading initially
    expect(screen.getByText("Loading...")).toBeInTheDocument();

    // Wait for loading to complete
    await waitFor(() => {
      expect(screen.getByTestId("user-status")).toHaveTextContent(
        "Not logged in"
      );
    });
  });

  it("should handle sign in", async () => {
    mockSupabase.auth.getSession.mockResolvedValue({
      data: { session: null },
      error: null,
    });

    mockSupabase.auth.onAuthStateChange.mockReturnValue({
      data: { subscription: { unsubscribe: vi.fn() } },
    });

    mockSupabase.auth.signInWithPassword.mockResolvedValue({
      data: { user: null, session: null },
      error: null,
    });

    render(
      <AuthProvider>
        <TestComponent />
      </AuthProvider>
    );

    await waitFor(() => {
      expect(screen.getByTestId("user-status")).toBeInTheDocument();
    });

    // Click sign in button
    const signInButton = screen.getByText("Sign In");
    signInButton.click();

    await waitFor(() => {
      expect(mockSupabase.auth.signInWithPassword).toHaveBeenCalledWith({
        email: "test@example.com",
        password: "password",
      });
    });
  });

  it("should throw error when useAuth is used outside AuthProvider", () => {
    // Suppress console.error for this test
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => {
      render(<TestComponent />);
    }).toThrow("useAuth must be used within an AuthProvider");

    consoleSpy.mockRestore();
  });
});
