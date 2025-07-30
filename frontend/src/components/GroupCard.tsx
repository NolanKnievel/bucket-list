import React from "react";
import { GroupSummary } from "../types";

interface GroupCardProps {
  group: GroupSummary;
  onViewGroup: (groupId: string) => void;
}

export const GroupCard: React.FC<GroupCardProps> = ({ group, onViewGroup }) => {
  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString();
  };

  const formatDeadline = (deadline?: string) => {
    if (!deadline) return null;

    const deadlineDate = new Date(deadline);
    const now = new Date();
    const diffTime = deadlineDate.getTime() - now.getTime();
    const diffDays = Math.ceil(diffTime / (1000 * 60 * 60 * 24));

    if (diffDays < 0) {
      return { text: "Expired", color: "text-red-600" };
    } else if (diffDays === 0) {
      return { text: "Due today", color: "text-red-600" };
    } else if (diffDays === 1) {
      return { text: "Due tomorrow", color: "text-orange-600" };
    } else if (diffDays <= 7) {
      return { text: `${diffDays} days left`, color: "text-orange-600" };
    } else {
      return { text: `${diffDays} days left`, color: "text-gray-600" };
    }
  };

  const deadline = formatDeadline(group.deadline);

  return (
    <div className="bg-white rounded-lg shadow-md hover:shadow-lg transition-shadow duration-200 p-6">
      <div className="flex justify-between items-start mb-4">
        <div className="flex-1">
          <h3 className="text-xl font-semibold text-gray-900 mb-2">
            {group.name}
          </h3>
          <div className="flex items-center space-x-4 text-sm text-gray-600">
            <span>
              {group.memberCount} member{group.memberCount !== 1 ? "s" : ""}
            </span>
            <span>
              {group.itemCount} item{group.itemCount !== 1 ? "s" : ""}
            </span>
            <span>Created {formatDate(group.createdAt)}</span>
          </div>
        </div>
        {deadline && (
          <div className={`text-sm font-medium ${deadline.color}`}>
            {deadline.text}
          </div>
        )}
      </div>

      {/* Progress Section */}
      <div className="mb-4">
        <div className="flex justify-between items-center mb-2">
          <span className="text-sm font-medium text-gray-700">Progress</span>
          <span className="text-sm text-gray-600">
            {group.completedCount}/{group.itemCount} completed (
            {Math.round(group.progressPercent)}%)
          </span>
        </div>
        <div className="w-full bg-gray-200 rounded-full h-2">
          <div
            className="bg-blue-600 h-2 rounded-full transition-all duration-300"
            style={{ width: `${group.progressPercent}%` }}
            role="progressbar"
            aria-valuenow={group.progressPercent}
            aria-valuemin={0}
            aria-valuemax={100}
          ></div>
        </div>
      </div>

      {/* Action Button */}
      <button
        onClick={() => onViewGroup(group.id)}
        className="w-full bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-200"
      >
        View Group
      </button>
    </div>
  );
};
