FROM oven/bun:latest as builder

WORKDIR /app

# Copy package files
COPY package.json bun.lock ./

# Install dependencies
RUN bun install --frozen-lockfile

# Copy source code
COPY . .

# Build the application
RUN bun build

# Production stage
FROM oven/bun:latest as runner

WORKDIR /app

# Copy the compiled binary
COPY --from=builder /app/server ./server

# Copy drizzle migrations
COPY --from=builder /app/drizzle ./drizzle

# Copy package.json for metadata
COPY --from=builder /app/package.json ./

# Create directory for SQLite database
RUN mkdir -p /app/data

EXPOSE 3000

# Run the compiled binary
CMD ["./server"] 