import React, { useState } from "react";
import { BucketListItem } from "../types";
import {
  CreateItemRequest,
  validateCreateItemRequest,
  sanitizeItemTitle,
  sanitizeItemDescription,
  ValidationError,
} from "../types";

interface AddItemFormWithWebSocketProps {
  groupId: string;
  memberId: string;
  onItemAdded: (item: BucketListItem) => void;
  onCancel?: () => void;
  onAddItemWithWebSocket: (itemData: {
    title: string;
    description?: string;
  }) => Promise<void>;
}

export const AddItemFormWithWebSocket: React.FC<
  AddItemFormWithWebSocketProps
> = ({
  groupId: _groupId,
  memberId,
  onItemAdded: _onItemAdded,
  onCancel,
  onAddItemWithWebSocket,
}) => {
  const [formData, setFormData] = useState({
    title: "",
    description: "",
  });
  const [validationErrors, setValidationErrors] = useState<ValidationError[]>(
    []
  );
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [apiError, setApiError] = useState<string | null>(null);

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
    const sanitizedData: CreateItemRequest = {
      title: sanitizeItemTitle(formData.title),
      description: formData.description
        ? sanitizeItemDescription(formData.description)
        : undefined,
      memberId,
    };

    // Validate form data
    const validation = validateCreateItemRequest(sanitizedData);
    if (!validation.isValid) {
      setValidationErrors(validation.errors);
      return;
    }

    setIsSubmitting(true);
    setApiError(null);
    setValidationErrors([]);

    try {
      // Use WebSocket-integrated add item function
      await onAddItemWithWebSocket({
        title: sanitizedData.title,
        description: sanitizedData.description,
      });

      // Reset form
      setFormData({ title: "", description: "" });
    } catch (error) {
      setApiError("Failed to add item. Please try again.");
      console.error("Error creating item:", error);
    } finally {
      setIsSubmitting(false);
    }
  };

  const handleCancel = () => {
    setFormData({ title: "", description: "" });
    setValidationErrors([]);
    setApiError(null);
    if (onCancel) {
      onCancel();
    }
  };

  const getFieldError = (fieldName: string): string | undefined => {
    const error = validationErrors.find((err) => err.field === fieldName);
    return error?.message;
  };

  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6 mb-6">
      <div className="mb-4">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">
          Add New Item
        </h3>
        <p className="text-sm text-gray-600">
          Add a new idea or experience to your group's bucket list.
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
                Error adding item
              </h3>
              <p className="mt-1 text-sm text-red-700">{apiError}</p>
            </div>
          </div>
        </div>
      )}

      <form onSubmit={handleSubmit} className="space-y-4">
        <div>
          <label
            htmlFor="title"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Title *
          </label>
          <input
            type="text"
            id="title"
            name="title"
            value={formData.title}
            onChange={handleInputChange}
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 ${
              getFieldError("title")
                ? "border-red-300 focus:ring-red-500 focus:border-red-500"
                : "border-gray-300"
            }`}
            placeholder="Enter the bucket list item title"
            maxLength={200}
            disabled={isSubmitting}
          />
          {getFieldError("title") && (
            <p className="mt-1 text-sm text-red-600">
              {getFieldError("title")}
            </p>
          )}
        </div>

        <div>
          <label
            htmlFor="description"
            className="block text-sm font-medium text-gray-700 mb-1"
          >
            Description (Optional)
          </label>
          <textarea
            id="description"
            name="description"
            value={formData.description}
            onChange={handleInputChange}
            rows={3}
            className={`w-full px-3 py-2 border rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500 resize-none ${
              getFieldError("description")
                ? "border-red-300 focus:ring-red-500 focus:border-red-500"
                : "border-gray-300"
            }`}
            placeholder="Add more details about this bucket list item (optional)"
            maxLength={1000}
            disabled={isSubmitting}
          />
          {getFieldError("description") && (
            <p className="mt-1 text-sm text-red-600">
              {getFieldError("description")}
            </p>
          )}
          <p className="mt-1 text-xs text-gray-500">
            {formData.description.length}/1000 characters
          </p>
        </div>

        <div className="flex space-x-3 pt-2">
          <button
            type="button"
            onClick={handleCancel}
            className="flex-1 bg-gray-500 hover:bg-gray-600 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-gray-500 focus:ring-offset-2 transition duration-200"
            disabled={isSubmitting}
          >
            Cancel
          </button>
          <button
            type="submit"
            className="flex-1 bg-blue-500 hover:bg-blue-600 text-white font-medium py-2 px-4 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2 transition duration-200 disabled:opacity-50 disabled:cursor-not-allowed"
            disabled={isSubmitting}
          >
            {isSubmitting ? (
              <div className="flex items-center justify-center">
                <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-white mr-2"></div>
                Adding...
              </div>
            ) : (
              "Add Item"
            )}
          </button>
        </div>
      </form>
    </div>
  );
};
