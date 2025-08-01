import React from "react";

interface ProgressBarProps {
  current: number;
  total: number;
  percentage: number;
  showLabel?: boolean;
  size?: "sm" | "md" | "lg";
  color?: "blue" | "green" | "yellow" | "red";
}

export const ProgressBar: React.FC<ProgressBarProps> = ({
  current,
  total,
  percentage,
  showLabel = true,
  size = "md",
  color = "blue",
}) => {
  const sizeClasses = {
    sm: "h-2",
    md: "h-3",
    lg: "h-4",
  };

  const colorClasses = {
    blue: "bg-blue-600",
    green: "bg-green-600",
    yellow: "bg-yellow-600",
    red: "bg-red-600",
  };

  const getProgressColor = () => {
    if (total === 0) return "blue";
    if (percentage === 100) return "green";
    if (percentage >= 75) return "blue";
    if (percentage >= 50) return "yellow";
    return "red";
  };

  const progressColor = color === "blue" ? getProgressColor() : color;

  return (
    <div className="w-full">
      {showLabel && (
        <div className="flex justify-between items-center mb-2">
          <span className="text-sm font-medium text-gray-700">
            Completion Progress
          </span>
          <span className="text-sm text-gray-600">
            {current}/{total} completed ({percentage}%)
          </span>
        </div>
      )}

      <div className="w-full bg-gray-200 rounded-full overflow-hidden">
        <div
          className={`${sizeClasses[size]} ${colorClasses[progressColor]} rounded-full transition-all duration-500 ease-out`}
          style={{ width: `${percentage}%` }}
          role="progressbar"
          aria-valuenow={percentage}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label={`${current} of ${total} items completed`}
        >
          {/* Optional inner glow effect */}
          <div className="h-full w-full bg-gradient-to-r from-transparent to-white opacity-30 rounded-full"></div>
        </div>
      </div>

      {total === 0 && (
        <p className="text-xs text-gray-500 mt-1 text-center">
          No items to track yet
        </p>
      )}
    </div>
  );
};
