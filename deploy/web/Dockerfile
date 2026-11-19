FROM node:24-alpine AS deps
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
ENV VP_GIT_HOOKS=0
RUN corepack enable
WORKDIR /app
COPY package.json pnpm-lock.yaml pnpm-workspace.yaml ./
RUN corepack install
COPY apps/web/package.json ./apps/web/
COPY apps/admin/package.json ./apps/admin/
RUN --mount=type=cache,id=pnpm,target=/pnpm/store pnpm install --frozen-lockfile

FROM deps AS builder
WORKDIR /app
ARG VITE_COMMIT_SHA
ARG VITE_BUILD_TIME
COPY apps/web/ ./apps/web/
COPY apps/admin/ ./apps/admin/
RUN VITE_COMMIT_SHA="${VITE_COMMIT_SHA}" \
    VITE_BUILD_TIME="${VITE_BUILD_TIME}" \
    pnpm build

FROM caddy:2.7.4
COPY deploy/web/Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/apps/web/dist /web
COPY --from=builder /app/apps/admin/dist /admin
