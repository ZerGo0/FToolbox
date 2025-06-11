import { runMigrations } from './src/db/migrate';

async function startServer() {
  try {
    await runMigrations();
  } catch (error) {
    console.error('Failed to run migrations:', error);
    process.exit(1);
  }

  const server = Bun.serve({
    port: 3000,
    fetch(_request) {
      return new Response('Welcome to Bun!');
    },
  });

  console.log(`Listening on ${server.url}`);
}

startServer();
