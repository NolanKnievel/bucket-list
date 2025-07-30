import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { GroupCard } from "./GroupCard";
import { GroupSummary } from "../types";

describe("GroupCard", () => {
  const mockGroup: GroupSummary = {
    id: "test-group-id",
    name: "Test Group",
    createdAt: "2024-01-01T00:00:00Z",
    createdBy: "user-1",
    deadline: "2024-12-31T23:59:59Z",
    memberCount: 3,
    itemCount: 10,
    completedCount: 4,
    progressPercent: 40,
  };

  const mockOnViewGroup = vi.fn();

  beforeEach(() => {
    mockOnViewGroup.mockClear();
  });

  it("renders group information correctly", () => {
    render(<GroupCard group={mockGroup} onViewGroup={mockOnViewGroup} />);

    expect(screen.getByText("Test Group")).toBeInTheDocument();
    expect(screen.getByText("3 members")).toBeInTheDocument();
    expect(screen.getByText("10 items")).toBeInTheDocument();
    expect(screen.getByText("4/10 completed (40%)")).toBeInTheDocument();
  });

  it("handles singular member and item counts", () => {
    const singleGroup: GroupSummary = {
      ...mockGroup,
      memberCount: 1,
      itemCount: 1,
    };

    render(<GroupCard group={singleGroup} onViewGroup={mockOnViewGroup} />);

    expect(screen.getByText("1 member")).toBeInTheDocument();
    expect(screen.getByText("1 item")).toBeInTheDocument();
  });

  it("displays deadline information", () => {
    render(<GroupCard group={mockGroup} onViewGroup={mockOnViewGroup} />);

    // Should show some deadline text (exact text depends on current date)
    expect(screen.getByText(/days left|Due|Expired/)).toBeInTheDocument();
  });

  it("handles group without deadline", () => {
    const groupWithoutDeadline: GroupSummary = {
      ...mockGroup,
      deadline: undefined,
    };

    render(
      <GroupCard group={groupWithoutDeadline} onViewGroup={mockOnViewGroup} />
    );

    // Should not show any deadline text
    expect(screen.queryByText(/days left|Due|Expired/)).not.toBeInTheDocument();
  });

  it("calls onViewGroup when View Group button is clicked", () => {
    render(<GroupCard group={mockGroup} onViewGroup={mockOnViewGroup} />);

    const viewButton = screen.getByText("View Group");
    fireEvent.click(viewButton);

    expect(mockOnViewGroup).toHaveBeenCalledWith("test-group-id");
  });

  it("displays progress bar with correct width", () => {
    render(<GroupCard group={mockGroup} onViewGroup={mockOnViewGroup} />);

    const progressBar = screen.getByRole("progressbar", { hidden: true });
    expect(progressBar).toHaveStyle({ width: "40%" });
  });
});
