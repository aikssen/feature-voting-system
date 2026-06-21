# syntax=docker/dockerfile:1
#
# Per DECISIONS.md C5, the frontend ships as the Vite dev server (review-friendly
# stack), not a production nginx build.

FROM node:22-alpine
RUN corepack enable
WORKDIR /app

# Install deps first so they cache across source edits. The pnpm-workspace.yaml
# carries allowBuilds for esbuild/@tailwindcss/oxide native postinstalls.
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile

COPY . .

EXPOSE 5173
CMD ["pnpm", "dev", "--host", "--port", "5173"]
