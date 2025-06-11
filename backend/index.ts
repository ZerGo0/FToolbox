import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { runMigrations } from './src/db/migrate';
import tagsRouter from './src/routes/tags';
import workersRouter from './src/routes/workers';
import { workerManager } from './src/workers/manager';
import { tagDiscoveryWorker } from './src/workers/tag-discovery';
import { tagUpdaterWorker } from './src/workers/tag-updater';

const app = new Hono();

// Enable CORS for frontend
app.use(
  '/*',
  cors({
    origin: 'http://localhost:5173',
    credentials: true,
  })
);

// Health check
app.get('/', (c) => c.json({ status: 'ok' }));

// Mount routes
app.route('/api/tags', tagsRouter);
app.route('/api/workers', workersRouter);

async function startServer() {
  try {
    await runMigrations();
  } catch (error) {
    console.error('Failed to run migrations:', error);
    process.exit(1);
  }

  const server = Bun.serve({
    port: 3000,
    fetch: app.fetch,
  });

  // Register workers
  await workerManager.register(tagUpdaterWorker);
  await workerManager.register(tagDiscoveryWorker);

  // Start workers if enabled
  if (process.env.WORKER_ENABLED !== 'false') {
    await workerManager.start('tag-updater');
    await workerManager.start('tag-discovery');
  }

  console.log(`Server running at ${server.url}`);
  console.log(`Workers enabled: ${process.env.WORKER_ENABLED !== 'false'}`);
}

// Graceful shutdown
process.on('SIGINT', async () => {
  console.log('Shutting down...');
  await workerManager.stopAll();
  process.exit(0);
});

startServer();
