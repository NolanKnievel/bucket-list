import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach } from "vitest";
import { BrowserRouter } from "react-router-dom";
import { GroupView } from "./GroupView";
import { AuthProvider } from "../contexts/AuthContext";
import { apiService } from "../utils/api";
import { GroupWithDetails, BucketListItem, Member } from "../types";

// Mock Supabase
vi.mock("../lib/supabase", () => ({
  supabase: {
    auth: {
      getSession: vi.fn().mockResolvedValue({ data: { session: null } }),
      onAuthStateChange: vi.fn().mockReturnValue({
        data: { subscription: { unsubscribe: vi.fn() } },
      }),
    },
  },
}));

// Mock the API service
vi.mock("../utils/api", () => ({
  apiService: {
    getGroup: vi.fn(),
    toggleItemCompletion: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    constructor(message: string, public code: string) {
      super(message);
    }
  },
}));

// Mock react-router-dom
vi.mock("react-router-dom", async () => {
  const actual = await vi.importActual("react-router-dom");
  return {
    ...actual,
    useParams: () => ({ groupId: "test-group-id" }),
    useNavigate: () => vi.fn(),
  };
});

// Mock the useAuth hook to return our mock user
vi.mock("../contexts/AuthContext", async () => {
  const actual = await vi.importActual("../contexts/AuthContext");
  return {
    ...actual,
    useAuth: () => ({
      user: { id: "user-1", email: "test@example.com" },
      session: null,
      loading: false,
      signIn: vi.fn(),
      signUp: vi.fn(),
      signOut: vi.fn(),
      resetPassword: vi.fn(),
    }),
  };
});

const mockUser = {
  id: "user-1",
  email: "test@example.com",
};

const mockAuthContext = {
  user: mockUser,
  signIn: vi.fn(),
  signOut: vi.fn(),
  loading: false,
};

const createMockMember = (
  id: string,
  userId?: string,
  isCreator = false
): Member => ({
  id,
  groupId: "test-group-id",
  userId,
  name: `Member ${id}`,
  joinedAt: new Date().toISOString(),
  isCreator,
});

const createMockItem = (id: string, completed: boolean): BucketListItem => ({
  id,
  groupId: "test-group-id",
  title: `Item ${id}`,
  description: `Description for item ${id}`,
  completed,
  createdBy: "member-1",
  createdAt: new Date().toISOString(),
});

const createMockGroup = (items: BucketListItem[]): GroupWithDetails => ({
  id: "test-group-id",
  name: "Test Group",
  createdAt: new Date().toISOString(),
  createdBy: "user-1",
  members: [
    createMockMember("member-1", "user-1", true),
    createMockMember("member-2", "user-2"),
  ],
  items,
});

const renderGroupView = () => {
  return render(
    <BrowserRouter>
      <GroupView />
    </BrowserRouter>
  );
};

