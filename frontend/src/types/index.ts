// Core TypeScript interfaces and types will be defined here
export interface Group {
  id: string;
  name: string;
  deadline?: string;
  createdAt: string;
  createdBy: string;
}

export interface Member {
  id: string;
  groupId: string;
  userId?: string;
  name: string;
  joinedAt: string;
  isCreator: boolean;
}

export interface BucketListItem {
  id: string;
  groupId: string;
  title: string;
  description?: string;
  completed: boolean;
  completedBy?: string;
  completedAt?: string;
  createdBy: string;
  createdAt: string;
}

export interface GroupWithDetails extends Group {
  members: Member[];
  items: BucketListItem[];
}

export interface GroupSummary extends Group {
  memberCount: number;
  itemCount: number;
  completedCount: number;
  progressPercent: number;
}

export interface SupabaseUser {
  id: string;
  email: string;
}

// Request/Response types for API
export interface CreateGroupRequest {
  name: string;
  deadline?: string;
}

export interface JoinGroupRequest {
  memberName: string;
  userId?: string;
}

export interface CreateItemRequest {
  title: string;
  description?: string;
  memberId: string;
}

export interface ToggleCompletionRequest {
  completed: boolean;
  memberId: string;
}

// WebSocket event types
export interface WebSocketEvent<T = any> {
  type: string;
  data: T;
}

export interface JoinGroupEvent {
  groupId: string;
  memberId: string;
}

export interface AddItemEvent {
  groupId: string;
  item: CreateItemRequest;
}

export interface ToggleCompletionEvent {
  groupId: string;
  itemId: string;
  completed: boolean;
}

// Validation error types
export interface ValidationError {
  field: string;
  message: string;
}

export interface ValidationResult {
  isValid: boolean;
  errors: ValidationError[];
}

// Validation functions
export const validateGroupName = (name: string): ValidationResult => {
  const errors: ValidationError[] = [];

  if (!name || name.trim().length === 0) {
    errors.push({ field: "name", message: "Group name is required" });
  } else if (name.trim().length < 2) {
    errors.push({
      field: "name",
      message: "Group name must be at least 2 characters long",
    });
  } else if (name.trim().length > 100) {
    errors.push({
      field: "name",
      message: "Group name must be less than 100 characters",
    });
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

export const validateMemberName = (name: string): ValidationResult => {
  const errors: ValidationError[] = [];

  if (!name || name.trim().length === 0) {
    errors.push({ field: "name", message: "Member name is required" });
  } else if (name.trim().length < 1) {
    errors.push({ field: "name", message: "Member name cannot be empty" });
  } else if (name.trim().length > 50) {
    errors.push({
      field: "name",
      message: "Member name must be less than 50 characters",
    });
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

export const validateItemTitle = (title: string): ValidationResult => {
  const errors: ValidationError[] = [];

  if (!title || title.trim().length === 0) {
    errors.push({ field: "title", message: "Item title is required" });
  } else if (title.trim().length < 1) {
    errors.push({ field: "title", message: "Item title cannot be empty" });
  } else if (title.trim().length > 200) {
    errors.push({
      field: "title",
      message: "Item title must be less than 200 characters",
    });
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

export const validateItemDescription = (
  description?: string
): ValidationResult => {
  const errors: ValidationError[] = [];

  if (description && description.length > 1000) {
    errors.push({
      field: "description",
      message: "Item description must be less than 1000 characters",
    });
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

export const validateDeadline = (deadline?: string): ValidationResult => {
  const errors: ValidationError[] = [];

  if (deadline) {
    const deadlineDate = new Date(deadline);
    const now = new Date();

    if (isNaN(deadlineDate.getTime())) {
      errors.push({ field: "deadline", message: "Invalid deadline format" });
    } else if (deadlineDate <= now) {
      errors.push({
        field: "deadline",
        message: "Deadline must be in the future",
      });
    }
  }

  return {
    isValid: errors.length === 0,
    errors,
  };
};

export const validateCreateGroupRequest = (
  request: CreateGroupRequest
): ValidationResult => {
  const nameValidation = validateGroupName(request.name);
  const deadlineValidation = validateDeadline(request.deadline);

  return {
    isValid: nameValidation.isValid && deadlineValidation.isValid,
    errors: [...nameValidation.errors, ...deadlineValidation.errors],
  };
};

export const validateJoinGroupRequest = (
  request: JoinGroupRequest
): ValidationResult => {
  return validateMemberName(request.memberName);
};

export const validateCreateItemRequest = (
  request: CreateItemRequest
): ValidationResult => {
  const titleValidation = validateItemTitle(request.title);
  const descriptionValidation = validateItemDescription(request.description);

  const errors: ValidationError[] = [];

  if (!request.memberId || request.memberId.trim().length === 0) {
    errors.push({ field: "memberId", message: "Member ID is required" });
  }

  return {
    isValid:
      titleValidation.isValid &&
      descriptionValidation.isValid &&
      errors.length === 0,
    errors: [
      ...titleValidation.errors,
      ...descriptionValidation.errors,
      ...errors,
    ],
  };
};

// Utility functions for data sanitization
export const sanitizeString = (input: string): string => {
  return input.trim().replace(/\s+/g, " ");
};

export const sanitizeGroupName = (name: string): string => {
  return sanitizeString(name);
};

export const sanitizeMemberName = (name: string): string => {
  return sanitizeString(name);
};

export const sanitizeItemTitle = (title: string): string => {
  return sanitizeString(title);
};

export const sanitizeItemDescription = (description: string): string => {
  return sanitizeString(description);
};
