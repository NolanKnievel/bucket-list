import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { BrowserRouter } from "react-router-dom";
import { vi, describe, it, expect, beforeEach } from "vitest";
import { JoinGroupForm } from "./JoinGroupForm";
import { apiService } from "../utils/api";
import type { GroupWithDetails, Member } from "../types";

// Mock the API service
vi.mock("../utils/api", () => ({
  apiService: {
    getGroup: vi.fn(),
    joinGroup: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    constructor(public code: string, message: string, public details?: any) {
      super(message);
      this.name = "ApiError";
    }
  },
}));

// Mock react-router-dom
const mockNavigate = vi.fn();
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useNavigate: () => mockNavigate,
    useParams: () => ({ groupId: "test-group-id" }),
  };
});

const mockGroup: GroupWithDetails = {
  id: "test-group-id",
  name: "Test Bucket List",
  createdAt: "2024-01-01T00:00:00Z",
  createdBy: "user-1",
  members: [
    {
      id: "member-1",
      groupId: "test-group-id",
      userId: "user-1",
      name: "Creator",
      joinedAt: "2024-01-01T00:00:00Z",
      isCreator: true,
    },
  ],
  items: [
    {
      id: "item-1",
      groupId: "test-group-id",
      title: "Test Item",
      completed: false,
      createdBy: "member-1",
      createdAt: "2024-01-01T00:00:00Z",
    },
  ],
};

const mockMember: Member = {
  id: "member-2",
  groupId: "test-group-id",
  name: "New Member",
  joinedAt: "2024-01-01T01:00:00Z",
  isCreator: false,
};

const renderJoinGroupForm = (props = {}) => {
  return render(
    <BrowserRouter>
      <JoinGroupForm {...props} />
    </BrowserRouter>
  );
};

