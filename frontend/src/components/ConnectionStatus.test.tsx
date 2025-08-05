import { render, screen, fireEvent } from "@testing-library/react";
import { vi } from "vitest";
import { ConnectionStatus } from "./ConnectionStatus";
import { ConnectionState } from "../contexts/WebSocketContext";

describe("ConnectionStatus", () => {
  it("should display connected state", () => {
    render(<ConnectionStatus connectionState={ConnectionState.CONNECTED} />);

    expect(screen.getByText("Connected")).toBeInTheDocument();
    expect(screen.getByText("●")).toBeInTheDocument();
  });

  it("should display connecting state", () => {
    render(<ConnectionStatus connectionState={ConnectionState.CONNECTING} />);

    expect(screen.getByText("Connecting...")).toBeInTheDocument();
    expect(screen.getByText("◐")).toBeInTheDocument();
  });

  it("should display reconnecting state", () => {
    render(<ConnectionStatus connectionState={ConnectionState.RECONNECTING} />);

    expect(screen.getByText("Reconnecting...")).toBeInTheDocument();
    expect(screen.getByText("◑")).toBeInTheDocument();
  });

  it("should display error state with retry button", () => {
    const mockReconnect = vi.fn();
    render(
      <ConnectionStatus
        connectionState={ConnectionState.ERROR}
        error="Connection failed"
        onReconnect={mockReconnect}
      />
    );

    expect(screen.getByText("Connection Error")).toBeInTheDocument();
    expect(screen.getByText("(Connection failed)")).toBeInTheDocument();

    const retryButton = screen.getByText("Retry");
    expect(retryButton).toBeInTheDocument();

    fireEvent.click(retryButton);
    expect(mockReconnect).toHaveBeenCalledTimes(1);
  });

  it("should display disconnected state with retry button", () => {
    const mockReconnect = vi.fn();
    render(
      <ConnectionStatus
        connectionState={ConnectionState.DISCONNECTED}
        onReconnect={mockReconnect}
      />
    );

    expect(screen.getByText("Disconnected")).toBeInTheDocument();
    expect(screen.getByText("○")).toBeInTheDocument();

    const retryButton = screen.getByText("Retry");
    expect(retryButton).toBeInTheDocument();

    fireEvent.click(retryButton);
    expect(mockReconnect).toHaveBeenCalledTimes(1);
  });

  it("should not show retry button when onReconnect is not provided", () => {
    render(<ConnectionStatus connectionState={ConnectionState.ERROR} />);

    expect(screen.queryByText("Retry")).not.toBeInTheDocument();
  });

  it("should apply custom className", () => {
    const { container } = render(
      <ConnectionStatus
        connectionState={ConnectionState.CONNECTED}
        className="custom-class"
      />
    );

    expect(container.firstChild).toHaveClass("custom-class");
  });

  it("should show error code when error is provided", () => {
    render(
      <ConnectionStatus
        connectionState={ConnectionState.ERROR}
        error="NETWORK_ERROR: Connection timeout"
      />
    );

    expect(screen.getByText("(NETWORK_ERROR)")).toBeInTheDocument();
  });
});
