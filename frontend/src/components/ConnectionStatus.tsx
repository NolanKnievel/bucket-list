import React from "react";
import { ConnectionState } from "../contexts/WebSocketContext";

interface ConnectionStatusProps {
  connectionState: ConnectionState;
  isOnline?: boolean;
  reconnectAttempts?: number;
  error?: string | null;
  onReconnect?: () => void;
  className?: string;
}

/**
 * Component that displays the current WebSocket connection status
 * with visual indicators and reconnection options
 */
export const ConnectionStatus: React.FC<ConnectionStatusProps> = ({
  connectionState,
  isOnline = true,
  reconnectAttempts = 0,
  error,
  onReconnect,
  className = "",
}) => {
  const getStatusConfig = () => {
    // Handle offline state first
    if (!isOnline) {
      return {
        color: "text-gray-600",
        bgColor: "bg-gray-100",
        icon: "⚠",
        text: "Offline",
        showReconnect: false,
      };
    }

    switch (connectionState) {
      case ConnectionState.CONNECTED:
        return {
          color: "text-green-600",
          bgColor: "bg-green-100",
          icon: "●",
          text: "Connected",
          showReconnect: false,
        };
      case ConnectionState.CONNECTING:
        return {
          color: "text-yellow-600",
          bgColor: "bg-yellow-100",
          icon: "◐",
          text: "Connecting...",
          showReconnect: false,
        };
      case ConnectionState.RECONNECTING:
        const attemptText =
          reconnectAttempts > 0 ? ` (${reconnectAttempts})` : "";
        return {
          color: "text-orange-600",
          bgColor: "bg-orange-100",
          icon: "◑",
          text: `Reconnecting${attemptText}...`,
          showReconnect: false,
        };
      case ConnectionState.ERROR:
        return {
          color: "text-red-600",
          bgColor: "bg-red-100",
          icon: "●",
          text: "Connection Error",
          showReconnect: true,
        };
      case ConnectionState.DISCONNECTED:
      default:
        return {
          color: "text-gray-600",
          bgColor: "bg-gray-100",
          icon: "○",
          text: "Disconnected",
          showReconnect: true,
        };
    }
  };

  const config = getStatusConfig();

  return (
    <div
      className={`flex items-center gap-2 px-3 py-1 rounded-full text-sm ${config.bgColor} ${className}`}
    >
      <span className={`${config.color} font-mono text-xs`} aria-hidden="true">
        {config.icon}
      </span>
      <span className={config.color}>{config.text}</span>

      {error && (
        <span className="text-red-600 text-xs" title={error}>
          ({error.split(":")[0]})
        </span>
      )}

      {config.showReconnect && onReconnect && (
        <button
          onClick={onReconnect}
          className="text-blue-600 hover:text-blue-800 text-xs underline ml-1"
          type="button"
        >
          Retry
        </button>
      )}
    </div>
  );
};
