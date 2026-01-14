FROM node:22 AS build

WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN corepack enable && corepack prepare pnpm@10.24.0 --activate
RUN pnpm install --frozen-lockfile

COPY . .
RUN pnpm build

FROM node:22-slim
WORKDIR /app

ENV NODE_ENV=production
ENV NITRO_PORT=3000
ENV NITRO_HOST=0.0.0.0

COPY --from=build /app/.output /app/.output

EXPOSE 3000
CMD ["node", "/app/.output/server/index.mjs"]
