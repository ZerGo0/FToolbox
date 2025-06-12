FROM oven/bun:latest

WORKDIR /app

# Copy package files
COPY package.json bun.lock ./

# Install production dependencies
RUN bun install --frozen-lockfile --production

# Copy source code
COPY . .

# Create directory for SQLite database
RUN mkdir -p /app/data

EXPOSE 3000

# Run with bun
CMD ["bun", "run", "index.ts"] 