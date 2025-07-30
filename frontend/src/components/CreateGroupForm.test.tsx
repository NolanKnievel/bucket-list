import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { CreateGroupForm } from "./CreateGroupForm";
import { apiService } from "../utils/api";

// Mock the API service
vi.mock("../utils/api", () => ({
  apiService: {
    createGroup: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    constructor(public code: string, message: string, public details?: any) {
      super(message);
      this.name = "ApiError";
    }
  },
}));

// Mock react-router-dom
vi.mock("react-router-dom", () => ({
  useNavigate: () => vi.fn(),
}));

describe("CreateGroupForm", () => {
  const mockOnSuccess = vi.fn();
  const mockOnCancel = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders form fields correctly", () => {
    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    expect(screen.getByText("Create New Group")).toBeInTheDocument();
    expect(screen.getByLabelText(/Group Name/)).toBeInTheDocument();
    expect(screen.getByLabelText(/Deadline/)).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Create Group" })
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
  });

  it("validates required group name", async () => {
    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Group name is required")).toBeInTheDocument();
    });

    expect(apiService.createGroup).not.toHaveBeenCalled();
  });

  it("validates group name length", async () => {
    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    const nameInput = screen.getByLabelText(/Group Name/);
    fireEvent.change(nameInput, { target: { value: "A" } });

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("Group name must be at least 2 characters long")
      ).toBeInTheDocument();
    });

    expect(apiService.createGroup).not.toHaveBeenCalled();
  });

  it("validates deadline is in the future", async () => {
    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    const nameInput = screen.getByLabelText(/Group Name/);
    const deadlineInput = screen.getByLabelText(/Deadline/);

    fireEvent.change(nameInput, { target: { value: "Test Group" } });
    fireEvent.change(deadlineInput, { target: { value: "2020-01-01T00:00" } });

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("Deadline must be in the future")
      ).toBeInTheDocument();
    });

    expect(apiService.createGroup).not.toHaveBeenCalled();
  });

  it("submits form with valid data", async () => {
    const mockResponse = {
      id: "test-group-id",
      name: "Test Group",
      shareLink: "http://localhost:3000/join/test-group-id",
      createdAt: "2024-01-01T00:00:00Z",
      createdBy: "user-1",
    };

    (apiService.createGroup as any).mockResolvedValue(mockResponse);

    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    const nameInput = screen.getByLabelText(/Group Name/);
    fireEvent.change(nameInput, { target: { value: "Test Group" } });

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(apiService.createGroup).toHaveBeenCalledWith({
        name: "Test Group",
        deadline: undefined,
      });
    });

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith(
        "test-group-id",
        "http://localhost:3000/join/test-group-id"
      );
    });
  });

  it("shows success state after group creation", async () => {
    const mockResponse = {
      id: "test-group-id",
      name: "Test Group",
      shareLink: "http://localhost:3000/join/test-group-id",
      createdAt: "2024-01-01T00:00:00Z",
      createdBy: "user-1",
    };

    (apiService.createGroup as any).mockResolvedValue(mockResponse);

    render(<CreateGroupForm />);

    const nameInput = screen.getByLabelText(/Group Name/);
    fireEvent.change(nameInput, { target: { value: "Test Group" } });

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("Group Created Successfully!")
      ).toBeInTheDocument();
      expect(
        screen.getByText('Your bucket list group "Test Group" is ready.')
      ).toBeInTheDocument();
      expect(
        screen.getByDisplayValue("http://localhost:3000/join/test-group-id")
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "View Group" })
      ).toBeInTheDocument();
      expect(
        screen.getByRole("button", { name: "Create Another" })
      ).toBeInTheDocument();
    });
  });

  it("handles API errors gracefully", async () => {
    const mockError = new Error("Network error");
    (apiService.createGroup as any).mockRejectedValue(mockError);

    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    const nameInput = screen.getByLabelText(/Group Name/);
    fireEvent.change(nameInput, { target: { value: "Test Group" } });

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("Failed to create group. Please try again.")
      ).toBeInTheDocument();
    });

    expect(mockOnSuccess).not.toHaveBeenCalled();
  });

  it("calls onCancel when cancel button is clicked", () => {
    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    const cancelButton = screen.getByRole("button", { name: "Cancel" });
    fireEvent.click(cancelButton);

    expect(mockOnCancel).toHaveBeenCalled();
  });

  it("shows loading state during submission", async () => {
    // Mock a delayed response
    (apiService.createGroup as any).mockImplementation(
      () => new Promise((resolve) => setTimeout(resolve, 100))
    );

    render(
      <CreateGroupForm onSuccess={mockOnSuccess} onCancel={mockOnCancel} />
    );

    const nameInput = screen.getByLabelText(/Group Name/);
    fireEvent.change(nameInput, { target: { value: "Test Group" } });

    const submitButton = screen.getByRole("button", { name: "Create Group" });
    fireEvent.click(submitButton);

    // Should show loading state
    expect(screen.getByText("Creating...")).toBeInTheDocument();
    expect(submitButton).toBeDisabled();
  });
});
