import { db } from '../db';
import { workers } from '../db/schema';
import { eq, sql } from 'drizzle-orm';
import { logger } from '../utils/logger';

export interface Worker {
  name: string;
  interval: number; // milliseconds
  run: () => Promise<void>;
}

export class WorkerManager {
  private workers = new Map<string, { worker: Worker; timer?: Timer; isRunning?: boolean }>();
  private isShuttingDown = false;

  constructor() {
    process.on('SIGINT', () => this.shutdown());
    process.on('SIGTERM', () => this.shutdown());
  }

  async register(worker: Worker) {
    if (this.workers.has(worker.name)) {
      throw new Error(`Worker ${worker.name} is already registered`);
    }

    this.workers.set(worker.name, { worker });
    logger.info(`Registered worker: ${worker.name}`);

    // Initialize worker in database
    try {
      await db
        .insert(workers)
        .values({
          name: worker.name,
          status: 'idle',
          runCount: 0,
          successCount: 0,
          failureCount: 0,
          isEnabled: true,
        })
        .onConflictDoNothing();
    } catch (error) {
      logger.error(`Failed to initialize worker ${worker.name} in database:`, error);
    }
  }

  async start(workerName: string) {
    const workerData = this.workers.get(workerName);
    if (!workerData) {
      throw new Error(`Worker ${workerName} not found`);
    }

    const { worker } = workerData;

    // Check if worker is enabled
    const workerRecord = await db.select().from(workers).where(eq(workers.name, workerName)).get();
    if (!workerRecord?.isEnabled) {
      logger.info(`Worker ${workerName} is disabled`);
      return;
    }

    // Run immediately (without blocking), then schedule
    this.runWorker(workerName).catch((error) => {
      logger.error(`Initial run of worker ${workerName} failed:`, error);
    });

    // Schedule periodic runs
    const timer = setInterval(async () => {
      if (!this.isShuttingDown) {
        // Run without blocking the interval
        this.runWorker(workerName).catch((error) => {
          logger.error(`Scheduled run of worker ${workerName} failed:`, error);
        });
      }
    }, worker.interval);

    this.workers.set(workerName, { worker, timer });
    logger.info(`Started worker: ${workerName} (interval: ${worker.interval}ms)`);
  }

  async stop(workerName: string) {
    const workerData = this.workers.get(workerName);
    if (!workerData || !workerData.timer) {
      return;
    }

    clearInterval(workerData.timer);
    workerData.timer = undefined;
    logger.info(`Stopped worker: ${workerName}`);
  }

  async runWorker(workerName: string) {
    const workerData = this.workers.get(workerName);
    if (!workerData) {
      throw new Error(`Worker ${workerName} not found`);
    }

    // Skip if worker is already running
    if (workerData.isRunning) {
      logger.warn(`Worker ${workerName} is already running, skipping this run`);
      return;
    }

    const { worker } = workerData;
    try {
      // Mark as running
      workerData.isRunning = true;
      this.workers.set(workerName, workerData);

      logger.info(`Running worker: ${worker.name}`);

      // Update status to running
      await db
        .update(workers)
        .set({
          status: 'running',
          lastRunAt: new Date(),
          updatedAt: new Date(),
        })
        .where(eq(workers.name, worker.name));

      // Run the worker
      await worker.run();

      // Update status to idle and increment success count
      await db
        .update(workers)
        .set({
          status: 'idle',
          runCount: sql`${workers.runCount} + 1`,
          successCount: sql`${workers.successCount} + 1`,
          lastError: null,
          nextRunAt: new Date(Date.now() + worker.interval),
          updatedAt: new Date(),
        })
        .where(eq(workers.name, worker.name));

      logger.info(`Worker ${worker.name} completed successfully`);
    } catch (error) {
      logger.error(`Worker ${worker.name} failed:`, error);

      // Update status to failed and increment failure count
      await db
        .update(workers)
        .set({
          status: 'failed',
          runCount: sql`${workers.runCount} + 1`,
          failureCount: sql`${workers.failureCount} + 1`,
          lastError: error instanceof Error ? error.message : String(error),
          nextRunAt: new Date(Date.now() + worker.interval),
          updatedAt: new Date(),
        })
        .where(eq(workers.name, worker.name));
    } finally {
      // Always mark as not running
      workerData.isRunning = false;
      this.workers.set(workerName, workerData);
    }
  }

  async startAll() {
    const enabled = process.env.WORKER_ENABLED !== 'false';
    if (!enabled) {
      logger.info('Workers are disabled by environment variable');
      return;
    }

    // Start all workers in parallel
    const startPromises = Array.from(this.workers.keys()).map(async (name) => {
      try {
        await this.start(name);
      } catch (error) {
        logger.error(`Failed to start worker ${name}:`, error);
      }
    });

    await Promise.all(startPromises);
  }

  async stopAll() {
    for (const [name] of this.workers) {
      await this.stop(name);
    }
  }

  async shutdown() {
    if (this.isShuttingDown) return;

    this.isShuttingDown = true;
    logger.info('Shutting down worker manager...');
    await this.stopAll();
    logger.info('Worker manager shutdown complete');
    process.exit(0);
  }

  async getStatus() {
    const workerRecords = await db.select().from(workers).all();
    return workerRecords.map((record) => ({
      ...record,
      isRunning:
        this.workers.has(record.name) && this.workers.get(record.name)?.timer !== undefined,
    }));
  }
}

export const workerManager = new WorkerManager();