describe("JoinGroupForm", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders loading state initially", () => {
    vi.mocked(apiService.getGroup).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    renderJoinGroupForm();

    expect(screen.getByText("Loading group details...")).toBeInTheDocument();
    expect(screen.getByRole("status")).toBeInTheDocument(); // Loading spinner
  });

  it("renders group details and form when group loads successfully", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(screen.getByText("Join Bucket List")).toBeInTheDocument();
    });

    expect(screen.getByText("Test Bucket List")).toBeInTheDocument();
    expect(screen.getByText("1 member • 1 item")).toBeInTheDocument();
    expect(screen.getByPlaceholderText("Enter your name")).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Join Group" })
    ).toBeInTheDocument();
  });

  it("renders error state when group not found", async () => {
    const { ApiError } = await import("../utils/api");
    const apiError = new ApiError("GROUP_NOT_FOUND", "Group not found");
    vi.mocked(apiService.getGroup).mockRejectedValue(apiError);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(screen.getByText("Group Not Found")).toBeInTheDocument();
    });

    expect(
      screen.getByText("This group doesn't exist or the link is invalid")
    ).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Go to Home" })
    ).toBeInTheDocument();
  });

  it("handles invalid group ID error", async () => {
    const { ApiError } = await import("../utils/api");
    const apiError = new ApiError("INVALID_GROUP_ID", "Invalid group ID");
    vi.mocked(apiService.getGroup).mockRejectedValue(apiError);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(screen.getByText("Invalid group link format")).toBeInTheDocument();
    });
  });

  it("validates member name input", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText("Enter your name")
      ).toBeInTheDocument();
    });

    const submitButton = screen.getByRole("button", { name: "Join Group" });
    const nameInput = screen.getByPlaceholderText("Enter your name");

    // Enter invalid name (too short after trimming)
    fireEvent.change(nameInput, { target: { value: "  " } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Member name is required")).toBeInTheDocument();
    });
  });

  it("successfully joins group with valid input", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);
    vi.mocked(apiService.joinGroup).mockResolvedValue(mockMember);

    const mockOnSuccess = vi.fn();
    renderJoinGroupForm({ onSuccess: mockOnSuccess });

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText("Enter your name")
      ).toBeInTheDocument();
    });

    const nameInput = screen.getByPlaceholderText("Enter your name");
    const submitButton = screen.getByRole("button", { name: "Join Group" });

    fireEvent.change(nameInput, { target: { value: "New Member" } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(vi.mocked(apiService.joinGroup)).toHaveBeenCalledWith(
        "test-group-id",
        {
          memberName: "New Member",
        }
      );
    });

    await waitFor(() => {
      expect(mockOnSuccess).toHaveBeenCalledWith("test-group-id", mockMember);
    });
  });

  it("navigates to group view when no onSuccess callback provided", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);
    vi.mocked(apiService.joinGroup).mockResolvedValue(mockMember);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText("Enter your name")
      ).toBeInTheDocument();
    });

    const nameInput = screen.getByPlaceholderText("Enter your name");
    const submitButton = screen.getByRole("button", { name: "Join Group" });

    fireEvent.change(nameInput, { target: { value: "New Member" } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(mockNavigate).toHaveBeenCalledWith("/groups/test-group-id");
    });
  });

  it("handles already member error", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);
    const { ApiError } = await import("../utils/api");
    const apiError = new ApiError("ALREADY_MEMBER", "Already a member");
    vi.mocked(apiService.joinGroup).mockRejectedValue(apiError);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText("Enter your name")
      ).toBeInTheDocument();
    });

    const nameInput = screen.getByPlaceholderText("Enter your name");
    const submitButton = screen.getByRole("button", { name: "Join Group" });

    fireEvent.change(nameInput, { target: { value: "Existing Member" } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(
        screen.getByText("You're already a member of this group")
      ).toBeInTheDocument();
    });
  });

  it("handles validation errors from server", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);
    const { ApiError } = await import("../utils/api");
    const apiError = new ApiError("VALIDATION_ERROR", "Validation failed", [
      { field: "memberName", message: "Name is too long" },
    ]);
    vi.mocked(apiService.joinGroup).mockRejectedValue(apiError);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText("Enter your name")
      ).toBeInTheDocument();
    });

    const nameInput = screen.getByPlaceholderText("Enter your name");
    const submitButton = screen.getByRole("button", { name: "Join Group" });

    fireEvent.change(nameInput, { target: { value: "Very Long Name" } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Name is too long")).toBeInTheDocument();
    });
  });

  it("clears errors when user starts typing", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText("Enter your name")
      ).toBeInTheDocument();
    });

    const nameInput = screen.getByPlaceholderText("Enter your name");
    const submitButton = screen.getByRole("button", { name: "Join Group" });

    // Trigger validation error
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Member name is required")).toBeInTheDocument();
    });

    // Start typing to clear error
    fireEvent.change(nameInput, { target: { value: "N" } });

    expect(
      screen.queryByText("Member name is required")
    ).not.toBeInTheDocument();
  });

  it("shows loading state during form submission", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);
    vi.mocked(apiService.joinGroup).mockImplementation(
      () => new Promise(() => {}) // Never resolves
    );

    renderJoinGroupForm();

    await waitFor(() => {
      expect(
        screen.getByPlaceholderText("Enter your name")
      ).toBeInTheDocument();
    });

    const nameInput = screen.getByPlaceholderText("Enter your name");
    const submitButton = screen.getByRole("button", { name: "Join Group" });

    fireEvent.change(nameInput, { target: { value: "Test User" } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText("Joining...")).toBeInTheDocument();
    });

    expect(submitButton).toBeDisabled();
    expect(nameInput).toBeDisabled();
  });

  it("navigates to home when back button is clicked", async () => {
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(screen.getByText("Back to Home")).toBeInTheDocument();
    });

    fireEvent.click(screen.getByText("Back to Home"));

    expect(mockNavigate).toHaveBeenCalledWith("/");
  });

  it("handles plural/singular text correctly", async () => {
    const singleItemGroup = {
      ...mockGroup,
      members: [mockGroup.members[0]],
      items: [mockGroup.items[0]],
    };

    vi.mocked(apiService.getGroup).mockResolvedValue(singleItemGroup);

    renderJoinGroupForm();

    await waitFor(() => {
      expect(screen.getByText("1 member • 1 item")).toBeInTheDocument();
    });

    // Test with multiple items
    const multipleItemsGroup = {
      ...mockGroup,
      members: [mockGroup.members[0], mockMember],
      items: [mockGroup.items[0], { ...mockGroup.items[0], id: "item-2" }],
    };

    vi.mocked(apiService.getGroup).mockResolvedValue(multipleItemsGroup);

    // Re-render component
    renderJoinGroupForm();

    await waitFor(() => {
      expect(screen.getByText("2 members • 2 items")).toBeInTheDocument();
    });
  });
});
