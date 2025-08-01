import React, { useState, useEffect } from "react";

interface CountdownTimerProps {
  deadline: string;
  createdAt?: string; // When the group was created to calculate elapsed time
  onExpired?: () => void;
}

interface TimeRemaining {
  days: number;
  hours: number;
  minutes: number;
  seconds: number;
  total: number;
}

export const CountdownTimer: React.FC<CountdownTimerProps> = ({
  deadline,
  createdAt,
  onExpired,
}) => {
  const [timeRemaining, setTimeRemaining] = useState<TimeRemaining>({
    days: 0,
    hours: 0,
    minutes: 0,
    seconds: 0,
    total: 0,
  });

  const calculateTimeRemaining = (): TimeRemaining => {
    const now = new Date().getTime();
    const deadlineTime = new Date(deadline).getTime();
    const difference = deadlineTime - now;

    if (difference <= 0) {
      return {
        days: 0,
        hours: 0,
        minutes: 0,
        seconds: 0,
        total: 0,
      };
    }

    const days = Math.floor(difference / (1000 * 60 * 60 * 24));
    const hours = Math.floor(
      (difference % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60)
    );
    const minutes = Math.floor((difference % (1000 * 60 * 60)) / (1000 * 60));
    const seconds = Math.floor((difference % (1000 * 60)) / 1000);

    return {
      days,
      hours,
      minutes,
      seconds,
      total: difference,
    };
  };

  useEffect(() => {
    const timer = setInterval(() => {
      const remaining = calculateTimeRemaining();
      setTimeRemaining(remaining);

      if (remaining.total <= 0 && onExpired) {
        onExpired();
      }
    }, 1000);

    // Set initial time
    setTimeRemaining(calculateTimeRemaining());

    return () => clearInterval(timer);
  }, [deadline, onExpired]);

  const getUrgencyColor = () => {
    const totalHours = timeRemaining.days * 24 + timeRemaining.hours;

    if (timeRemaining.total <= 0) return "text-red-600";
    if (totalHours <= 24) return "text-red-600";
    if (totalHours <= 72) return "text-orange-600";
    if (timeRemaining.days <= 7) return "text-yellow-600";
    return "text-gray-700";
  };

  const getProgressBarColor = () => {
    const totalHours = timeRemaining.days * 24 + timeRemaining.hours;

    if (timeRemaining.total <= 0) return "bg-red-600";
    if (totalHours <= 24) return "bg-red-500";
    if (totalHours <= 72) return "bg-orange-500";
    if (timeRemaining.days <= 7) return "bg-yellow-500";
    return "bg-blue-500";
  };

  const calculateElapsedPercentage = () => {
    const now = new Date().getTime();
    const deadlineTime = new Date(deadline).getTime();

    // Use actual creation time if provided, otherwise assume created 30 days ago as fallback
    const createdTime = createdAt
      ? new Date(createdAt).getTime()
      : deadlineTime - 30 * 24 * 60 * 60 * 1000;

    const totalDuration = deadlineTime - createdTime;
    const elapsed = now - createdTime;

    return Math.min(Math.max((elapsed / totalDuration) * 100, 0), 100);
  };

  if (timeRemaining.total <= 0) {
    return (
      <div className="text-center">
        <div className="text-2xl font-bold text-red-600 mb-2">
          ⏰ Time's Up!
        </div>
        <p className="text-sm text-red-600">
          The deadline for this bucket list has passed
        </p>
        <div className="w-full bg-gray-200 rounded-full h-2 mt-4">
          <div className="bg-red-600 h-2 rounded-full w-full"></div>
        </div>
      </div>
    );
  }

  return (
    <div className="text-center">
      {/* Time Display */}
      <div className={`grid grid-cols-3 gap-4 mb-4 ${getUrgencyColor()}`}>
        <div className="text-center">
          <div className="text-2xl font-bold">
            {timeRemaining.days.toString().padStart(2, "0")}
          </div>
          <div className="text-xs uppercase tracking-wide">
            Day{timeRemaining.days !== 1 ? "s" : ""}
          </div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold">
            {timeRemaining.hours.toString().padStart(2, "0")}
          </div>
          <div className="text-xs uppercase tracking-wide">
            Hour{timeRemaining.hours !== 1 ? "s" : ""}
          </div>
        </div>
        <div className="text-center">
          <div className="text-2xl font-bold">
            {timeRemaining.minutes.toString().padStart(2, "0")}
          </div>
          <div className="text-xs uppercase tracking-wide">
            Min{timeRemaining.minutes !== 1 ? "s" : ""}
          </div>
        </div>
      </div>

      {/* Deadline Info */}
      <p className="text-sm text-gray-600 mb-3">
        Deadline: {new Date(deadline).toLocaleDateString()} at{" "}
        {new Date(deadline).toLocaleTimeString([], {
          hour: "2-digit",
          minute: "2-digit",
        })}
      </p>

      {/* Progress Bar */}
      <div className="w-full bg-gray-200 rounded-full h-2">
        <div
          className={`h-2 rounded-full transition-all duration-1000 ${getProgressBarColor()}`}
          style={{ width: `${calculateElapsedPercentage()}%` }}
          role="progressbar"
          aria-valuenow={calculateElapsedPercentage()}
          aria-valuemin={0}
          aria-valuemax={100}
          aria-label="Time elapsed progress"
        ></div>
      </div>
    </div>
  );
};
