import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { apiService, ApiError } from "../utils/api";
import { GroupWithDetails, BucketListItem } from "../types";
import { MembersList } from "./MembersList";
import { ProgressBar } from "./ProgressBar";
import { CountdownTimer } from "./CountdownTimer";
import { ErrorBoundary } from "./ErrorBoundary";
import { BucketListItem as BucketListItemComponent } from "./BucketListItem";
import { AddItemForm } from "./AddItemForm";
import { useAuth } from "../contexts/AuthContext";

export const GroupView: React.FC = () => {
  const { groupId } = useParams<{ groupId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();

  const [group, setGroup] = useState<GroupWithDetails | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddForm, setShowAddForm] = useState(false);

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

  const calculateProgress = () => {
    if (!group) return 0;
    const items = group.items || [];
    if (items.length === 0) return 0;
    const completedItems = items.filter((item) => item.completed).length;
    return Math.round((completedItems / items.length) * 100);
  };

  const getCompletedCount = () => {
    if (!group) return 0;
    const items = group.items || [];
    return items.filter((item) => item.completed).length;
  };

  // Find current member ID based on authenticated user
  const getCurrentMemberId = () => {
    if (!user || !group?.members) return undefined;
    const currentMember = group.members.find(
      (member) => member.userId === user.id
    );
    return currentMember?.id;
  };

  // Handle toggling item completion
  const handleToggleCompletion = async (itemId: string, completed: boolean) => {
    const currentMemberId = getCurrentMemberId();
    if (!currentMemberId) {
      throw new Error("You must be a member of this group to toggle items");
    }

    try {
      const updatedItem = await apiService.toggleItemCompletion(itemId, {
        completed,
        memberId: currentMemberId,
      });

      // Update the local state with the updated item
      setGroup((prevGroup) => {
        if (!prevGroup) return prevGroup;

        const updatedItems = prevGroup.items.map((item) =>
          item.id === itemId ? updatedItem : item
        );

        return {
          ...prevGroup,
          items: updatedItems,
        };
      });
    } catch (error) {
      console.error("Failed to toggle item completion:", error);
      throw error;
    }
  };

  // Handle adding new item
  const handleItemAdded = (newItem: BucketListItem) => {
    setGroup((prevGroup) => {
      if (!prevGroup) return prevGroup;

      return {
        ...prevGroup,
        items: [newItem, ...prevGroup.items], // Add new item at the beginning
      };
    });

    // Hide the form after successful addition
    setShowAddForm(false);
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

  // Ensure items is always an array
  const items = group.items || [];

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
                </div>
              </div>
              <button
                onClick={() => navigate("/dashboard")}
                className="bg-gray-500 hover:bg-gray-600 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition duration-200"
              >
                Back to Dashboard
              </button>
            </div>

            {/* Progress and Countdown Section */}
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6 mb-6">
              {/* Progress Bar */}
              <div className="bg-white rounded-lg shadow-sm p-6">
                <h3 className="text-lg font-semibold text-gray-900 mb-4">
                  Progress
                </h3>
                <ErrorBoundary>
                  <ProgressBar
                    current={getCompletedCount()}
                    total={items.length}
                    percentage={calculateProgress()}
                  />
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
                  {getCurrentMemberId() && !showAddForm && (
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
                {showAddForm && getCurrentMemberId() && (
                  <ErrorBoundary>
                    <AddItemForm
                      groupId={group.id}
                      memberId={getCurrentMemberId()!}
                      onItemAdded={handleItemAdded}
                      onCancel={() => setShowAddForm(false)}
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
                            currentMemberId={getCurrentMemberId()}
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
