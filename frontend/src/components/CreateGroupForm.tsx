import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { apiService, ApiError } from "../utils/api";
import {
  CreateGroupRequest,
  validateCreateGroupRequest,
  sanitizeGroupName,
  ValidationError,
} from "../types";

interface CreateGroupFormProps {
  onSuccess?: (groupId: string, shareLink: string) => void;
  onCancel?: () => void;
}

export const CreateGroupForm: React.FC<CreateGroupFormProps> = ({
  onSuccess,
  onCancel,
}) => {
  const navigate = useNavigate();
  const [formData, setFormData] = useState<CreateGroupRequest>({
    name: "",
    deadline: "",
  });
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>(
    []
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);
  const [createdGroup, setCreatedGroup] = useState<{
    id: string;
    name: string;
    shareLink: string;
  } | null>(null);

  const handleInputChange = (
    e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>
  ) => {
    const { name, value } = e.target;
    setFormData((prev) => ({
      ...prev,
      [name]: value,
    }));

    // Clear validation errors for this field
    setValidationErrors((prev) => prev.filter((error) => error.field !== name));
    setApiError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    // Sanitize input data
    const sanitizedData: CreateGroupRequest = {
      name: sanitizeGroupName(formData.name),
      deadline: formData.deadline
        ? new Date(formData.deadline).toISOString()
        : undefined,
    };

    // Validate form data
    const validation = validateCreateGroupRequest(sanitizedData);
    if (!validation.isValid) {
      setValidationErrors(validation.errors);
      return;
    }

    setIsSubmitting(true);
    setApiError(null);
    setValidationErrors([]);

    try {
      const result = await apiService.createGroup(sanitizedData);

      setCreatedGroup({
        id: result.id,
        name: result.name,
        shareLink: result.shareLink,
      });

      // Call success callback if provided
      if (onSuccess) {
        onSuccess(result.id, result.shareLink);
      }
    } catch (error) {
      if (error instanceof ApiError) {
        setApiError(error.message);
      } else {
        setApiError("Failed to create group. Please try again.");
      }
      console.error("Error creating group:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleViewGroup = () => {
    if (createdGroup) {
      // TODO: Navigate to group view - will be implemented in future tasks
      console.log("Navigate to group:", createdGroup.id);
      navigate("/dashboard");
    }
  };

  const handleCreateAnother = () => {
    setCreatedGroup(null);
    setFormData({ name: "", deadline: "" });
    setValidationErrors([]);
    setApiError(null);
  };

  const handleCancel = () => {
    if (onCancel) {
      onCancel();
    } else {
      navigate("/dashboard");
    }
  };

  const copyShareLink = async () => {
    if (createdGroup?.shareLink) {
      try {
        await navigator.clipboard.writeText(createdGroup.shareLink);
        // TODO: Show toast notification for successful copy
        console.log("Share link copied to clipboard");
      } catch (error) {
        console.error("Failed to copy share link:", error);
        // Fallback: select the text
        const linkElement = document.getElementById("share-link");
        if (linkElement) {
          const range = document.createRange();
          range.selectNode(linkElement);
          window.getSelection()?.removeAllRanges();
          window.getSelection()?.addRange(range);
        }
      }
    }
  };

  const getFieldError = (fieldName: string): string | undefined => {
    const error = validationErrors.find((err) => err.field === fieldName);
    return error?.message;
  };

  // Success state - show created group details and share link
  if (createdGroup) {
    return (
      <div className="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
        <div className="text-center mb-6">
          <div className="mx-auto flex items-center justify-center h-12 w-12 rounded-full bg-green-100 mb-4">
            <svg
              className="h-6 w-6 text-green-600"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M5 13l4 4L19 7"
              />
            </svg>
          </div>
          <h2 className="text-2xl font-bold text-gray-900 mb-2">
            Group Created Successfully!
          </h2>
          <p className="text-gray-600">
            Your bucket list group "{createdGroup.name}" is ready.
          </p>
        </div>

        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-2">
            Share this link with your friends:
          </label>
          <div className="flex items-center space-x-2">
            <input
              id="share-link"
              type="text"
              value={createdGroup.shareLink}
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

        <div className="flex space-x-3">
          <button
            onClick={handleViewGroup}
            className="flex-1 bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline transition duration-200"
          >
            View Group
          </button>
          <button
            onClick={handleCreateAnother}
            className="flex-1 bg-gray-500 hover:bg-gray-600 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline transition duration-200"
          >
            Create Another
          </button>
        </div>
      </div>
    );
  }

  // Form state - show create group form
  return (
    <div className="max-w-md mx-auto bg-white rounded-lg shadow-md p-6">
      <div className="mb-6">
        <h2 className="text-2xl font-bold text-gray-900 mb-2">
          Create New Group
        </h2>
        <p className="text-gray-600">
          Start a collaborative bucket list with your friends and family.
        </p>
      </div>

      {apiError && (
        <div className="mb-4 bg-red-50 border border-red-200 rounded-lg p-4">
          <div className="flex">
            <div className="flex-shrink-0">
              <svg
                className="h-5 w-5 text-red-400"
                viewBox="0 0 20 20"
                fill="currentColor"
              >
                <path
                  fillRule="evenodd"
                  d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z"
                  clipRule="evenodd"
                />
              </svg>
            </div>
            <div className="ml-3">
              <h3 className="text-sm font-medium text-red-800">
                Error creating group
              </h3>
              <p className="mt-1 text-sm text-red-700">{apiError}</p>
            </div>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label
            htmlFor="name"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Group Name *
          </label>
          <input
            type="text"
            id="name"
            name="name"
            value={formData.name}
            onChange={handleInputChange}
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
              getFieldError("name")
                ? "border-red-300 focus:ring-red-500 focus:border-red-500"
                : "border-gray-300"
            }`}
            placeholder="Enter a name for your bucket list group"
            maxLength={100}
            disabled={isSubmitting}
          />
          {getFieldError("name") && (
            <p className="mt-1 text-sm text-red-600">{getFieldError("name")}</p>
          )}
        </div>

        <div>
          <label
            htmlFor="deadline"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Deadline (Optional)
          </label>
          <input
            type="datetime-local"
            id="deadline"
            name="deadline"
            value={formData.deadline}
            onChange={handleInputChange}
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
              getFieldError("deadline")
                ? "border-red-300 focus:ring-red-500 focus:border-red-500"
                : "border-gray-300"
            }`}
            disabled={isSubmitting}
          />
          {getFieldError("deadline") && (
            <p className="mt-1 text-sm text-red-600">
              {getFieldError("deadline")}
            </p>
          )}
          <p className="mt-1 text-xs text-gray-500">
            Set a deadline to add urgency to your bucket list goals.
          </p>
        </div>

        <div className="flex space-x-3 pt-4">
          <button
            type="button"
            onClick={handleCancel}
            className="flex-1 bg-gray-500 hover:bg-gray-600 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline transition duration-200"
            disabled={isSubmitting}
          >
            Cancel
          </button>
          <button
            type="submit"
            className="flex-1 bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Creating...
              </div>
            ) : (
              "Create Group"
            )}
          </button>
        </div>
      </form>
    </div>
  );
};
