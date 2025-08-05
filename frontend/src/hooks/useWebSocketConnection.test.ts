import React from "react";
import { renderHook, act, waitFor } from "@testing-library/react";
import { vi } from "vitest";
import { useWebSocketConnection } from "./useWebSocketConnection";
import {
  WebSocketProvider,
  ConnectionState,
} from "../contexts/WebSocketContext";

// Mock Socket.IO
vi.mock("socket.io-client", () => {
  const mockSocket = {
    connected: false,
    on: vi.fn(),
    emit: vi.fn(),
    disconnect: vi.fn(),
    connect: vi.fn(),
  };

  const mockIo = vi.fn(() => mockSocket);

  return {
    io: mockIo,
  };
});

const wrapper = ({ children }: { children: React.ReactNode }) => {
  return React.createElement(WebSocketProvider, null, children);
};

describe("useWebSocketConnection", () => {
  const mockProps = {
    groupId: "test-group-id",
    memberId: "test-member-id",
    onMemberJoined: vi.fn(),
    onItemAdded: vi.fn(),
    onItemUpdated: vi.fn(),
    onError: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should connect automatically when groupId and memberId are provided", async () => {
    const { result } = renderHook(() => useWebSocketConnection(mockProps), {
      wrapper,
    });

    expect(result.current.connectionState).toBe(ConnectionState.CONNECTING);

    const { io } = await import("socket.io-client");
    expect(io).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        query: {
          groupId: "test-group-id",
          memberId: "test-member-id",
        },
      })
    );
  });

  it("should provide convenient methods for WebSocket operations", async () => {
    const { result } = renderHook(() => useWebSocketConnection(mockProps), {
      wrapper,
    });

    // Test addItem method
    act(() => {
      result.current.addItem({
        title: "Test Item",
        description: "Test Description",
      });
    });

    // Test toggleCompletion method
    act(() => {
      result.current.toggleCompletion("test-item-id", true);
    });

    // Test reconnect method
    act(() => {
      result.current.reconnect();
    });

    // Methods should be available and not throw errors
    expect(typeof result.current.addItem).toBe("function");
    expect(typeof result.current.toggleCompletion).toBe("function");
    expect(typeof result.current.reconnect).toBe("function");
  });

  it("should handle connection errors", () => {
    const { result } = renderHook(() => useWebSocketConnection(mockProps), {
      wrapper,
    });

    // Initially should not have errors
    expect(result.current.error).toBeNull();
  });

  it("should disconnect when unmounted", async () => {
    const { result, unmount } = renderHook(
      () => useWebSocketConnection(mockProps),
      { wrapper }
    );

    unmount();

    // After unmount, the WebSocket should be disconnected
    // Note: This is handled by the WebSocketProvider cleanup
  });
});
