# Environment Configuration Guide

This document describes the environment variables required for the Collaborative Bucket List application in different environments.

## Overview

The application consists of two main components:

- **Frontend**: React application built with Vite
- **Backend**: Go API server with WebSocket support

## Frontend Environment Variables

### Required Variables

| Variable                 | Description            | Example                                   | Environment |
| ------------------------ | ---------------------- | ----------------------------------------- | ----------- |
| `VITE_SUPABASE_URL`      | Supabase project URL   | `https://your-project.supabase.co`        | All         |
| `VITE_SUPABASE_ANON_KEY` | Supabase anonymous key | `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` | All         |
| `VITE_API_URL`           | Backend API base URL   | `https://api.yourdomain.com`              | All         |
| `VITE_WS_URL`            | WebSocket server URL   | `wss://api.yourdomain.com`                | All         |

### Optional Variables

| Variable            | Description               | Default                     | Environment |
| ------------------- | ------------------------- | --------------------------- | ----------- |
| `VITE_APP_NAME`     | Application display name  | `Collaborative Bucket List` | All         |
| `VITE_APP_VERSION`  | Application version       | `1.0.0`                     | All         |
| `VITE_ENVIRONMENT`  | Current environment       | `development`               | All         |
| `VITE_ANALYTICS_ID` | Analytics tracking ID     | -                           | Production  |
| `VITE_SENTRY_DSN`   | Sentry error tracking DSN | -                           | Production  |

## Backend Environment Variables

### Required Variables

| Variable                    | Description                  | Example                                   | Environment |
| --------------------------- | ---------------------------- | ----------------------------------------- | ----------- |
| `PORT`                      | Server port                  | `8080`                                    | All         |
| `DATABASE_URL`              | PostgreSQL connection string | `postgresql://user:pass@host:5432/db`     | All         |
| `SUPABASE_URL`              | Supabase project URL         | `https://your-project.supabase.co`        | All         |
| `SUPABASE_SERVICE_ROLE_KEY` | Supabase service role key    | `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...` | All         |
| `SUPABASE_JWT_SECRET`       | Supabase JWT secret          | `your-jwt-secret`                         | All         |
| `ALLOWED_ORIGINS`           | CORS allowed origins         | `https://yourdomain.com`                  | All         |
| `FRONTEND_URL`              | Frontend application URL     | `https://yourdomain.com`                  | All         |

### Optional Variables

| Variable                         | Description               | Default     | Environment |
| -------------------------------- | ------------------------- | ----------- | ----------- |
| `GIN_MODE`                       | Gin framework mode        | `debug`     | All         |
| `HOST`                           | Server bind address       | `localhost` | All         |
| `JWT_SECRET`                     | Additional JWT secret     | -           | Production  |
| `RATE_LIMIT_REQUESTS_PER_MINUTE` | Rate limiting             | `60`        | Production  |
| `LOG_LEVEL`                      | Logging level             | `info`      | All         |
| `LOG_FORMAT`                     | Log output format         | `text`      | All         |
| `HEALTH_CHECK_ENABLED`           | Enable health checks      | `true`      | Production  |
| `HEALTH_CHECK_PATH`              | Health check endpoint     | `/health`   | Production  |
| `SENTRY_DSN`                     | Error tracking DSN        | -           | Production  |
| `METRICS_ENABLED`                | Enable metrics collection | `false`     | Production  |
| `METRICS_PORT`                   | Metrics server port       | `9090`      | Production  |
| `SSL_ENABLED`                    | Enable HTTPS              | `false`     | Production  |
| `SSL_CERT_PATH`                  | SSL certificate path      | -           | Production  |
| `SSL_KEY_PATH`                   | SSL private key path      | -           | Production  |

## Environment Setup

### Development

1. **Frontend**:

   ```bash
   cd frontend
   cp .env.example .env.local
   # Edit .env.local with your development values
   ```

2. **Backend**:
   ```bash
   cd backend
   cp .env.example .env
   # Edit .env with your development values
   ```

### Production

1. **Frontend**:

   ```bash
   cd frontend
   cp .env.production .env.local
   # Edit .env.local with your production values
   npm run build:prod
   ```

2. **Backend**:
   ```bash
   cd backend
   cp .env.production .env
   # Edit .env with your production values
   make build-prod
   ```

## Security Considerations

### Sensitive Variables

The following variables contain sensitive information and should be kept secure:

- `VITE_SUPABASE_ANON_KEY` (Frontend)
- `SUPABASE_SERVICE_ROLE_KEY` (Backend)
- `SUPABASE_JWT_SECRET` (Backend)
- `JWT_SECRET` (Backend)
- `DATABASE_URL` (Backend)
- `SENTRY_DSN` (Both)

### Best Practices

1. **Never commit `.env` files** to version control
2. **Use different keys** for different environments
3. **Rotate secrets regularly** in production
4. **Use environment-specific Supabase projects**
5. **Enable row-level security** in Supabase
6. **Use HTTPS** in production
7. **Implement proper CORS** configuration
8. **Enable rate limiting** in production

## Deployment Checklist

### Frontend Deployment

- [ ] Set `VITE_ENVIRONMENT=production`
- [ ] Configure production Supabase project
- [ ] Set correct API and WebSocket URLs
- [ ] Enable analytics and error tracking
- [ ] Build with `npm run build:prod`
- [ ] Serve static files with proper caching headers

### Backend Deployment

- [ ] Set `GIN_MODE=release`
- [ ] Configure production database
- [ ] Set secure JWT secrets
- [ ] Configure CORS for production domains
- [ ] Enable health checks
- [ ] Set up error tracking
- [ ] Configure SSL certificates (if applicable)
- [ ] Set appropriate rate limits
- [ ] Enable structured logging

## Troubleshooting

### Common Issues

1. **CORS Errors**: Ensure `ALLOWED_ORIGINS` includes your frontend domain
2. **WebSocket Connection Failed**: Check `VITE_WS_URL` and firewall settings
3. **Database Connection Failed**: Verify `DATABASE_URL` and network access
4. **Authentication Errors**: Check Supabase configuration and JWT secrets
5. **Build Failures**: Ensure all required environment variables are set

### Environment Validation

The application includes environment validation on startup. Check the logs for any missing or invalid configuration values.

## Support

For additional help with environment configuration:

1. Check the application logs for specific error messages
2. Verify all required variables are set
3. Test database and Supabase connectivity
4. Ensure network access between components
5. Review Supabase project settings and RLS policies
