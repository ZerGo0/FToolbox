import { Hono } from 'hono';
import { workerManager } from '../workers/manager';
import { logger } from '../utils/logger';

const app = new Hono();

// Get worker status
app.get('/status', async (c) => {
  try {
    const status = await workerManager.getStatus();
    return c.json({ success: true, data: status });
  } catch (error) {
    logger.error('Failed to get worker status:', error);
    return c.json({ success: false, error: 'Failed to get worker status' }, 500);
  }
});

// Trigger tag update worker manually
app.post('/tag-updater/trigger', async (c) => {
  try {
    // TODO: Add authentication
    await workerManager.runWorker('tag-updater');
    return c.json({ success: true, message: 'Tag updater triggered' });
  } catch (error) {
    logger.error('Failed to trigger tag updater:', error);
    return c.json({ success: false, error: 'Failed to trigger tag updater' }, 500);
  }
});

// Trigger tag discovery worker manually
app.post('/tag-discovery/trigger', async (c) => {
  try {
    // TODO: Add authentication
    await workerManager.runWorker('tag-discovery');
    return c.json({ success: true, message: 'Tag discovery triggered' });
  } catch (error) {
    logger.error('Failed to trigger tag discovery:', error);
    return c.json({ success: false, error: 'Failed to trigger tag discovery' }, 500);
  }
});

// Enable/disable worker
app.patch('/:workerName/enable', async (c) => {
  try {
    const workerName = c.req.param('workerName');
    const { enabled } = await c.req.json();

    // TODO: Add authentication
    // TODO: Update worker enabled status in database

    if (enabled) {
      await workerManager.start(workerName);
    } else {
      await workerManager.stop(workerName);
    }

    return c.json({
      success: true,
      message: `Worker ${workerName} ${enabled ? 'enabled' : 'disabled'}`,
    });
  } catch (error) {
    logger.error('Failed to update worker status:', error);
    return c.json({ success: false, error: 'Failed to update worker status' }, 500);
  }
});

export default app;
