import React, {
  createContext,
  useContext,
  useEffect,
  useState,
  useCallback,
  useRef,
} from "react";
import { io, Socket } from "socket.io-client";
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

// Socket.IO connection options
interface SocketOptions {
  groupId: string;
  memberId: string;
}

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
// WebSocket context interface
interface WebSocketContextType {
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

const WebSocketContext = createContext<WebSocketContextType | null>(null);

// WebSocket provider component
export const WebSocketProvider: React.FC<{ children: React.ReactNode }> = ({
  children,
}) => {
  const [connectionState, setConnectionState] = useState<ConnectionState>(
    ConnectionState.DISCONNECTED
  );
  const [error, setError] = useState<string | null>(null);
  const [isOnline, setIsOnline] = useState(navigator.onLine);

  const socketRef = useRef<Socket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const lastConnectionOptionsRef = useRef<SocketOptions | null>(null);
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

  // Get Socket.IO server URL from environment or default
  const getSocketUrl = useCallback(() => {
    const protocol = window.location.protocol === "https:" ? "https:" : "http:";
    const host = import.meta.env.VITE_WS_HOST || window.location.host;
    return `${protocol}//${host}`;
  }, []);

  // Send message via Socket.IO
  const sendMessage = useCallback((eventType: string, data: any) => {
    if (socketRef.current && socketRef.current.connected) {
      socketRef.current.emit(eventType, data);
    } else {
      console.warn(
        "Socket.IO is not connected. Cannot send message:",
        eventType,
        data
      );
    }
  }, []); // Setup Socket.IO event listeners
  const setupSocketListeners = useCallback((socket: Socket) => {
    // Handle member joined events
    socket.on(WS_EVENTS.MEMBER_JOINED, (member: Member) => {
      memberJoinedListeners.current.forEach((listener) => listener(member));
    });

    // Handle item added events
    socket.on(WS_EVENTS.ITEM_ADDED, (item: BucketListItem) => {
      itemAddedListeners.current.forEach((listener) => listener(item));
    });

    // Handle item updated events
    socket.on(WS_EVENTS.ITEM_UPDATED, (item: BucketListItem) => {
      itemUpdatedListeners.current.forEach((listener) => listener(item));
    });

    // Handle error events
    socket.on(WS_EVENTS.ERROR, (errorData: ErrorPayload) => {
      setError(`${errorData.code}: ${errorData.message}`);
      errorListeners.current.forEach((listener) => listener(errorData));
    });

    // Handle generic errors
    socket.on("error", (err: any) => {
      console.error("Socket.IO error:", err);
      setError("Socket.IO connection error");
      setConnectionState(ConnectionState.ERROR);
    });
  }, []);

  // Handle Socket.IO connection events
  const handleConnect = useCallback(() => {
    console.log("Socket.IO connected");
    setConnectionState(ConnectionState.CONNECTED);
    setError(null);
    reconnectAttemptsRef.current = 0;
  }, []);

  const handleDisconnect = useCallback((reason: string) => {
    console.log("Socket.IO disconnected:", reason);
    setConnectionState(ConnectionState.DISCONNECTED);

    // Socket.IO handles automatic reconnection, but we track the state
    if (reason === "io server disconnect") {
      // Server initiated disconnect, don't try to reconnect
      setError("Server disconnected");
      setConnectionState(ConnectionState.ERROR);
    }
  }, []);

  const handleConnectError = useCallback((error: Error) => {
    console.error("Socket.IO connection error:", error);
    setError(`Connection failed: ${error.message}`);
    setConnectionState(ConnectionState.ERROR);
  }, []);

  const handleReconnect = useCallback((attemptNumber: number) => {
    console.log(`Socket.IO reconnected after ${attemptNumber} attempts`);
    setConnectionState(ConnectionState.CONNECTED);
    setError(null);
    reconnectAttemptsRef.current = 0;
  }, []);

  const handleReconnectAttempt = useCallback((attemptNumber: number) => {
    console.log(`Socket.IO reconnection attempt ${attemptNumber}`);
    setConnectionState(ConnectionState.RECONNECTING);
    reconnectAttemptsRef.current = attemptNumber;
  }, []);

  const handleReconnectError = useCallback((error: Error) => {
    console.error("Socket.IO reconnection error:", error);
    setError(`Reconnection failed: ${error.message}`);
  }, []);

  const handleReconnectFailed = useCallback(() => {
    console.error("Socket.IO reconnection failed after all attempts");
    setError("Failed to reconnect after maximum attempts");
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
  }, [handleOnline, handleOffline]); // Connect to Socket.IO server
  const connect = useCallback(
    (groupId: string, memberId: string) => {
      // Disconnect existing socket if any
      if (socketRef.current) {
        socketRef.current.disconnect();
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

        const socketUrl = getSocketUrl();
        const socket = io(socketUrl, {
          query: {
            groupId,
            memberId,
          },
          transports: ["websocket", "polling"], // Fallback to polling if websocket fails
          timeout: 10000, // 10 second timeout
          reconnection: true,
          reconnectionAttempts: maxReconnectAttempts,
          reconnectionDelay: reconnectDelay,
          reconnectionDelayMax: 5000,
          maxReconnectionAttempts: maxReconnectAttempts,
          forceNew: true, // Force a new connection
        });

        // Setup connection event listeners
        socket.on("connect", handleConnect);
        socket.on("disconnect", handleDisconnect);
        socket.on("connect_error", handleConnectError);
        socket.on("reconnect", handleReconnect);
        socket.on("reconnect_attempt", handleReconnectAttempt);
        socket.on("reconnect_error", handleReconnectError);
        socket.on("reconnect_failed", handleReconnectFailed);

        // Setup application event listeners
        setupSocketListeners(socket);

        socketRef.current = socket;
      } catch (err) {
        console.error("Failed to create Socket.IO connection:", err);
        setError("Failed to create Socket.IO connection");
        setConnectionState(ConnectionState.ERROR);
      }
    },
    [
      getSocketUrl,
      handleConnect,
      handleDisconnect,
      handleConnectError,
      handleReconnect,
      handleReconnectAttempt,
      handleReconnectError,
      handleReconnectFailed,
      setupSocketListeners,
    ]
  );

  // Disconnect from Socket.IO server
  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (socketRef.current) {
      socketRef.current.disconnect();
      socketRef.current = null;
    }

    setConnectionState(ConnectionState.DISCONNECTED);
    setError(null);
    reconnectAttemptsRef.current = 0;
    lastConnectionOptionsRef.current = null;
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

  // Socket.IO event methods
  const joinGroup = useCallback(
    (groupId: string, memberId: string) => {
      const payload: JoinGroupPayload = { groupId, memberId };
      sendMessage(WS_EVENTS.JOIN_GROUP, payload);
    },
    [sendMessage]
  );

  const addItem = useCallback(
    (
      groupId: string,
      item: { title: string; description?: string; memberId: string }
    ) => {
      const payload: AddItemPayload = { groupId, item };
      sendMessage(WS_EVENTS.ADD_ITEM, payload);
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
      sendMessage(WS_EVENTS.TOGGLE_COMPLETION, payload);
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

  const contextValue: WebSocketContextType = {
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
    <WebSocketContext.Provider value={contextValue}>
      {children}
    </WebSocketContext.Provider>
  );
};

// Custom hook to use WebSocket context
export const useWebSocket = (): WebSocketContextType => {
  const context = useContext(WebSocketContext);
  if (!context) {
    throw new Error("useWebSocket must be used within a WebSocketProvider");
  }
  return context;
};
