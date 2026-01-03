FROM golang:1.25 AS backend-build

WORKDIR /backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o brainbook-api .

FROM node:22 AS frontend-build

WORKDIR /frontend
COPY frontend/package.json frontend/pnpm-lock.yaml frontend/pnpm-workspace.yaml ./
RUN corepack enable && corepack prepare pnpm@10.24.0 --activate
RUN pnpm install --frozen-lockfile
COPY frontend ./
RUN pnpm build

FROM node:22-slim
WORKDIR /app

ENV NODE_ENV=production
ENV NITRO_PORT=3000
ENV NITRO_HOST=0.0.0.0
ENV HTTP_PORT=8080

COPY --from=backend-build /backend/brainbook-api /app/brainbook-api
COPY --from=frontend-build /frontend/.output /app/frontend/.output

EXPOSE 3000 8080

CMD ["sh", "-c", "/app/brainbook-api & node /app/frontend/.output/server/index.mjs"]
