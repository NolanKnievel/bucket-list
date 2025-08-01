import {
  GroupSummary,
  CreateGroupRequest,
  Group,
  GroupWithDetails,
  JoinGroupRequest,
  Member,
  CreateItemRequest,
  ToggleCompletionRequest,
  BucketListItem,
} from "../types";
import { supabase } from "../lib/supabase";

const API_BASE_URL =
  import.meta.env.VITE_API_URL || "http://localhost:8080/api";

class ApiError extends Error {
  constructor(public code: string, message: string, public details?: any) {
    super(message);
    this.name = "ApiError";
  }
}

const getAuthHeaders = async (): Promise<HeadersInit> => {
  const {
    data: { session },
  } = await supabase.auth.getSession();

  if (!session?.access_token) {
    throw new ApiError("NO_TOKEN", "No authentication token found");
  }

  return {
    Authorization: `Bearer ${session.access_token}`,
    "Content-Type": "application/json",
  };
};

const handleResponse = async <T>(response: Response): Promise<T> => {
  if (!response.ok) {
    let errorData;
    try {
      errorData = await response.json();
    } catch {
      throw new ApiError(
        "NETWORK_ERROR",
        `HTTP ${response.status}: ${response.statusText}`
      );
    }

    if (errorData.error) {
      throw new ApiError(
        errorData.error.code,
        errorData.error.message,
        errorData.error.details
      );
    }

    throw new ApiError(
      "UNKNOWN_ERROR",
      `HTTP ${response.status}: ${response.statusText}`
    );
  }

  try {
    return await response.json();
  } catch (error) {
    throw new ApiError("PARSE_ERROR", "Failed to parse response JSON");
  }
};

export const apiService = {
  async getUserGroups(): Promise<GroupSummary[]> {
    try {
      const headers = await getAuthHeaders();
      const response = await fetch(`${API_BASE_URL}/users/groups`, {
        method: "GET",
        headers,
      });

      const data = await handleResponse<{ groups: GroupSummary[] }>(response);
      return data.groups || [];
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("NETWORK_ERROR", "Failed to fetch user groups");
    }
  },

  async createGroup(
    request: CreateGroupRequest
  ): Promise<Group & { shareLink: string }> {
    try {
      const headers = await getAuthHeaders();
      const response = await fetch(`${API_BASE_URL}/groups`, {
        method: "POST",
        headers,
        body: JSON.stringify(request),
      });

      const data = await handleResponse<Group & { shareLink: string }>(
        response
      );
      return data;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("NETWORK_ERROR", "Failed to create group");
    }
  },

  async getGroup(groupId: string): Promise<GroupWithDetails> {
    try {
      const response = await fetch(`${API_BASE_URL}/groups/${groupId}`, {
        method: "GET",
        headers: {
          "Content-Type": "application/json",
        },
      });

      const data = await handleResponse<GroupWithDetails>(response);
      return data;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("NETWORK_ERROR", "Failed to fetch group details");
    }
  },

  async joinGroup(groupId: string, request: JoinGroupRequest): Promise<Member> {
    try {
      // Get current user if authenticated
      const {
        data: { session },
      } = await supabase.auth.getSession();

      const requestBody = {
        ...request,
        userId: session?.user?.id || undefined,
      };

      const response = await fetch(`${API_BASE_URL}/groups/${groupId}/join`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(requestBody),
      });

      const data = await handleResponse<{ member: Member }>(response);
      return data.member;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("NETWORK_ERROR", "Failed to join group");
    }
  },

  async createItem(
    groupId: string,
    request: CreateItemRequest
  ): Promise<BucketListItem> {
    try {
      const response = await fetch(`${API_BASE_URL}/groups/${groupId}/items`, {
        method: "POST",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      const data = await handleResponse<{ item: BucketListItem }>(response);
      return data.item;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("NETWORK_ERROR", "Failed to create item");
    }
  },

  async toggleItemCompletion(
    itemId: string,
    request: ToggleCompletionRequest
  ): Promise<BucketListItem> {
    try {
      const response = await fetch(`${API_BASE_URL}/items/${itemId}/complete`, {
        method: "PATCH",
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(request),
      });

      const data = await handleResponse<{ item: BucketListItem }>(response);
      return data.item;
    } catch (error) {
      if (error instanceof ApiError) {
        throw error;
      }
      throw new ApiError("NETWORK_ERROR", "Failed to toggle item completion");
    }
  },
};

export { ApiError };
