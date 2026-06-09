FROM node:22-alpine AS base
RUN npm install -g pnpm
WORKDIR /app
COPY package.json pnpm-lock.yaml* ./
RUN pnpm install --frozen-lockfile

# Development stage (used by docker-compose)
FROM base AS dev
COPY . .
EXPOSE 5173
CMD ["pnpm", "dev", "--host"]

# Production build
FROM base AS builder
COPY . .
RUN pnpm build

FROM node:22-alpine AS prod
RUN npm install -g pnpm
WORKDIR /app
COPY --from=builder /app/build ./build
COPY --from=builder /app/package.json .
RUN pnpm install --prod
EXPOSE 3001
CMD ["node", "build"]
