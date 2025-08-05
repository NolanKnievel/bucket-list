import React, { useState, useEffect } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { apiService, ApiError } from "../utils/api";
import { validateMemberName, sanitizeMemberName } from "../types";
import type { GroupWithDetails, Member } from "../types";

interface JoinGroupFormProps {
  onSuccess?: (groupId: string, member: Member) => void;
}

export const JoinGroupForm: React.FC<JoinGroupFormProps> = ({ onSuccess }) => {
  const { groupId } = useParams<{ groupId: string }>();
  const navigate = useNavigate();

  console.log("JoinGroupForm: Component rendered with groupId:", groupId);

  const [memberName, setMemberName] = useState("");
  const [isLoading, setIsLoading] = useState(false);
  const [isLoadingGroup, setIsLoadingGroup] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [validationErrors, setValidationErrors] = useState<string[]>([]);
  const [group, setGroup] = useState<GroupWithDetails | null>(null);

  // Load group details when component mounts
  useEffect(() => {
    const loadGroup = async () => {
      console.log("JoinGroupForm: Loading group with ID:", groupId);

      if (!groupId) {
        console.log("JoinGroupForm: No group ID found");
        setError("Invalid group link - no group ID found");
        setIsLoadingGroup(false);
        return;
      }

      try {
        setIsLoadingGroup(true);
        console.log("JoinGroupForm: Calling API to get group");
        const groupData = await apiService.getGroup(groupId);
        console.log("JoinGroupForm: Group data received:", groupData);
        setGroup(groupData);
        setError(null);
      } catch (err) {
        console.log("JoinGroupForm: Error loading group:", err);
        if (err instanceof ApiError) {
          switch (err.code) {
            case "GROUP_NOT_FOUND":
              setError("This group doesn't exist or the link is invalid");
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
        setIsLoadingGroup(false);
      }
    };

    loadGroup();
  }, [groupId]);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!groupId || !group) {
      setError("Invalid group");
      return;
    }

    const sanitizedName = sanitizeMemberName(memberName);

    // Validate member name
    const validation = validateMemberName(sanitizedName);
    if (!validation.isValid) {
      setValidationErrors(validation.errors.map((err) => err.message));
      return;
    }

    setValidationErrors([]);
    setIsLoading(true);
    setError(null);

    try {
      const member = await apiService.joinGroup(groupId, {
        memberName: sanitizedName,
      });

      // Store member info in localStorage for guest access
      localStorage.setItem(
        `member_${groupId}`,
        JSON.stringify({
          id: member.id,
          name: member.name,
          joinedAt: member.joinedAt,
        })
      );

      // Call success callback if provided
      if (onSuccess) {
        onSuccess(groupId, member);
      } else {
        // Default behavior: navigate to group view
        navigate(`/groups/${groupId}`);
      }
    } catch (err) {
      if (err instanceof ApiError) {
        switch (err.code) {
          case "GROUP_NOT_FOUND":
            setError("This group no longer exists");
            break;
          case "ALREADY_MEMBER":
            setError("You're already a member of this group");
            break;
          case "VALIDATION_ERROR":
            if (err.details && Array.isArray(err.details)) {
              setValidationErrors(
                err.details.map((detail: any) => detail.message)
              );
            } else {
              setError("Please check your input and try again");
            }
            break;
          default:
            setError("Failed to join group. Please try again.");
        }
      } else {
        setError("Failed to join group. Please try again.");
      }
    } finally {
      setIsLoading(false);
    }
  };

  const handleNameChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setMemberName(e.target.value);
    // Clear validation errors when user starts typing
    if (validationErrors.length > 0) {
      setValidationErrors([]);
    }
    if (error) {
      setError(null);
    }
  };

  console.log(
    "JoinGroupForm: Render state - isLoadingGroup:",
    isLoadingGroup,
    "error:",
    error,
    "group:",
    group
  );

  // Show loading state while fetching group
  if (isLoadingGroup) {
    console.log("JoinGroupForm: Rendering loading state");
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          <div className="text-center">
            <div
              className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600 mx-auto"
              role="status"
              aria-label="Loading"
            ></div>
            <p className="mt-4 text-gray-600">Loading group details...</p>
          </div>
        </div>
      </div>
    );
  }

  // Show error state if group couldn't be loaded
  if (error && !group) {
    return (
      <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
        <div className="max-w-md w-full space-y-8">
          <div className="text-center">
            <div className="mx-auto h-12 w-12 text-red-500">
              <svg fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 18.5c-.77.833.192 2.5 1.732 2.5z"
                />
              </svg>
            </div>
            <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
              Group Not Found
            </h2>
            <p className="mt-2 text-center text-sm text-red-600">{error}</p>
            <div className="mt-6">
              <button
                onClick={() => navigate("/")}
                className="w-full flex justify-center py-2 px-4 border border-transparent rounded-md shadow-sm text-sm font-medium text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500"
              >
                Go to Home
              </button>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gray-50 flex items-center justify-center py-12 px-4 sm:px-6 lg:px-8">
      <div className="max-w-md w-full space-y-8">
        <div>
          <h2 className="mt-6 text-center text-3xl font-extrabold text-gray-900">
            Join Bucket List
          </h2>
          {group && (
            <div className="mt-4 text-center">
              <h3 className="text-xl font-semibold text-gray-800">
                {group.name}
              </h3>
              <p className="text-sm text-gray-600 mt-1">
                {group.members?.length || 0} member
                {(group.members?.length || 0) !== 1 ? "s" : ""} •{" "}
                {group.items?.length || 0} item
                {(group.items?.length || 0) !== 1 ? "s" : ""}
              </p>
            </div>
          )}
        </div>

        <form className="mt-8 space-y-6" onSubmit={handleSubmit}>
          <div>
            <label htmlFor="memberName" className="sr-only">
              Your Name
            </label>
            <input
              id="memberName"
              name="memberName"
              type="text"
              value={memberName}
              onChange={handleNameChange}
              className="appearance-none rounded-md relative block w-full px-3 py-2 border border-gray-300 placeholder-gray-500 text-gray-900 focus:outline-none focus:ring-blue-500 focus:border-blue-500 focus:z-10 sm:text-sm"
              placeholder="Enter your name"
              maxLength={50}
              disabled={isLoading}
            />
          </div>

          {/* Validation Errors */}
          {validationErrors.length > 0 && (
            <div className="rounded-md bg-red-50 p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg
                    className="h-5 w-5 text-red-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 18.5c-.77.833.192 2.5 1.732 2.5z"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <h3 className="text-sm font-medium text-red-800">
                    Please fix the following errors:
                  </h3>
                  <div className="mt-2 text-sm text-red-700">
                    <ul className="list-disc pl-5 space-y-1">
                      {validationErrors.map((error, index) => (
                        <li key={index}>{error}</li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* General Error */}
          {error && group && (
            <div className="rounded-md bg-red-50 p-4">
              <div className="flex">
                <div className="flex-shrink-0">
                  <svg
                    className="h-5 w-5 text-red-400"
                    fill="none"
                    stroke="currentColor"
                    viewBox="0 0 24 24"
                  >
                    <path
                      strokeLinecap="round"
                      strokeLinejoin="round"
                      strokeWidth={2}
                      d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-2.5L13.732 4c-.77-.833-1.732-.833-2.5 0L4.268 18.5c-.77.833.192 2.5 1.732 2.5z"
                    />
                  </svg>
                </div>
                <div className="ml-3">
                  <p className="text-sm text-red-700">{error}</p>
                </div>
              </div>
            </div>
          )}

          <div>
            <button
              type="submit"
              disabled={isLoading}
              className="group relative w-full flex justify-center py-2 px-4 border border-transparent text-sm font-medium rounded-md text-white bg-blue-600 hover:bg-blue-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-blue-500 disabled:opacity-50 disabled:cursor-not-allowed"
            >
              {isLoading ? (
                <>
                  <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                  Joining...
                </>
              ) : (
                "Join Group"
              )}
            </button>
          </div>

          <div className="text-center">
            <button
              type="button"
              onClick={() => navigate("/")}
              className="text-sm text-blue-600 hover:text-blue-500"
            >
              Back to Home
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};
