import "@testing-library/jest-dom";

// Mock environment variables for tests
Object.defineProperty(import.meta, "env", {
  value: {
    VITE_SUPABASE_URL: "https://test.supabase.co",
    VITE_SUPABASE_ANON_KEY: "test-anon-key",
    VITE_API_URL: "http://localhost:8080",
    VITE_WS_URL: "ws://localhost:8080",
  },
  writable: true,
});
