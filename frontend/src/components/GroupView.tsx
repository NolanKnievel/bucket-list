import React, {
  useState,
  useEffect,
  useCallback,
  useRef,
  useMemo,
} from "react";
import { useParams, useNavigate } from "react-router-dom";
import { apiService, ApiError } from "../utils/api";

import { GroupWithDetails, BucketListItem, Member } from "../types";
import { MembersList } from "./MembersList";
import { ProgressBar } from "./ProgressBar";
import { CountdownTimer } from "./CountdownTimer";
import { ErrorBoundary } from "./ErrorBoundary";
import { BucketListItem as BucketListItemComponent } from "./BucketListItem";

import { AddItemFormWithWebSocket } from "./AddItemFormWithWebSocket";
import { ConnectionStatus } from "./ConnectionStatus";
import { useAuth } from "../contexts/AuthContext";
import { useNativeWebSocketConnection } from "../hooks/useNativeWebSocketConnection";
import { useRealTimeProgress } from "../hooks/useRealTimeProgress";

export const GroupView: React.FC = () => {
  const { groupId } = useParams<{ groupId: string }>();
  const navigate = useNavigate();
  const { user, loading: authLoading } = useAuth();

  const [group, setGroup] = useState<GroupWithDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddForm, setShowAddForm] = useState(false);

  // Track optimistic updates for rollback functionality
  const optimisticUpdatesRef = useRef<Map<string, any>>(new Map());
  const [wsError, setWsError] = useState<string | null>(null);

  // Get current member ID based on authenticated user
  const getCurrentMemberId = useCallback(() => {
    // First, try to find member by authenticated user ID
    if (user && group?.members) {
      const currentMember = group.members.find(
        (member) => member.userId === user.id
      );
      if (currentMember) {
        return currentMember.id;
      }
    }

    // If no authenticated user or no matching member, check localStorage for guest member
    if (groupId && group?.members) {
      try {
        const storedMemberData = localStorage.getItem(`member_${groupId}`);
        if (storedMemberData) {
          const memberInfo = JSON.parse(storedMemberData);
          // Verify the member still exists in the group
          const guestMember = group.members.find(
            (member) => member.id === memberInfo.id
          );
          if (guestMember) {
            return guestMember.id;
          } else {
            // Member no longer exists, clean up localStorage
            localStorage.removeItem(`member_${groupId}`);
          }
        }
      } catch (error) {
        console.error("Error reading member info from localStorage:", error);
        // Clean up corrupted data
        if (groupId) {
          localStorage.removeItem(`member_${groupId}`);
        }
      }
    }

    return undefined;
  }, [user, group, groupId]);

  // Real-time event handlers
  const handleMemberJoined = useCallback((newMember: Member) => {
    console.log("Real-time: Member joined", newMember);
    setGroup((prevGroup) => {
      if (!prevGroup) return prevGroup;

      // Check if member already exists to avoid duplicates
      const memberExists = prevGroup.members.some((m) => m.id === newMember.id);
      if (memberExists) return prevGroup;

      return {
        ...prevGroup,
        members: [...prevGroup.members, newMember],
      };
    });
  }, []);

  const handleItemAdded = useCallback((newItem: BucketListItem) => {
    console.log("Real-time: Item added", newItem);
    setGroup((prevGroup) => {
      if (!prevGroup) return prevGroup;

      // Check if item already exists to avoid duplicates
      const items = prevGroup.items || [];
      const itemExists = items.some((item) => item.id === newItem.id);
      if (itemExists) return prevGroup;

      return {
        ...prevGroup,
        items: [newItem, ...items], // Add new item at the beginning
      };
    });
  }, []);

  const handleItemUpdated = useCallback((updatedItem: BucketListItem) => {
    console.log("Real-time: Item updated", updatedItem);
    setGroup((prevGroup) => {
      if (!prevGroup) return prevGroup;

      const updatedItems = (prevGroup.items || []).map((item) =>
        item.id === updatedItem.id ? updatedItem : item
      );

      return {
        ...prevGroup,
        items: updatedItems,
      };
    });
  }, []);

  const handleWebSocketError = useCallback(
    (error: { code: string; message: string; details?: string }) => {
      console.error("WebSocket error:", error);
      setWsError(`${error.code}: ${error.message}`);

      // Clear error after 5 seconds
      setTimeout(() => setWsError(null), 5000);
    },
    []
  );

  // Get current member ID
  const currentMemberId = getCurrentMemberId();

  // Debug logging
  console.log("WebSocket connection check:", {
    groupId,
    currentMemberId,
    hasUser: !!user,
    hasGroup: !!group,
    hasMembers: !!group?.members,
    authLoading,
  });

  // WebSocket connection - always call the hook to avoid Rules of Hooks violation
  let webSocket: ReturnType<typeof useNativeWebSocketConnection> | null = null;
  try {
    // Always call the hook, but pass empty values if data isn't ready
    // Wait for auth to complete and ensure we have all required data
    const shouldConnect = !!(groupId && currentMemberId && !authLoading);

    if (shouldConnect) {
      console.log("Establishing WebSocket connection with:", {
        groupId,
        currentMemberId,
      });
    } else {
      console.log(
        "WebSocket hook called but connection skipped - missing required data or auth loading"
      );
    }

    webSocket = useNativeWebSocketConnection({
      groupId: groupId || "",
      memberId: currentMemberId || "",
      onMemberJoined: handleMemberJoined,
      onItemAdded: handleItemAdded,
      onItemUpdated: handleItemUpdated,
      onError: handleWebSocketError,
    });
  } catch (error) {
    // WebSocket not available (e.g., in tests without provider)
    console.warn("WebSocket not available:", error);
    webSocket = {
      connectionState: "disconnected" as any,
      isConnected: false,
      isOnline: true,
      error: null,
      reconnectAttempts: 0,
      reconnect: () => {},
      addItem: () => {},
      toggleCompletion: () => {},
    };
  }

  useEffect(() => {
    const loadGroup = async () => {
      console.log("GroupView: Loading group with ID:", groupId);

      if (!groupId) {
        console.log("GroupView: No group ID provided");
        setError("Invalid group ID");
        setLoading(false);
        return;
      }

      try {
        setLoading(true);
        setError(null);
        console.log("GroupView: Calling API to get group");
        const groupData = await apiService.getGroup(groupId);
        console.log("GroupView: Group data received:", groupData);
        setGroup(groupData);
      } catch (err) {
        console.log("GroupView: Error loading group:", err);
        if (err instanceof ApiError) {
          switch (err.code) {
            case "GROUP_NOT_FOUND":
              setError("This group doesn't exist or has been deleted");
              break;
            case "INVALID_GROUP_ID":
              setError("Invalid group link format");
              break;
            default:
              setError("Failed to load group details. Please try again.");
          }
        } else {
          setError("Failed to load group details. Please try again.");
        }
      } finally {
        setLoading(false);
      }
    };

    loadGroup();
  }, [groupId]);

  // Get real-time progress data - stabilize the items array
  const items = useMemo(() => {
    return group && group.items ? group.items : [];
  }, [group]);

  const progressData = useRealTimeProgress({
    items,
  });

  // Handle toggling item completion with optimistic updates
  const handleToggleCompletion = async (itemId: string, completed: boolean) => {
    if (!currentMemberId) {
      throw new Error("You must be a member of this group to toggle items");
    }

    // Store original item state for potential rollback
    const originalItem = group?.items.find((item) => item.id === itemId);
    if (!originalItem) return;

    const optimisticUpdateId = `toggle-${itemId}-${Date.now()}`;
    optimisticUpdatesRef.current.set(optimisticUpdateId, originalItem);

    // Optimistic update - immediately update UI
    const optimisticItem: BucketListItem = {
      ...originalItem,
      completed,
      completedBy: completed ? currentMemberId : undefined,
      completedAt: completed ? new Date().toISOString() : undefined,
    };

    setGroup((prevGroup) => {
      if (!prevGroup) return prevGroup;

      const updatedItems = (prevGroup.items || []).map((item) =>
        item.id === itemId ? optimisticItem : item
      );

      return {
        ...prevGroup,
        items: updatedItems,
      };
    });

    try {
      // Send WebSocket event for real-time updates to other clients
      if (webSocket?.isConnected) {
        webSocket.toggleCompletion(itemId, completed);
      }

      // Also call API for persistence (backend will handle WebSocket broadcast)
      await apiService.toggleItemCompletion(itemId, {
        completed,
        memberId: currentMemberId,
      });

      // Success - remove optimistic update tracking
      optimisticUpdatesRef.current.delete(optimisticUpdateId);
    } catch (error) {
      console.error("Failed to toggle item completion:", error);

      // Rollback optimistic update
      setGroup((prevGroup) => {
        if (!prevGroup) return prevGroup;

        const rolledBackItems = (prevGroup.items || []).map((item) =>
          item.id === itemId ? originalItem : item
        );

        return {
          ...prevGroup,
          items: rolledBackItems,
        };
      });

      optimisticUpdatesRef.current.delete(optimisticUpdateId);
      throw error;
    }
  };

  // Handle adding new item from form (legacy callback)
  const handleItemAddedFromForm = async (_newItem: BucketListItem) => {
    // This is called from AddItemForm after successful API call
    // The real-time update will be handled by WebSocket events

    // Hide the form after successful addition
    setShowAddForm(false);
  };

  // Handle adding new item with WebSocket integration
  const handleAddItemWithWebSocket = async (itemData: {
    title: string;
    description?: string;
  }) => {
    if (!currentMemberId || !groupId) return;

    const optimisticUpdateId = `add-item-${Date.now()}`;

    // Create optimistic item
    const optimisticItem: BucketListItem = {
      id: `temp-${Date.now()}`, // Temporary ID
      groupId,
      title: itemData.title,
      description: itemData.description,
      completed: false,
      createdBy: currentMemberId,
      createdAt: new Date().toISOString(),
    };

    optimisticUpdatesRef.current.set(optimisticUpdateId, null); // No original to rollback to

    // Optimistic update - immediately add to UI
    setGroup((prevGroup) => {
      if (!prevGroup) return prevGroup;

      return {
        ...prevGroup,
        items: [optimisticItem, ...(prevGroup.items || [])],
      };
    });

    try {
      // Send WebSocket event for real-time updates
      if (webSocket?.isConnected) {
        webSocket.addItem(itemData);
      }

      // Call API for persistence
      const actualItem = await apiService.createItem(groupId, {
        ...itemData,
        memberId: currentMemberId,
      });

      // Replace optimistic item with actual item
      setGroup((prevGroup) => {
        if (!prevGroup) return prevGroup;

        const updatedItems = (prevGroup.items || []).map((item) =>
          item.id === optimisticItem.id ? actualItem : item
        );

        return {
          ...prevGroup,
          items: updatedItems,
        };
      });

      optimisticUpdatesRef.current.delete(optimisticUpdateId);
      setShowAddForm(false);
    } catch (error) {
      console.error("Failed to add item:", error);

      // Rollback optimistic update
      setGroup((prevGroup) => {
        if (!prevGroup) return prevGroup;

        const rolledBackItems = (prevGroup.items || []).filter(
          (item) => item.id !== optimisticItem.id
        );

        return {
          ...prevGroup,
          items: rolledBackItems,
        };
      });

      optimisticUpdatesRef.current.delete(optimisticUpdateId);
      throw error;
    }
  };

  console.log(
    "GroupView: Render state - loading:",
    loading,
    "error:",
    error,
    "group:",
    group
  );

  // Loading state
  if (loading) {
    console.log("GroupView: Rendering loading state");
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center">
        <div className="text-center">
          <div
            className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"
            role="status"
            aria-label="Loading group"
          ></div>
          <p className="mt-4 text-gray-600">Loading group details...</p>
        </div>
      </div>
    );
  }

  // Error state
  if (error) {
    console.log("GroupView: Rendering error state");
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full text-center">
          <div className="mx-auto h-12 w-12 text-red-500 mb-4">
            <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 18.5c-.77.833.192 2.5 1.732 2.5z"
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">
            Unable to Load Group
          </h2>
          <p className="text-gray-600 mb-6">{error}</p>
          <div className="space-y-3">
            <button
              onClick={() => window.location.reload()}
              className="w-full bg-blue-600 hover:bg-blue-700 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-200"
            >
              Try Again
            </button>
            <button
              onClick={() => navigate("/")}
              className="w-full bg-gray-300 hover:bg-gray-400 text-gray-700 font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition duration-200"
            >
              Go to Home
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Main group view
  if (!group || !group.members) {
    console.log("GroupView: Returning null - group check failed", {
      hasGroup: !!group,
      hasMembers: !!group?.members,
      hasItems: !!group?.items,
    });
    return null;
  }

  // items is already defined above with useMemo

  console.log("GroupView: Rendering main content");

  // Generate share link
  const shareLink = `${window.location.origin}/groups/${group.id}/join`;

  const copyShareLink = async () => {
    try {
      await navigator.clipboard.writeText(shareLink);
      // TODO: Show toast notification for successful copy
      console.log("Share link copied to clipboard");
    } catch (error) {
      console.error("Failed to copy share link:", error);
      // Fallback: select the text
      const linkElement = document.getElementById("group-share-link");
      if (linkElement) {
        const range = document.createRange();
        range.selectNode(linkElement);
        window.getSelection()?.removeAllRanges();
        window.getSelection()?.addRange(range);
      }
    }
  };

  return (
    <ErrorBoundary>
      <div className="min-h-screen bg-gray-50">
        <div className="container mx-auto px-4 py-6 max-w-6xl">
          {/* Header */}
          <header className="mb-8">
            <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between mb-6">
              <div className="mb-4 sm:mb-0">
                <h1 className="text-3xl font-bold text-gray-900 mb-2">
                  {group.name}
                </h1>
                <div className="flex flex-wrap items-center gap-4 text-sm text-gray-600">
                  <span>
                    {group.members?.length || 0} member
                    {(group.members?.length || 0) !== 1 ? "s" : ""}
                  </span>
                  <span>
                    {group.items?.length || 0} item
                    {(group.items?.length || 0) !== 1 ? "s" : ""}
                  </span>
                  <span>
                    Created {new Date(group.createdAt).toLocaleDateString()}
                  </span>
                  {/* WebSocket Connection Status */}
                  {webSocket && (
                    <ConnectionStatus
                      connectionState={webSocket.connectionState}
                      isOnline={webSocket.isOnline}
                      reconnectAttempts={webSocket.reconnectAttempts}
                      error={webSocket.error}
                      onReconnect={webSocket.reconnect}
                      className="text-xs"
                    />
                  )}
                </div>
              </div>
              <button
                onClick={() => navigate("/dashboard")}
                className="bg-gray-500 hover:bg-gray-600 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition duration-200"
              >
                Back to Dashboard
              </button>
            </div>

            {/* WebSocket Error Display */}
            {wsError && (
              <div className="mb-4 bg-yellow-50 border border-yellow-200 rounded-lg p-4">
                <div className="flex">
                  <div className="flex-shrink-0">
                    <svg
                      className="h-5 w-5 text-yellow-400"
                      viewBox="0 0 20 20"
                      fill="currentColor"
                    >
                      <path
                        fillRule="evenodd"
                        d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z"
                        clipRule="evenodd"
                      />
                    </svg>
                  </div>
                  <div className="ml-3">
                    <h3 className="text-sm font-medium text-yellow-800">
                      Connection Issue
                    </h3>
                    <p className="mt-1 text-sm text-yellow-700">{wsError}</p>
                    {webSocket && !webSocket.isConnected && (
                      <button
                        onClick={webSocket.reconnect}
                        className="mt-2 text-sm text-yellow-800 underline hover:text-yellow-900"
                      >
                        Try to reconnect
                      </button>
                    )}
                  </div>
                </div>
              </div>
            )}

            {/* Progress and Countdown Section */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
              {/* Progress Bar */}
              <div className="bg-white rounded-lg shadow-sm p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">
                  Progress
                </h3>
                <ErrorBoundary>
                  <ProgressBar {...progressData} />
                </ErrorBoundary>
              </div>

              {/* Countdown Timer */}
              {group.deadline && (
                <div className="bg-white rounded-lg shadow-sm p-6">
                  <h3 className="text-lg font-semibold text-gray-900 mb-4">
                    Time Remaining
                  </h3>
                  <ErrorBoundary>
                    <CountdownTimer
                      deadline={group.deadline}
                      createdAt={group.createdAt}
                    />
                  </ErrorBoundary>
                </div>
              )}
            </div>

            {/* Share Link Section */}
            <div className="bg-white rounded-lg shadow-sm p-6 mb-6">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">
                Invite Friends
              </h3>
              <div className="mb-2">
                <label className="block text-sm font-medium text-gray-700 mb-2">
                  Share this link with your friends:
                </label>
                <div className="flex items-center space-x-2">
                  <input
                    id="group-share-link"
                    type="text"
                    value={shareLink}
                    readOnly
                    className="flex-1 px-3 py-2 border border-gray-300 rounded-md bg-gray-50 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                  />
                  <button
                    onClick={copyShareLink}
                    className="px-3 py-2 bg-blue-500 text-white rounded-md hover:bg-blue-600 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-200"
                    title="Copy to clipboard"
                  >
                    <svg
                      className="h-4 w-4"
                      fill="none"
                      viewBox="0 0 24 24"
                      stroke="currentColor"
                    >
                      <path
                        strokeLinecap="round"
                        strokeLinejoin="round"
                        strokeWidth={2}
                        d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z"
                      />
                    </svg>
                  </button>
                </div>
                <p className="text-xs text-gray-500 mt-1">
                  Anyone with this link can join your group and add items to the
                  bucket list.
                </p>
              </div>
            </div>
          </header>

          {/* Main Content */}
          <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
            {/* Members List */}
            <div className="lg:col-span-1">
              <div className="bg-white rounded-lg shadow-sm p-6">
                <h2 className="text-xl font-semibold text-gray-900 mb-4">
                  Members
                </h2>
                <ErrorBoundary>
                  <MembersList members={group.members} />
                </ErrorBoundary>
              </div>
            </div>

            {/* Bucket List Items */}
            <div className="lg:col-span-2">
              <div className="bg-white rounded-lg shadow-sm p-6">
                <div className="flex items-center justify-between mb-4">
                  <h2 className="text-xl font-semibold text-gray-900">
                    Bucket List
                  </h2>
                  {currentMemberId && !showAddForm && (
                    <button
                      onClick={() => setShowAddForm(true)}
                      className="bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-200 flex items-center"
                    >
                      <svg
                        className="h-4 w-4 mr-2"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M12 4v16m8-8H4"
                        />
                      </svg>
                      Add Item
                    </button>
                  )}
                </div>

                {/* Add Item Form */}
                {showAddForm && currentMemberId && (
                  <ErrorBoundary>
                    <AddItemFormWithWebSocket
                      groupId={group.id}
                      memberId={currentMemberId}
                      onItemAdded={handleItemAddedFromForm}
                      onCancel={() => setShowAddForm(false)}
                      onAddItemWithWebSocket={handleAddItemWithWebSocket}
                    />
                  </ErrorBoundary>
                )}

                <ErrorBoundary>
                  {items.length === 0 && !showAddForm ? (
                    <div className="text-center py-12">
                      <svg
                        className="mx-auto h-12 w-12 text-gray-400 mb-4"
                        fill="none"
                        viewBox="0 0 24 24"
                        stroke="currentColor"
                      >
                        <path
                          strokeLinecap="round"
                          strokeLinejoin="round"
                          strokeWidth={2}
                          d="M9 5H7a2 2 0 00-2 2v10a2 2 0 002 2h8a2 2 0 002-2V7a2 2 0 00-2-2h-2M9 5a2 2 0 002 2h2a2 2 0 002-2M9 5a2 2 0 012-2h2a2 2 0 012 2"
                        />
                      </svg>
                      <h3 className="text-lg font-medium text-gray-900 mb-2">
                        No items yet
                      </h3>
                      <p className="text-gray-600">
                        Be the first to add an item to this bucket list!
                      </p>
                    </div>
                  ) : (
                    <div className="space-y-4">
                      {items
                        .sort(
                          (a, b) =>
                            new Date(b.createdAt).getTime() -
                            new Date(a.createdAt).getTime()
                        )
                        .map((item) => (
                          <BucketListItemComponent
                            key={item.id}
                            item={item}
                            members={group.members}
                            currentMemberId={currentMemberId}
                            onToggleCompletion={handleToggleCompletion}
                          />
                        ))}
                    </div>
                  )}
                </ErrorBoundary>
              </div>
            </div>
          </div>
        </div>
      </div>
    </ErrorBoundary>
  );
};
