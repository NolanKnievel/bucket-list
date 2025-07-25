# Collaborative Bucket List

A real-time web application that enables groups to collaboratively create and manage shared bucket lists.

## Project Structure

```
├── frontend/                 # React TypeScript frontend
│   ├── src/
│   │   ├── components/      # React components
│   │   ├── contexts/        # React contexts
│   │   ├── hooks/           # Custom React hooks
│   │   ├── types/           # TypeScript type definitions
│   │   ├── utils/           # Utility functions
│   │   └── test/            # Test setup and utilities
│   ├── package.json         # Frontend dependencies
│   ├── vite.config.ts       # Vite configuration
│   ├── tailwind.config.js   # Tailwind CSS configuration
│   └── tsconfig.json        # TypeScript configuration
│
└── backend/                 # Go backend API
    ├── cmd/                 # Application entry points
    ├── internal/
    │   ├── handlers/        # HTTP request handlers
    │   ├── middleware/      # HTTP middleware
    │   ├── models/          # Data models
    │   ├── repositories/    # Data access layer
    │   └── services/        # Business logic layer
    ├── pkg/                 # Public packages
    └── go.mod               # Go module dependencies
```

## Technology Stack

- **Frontend**: React, TypeScript, Vite, Tailwind CSS
- **Backend**: Go, Gin framework, Gorilla WebSocket
- **Database**: PostgreSQL (via Supabase)
- **Authentication**: Supabase Auth
- **Real-time**: WebSockets

## Getting Started

### Prerequisites

- Node.js 18+ and npm
- Go 1.21+
- PostgreSQL database (or Supabase account)

### Frontend Setup

```bash
cd frontend
npm install
cp .env.example .env
# Edit .env with your configuration
npm run dev
```

### Backend Setup

```bash
cd backend
go mod tidy
cp .env.example .env
# Edit .env with your configuration
go run cmd/main.go
```

## Environment Variables

See `.env.example` files in both `frontend/` and `backend/` directories for required configuration.
