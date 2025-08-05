import { useMemo } from "react";
import { BucketListItem } from "../types";
import { getProgressData } from "../utils/progressCalculation";

interface UseRealTimeProgressProps {
  items: BucketListItem[];
}

interface UseRealTimeProgressReturn {
  current: number;
  total: number;
  percentage: number;
}

/**
 * Custom hook for calculating real-time progress data
 * Automatically recalculates when items change due to WebSocket updates
 */
export const useRealTimeProgress = ({
  items,
}: UseRealTimeProgressProps): UseRealTimeProgressReturn => {
  const progressData = useMemo(() => {
    return getProgressData(items);
  }, [items]);

  return progressData;
};
