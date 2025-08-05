import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useRef,
} from "react";
import { BucketListItem, Member } from "../types";

// WebSocket connection states
export enum ConnectionState {
  DISCONNECTED = "disconnected",
  CONNECTING = "connecting",
  CONNECTED = "connected",
  RECONNECTING = "reconnecting",
  ERROR = "error",
}

// WebSocket event types (matching backend)
export const WS_EVENTS = {
  // Client to Server
  JOIN_GROUP: "join-group",
  ADD_ITEM: "add-item",
  TOGGLE_COMPLETION: "toggle-completion",

  // Server to Client
  MEMBER_JOINED: "member-joined",
  ITEM_ADDED: "item-added",
  ITEM_UPDATED: "item-updated",
  ERROR: "error",
} as const;

// Event payload types
interface JoinGroupPayload {
  groupId: string;
  memberId: string;
}

interface AddItemPayload {
  groupId: string;
  item: {
    title: string;
    description?: string;
    memberId: string;
  };
}

interface ToggleCompletionPayload {
  groupId: string;
  itemId: string;
  completed: boolean;
  memberId: string;
}

interface ErrorPayload {
  code: string;
  message: string;
  details?: string;
}

interface WebSocketMessage {
  type: string;
  roomId: string;
  memberId: string;
  data: any;
}

// WebSocket context interface
interface NativeWebSocketContextType {
  connectionState: ConnectionState;
  isConnected: boolean;
  error: string | null;
  connect: (groupId: string, memberId: string) => void;
  disconnect: () => void;
  joinGroup: (groupId: string, memberId: string) => void;
  addItem: (
    groupId: string,
    item: { title: string; description?: string; memberId: string }
  ) => void;
  toggleCompletion: (
    groupId: string,
    itemId: string,
    completed: boolean,
    memberId: string
  ) => void;
  isOnline: boolean;
  reconnectAttempts: number;
  reconnect: () => void;
  onMemberJoined: (callback: (member: Member) => void) => () => void;
  onItemAdded: (callback: (item: BucketListItem) => void) => () => void;
  onItemUpdated: (callback: (item: BucketListItem) => void) => () => void;
  onError: (callback: (error: ErrorPayload) => void) => () => void;
}

const NativeWebSocketContext = createContext<NativeWebSocketContextType | null>(
  null
);

