import { BucketListItem } from "../types";

/**
 * Calculate completion progress percentage based on bucket list items
 * @param items Array of bucket list items
 * @returns Progress percentage (0-100)
 */
export const calculateCompletionProgress = (
  items: BucketListItem[]
): number => {
  if (!items || items.length === 0) return 0;

  const completedItems = items.filter((item) => item.completed).length;
  return Math.round((completedItems / items.length) * 100);
};

/**
 * Get count of completed items
 * @param items Array of bucket list items
 * @returns Number of completed items
 */
export const getCompletedItemsCount = (items: BucketListItem[]): number => {
  if (!items) return 0;
  return items.filter((item) => item.completed).length;
};

/**
 * Get progress data for display in ProgressBar component
 * @param items Array of bucket list items
 * @returns Object with current, total, and percentage values
 */
export const getProgressData = (items: BucketListItem[]) => {
  const total = items?.length || 0;
  const current = getCompletedItemsCount(items);
  const percentage = calculateCompletionProgress(items);

  return {
    current,
    total,
    percentage,
  };
};
