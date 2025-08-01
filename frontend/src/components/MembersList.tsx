import React from "react";
import { Member } from "../types";

interface MembersListProps {
  members: Member[];
}

export const MembersList: React.FC<MembersListProps> = ({ members }) => {
  if (members.length === 0) {
    return (
      <div className="text-center py-8">
        <svg
          className="mx-auto h-8 w-8 text-gray-400 mb-2"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            strokeLinecap="round"
            strokeLinejoin="round"
            strokeWidth={2}
            d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"
          />
        </svg>
        <p className="text-sm text-gray-500">No members yet</p>
      </div>
    );
  }

  // Sort members to show creator first, then by join date
  const sortedMembers = [...members].sort((a, b) => {
    if (a.isCreator && !b.isCreator) return -1;
    if (!a.isCreator && b.isCreator) return 1;
    return new Date(a.joinedAt).getTime() - new Date(b.joinedAt).getTime();
  });

  return (
    <div className="space-y-3">
      {sortedMembers.map((member) => (
        <div
          key={member.id}
          className="flex items-center justify-between p-3 bg-gray-50 rounded-lg"
        >
          <div className="flex items-center space-x-3">
            {/* Avatar */}
            <div className="flex-shrink-0">
              <div className="h-8 w-8 bg-blue-500 rounded-full flex items-center justify-center">
                <span className="text-sm font-medium text-white">
                  {member.name.charAt(0).toUpperCase()}
                </span>
              </div>
            </div>

            {/* Member Info */}
            <div className="flex-1 min-w-0">
              <div className="flex items-center space-x-2">
                <p className="text-sm font-medium text-gray-900 truncate">
                  {member.name}
                </p>
                {member.isCreator && (
                  <span className="inline-flex items-center px-2 py-0.5 rounded-full text-xs font-medium bg-blue-100 text-blue-800">
                    Creator
                  </span>
                )}
              </div>
              <p className="text-xs text-gray-500">
                Joined {new Date(member.joinedAt).toLocaleDateString()}
              </p>
            </div>
          </div>

          {/* Status indicator */}
          <div className="flex-shrink-0">
            <div
              className="h-2 w-2 bg-green-400 rounded-full"
              title="Active"
            ></div>
          </div>
        </div>
      ))}

      {/* Member count summary */}
      <div className="pt-3 border-t border-gray-200">
        <p className="text-xs text-gray-500 text-center">
          {members.length} member{members.length !== 1 ? "s" : ""} total
        </p>
      </div>
    </div>
  );
};