// WebSocket provider component
export const NativeWebSocketProvider: React.FC<{
  children: React.ReactNode;
}> = ({ children }) => {
  const [connectionState, setConnectionState] = useState<ConnectionState>(
    ConnectionState.DISCONNECTED
  );
  const [error, setError] = useState<string | null>(null);
  const [isOnline, setIsOnline] = useState(navigator.onLine);

  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const lastConnectionOptionsRef = useRef<{
    groupId: string;
    memberId: string;
  } | null>(null);
  const currentMemberIdRef = useRef<string>("");
  const maxReconnectAttempts = 5;
  const reconnectDelay = 1000; // Start with 1 second

  // Event listeners
  const memberJoinedListeners = useRef<Set<(member: Member) => void>>(
    new Set()
  );
  const itemAddedListeners = useRef<Set<(item: BucketListItem) => void>>(
    new Set()
  );
  const itemUpdatedListeners = useRef<Set<(item: BucketListItem) => void>>(
    new Set()
  );
  const errorListeners = useRef<Set<(error: ErrorPayload) => void>>(new Set());

  // Get WebSocket server URL from environment or default
  const getWebSocketUrl = useCallback((groupId: string, memberId: string) => {
    const wsUrl = import.meta.env.VITE_WS_URL;
    if (wsUrl) {
      // Use the configured WebSocket URL and convert http/https to ws/wss
      const protocol = wsUrl.startsWith("https:") ? "wss:" : "ws:";
      const host = wsUrl.replace(/^https?:\/\//, "");
      return `${protocol}//${host}/api/ws/groups/${groupId}?memberId=${memberId}`;
    } else {
      // Fallback to localhost:8080
      const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
      return `${protocol}//localhost:8080/api/ws/groups/${groupId}?memberId=${memberId}`;
    }
  }, []);

  // Send message via WebSocket
  const sendMessage = useCallback(
    (messageType: string, data: any, roomId: string, memberId: string) => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        const message: WebSocketMessage = {
          type: messageType,
          roomId,
          memberId,
          data,
        };
        wsRef.current.send(JSON.stringify(message));
      } else {
        console.warn(
          "WebSocket is not connected. Cannot send message:",
          messageType,
          data
        );
      }
    },
    []
  );

  // Handle incoming WebSocket messages
  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);

      switch (message.type) {
        case WS_EVENTS.MEMBER_JOINED:
          memberJoinedListeners.current.forEach((listener) =>
            listener(message.data)
          );
          break;
        case WS_EVENTS.ITEM_ADDED:
          itemAddedListeners.current.forEach((listener) =>
            listener(message.data)
          );
          break;
        case WS_EVENTS.ITEM_UPDATED:
          itemUpdatedListeners.current.forEach((listener) =>
            listener(message.data)
          );
          break;
        case WS_EVENTS.ERROR:
          setError(`${message.data.code}: ${message.data.message}`);
          errorListeners.current.forEach((listener) => listener(message.data));
          break;
        default:
          console.log("Unknown message type:", message.type);
      }
    } catch (err) {
      console.error("Error parsing WebSocket message:", err);
    }
  }, []);

  // Handle WebSocket connection events
  const handleOpen = useCallback(() => {
    console.log("WebSocket connected");
    setConnectionState(ConnectionState.CONNECTED);
    setError(null);
    reconnectAttemptsRef.current = 0;
  }, []);

  const handleClose = useCallback((event: CloseEvent) => {
    console.log("WebSocket disconnected:", event.code, event.reason);
    setConnectionState(ConnectionState.DISCONNECTED);

    // Attempt to reconnect if it wasn't a clean close
    if (
      event.code !== 1000 &&
      reconnectAttemptsRef.current < maxReconnectAttempts
    ) {
      setConnectionState(ConnectionState.RECONNECTING);
      reconnectAttemptsRef.current++;

      const delay =
        reconnectDelay * Math.pow(2, reconnectAttemptsRef.current - 1);
      reconnectTimeoutRef.current = setTimeout(() => {
        if (lastConnectionOptionsRef.current) {
          const { groupId, memberId } = lastConnectionOptionsRef.current;
          connect(groupId, memberId);
        }
      }, delay);
    }
  }, []);

  const handleError = useCallback((event: Event) => {
    console.error("WebSocket error:", event);
    setError("WebSocket connection error");
    setConnectionState(ConnectionState.ERROR);
  }, []);

  // Handle online/offline events
  const handleOnline = useCallback(() => {
    console.log("Browser came online");
    setIsOnline(true);
    setError(null);

    // Attempt to reconnect if we have previous connection options
    if (
      lastConnectionOptionsRef.current &&
      connectionState !== ConnectionState.CONNECTED
    ) {
      const { groupId, memberId } = lastConnectionOptionsRef.current;
      connect(groupId, memberId);
    }
  }, [connectionState]);

  const handleOffline = useCallback(() => {
    console.log("Browser went offline");
    setIsOnline(false);
    setError("You are offline");
    setConnectionState(ConnectionState.DISCONNECTED);
  }, []);

  // Setup online/offline event listeners
  useEffect(() => {
    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, [handleOnline, handleOffline]);

  // Connect to WebSocket server
  const connect = useCallback(
    (groupId: string, memberId: string) => {
      // Disconnect existing WebSocket if any
      if (wsRef.current) {
        wsRef.current.close();
      }

      // Clear any pending reconnect timeout
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      try {
        setConnectionState(ConnectionState.CONNECTING);
        setError(null);

        // Store connection options for potential reconnection
        lastConnectionOptionsRef.current = { groupId, memberId };
        currentMemberIdRef.current = memberId;

        const wsUrl = getWebSocketUrl(groupId, memberId);
        console.log("Connecting to WebSocket:", wsUrl);

        const ws = new WebSocket(wsUrl);

        ws.onopen = handleOpen;
        ws.onclose = handleClose;
        ws.onerror = handleError;
        ws.onmessage = handleMessage;

        wsRef.current = ws;
      } catch (err) {
        console.error("Failed to create WebSocket connection:", err);
        setError("Failed to create WebSocket connection");
        setConnectionState(ConnectionState.ERROR);
      }
    },
    [getWebSocketUrl, handleOpen, handleClose, handleError, handleMessage]
  );

  // Disconnect from WebSocket server
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close(1000, "Client disconnect");
      wsRef.current = null;
    }

    setConnectionState(ConnectionState.DISCONNECTED);
    setError(null);
    reconnectAttemptsRef.current = 0;
    lastConnectionOptionsRef.current = null;
    currentMemberIdRef.current = "";
  }, []);

  // Manual reconnect method
  const reconnect = useCallback(() => {
    if (lastConnectionOptionsRef.current) {
      const { groupId, memberId } = lastConnectionOptionsRef.current;
      connect(groupId, memberId);
    } else {
      console.warn("No previous connection options available for reconnection");
      setError("Cannot reconnect: no previous connection found");
    }
  }, [connect]);

  // WebSocket event methods
  const joinGroup = useCallback(
    (groupId: string, memberId: string) => {
      const payload: JoinGroupPayload = { groupId, memberId };
      sendMessage(WS_EVENTS.JOIN_GROUP, payload, groupId, memberId);
    },
    [sendMessage]
  );

  const addItem = useCallback(
    (
      groupId: string,
      item: { title: string; description?: string; memberId: string }
    ) => {
      const payload: AddItemPayload = { groupId, item };
      sendMessage(WS_EVENTS.ADD_ITEM, payload, groupId, item.memberId);
    },
    [sendMessage]
  );

  const toggleCompletion = useCallback(
    (groupId: string, itemId: string, completed: boolean, memberId: string) => {
      const payload: ToggleCompletionPayload = {
        groupId,
        itemId,
        completed,
        memberId,
      };
      sendMessage(WS_EVENTS.TOGGLE_COMPLETION, payload, groupId, memberId);
    },
    [sendMessage]
  );

  // Event listener registration methods
  const onMemberJoined = useCallback((callback: (member: Member) => void) => {
    memberJoinedListeners.current.add(callback);
    return () => memberJoinedListeners.current.delete(callback);
  }, []);

  const onItemAdded = useCallback(
    (callback: (item: BucketListItem) => void) => {
      itemAddedListeners.current.add(callback);
      return () => itemAddedListeners.current.delete(callback);
    },
    []
  );

  const onItemUpdated = useCallback(
    (callback: (item: BucketListItem) => void) => {
      itemUpdatedListeners.current.add(callback);
      return () => itemUpdatedListeners.current.delete(callback);
    },
    []
  );

  const onError = useCallback((callback: (error: ErrorPayload) => void) => {
    errorListeners.current.add(callback);
    return () => errorListeners.current.delete(callback);
  }, []);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      disconnect();
    };
  }, [disconnect]);

  const contextValue: NativeWebSocketContextType = {
    connectionState,
    isConnected: connectionState === ConnectionState.CONNECTED,
    isOnline,
    error,
    reconnectAttempts: reconnectAttemptsRef.current,
    connect,
    disconnect,
    reconnect,
    joinGroup,
    addItem,
    toggleCompletion,
    onMemberJoined,
    onItemAdded,
    onItemUpdated,
    onError,
  };

  return (
    <NativeWebSocketContext.Provider value={contextValue}>
      {children}
    </NativeWebSocketContext.Provider>
  );
};

// Custom hook to use WebSocket context
export const useNativeWebSocket = (): NativeWebSocketContextType => {
  const context = useContext(NativeWebSocketContext);
  if (!context) {
    throw new Error(
      "useNativeWebSocket must be used within a NativeWebSocketProvider"
    );
  }
  return context;
};
