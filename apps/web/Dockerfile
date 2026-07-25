FROM oven/bun:1.3.14-alpine
WORKDIR /app
COPY package.json bun.lock ./
RUN bun install --frozen-lockfile

FROM oven/bun:1.3.14-alpine
WORKDIR /app
COPY --from=0 /app/node_modules ./node_modules
COPY . .
RUN bun run build

FROM oven/bun:1.3.14-alpine
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=3000
ENV HOSTNAME=0.0.0.0
RUN addgroup --system --gid 1001 app && adduser --system --uid 1001 app
COPY --from=1 --chown=app:app /app/.next/standalone ./
COPY --from=1 --chown=app:app /app/.next/static ./.next/static
USER app
EXPOSE 3000
CMD ["bun", "server.js"]
