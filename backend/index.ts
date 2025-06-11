import { Hono } from 'hono';
import { cors } from 'hono/cors';
import { runMigrations } from './src/db/migrate';
import tagsRouter from './src/routes/tags';
import { workerManager } from './src/workers/manager';
import { rankCalculatorWorker } from './src/workers/rank-calculator';
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

// Worker status endpoint
app.get('/api/workers/status', async (c) => {
  const workers = await workerManager.getStatus();

  // Check if any worker is running
  const isRunning = workers.some((w) => w.status === 'running');
  // Check if any worker has failed
  const hasFailed = workers.some((w) => w.status === 'failed');

  let status: 'idle' | 'running' | 'failed' = 'idle';
  if (hasFailed) {
    status = 'failed';
  } else if (isRunning) {
    status = 'running';
  }

  return c.json({ status });
});

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
  await workerManager.register(rankCalculatorWorker);

  // Start workers if enabled (in parallel)
  if (process.env.WORKER_ENABLED !== 'false') {
    await Promise.all([
      workerManager.start('tag-updater'),
      workerManager.start('tag-discovery'),
      workerManager.start('rank-calculator'),
    ]);
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
