import { describe, it, expect } from "vitest";
import {
  calculateCompletionProgress,
  getCompletedItemsCount,
  getProgressData,
} from "./progressCalculation";
import { BucketListItem } from "../types";

// Mock bucket list items for testing
const createMockItem = (id: string, completed: boolean): BucketListItem => ({
  id,
  groupId: "group-1",
  title: `Item ${id}`,
  description: `Description for item ${id}`,
  completed,
  createdBy: "member-1",
  createdAt: new Date().toISOString(),
});

describe("Progress Calculation Utils", () => {
  describe("calculateCompletionProgress", () => {
    it("returns 0 when no items exist", () => {
      expect(calculateCompletionProgress([])).toBe(0);
      expect(calculateCompletionProgress(undefined as any)).toBe(0);
    });

    it("returns 0 when no items are completed", () => {
      const items = [
        createMockItem("1", false),
        createMockItem("2", false),
        createMockItem("3", false),
      ];
      expect(calculateCompletionProgress(items)).toBe(0);
    });

    it("returns 100 when all items are completed", () => {
      const items = [
        createMockItem("1", true),
        createMockItem("2", true),
        createMockItem("3", true),
      ];
      expect(calculateCompletionProgress(items)).toBe(100);
    });

    it("calculates correct percentage for partial completion", () => {
      const items = [
        createMockItem("1", true),
        createMockItem("2", false),
        createMockItem("3", true),
        createMockItem("4", false),
      ];
      // 2 out of 4 completed = 50%
      expect(calculateCompletionProgress(items)).toBe(50);
    });

    it("rounds percentage to nearest integer", () => {
      const items = [
        createMockItem("1", true),
        createMockItem("2", false),
        createMockItem("3", false),
      ];
      // 1 out of 3 completed = 33.33% -> rounds to 33%
      expect(calculateCompletionProgress(items)).toBe(33);
    });
  });

  describe("getCompletedItemsCount", () => {
    it("returns 0 when no items exist", () => {
      expect(getCompletedItemsCount([])).toBe(0);
      expect(getCompletedItemsCount(undefined as any)).toBe(0);
    });

    it("returns correct count of completed items", () => {
      const items = [
        createMockItem("1", true),
        createMockItem("2", false),
        createMockItem("3", true),
        createMockItem("4", true),
        createMockItem("5", false),
      ];
      expect(getCompletedItemsCount(items)).toBe(3);
    });

    it("returns 0 when no items are completed", () => {
      const items = [createMockItem("1", false), createMockItem("2", false)];
      expect(getCompletedItemsCount(items)).toBe(0);
    });
  });

  describe("getProgressData", () => {
    it("returns correct progress data for empty list", () => {
      const result = getProgressData([]);
      expect(result).toEqual({
        current: 0,
        total: 0,
        percentage: 0,
      });
    });

    it("returns correct progress data for partial completion", () => {
      const items = [
        createMockItem("1", true),
        createMockItem("2", false),
        createMockItem("3", true),
        createMockItem("4", false),
        createMockItem("5", false),
      ];
      const result = getProgressData(items);
      expect(result).toEqual({
        current: 2,
        total: 5,
        percentage: 40,
      });
    });

    it("returns correct progress data for full completion", () => {
      const items = [createMockItem("1", true), createMockItem("2", true)];
      const result = getProgressData(items);
      expect(result).toEqual({
        current: 2,
        total: 2,
        percentage: 100,
      });
    });

    it("handles undefined items array", () => {
      const result = getProgressData(undefined as any);
      expect(result).toEqual({
        current: 0,
        total: 0,
        percentage: 0,
      });
    });
  });
});
