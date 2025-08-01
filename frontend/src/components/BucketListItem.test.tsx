import React from "react";
import { render, screen, fireEvent, waitFor } from "@testing-library/react";
import { BucketListItem } from "./BucketListItem";
import { BucketListItem as BucketListItemType, Member } from "../types";

// Mock data
const mockMembers: Member[] = [
  {
    id: "member-1",
    groupId: "group-1",
    userId: "user-1",
    name: "John Doe",
    joinedAt: "2024-01-01T00:00:00Z",
    isCreator: true,
  },
  {
    id: "member-2",
    groupId: "group-1",
    userId: "user-2",
    name: "Jane Smith",
    joinedAt: "2024-01-02T00:00:00Z",
    isCreator: false,
  },
];

const mockIncompleteItem: BucketListItemType = {
  id: "item-1",
  groupId: "group-1",
  title: "Visit Paris",
  description: "Explore the city of lights and see the Eiffel Tower",
  completed: false,
  createdBy: "member-1",
  createdAt: "2024-01-01T12:00:00Z",
};

const mockCompletedItem: BucketListItemType = {
  id: "item-2",
  groupId: "group-1",
  title: "Learn to cook pasta",
  description: "Master the art of Italian cuisine",
  completed: true,
  completedBy: "member-2",
  completedAt: "2024-01-03T15:30:00Z",
  createdBy: "member-1",
  createdAt: "2024-01-01T10:00:00Z",
};

import { vi } from "vitest";

const mockOnToggleCompletion = vi.fn();

describe("BucketListItem", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("renders incomplete item correctly", () => {
    render(
      <BucketListItem
        item={mockIncompleteItem}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    expect(screen.getByText("Visit Paris")).toBeInTheDocument();
    expect(
      screen.getByText("Explore the city of lights and see the Eiffel Tower")
    ).toBeInTheDocument();
    expect(screen.getByText("Added by John Doe")).toBeInTheDocument();
    expect(screen.getByText("1/1/2024")).toBeInTheDocument();

    // Should show incomplete circle icon
    const toggleButton = screen.getByRole("button", {
      name: /mark as complete/i,
    });
    expect(toggleButton).toBeInTheDocument();
  });

  it("renders completed item correctly", () => {
    render(
      <BucketListItem
        item={mockCompletedItem}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    expect(screen.getByText("Learn to cook pasta")).toBeInTheDocument();
    expect(
      screen.getByText("Master the art of Italian cuisine")
    ).toBeInTheDocument();
    expect(screen.getByText("Added by John Doe")).toBeInTheDocument();
    expect(screen.getByText(/Completed.*by Jane Smith/)).toBeInTheDocument();

    // Should show completed checkmark icon
    const toggleButton = screen.getByRole("button", {
      name: /mark as incomplete/i,
    });
    expect(toggleButton).toBeInTheDocument();
  });

  it("renders item without description", () => {
    const itemWithoutDescription = {
      ...mockIncompleteItem,
      description: undefined,
    };

    render(
      <BucketListItem
        item={itemWithoutDescription}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    expect(screen.getByText("Visit Paris")).toBeInTheDocument();
    expect(
      screen.queryByText("Explore the city of lights and see the Eiffel Tower")
    ).not.toBeInTheDocument();
  });

  it("handles toggle completion when user is a member", async () => {
    mockOnToggleCompletion.mockResolvedValue(undefined);

    render(
      <BucketListItem
        item={mockIncompleteItem}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    const toggleButton = screen.getByRole("button", {
      name: /mark as complete/i,
    });
    fireEvent.click(toggleButton);

    await waitFor(() => {
      expect(mockOnToggleCompletion).toHaveBeenCalledWith("item-1", true);
    });
  });

  it("shows loading state during toggle", async () => {
    let resolveToggle: () => void;
    const togglePromise = new Promise<void>((resolve) => {
      resolveToggle = resolve;
    });
    mockOnToggleCompletion.mockReturnValue(togglePromise);

    render(
      <BucketListItem
        item={mockIncompleteItem}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    const toggleButton = screen.getByRole("button", {
      name: /mark as complete/i,
    });
    fireEvent.click(toggleButton);

    // Should show loading spinner
    expect(screen.getByRole("button")).toHaveClass(
      "opacity-50",
      "cursor-not-allowed"
    );

    // Resolve the promise
    resolveToggle!();
    await waitFor(() => {
      expect(screen.getByRole("button")).not.toHaveClass("opacity-50");
    });
  });

  it("does not show toggle button when user is not a member", () => {
    render(
      <BucketListItem
        item={mockIncompleteItem}
        members={mockMembers}
        currentMemberId={undefined}
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    expect(screen.queryByRole("button")).not.toBeInTheDocument();
    // Should show static icon instead
    expect(screen.getByTitle("Not completed")).toBeInTheDocument();
  });

  it("handles unknown creator gracefully", () => {
    const itemWithUnknownCreator = {
      ...mockIncompleteItem,
      createdBy: "unknown-member",
    };

    render(
      <BucketListItem
        item={itemWithUnknownCreator}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    expect(screen.getByText("Added by Unknown")).toBeInTheDocument();
  });

  it("applies correct styling for completed items", () => {
    const { container } = render(
      <BucketListItem
        item={mockCompletedItem}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    const itemContainer = container.firstChild as HTMLElement;
    expect(itemContainer).toHaveClass("bg-green-50", "border-green-200");

    const title = screen.getByText("Learn to cook pasta");
    expect(title).toHaveClass("text-green-800", "line-through");
  });

  it("applies correct styling for incomplete items", () => {
    const { container } = render(
      <BucketListItem
        item={mockIncompleteItem}
        members={mockMembers}
        currentMemberId="member-1"
        onToggleCompletion={mockOnToggleCompletion}
      />
    );

    const itemContainer = container.firstChild as HTMLElement;
    expect(itemContainer).toHaveClass("bg-white", "border-gray-200");

    const title = screen.getByText("Visit Paris");
    expect(title).toHaveClass("text-gray-900");
    expect(title).not.toHaveClass("line-through");
  });
});
