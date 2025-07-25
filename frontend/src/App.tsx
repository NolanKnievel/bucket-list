import { AuthProvider } from "./contexts/AuthContext";
import { AuthForm } from "./components/AuthForm";
import { BackendTest } from "./components/BackendTest";

function App() {
  return (
    <AuthProvider>
      <div className="min-h-screen bg-gray-50">
        <div className="container mx-auto px-4 py-8">
          <header className="text-center mb-8">
            <h1 className="text-4xl font-bold text-gray-900 mb-2">
              Collaborative Bucket List
            </h1>
            <p className="text-lg text-gray-600">
              Authentication System Testing
            </p>
          </header>

          <main className="max-w-4xl mx-auto space-y-6">
            <AuthForm />
            <BackendTest />
          </main>
        </div>
      </div>
    </AuthProvider>
  );
}

export default App;
