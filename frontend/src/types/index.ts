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

export interface SupabaseUser {
  id: string;
  email: string;
}
