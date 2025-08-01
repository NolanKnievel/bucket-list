import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { vi } from "vitest";
import { AddItemForm } from "./AddItemForm";
import { apiService } from "../utils/api";
import { BucketListItem } from "../types";

// Create a mock ApiError class using vi.hoisted
const { MockApiError } = vi.hoisted(() => ({
  MockApiError: class MockApiError extends Error {
    constructor(public code: string, message: string, public details?: any) {
      super(message);
      this.name = "ApiError";
    }
  },
}));

// Mock the API service
vi.mock("../utils/api", () => ({
  apiService: {
    createItem: vi.fn(),
  },
  ApiError: MockApiError,
}));

const mockApiService = apiService as any;

describe("AddItemForm", () => {
  const defaultProps = {
    groupId: "test-group-id",
    memberId: "test-member-id",
    onItemAdded: vi.fn(),
    onCancel: vi.fn(),
  };

  const mockItem: BucketListItem = {
    id: "test-item-id",
    groupId: "test-group-id",
    title: "Test Item",
    description: "Test Description",
    completed: false,
    createdBy: "test-member-id",
    createdAt: "2024-01-01T00:00:00Z",
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders the form with title and description fields", () => {
    render(<AddItemForm {...defaultProps} />);

    expect(screen.getByText("Add New Item")).toBeInTheDocument();
    expect(screen.getByLabelText("Title *")).toBeInTheDocument();
    expect(screen.getByLabelText("Description (Optional)")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Add Item" })
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
  });

  it("shows validation error when title is empty", async () => {
    render(<AddItemForm {...defaultProps} />);

    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Item title is required")).toBeInTheDocument();
    });

    expect(mockApiService.createItem).not.toHaveBeenCalled();
    expect(defaultProps.onItemAdded).not.toHaveBeenCalled();
  });

  it("shows validation error when title is too long", async () => {
    render(<AddItemForm {...defaultProps} />);

    const titleInput = screen.getByLabelText("Title *");
    fireEvent.change(titleInput, {
      target: { value: "a".repeat(201) }, // Exceeds 200 character limit
    });

    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("Item title must be less than 200 characters")
      ).toBeInTheDocument();
    });

    expect(mockApiService.createItem).not.toHaveBeenCalled();
  });

  it("shows validation error when description is too long", async () => {
    render(<AddItemForm {...defaultProps} />);

    const titleInput = screen.getByLabelText("Title *");
    const descriptionInput = screen.getByLabelText("Description (Optional)");

    fireEvent.change(titleInput, { target: { value: "Valid Title" } });
    fireEvent.change(descriptionInput, {
      target: { value: "a".repeat(1001) }, // Exceeds 1000 character limit
    });

    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("Item description must be less than 1000 characters")
      ).toBeInTheDocument();
    });

    expect(mockApiService.createItem).not.toHaveBeenCalled();
  });

  it("successfully submits form with valid data", async () => {
    mockApiService.createItem.mockResolvedValue(mockItem);

    render(<AddItemForm {...defaultProps} />);

    const titleInput = screen.getByLabelText("Title *");
    const descriptionInput = screen.getByLabelText("Description (Optional)");

    fireEvent.change(titleInput, { target: { value: "Test Item" } });
    fireEvent.change(descriptionInput, {
      target: { value: "Test Description" },
    });

    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockApiService.createItem).toHaveBeenCalledWith("test-group-id", {
        title: "Test Item",
        description: "Test Description",
        memberId: "test-member-id",
      });
    });

    await waitFor(() => {
      expect(defaultProps.onItemAdded).toHaveBeenCalledWith(mockItem);
    });

    // Form should be reset after successful submission
    await waitFor(() => {
      expect(titleInput).toHaveValue("");
      expect(descriptionInput).toHaveValue("");
    });
  });

  it("successfully submits form with only title (no description)", async () => {
    mockApiService.createItem.mockResolvedValue(mockItem);

    render(<AddItemForm {...defaultProps} />);

    const titleInput = screen.getByLabelText("Title *");
    fireEvent.change(titleInput, { target: { value: "Test Item" } });

    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockApiService.createItem).toHaveBeenCalledWith("test-group-id", {
        title: "Test Item",
        description: undefined,
        memberId: "test-member-id",
      });
    });

    await waitFor(() => {
      expect(defaultProps.onItemAdded).toHaveBeenCalledWith(mockItem);
    });
  });

  it("handles API error gracefully", async () => {
    const apiError = new MockApiError("VALIDATION_ERROR", "Invalid item data");
    mockApiService.createItem.mockRejectedValue(apiError);

    render(<AddItemForm {...defaultProps} />);

    const titleInput = screen.getByLabelText("Title *");
    fireEvent.change(titleInput, { target: { value: "Test Item" } });

    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Invalid item data")).toBeInTheDocument();
    });

    expect(defaultProps.onItemAdded).not.toHaveBeenCalled();
  });

  it("shows loading state during submission", async () => {
    // Create a promise that we can control
    let resolvePromise: (value: any) => void;
    const promise = new Promise((resolve) => {
      resolvePromise = resolve;
    });
    mockApiService.createItem.mockReturnValue(promise);

    render(<AddItemForm {...defaultProps} />);

    const titleInput = screen.getByLabelText("Title *");
    fireEvent.change(titleInput, { target: { value: "Test Item" } });

    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    // Should show loading state
    await waitFor(() => {
      expect(screen.getByText("Adding...")).toBeInTheDocument();
    });

    // Form fields should be disabled
    expect(titleInput).toBeDisabled();
    expect(screen.getByLabelText("Description (Optional)")).toBeDisabled();

    // Resolve the promise
    resolvePromise!(mockItem);

    await waitFor(() => {
      expect(
        screen.getByRole("button", { name: "Add Item" })
      ).toBeInTheDocument();
    });
  });

  it("calls onCancel when cancel button is clicked", () => {
    render(<AddItemForm {...defaultProps} />);

    const cancelButton = screen.getByRole("button", { name: "Cancel" });
    fireEvent.click(cancelButton);

    expect(defaultProps.onCancel).toHaveBeenCalled();
  });

  it("clears form when cancel is clicked", () => {
    render(<AddItemForm {...defaultProps} />);

    const titleInput = screen.getByLabelText("Title *");
    const descriptionInput = screen.getByLabelText("Description (Optional)");

    fireEvent.change(titleInput, { target: { value: "Test Title" } });
    fireEvent.change(descriptionInput, {
      target: { value: "Test Description" },
    });

    const cancelButton = screen.getByRole("button", { name: "Cancel" });
    fireEvent.click(cancelButton);

    expect(titleInput).toHaveValue("");
    expect(descriptionInput).toHaveValue("");
  });

  it("shows character count for description", () => {
    render(<AddItemForm {...defaultProps} />);

    const descriptionInput = screen.getByLabelText("Description (Optional)");

    // Initially should show 0/1000
    expect(screen.getByText("0/1000 characters")).toBeInTheDocument();

    fireEvent.change(descriptionInput, {
      target: { value: "Hello world" },
    });

    expect(screen.getByText("11/1000 characters")).toBeInTheDocument();
  });

  it("clears validation errors when user starts typing", async () => {
    render(<AddItemForm {...defaultProps} />);

    // First trigger validation error
    const submitButton = screen.getByRole("button", { name: "Add Item" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Item title is required")).toBeInTheDocument();
    });

    // Then start typing in title field
    const titleInput = screen.getByLabelText("Title *");
    fireEvent.change(titleInput, { target: { value: "T" } });

    // Error should be cleared
    expect(
      screen.queryByText("Item title is required")
    ).not.toBeInTheDocument();
  });
});
