# Stage 1: Build SvelteKit static site
FROM node:24-alpine AS frontend
WORKDIR /app
COPY package.json package-lock.json ./
COPY secco-render-1.0.0.tgz ./
RUN sed -i 's|file:../secco_renderer/secco-render-1.0.0.tgz|file:./secco-render-1.0.0.tgz|' package.json
RUN npm ci
COPY svelte.config.js vite.config.ts tsconfig.json ./
COPY src/ src/
RUN npm run build

# Stage 2: Build Go binary with embedded static files
FROM golang:1.25-alpine AS backend
WORKDIR /app
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/*.go ./
COPY --from=frontend /app/build ./build/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o secco-cms .

# Stage 3: Minimal runtime
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/morawskiOZ/secco-monorepo"
LABEL org.opencontainers.image.description="Secco CMS - Content Management System"
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /app/secco-cms /secco-cms
EXPOSE 8081
ENTRYPOINT ["/secco-cms"]
