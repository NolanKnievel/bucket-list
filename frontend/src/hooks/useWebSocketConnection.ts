import { useEffect, useCallback, useState } from "react";
import { useWebSocket, ConnectionState } from "../contexts/WebSocketContext";
import { BucketListItem, Member } from "../types";

interface UseWebSocketConnectionProps {
  groupId: string;
  memberId: string;
  onMemberJoined?: (member: Member) => void;
  onItemAdded?: (item: BucketListItem) => void;
  onItemUpdated?: (item: BucketListItem) => void;
  onError?: (error: {
    code: string;
    message: string;
    details?: string;
  }) => void;
}

interface UseWebSocketConnectionReturn {
  connectionState: ConnectionState;
  isConnected: boolean;
  isOnline: boolean;
  error: string | null;
  reconnectAttempts: number;
  reconnect: () => void;
  addItem: (item: { title: string; description?: string }) => void;
  toggleCompletion: (itemId: string, completed: boolean) => void;
}

/**
 * Custom hook for managing WebSocket connection to a specific group
 * Handles automatic connection, reconnection, and provides convenient methods
 */
export const useWebSocketConnection = ({
  groupId,
  memberId,
  onMemberJoined,
  onItemAdded,
  onItemUpdated,
  onError,
}: UseWebSocketConnectionProps): UseWebSocketConnectionReturn => {
  const {
    connectionState,
    isConnected,
    isOnline,
    error,
    reconnectAttempts,
    connect,
    disconnect,
    reconnect: wsReconnect,
    joinGroup,
    addItem: wsAddItem,
    toggleCompletion: wsToggleCompletion,
    onMemberJoined: registerMemberJoinedListener,
    onItemAdded: registerItemAddedListener,
    onItemUpdated: registerItemUpdatedListener,
    onError: registerErrorListener,
  } = useWebSocket();

  const [hasJoinedGroup, setHasJoinedGroup] = useState(false);

  // Connect to WebSocket when component mounts or groupId/memberId changes
  useEffect(() => {
    if (groupId && memberId) {
      connect(groupId, memberId);
      setHasJoinedGroup(false);
    }

    return () => {
      disconnect();
      setHasJoinedGroup(false);
    };
  }, [groupId, memberId, connect, disconnect]);

  // Join the group once connected
  useEffect(() => {
    if (isConnected && groupId && memberId && !hasJoinedGroup) {
      joinGroup(groupId, memberId);
      setHasJoinedGroup(true);
    }
  }, [isConnected, groupId, memberId, hasJoinedGroup, joinGroup]);

  // Register event listeners
  useEffect(() => {
    const unsubscribers: (() => void)[] = [];

    if (onMemberJoined) {
      unsubscribers.push(registerMemberJoinedListener(onMemberJoined));
    }

    if (onItemAdded) {
      unsubscribers.push(registerItemAddedListener(onItemAdded));
    }

    if (onItemUpdated) {
      unsubscribers.push(registerItemUpdatedListener(onItemUpdated));
    }

    if (onError) {
      unsubscribers.push(registerErrorListener(onError));
    }

    return () => {
      unsubscribers.forEach((unsubscribe) => unsubscribe());
    };
  }, [
    onMemberJoined,
    onItemAdded,
    onItemUpdated,
    onError,
    registerMemberJoinedListener,
    registerItemAddedListener,
    registerItemUpdatedListener,
    registerErrorListener,
  ]);

  // Reconnect function
  const reconnect = useCallback(() => {
    wsReconnect();
    setHasJoinedGroup(false);
  }, [wsReconnect]);

  // Convenient wrapper for adding items
  const addItem = useCallback(
    (item: { title: string; description?: string }) => {
      if (groupId && memberId) {
        wsAddItem(groupId, { ...item, memberId });
      }
    },
    [groupId, memberId, wsAddItem]
  );

  // Convenient wrapper for toggling completion
  const toggleCompletion = useCallback(
    (itemId: string, completed: boolean) => {
      if (groupId && memberId) {
        wsToggleCompletion(groupId, itemId, completed, memberId);
      }
    },
    [groupId, memberId, wsToggleCompletion]
  );

  return {
    connectionState,
    isConnected,
    isOnline,
    error,
    reconnectAttempts,
    reconnect,
    addItem,
    toggleCompletion,
  };
};
