import {
  BrowserRouter as Router,
  Routes,
  Route,
  Navigate,
} from "react-router-dom";
import { AuthProvider } from "./contexts/AuthContext";
import { NativeWebSocketProvider } from "./contexts/NativeWebSocketContext";
import { HomePage } from "./components/HomePage";
import { Dashboard } from "./components/Dashboard";
import { CreateGroupPage } from "./components/CreateGroupPage";
import { JoinGroupForm } from "./components/JoinGroupForm";
import { GroupView } from "./components/GroupView";
import { ProtectedRoute } from "./components/ProtectedRoute";

function App() {
  return (
    <AuthProvider>
      <NativeWebSocketProvider>
        <Router>
          <Routes>
            {/* Public route - Home page with authentication */}
            <Route path="/" element={<HomePage />} />

            {/* Protected routes - require authentication */}
            <Route
              path="/dashboard"
              element={
                <ProtectedRoute>
                  <Dashboard />
                </ProtectedRoute>
              }
            />
            <Route
              path="/create-group"
              element={
                <ProtectedRoute>
                  <CreateGroupPage />
                </ProtectedRoute>
              }
            />

            {/* Public routes - Group access */}
            <Route path="/groups/:groupId/join" element={<JoinGroupForm />} />
            <Route path="/groups/:groupId" element={<GroupView />} />

            {/* Catch all route - redirect to home */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </Router>
      </NativeWebSocketProvider>
    </AuthProvider>
  );
}

export default App;
