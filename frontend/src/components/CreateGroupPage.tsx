import React from "react";
import { useNavigate } from "react-router-dom";
import { CreateGroupForm } from "./CreateGroupForm";

export const CreateGroupPage: React.FC = () => {
  const navigate = useNavigate();

  const handleSuccess = (groupId: string, shareLink: string) => {
    // For now, just stay on the success screen
    // In future tasks, we might navigate to the group view
    console.log("Group created successfully:", { groupId, shareLink });
  };

  const handleCancel = () => {
    navigate("/dashboard");
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="container mx-auto px-4 py-8">
        <header className="mb-8">
          <button
            onClick={handleCancel}
            className="inline-flex items-center text-blue-600 hover:text-blue-800 font-medium"
          >
            <svg
              className="w-4 h-4 mr-2"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M15 19l-7-7 7-7"
              />
            </svg>
            Back to Dashboard
          </button>
        </header>

        <main className="max-w-2xl mx-auto">
          <CreateGroupForm onSuccess={handleSuccess} onCancel={handleCancel} />
        </main>
      </div>
    </div>
  );
};
