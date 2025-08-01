import React, { useState } from "react";
import { BucketListItem as BucketListItemType, Member } from "../types";

interface BucketListItemProps {
  item: BucketListItemType;
  members: Member[];
  currentMemberId?: string;
  onToggleCompletion: (itemId: string, completed: boolean) => Promise<void>;
}

export const BucketListItem: React.FC<BucketListItemProps> = ({
  item,
  members,
  currentMemberId,
  onToggleCompletion,
}) => {
  const [isToggling, setIsToggling] = useState(false);

  const handleToggleCompletion = async () => {
    if (!currentMemberId || isToggling) return;

    setIsToggling(true);
    try {
      await onToggleCompletion(item.id, !item.completed);
    } catch (error) {
      console.error("Failed to toggle completion:", error);
      // TODO: Show error toast
    } finally {
      setIsToggling(false);
    }
  };

  // Find the member who created this item
  const creator = members.find((m) => m.id === item.createdBy);
  const creatorName = creator?.name || "Unknown";

  // Find the member who completed this item (if completed)
  const completedByMember = item.completedBy
    ? members.find((m) => m.id === item.completedBy)
    : null;

  return (
    <div
      className={`p-4 border rounded-lg transition-colors duration-200 ${
        item.completed
          ? "bg-green-50 border-green-200"
          : "bg-white border-gray-200 hover:border-gray-300"
      }`}
    >
      <div className="flex items-start justify-between">
        <div className="flex-1 min-w-0">
          {/* Title */}
          <h4
            className={`font-medium text-lg leading-tight ${
              item.completed ? "text-green-800 line-through" : "text-gray-900"
            }`}
          >
            {item.title}
          </h4>

          {/* Description */}
          {item.description && (
            <p
              className={`mt-2 text-sm leading-relaxed ${
                item.completed ? "text-green-600" : "text-gray-600"
              }`}
            >
              {item.description}
            </p>
          )}

          {/* Metadata */}
          <div className="mt-3 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-gray-500">
            <span className="flex items-center">
              <svg
                className="w-3 h-3 mr-1"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M16 7a4 4 0 11-8 0 4 4 0 018 0zM12 14a7 7 0 00-7 7h14a7 7 0 00-7-7z"
                />
              </svg>
              Added by {creatorName}
            </span>

            <span className="flex items-center">
              <svg
                className="w-3 h-3 mr-1"
                fill="none"
                stroke="currentColor"
                viewBox="0 0 24 24"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
                />
              </svg>
              {new Date(item.createdAt).toLocaleDateString()}
            </span>

            {item.completed && item.completedAt && (
              <>
                <span className="text-green-600 flex items-center">
                  <svg
                    className="w-3 h-3 mr-1"
                    fill="currentColor"
                    viewBox="0 0 20 20"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                      clipRule="evenodd"
                    />
                  </svg>
                  Completed {new Date(item.completedAt).toLocaleDateString()}
                  {completedByMember && ` by ${completedByMember.name}`}
                </span>
              </>
            )}
          </div>
        </div>

        {/* Completion Toggle Button */}
        <div className="ml-4 flex-shrink-0">
          {currentMemberId ? (
            <button
              onClick={handleToggleCompletion}
              disabled={isToggling}
              className={`p-2 rounded-full transition-all duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2 ${
                item.completed
                  ? "text-green-600 hover:text-green-700 hover:bg-green-100 focus:ring-green-500"
                  : "text-gray-400 hover:text-gray-600 hover:bg-gray-100 focus:ring-gray-500"
              } ${
                isToggling ? "opacity-50 cursor-not-allowed" : "cursor-pointer"
              }`}
              title={item.completed ? "Mark as incomplete" : "Mark as complete"}
              aria-label={
                item.completed ? "Mark as incomplete" : "Mark as complete"
              }
            >
              {isToggling ? (
                <div className="animate-spin h-5 w-5 border-2 border-current border-t-transparent rounded-full" />
              ) : item.completed ? (
                <svg
                  className="h-5 w-5"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
              ) : (
                <svg
                  className="h-5 w-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <circle cx="12" cy="12" r="10" strokeWidth={2} />
                </svg>
              )}
            </button>
          ) : (
            <div
              className={`p-2 ${
                item.completed ? "text-green-600" : "text-gray-400"
              }`}
              title={item.completed ? "Completed" : "Not completed"}
            >
              {item.completed ? (
                <svg
                  className="h-5 w-5"
                  fill="currentColor"
                  viewBox="0 0 20 20"
                >
                  <path
                    fillRule="evenodd"
                    d="M10 18a8 8 0 100-16 8 8 0 000 16zm3.707-9.293a1 1 0 00-1.414-1.414L9 10.586 7.707 9.293a1 1 0 00-1.414 1.414l2 2a1 1 0 001.414 0l4-4z"
                    clipRule="evenodd"
                  />
                </svg>
              ) : (
                <svg
                  className="h-5 w-5"
                  fill="none"
                  stroke="currentColor"
                  viewBox="0 0 24 24"
                >
                  <circle cx="12" cy="12" r="10" strokeWidth={2} />
                </svg>
              )}
            </div>
          )}
        </div>
      </div>
    </div>
  );
};