describe("GroupView Progress Tracking Integration", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("displays 0% progress when no items exist (Requirement 8.5)", async () => {
    const mockGroup = createMockGroup([]);
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderGroupView();

    await waitFor(() => {
      expect(screen.getByText("0/0 completed (0%)")).toBeInTheDocument();
      expect(screen.getByText("No items to track yet")).toBeInTheDocument();
    });

    // Verify progress bar shows 0%
    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "0");
    expect(progressBar).toHaveStyle({ width: "0%" });
  });

  it("displays correct progress percentage with numerical values (Requirement 8.1, 8.4)", async () => {
    const items = [
      createMockItem("1", true), // completed
      createMockItem("2", false), // not completed
      createMockItem("3", true), // completed
      createMockItem("4", false), // not completed
      createMockItem("5", false), // not completed
    ];
    const mockGroup = createMockGroup(items);
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderGroupView();

    await waitFor(() => {
      // 2 out of 5 completed = 40%
      expect(screen.getByText("2/5 completed (40%)")).toBeInTheDocument();
      expect(screen.getByText("Completion Progress")).toBeInTheDocument();
    });

    // Verify progress bar shows correct percentage
    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "40");
    expect(progressBar).toHaveStyle({ width: "40%" });
  });

  it("updates progress immediately when item is marked complete (Requirement 8.2)", async () => {
    const items = [
      createMockItem("1", false), // will be marked complete
      createMockItem("2", false),
      createMockItem("3", false),
    ];
    const mockGroup = createMockGroup(items);
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    // Mock the toggle completion API call
    const updatedItem = { ...items[0], completed: true };
    vi.mocked(apiService.toggleItemCompletion).mockResolvedValue(updatedItem);

    renderGroupView();

    // Wait for initial render
    await waitFor(() => {
      expect(screen.getByText("0/3 completed (0%)")).toBeInTheDocument();
    });

    // Find and click the completion toggle for the first item
    const toggleButtons = screen.getAllByLabelText("Mark as complete");
    fireEvent.click(toggleButtons[0]);

    // Wait for the progress to update
    await waitFor(() => {
      expect(screen.getByText("1/3 completed (33%)")).toBeInTheDocument();
    });

    // Verify progress bar updated
    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "33");
    expect(progressBar).toHaveStyle({ width: "33%" });
  });

  it("updates progress immediately when item is unmarked as complete (Requirement 8.3)", async () => {
    const items = [
      createMockItem("1", true), // will be unmarked
      createMockItem("2", true),
      createMockItem("3", false),
    ];
    const mockGroup = createMockGroup(items);
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    // Mock the toggle completion API call
    const updatedItem = { ...items[0], completed: false };
    vi.mocked(apiService.toggleItemCompletion).mockResolvedValue(updatedItem);

    renderGroupView();

    // Wait for initial render - 2 out of 3 completed = 67%
    await waitFor(() => {
      expect(screen.getByText("2/3 completed (67%)")).toBeInTheDocument();
    });

    // Find and click the completion toggle for the first item to unmark it
    const toggleButtons = screen.getAllByLabelText("Mark as incomplete");
    fireEvent.click(toggleButtons[0]);

    // Wait for the progress to update - 1 out of 3 completed = 33%
    await waitFor(() => {
      expect(screen.getByText("1/3 completed (33%)")).toBeInTheDocument();
    });

    // Verify progress bar updated
    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "33");
    expect(progressBar).toHaveStyle({ width: "33%" });
  });

  it("shows both visual progress bar and numerical percentage (Requirement 8.4)", async () => {
    const items = [
      createMockItem("1", true),
      createMockItem("2", true),
      createMockItem("3", false),
      createMockItem("4", false),
    ];
    const mockGroup = createMockGroup(items);
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderGroupView();

    await waitFor(() => {
      // Check numerical percentage display
      expect(screen.getByText("2/4 completed (50%)")).toBeInTheDocument();
      expect(screen.getByText("Completion Progress")).toBeInTheDocument();

      // Check visual progress bar
      const progressBar = screen.getByRole("progressbar");
      expect(progressBar).toBeInTheDocument();
      expect(progressBar).toHaveAttribute("aria-valuenow", "50");
      expect(progressBar).toHaveStyle({ width: "50%" });

      // Check accessibility attributes
      expect(progressBar).toHaveAttribute(
        "aria-label",
        "2 of 4 items completed"
      );
      expect(progressBar).toHaveAttribute("aria-valuemin", "0");
      expect(progressBar).toHaveAttribute("aria-valuemax", "100");
    });
  });

  it("shows 100% progress when all items are completed", async () => {
    const items = [
      createMockItem("1", true),
      createMockItem("2", true),
      createMockItem("3", true),
    ];
    const mockGroup = createMockGroup(items);
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderGroupView();

    await waitFor(() => {
      expect(screen.getByText("3/3 completed (100%)")).toBeInTheDocument();
    });

    // Verify progress bar shows 100% and has green color
    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "100");
    expect(progressBar).toHaveStyle({ width: "100%" });
    expect(progressBar).toHaveClass("bg-green-600");
  });

  it("handles progress calculation edge cases correctly", async () => {
    // Test with 1 out of 3 items completed (33.33% -> rounds to 33%)
    const items = [
      createMockItem("1", true),
      createMockItem("2", false),
      createMockItem("3", false),
    ];
    const mockGroup = createMockGroup(items);
    vi.mocked(apiService.getGroup).mockResolvedValue(mockGroup);

    renderGroupView();

    await waitFor(() => {
      expect(screen.getByText("1/3 completed (33%)")).toBeInTheDocument();
    });

    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "33");
    expect(progressBar).toHaveStyle({ width: "33%" });
  });
});
