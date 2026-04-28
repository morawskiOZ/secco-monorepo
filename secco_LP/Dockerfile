# Stage 1: Build SvelteKit static site
FROM node:24-alpine AS frontend
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY svelte.config.js vite.config.ts tsconfig.json ./
COPY src/ src/
COPY static/ static/
ARG PUBLIC_TURNSTILE_SITE_KEY
ENV PUBLIC_TURNSTILE_SITE_KEY=${PUBLIC_TURNSTILE_SITE_KEY}
RUN npm run build

# Stage 2: Build Go binary with embedded static files
FROM golang:1.23-alpine AS backend
WORKDIR /app
COPY backend/go.mod ./
RUN go mod download
COPY backend/*.go ./
COPY --from=frontend /app/build ./build/
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o secco-studio .

# Stage 3: Minimal runtime
FROM scratch
LABEL org.opencontainers.image.source="https://github.com/morawskiOZ/secco-monorepo"
LABEL org.opencontainers.image.description="Secco Studio - Website & Valuation Form"
COPY --from=backend /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=backend /app/secco-studio /secco-studio
EXPOSE 8080
ENTRYPOINT ["/secco-studio"]
