import { migrate } from 'drizzle-orm/libsql/migrator';
import { db } from './index';

export async function runMigrations() {
  console.log('Running migrations...');

  try {
    await migrate(db, { migrationsFolder: './drizzle' });
    console.log('Migrations completed successfully');
  } catch (error) {
    console.error('Migration failed:', error);
    throw error;
  }
}
