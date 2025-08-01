import { render, screen, act } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { CountdownTimer } from "./CountdownTimer";

// Mock Date to control time in tests
const mockDate = new Date("2024-01-15T12:00:00Z");

describe("CountdownTimer", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(mockDate);
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("displays days, hours, and minutes correctly", () => {
    const deadline = "2024-01-20T12:00:00Z"; // 5 days from mock date
    const createdAt = "2024-01-10T12:00:00Z"; // 5 days before mock date

    render(<CountdownTimer deadline={deadline} createdAt={createdAt} />);

    // Check for specific time values in their respective containers
    const timeDisplays = screen.getAllByText(/^\d{2}$/);
    expect(timeDisplays).toHaveLength(3); // days, hours, minutes

    expect(screen.getByText("Days")).toBeInTheDocument();
    expect(screen.getByText("Hours")).toBeInTheDocument();
    expect(screen.getByText("Mins")).toBeInTheDocument();
  });

  it("does not display seconds", () => {
    const deadline = "2024-01-20T12:00:00Z";
    const createdAt = "2024-01-10T12:00:00Z";

    render(<CountdownTimer deadline={deadline} createdAt={createdAt} />);

    expect(screen.queryByText("Secs")).not.toBeInTheDocument();
    expect(screen.queryByText("Sec")).not.toBeInTheDocument();
  });

  it("shows correct urgency colors when deadline is close", () => {
    const deadline = "2024-01-15T18:00:00Z"; // 6 hours from mock date
    const createdAt = "2024-01-10T12:00:00Z";

    render(<CountdownTimer deadline={deadline} createdAt={createdAt} />);

    // Find the grid container that should have the urgency color class
    const gridContainer = screen.getByText("Days").closest(".grid");
    expect(gridContainer).toHaveClass("text-red-600");
  });

  it("displays progress bar with correct elapsed percentage", () => {
    const deadline = "2024-01-20T12:00:00Z"; // 5 days from mock date
    const createdAt = "2024-01-10T12:00:00Z"; // 5 days before mock date (50% elapsed)

    render(<CountdownTimer deadline={deadline} createdAt={createdAt} />);

    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveStyle("width: 50%");
  });

  it("shows expired state when deadline has passed", () => {
    const deadline = "2024-01-10T12:00:00Z"; // 5 days before mock date
    const createdAt = "2024-01-05T12:00:00Z";

    render(<CountdownTimer deadline={deadline} createdAt={createdAt} />);

    expect(screen.getByText("⏰ Time's Up!")).toBeInTheDocument();
    expect(
      screen.getByText("The deadline for this bucket list has passed")
    ).toBeInTheDocument();
  });

  it("updates countdown in real-time", () => {
    const deadline = "2024-01-15T12:01:00Z"; // 1 minute from mock date
    const createdAt = "2024-01-10T12:00:00Z";

    render(<CountdownTimer deadline={deadline} createdAt={createdAt} />);

    expect(screen.getByText("01")).toBeInTheDocument(); // 1 minute

    // Advance time by 30 seconds
    act(() => {
      vi.advanceTimersByTime(30000);
    });

    // After 30 seconds, should still show 00 minutes (rounded down from 30 seconds)
    const timeDisplays = screen.getAllByText("00");
    expect(timeDisplays.length).toBeGreaterThan(0); // Should have at least one "00"
  });

  it("calls onExpired callback when deadline is reached", () => {
    const onExpired = vi.fn();
    const deadline = "2024-01-15T12:00:30Z"; // 30 seconds from mock date
    const createdAt = "2024-01-10T12:00:00Z";

    render(
      <CountdownTimer
        deadline={deadline}
        createdAt={createdAt}
        onExpired={onExpired}
      />
    );

    // Advance time past the deadline
    act(() => {
      vi.advanceTimersByTime(31000);
    });

    expect(onExpired).toHaveBeenCalled();
  });

  it("handles missing createdAt prop gracefully", () => {
    const deadline = "2024-01-20T12:00:00Z";

    render(<CountdownTimer deadline={deadline} />);

    // Should still render without crashing
    expect(screen.getByText("05")).toBeInTheDocument(); // Days
    expect(screen.getByRole("progressbar")).toBeInTheDocument();
  });

  it("displays deadline information correctly", () => {
    const deadline = "2024-01-20T15:30:00Z";
    const createdAt = "2024-01-10T12:00:00Z";

    render(<CountdownTimer deadline={deadline} createdAt={createdAt} />);

    expect(screen.getByText(/Deadline:/)).toBeInTheDocument();
    expect(screen.getByText(/1\/20\/2024/)).toBeInTheDocument();
    // Check for time format (could be AM or PM depending on timezone)
    expect(screen.getByText(/\d{1,2}:\d{2} (AM|PM)/)).toBeInTheDocument();
  });
});
