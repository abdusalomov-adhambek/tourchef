# ---- build ----
FROM node:20-alpine AS builder
WORKDIR /app
# NUXT routeRules are compiled at build time, so the proxy target must be
# provided as a build argument (a runtime env var alone is not enough).
ARG API_PROXY_TARGET=http://tourchef-backend:8080
ENV API_PROXY_TARGET=$API_PROXY_TARGET
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
RUN npm run build

# ---- runtime ----
FROM node:20-alpine AS runtime
WORKDIR /app
ENV NODE_ENV=production \
    NITRO_HOST=0.0.0.0 \
    NITRO_PORT=3000 \
    PORT=3000
COPY --from=builder /app/.output ./.output
EXPOSE 3000
CMD ["node", ".output/server/index.mjs"]
