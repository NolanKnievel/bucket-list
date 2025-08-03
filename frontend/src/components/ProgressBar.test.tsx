import React from "react";
import { render, screen } from "@testing-library/react";
import { describe, it, expect } from "vitest";
import { ProgressBar } from "./ProgressBar";

describe("ProgressBar", () => {
  it("displays progress bar with correct percentage and numerical values", () => {
    render(<ProgressBar current={3} total={10} percentage={30} />);

    // Check that the label is displayed
    expect(screen.getByText("Completion Progress")).toBeInTheDocument();

    // Check that numerical percentage is displayed
    expect(screen.getByText("3/10 completed (30%)")).toBeInTheDocument();

    // Check that progress bar has correct aria attributes
    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveAttribute("aria-valuenow", "30");
    expect(progressBar).toHaveAttribute("aria-valuemin", "0");
    expect(progressBar).toHaveAttribute("aria-valuemax", "100");
    expect(progressBar).toHaveAttribute(
      "aria-label",
      "3 of 10 items completed"
    );
  });

  it("shows 0% progress when no items exist", () => {
    render(<ProgressBar current={0} total={0} percentage={0} />);

    // Check that 0% is displayed
    expect(screen.getByText("0/0 completed (0%)")).toBeInTheDocument();

    // Check that helpful message is shown
    expect(screen.getByText("No items to track yet")).toBeInTheDocument();
  });

  it("shows 100% progress when all items are completed", () => {
    render(<ProgressBar current={5} total={5} percentage={100} />);

    expect(screen.getByText("5/5 completed (100%)")).toBeInTheDocument();

    // Progress bar should be green when 100% complete
    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("bg-green-600");
  });

  it("applies correct color based on progress percentage", () => {
    // Test different progress levels
    const { rerender } = render(
      <ProgressBar current={1} total={10} percentage={10} />
    );

    let progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("bg-red-600"); // Less than 50%

    rerender(<ProgressBar current={6} total={10} percentage={60} />);
    progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("bg-yellow-600"); // 50-74%

    rerender(<ProgressBar current={8} total={10} percentage={80} />);
    progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("bg-blue-600"); // 75-99%

    rerender(<ProgressBar current={10} total={10} percentage={100} />);
    progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("bg-green-600"); // 100%
  });

  it("can hide label when showLabel is false", () => {
    render(
      <ProgressBar current={3} total={10} percentage={30} showLabel={false} />
    );

    expect(screen.queryByText("Completion Progress")).not.toBeInTheDocument();
    expect(screen.queryByText("3/10 completed (30%)")).not.toBeInTheDocument();
  });

  it("supports different sizes", () => {
    const { rerender } = render(
      <ProgressBar current={3} total={10} percentage={30} size="sm" />
    );

    let progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("h-2");

    rerender(<ProgressBar current={3} total={10} percentage={30} size="md" />);
    progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("h-3");

    rerender(<ProgressBar current={3} total={10} percentage={30} size="lg" />);
    progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass("h-4");
  });

  it("has smooth transition animation", () => {
    render(<ProgressBar current={3} total={10} percentage={30} />);

    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveClass(
      "transition-all",
      "duration-500",
      "ease-out"
    );
  });

  it("sets correct width style based on percentage", () => {
    render(<ProgressBar current={3} total={10} percentage={30} />);

    const progressBar = screen.getByRole("progressbar");
    expect(progressBar).toHaveStyle({ width: "30%" });
  });
});
