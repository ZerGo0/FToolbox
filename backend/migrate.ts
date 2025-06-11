import { runMigrations } from './src/db/migrate';

runMigrations()
  .then(() => process.exit(0))
  .catch((error) => {
    console.error('Migration failed:', error);
    process.exit(1);
  });