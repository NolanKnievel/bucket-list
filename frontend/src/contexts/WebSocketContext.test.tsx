import { render, screen, act, waitFor } from "@testing-library/react";
import { vi } from "vitest";
import {
  WebSocketProvider,
  useWebSocket,
  ConnectionState,
  WS_EVENTS,
} from "./WebSocketContext";
import { BucketListItem, Member } from "../types";

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

// Test component that uses the WebSocket context
const TestComponent = () => {
  const {
    connectionState,
    isConnected,
    isOnline,
    error,
    reconnectAttempts,
    connect,
    disconnect,
    reconnect,
  } = useWebSocket();

  return (
    <div>
      <div data-testid="connection-state">{connectionState}</div>
      <div data-testid="is-connected">{isConnected.toString()}</div>
      <div data-testid="is-online">{isOnline.toString()}</div>
      <div data-testid="error">{error || "null"}</div>
      <div data-testid="reconnect-attempts">{reconnectAttempts}</div>
      <button onClick={() => connect("test-group", "test-member")}>
        Connect
      </button>
      <button onClick={disconnect}>Disconnect</button>
      <button onClick={reconnect}>Reconnect</button>
    </div>
  );
};

describe("WebSocketContext", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("should provide initial state", () => {
    render(
      <WebSocketProvider>
        <TestComponent />
      </WebSocketProvider>
    );

    expect(screen.getByTestId("connection-state")).toHaveTextContent(
      ConnectionState.DISCONNECTED
    );
    expect(screen.getByTestId("is-connected")).toHaveTextContent("false");
    expect(screen.getByTestId("is-online")).toHaveTextContent("true");
    expect(screen.getByTestId("error")).toHaveTextContent("null");
    expect(screen.getByTestId("reconnect-attempts")).toHaveTextContent("0");
  });

  it("should connect to Socket.IO", async () => {
    const { io } = await import("socket.io-client");

    render(
      <WebSocketProvider>
        <TestComponent />
      </WebSocketProvider>
    );

    const connectButton = screen.getByText("Connect");
    act(() => {
      connectButton.click();
    });

    expect(io).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({
        query: {
          groupId: "test-group",
          memberId: "test-member",
        },
        transports: ["websocket", "polling"],
        timeout: 10000,
        reconnection: true,
      })
    );
  });

  it("should handle disconnect", () => {
    render(
      <WebSocketProvider>
        <TestComponent />
      </WebSocketProvider>
    );

    const disconnectButton = screen.getByText("Disconnect");
    act(() => {
      disconnectButton.click();
    });

    expect(screen.getByTestId("connection-state")).toHaveTextContent(
      ConnectionState.DISCONNECTED
    );
  });

  it("should throw error when used outside provider", () => {
    // Suppress console.error for this test
    const consoleSpy = vi.spyOn(console, "error").mockImplementation(() => {});

    expect(() => {
      render(<TestComponent />);
    }).toThrow("useWebSocket must be used within a WebSocketProvider");

    consoleSpy.mockRestore();
  });
});
