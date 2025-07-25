import React, { useState } from "react";
import { useAuth } from "../contexts/AuthContext";
import { createAuthHeaders } from "../utils/auth";

export const BackendTest: React.FC = () => {
  const { user } = useAuth();
  const [testResult, setTestResult] = useState<string>("");
  const [isLoading, setIsLoading] = useState(false);

  const testBackendAuth = async () => {
    setIsLoading(true);
    setTestResult("");

    try {
      const headers = await createAuthHeaders();

      if (!headers.Authorization) {
        setTestResult("❌ No auth token available. Please sign in first.");
        return;
      }

      const response = await fetch(
        `${import.meta.env.VITE_API_URL}/api/auth/verify`,
        {
          method: "POST",
          headers,
        }
      );

      const data = await response.json();

      if (response.ok) {
        setTestResult(
          `✅ Backend authentication successful!\n\nResponse:\n${JSON.stringify(
            data,
            null,
            2
          )}`
        );
      } else {
        setTestResult(
          `❌ Backend authentication failed!\n\nError:\n${JSON.stringify(
            data,
            null,
            2
          )}`
        );
      }
    } catch (error) {
      setTestResult(
        `❌ Network error: ${
          error instanceof Error ? error.message : "Unknown error"
        }`
      );
    } finally {
      setIsLoading(false);
    }
  };

  const testHealthEndpoint = async () => {
    setIsLoading(true);
    setTestResult("");

    try {
      const response = await fetch(`${import.meta.env.VITE_API_URL}/health`);
      const data = await response.json();

      if (response.ok) {
        setTestResult(
          `✅ Backend health check successful!\n\nResponse:\n${JSON.stringify(
            data,
            null,
            2
          )}`
        );
      } else {
        setTestResult(
          `❌ Backend health check failed!\n\nError:\n${JSON.stringify(
            data,
            null,
            2
          )}`
        );
      }
    } catch (error) {
      setTestResult(
        `❌ Network error: ${
          error instanceof Error ? error.message : "Unknown error"
        }\n\nMake sure your backend is running on ${
          import.meta.env.VITE_API_URL
        }`
      );
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="max-w-2xl mx-auto bg-white rounded-lg shadow-md p-6 mt-6">
      <h3 className="text-xl font-bold text-center mb-6">
        Backend API Testing
      </h3>

      <div className="space-y-4">
        <div className="flex gap-4">
          <button
            onClick={testHealthEndpoint}
            disabled={isLoading}
            className="flex-1 bg-green-500 hover:bg-green-600 disabled:bg-green-300 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline transition duration-200"
          >
            {isLoading ? "Testing..." : "Test Health Endpoint"}
          </button>

          <button
            onClick={testBackendAuth}
            disabled={isLoading || !user}
            className="flex-1 bg-purple-500 hover:bg-purple-600 disabled:bg-purple-300 text-white font-bold py-2 px-4 rounded focus:outline-none focus:shadow-outline transition duration-200"
          >
            {isLoading ? "Testing..." : "Test Auth Endpoint"}
          </button>
        </div>

        {!user && (
          <p className="text-sm text-gray-600 text-center">
            Sign in to test the authenticated endpoint
          </p>
        )}

        {testResult && (
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-4">
            <h4 className="font-semibold mb-2">Test Result:</h4>
            <pre className="text-sm whitespace-pre-wrap overflow-x-auto">
              {testResult}
            </pre>
          </div>
        )}
      </div>
    </div>
  );
};
