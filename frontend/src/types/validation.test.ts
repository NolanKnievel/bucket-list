import {
  validateGroupName,
  validateMemberName,
  validateItemTitle,
  validateItemDescription,
  validateDeadline,
  validateCreateGroupRequest,
  validateJoinGroupRequest,
  validateCreateItemRequest,
  sanitizeString,
  sanitizeGroupName,
  sanitizeMemberName,
  sanitizeItemTitle,
  sanitizeItemDescription,
} from "./index";

// Test validation functions
describe("Validation Functions", () => {
  describe("validateGroupName", () => {
    test("should validate valid group names", () => {
      const result = validateGroupName("My Group");
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject empty group names", () => {
      const result = validateGroupName("");
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe("Group name is required");
    });

    test("should reject group names that are too short", () => {
      const result = validateGroupName("A");
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe(
        "Group name must be at least 2 characters long"
      );
    });

    test("should reject group names that are too long", () => {
      const longName = "A".repeat(101);
      const result = validateGroupName(longName);
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe(
        "Group name must be less than 100 characters"
      );
    });
  });

  describe("validateMemberName", () => {
    test("should validate valid member names", () => {
      const result = validateMemberName("John Doe");
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject empty member names", () => {
      const result = validateMemberName("");
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe("Member name is required");
    });

    test("should reject member names that are too long", () => {
      const longName = "A".repeat(51);
      const result = validateMemberName(longName);
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe(
        "Member name must be less than 50 characters"
      );
    });
  });

  describe("validateItemTitle", () => {
    test("should validate valid item titles", () => {
      const result = validateItemTitle("Visit Paris");
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject empty item titles", () => {
      const result = validateItemTitle("");
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe("Item title is required");
    });

    test("should reject item titles that are too long", () => {
      const longTitle = "A".repeat(201);
      const result = validateItemTitle(longTitle);
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe(
        "Item title must be less than 200 characters"
      );
    });
  });

  describe("validateItemDescription", () => {
    test("should validate valid descriptions", () => {
      const result = validateItemDescription("A nice description");
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should validate undefined descriptions", () => {
      const result = validateItemDescription(undefined);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject descriptions that are too long", () => {
      const longDescription = "A".repeat(1001);
      const result = validateItemDescription(longDescription);
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe(
        "Item description must be less than 1000 characters"
      );
    });
  });

  describe("validateDeadline", () => {
    test("should validate future deadlines", () => {
      const futureDate = new Date(Date.now() + 86400000).toISOString(); // Tomorrow
      const result = validateDeadline(futureDate);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should validate undefined deadlines", () => {
      const result = validateDeadline(undefined);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject past deadlines", () => {
      const pastDate = new Date(Date.now() - 86400000).toISOString(); // Yesterday
      const result = validateDeadline(pastDate);
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe("Deadline must be in the future");
    });

    test("should reject invalid date formats", () => {
      const result = validateDeadline("invalid-date");
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe("Invalid deadline format");
    });
  });
});

describe("Request Validation Functions", () => {
  describe("validateCreateGroupRequest", () => {
    test("should validate valid create group requests", () => {
      const request = {
        name: "My Group",
        deadline: new Date(Date.now() + 86400000).toISOString(),
      };
      const result = validateCreateGroupRequest(request);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject invalid create group requests", () => {
      const request = {
        name: "",
        deadline: new Date(Date.now() - 86400000).toISOString(),
      };
      const result = validateCreateGroupRequest(request);
      expect(result.isValid).toBe(false);
      expect(result.errors).toHaveLength(2);
    });
  });

  describe("validateJoinGroupRequest", () => {
    test("should validate valid join group requests", () => {
      const request = { memberName: "John Doe" };
      const result = validateJoinGroupRequest(request);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject invalid join group requests", () => {
      const request = { memberName: "" };
      const result = validateJoinGroupRequest(request);
      expect(result.isValid).toBe(false);
      expect(result.errors[0].message).toBe("Member name is required");
    });
  });

  describe("validateCreateItemRequest", () => {
    test("should validate valid create item requests", () => {
      const request = {
        title: "Visit Paris",
        description: "A wonderful trip",
        memberId: "member-123",
      };
      const result = validateCreateItemRequest(request);
      expect(result.isValid).toBe(true);
      expect(result.errors).toHaveLength(0);
    });

    test("should reject invalid create item requests", () => {
      const request = {
        title: "",
        description: "A".repeat(1001),
        memberId: "",
      };
      const result = validateCreateItemRequest(request);
      expect(result.isValid).toBe(false);
      expect(result.errors).toHaveLength(3);
    });
  });
});

describe("Sanitization Functions", () => {
  describe("sanitizeString", () => {
    test("should trim whitespace", () => {
      expect(sanitizeString("  hello world  ")).toBe("hello world");
    });

    test("should normalize multiple spaces", () => {
      expect(sanitizeString("hello    world")).toBe("hello world");
    });

    test("should handle mixed whitespace", () => {
      expect(sanitizeString("  hello   world  ")).toBe("hello world");
    });
  });

  describe("sanitizeGroupName", () => {
    test("should sanitize group names", () => {
      expect(sanitizeGroupName("  My   Group  ")).toBe("My Group");
    });
  });

  describe("sanitizeMemberName", () => {
    test("should sanitize member names", () => {
      expect(sanitizeMemberName("  John   Doe  ")).toBe("John Doe");
    });
  });

  describe("sanitizeItemTitle", () => {
    test("should sanitize item titles", () => {
      expect(sanitizeItemTitle("  Visit   Paris  ")).toBe("Visit Paris");
    });
  });

  describe("sanitizeItemDescription", () => {
    test("should sanitize item descriptions", () => {
      expect(sanitizeItemDescription("  A   wonderful   trip  ")).toBe(
        "A wonderful trip"
      );
    });
  });
});
