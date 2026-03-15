# Build stage
FROM node:20-alpine AS build

WORKDIR /app

# Copy package files
COPY package*.json ./

# Install dependencies
RUN npm ci

# Copy source files
COPY . .

# Build the app (no hardcoded API URL — nginx proxies /api/ to backend)
RUN npm run build

# Production stage
FROM nginx:alpine

# Copy built assets from build stage
COPY --from=build /app/dist /usr/share/nginx/html

# Fix permissions so nginx (non-root) can read all static files
RUN chmod -R 644 /usr/share/nginx/html && \
    find /usr/share/nginx/html -type d -exec chmod 755 {} \;

# Copy nginx config template
COPY nginx.conf /etc/nginx/templates/default.conf.template

# Expose port 8080 (Cloud Run requirement)
EXPOSE 8080

# API_UPSTREAM defaults to Cloud Run URL; override to http://api:8080 in docker compose
ENV API_UPSTREAM=https://agent-united-api-961985674922.us-central1.run.app

# nginx:alpine entrypoint handles envsubst on /etc/nginx/templates/*.template automatically
CMD ["nginx", "-g", "daemon off;"]
