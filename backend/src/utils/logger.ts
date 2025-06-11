export class Logger {
  constructor(private context: string) {}

  info(message: string, ...args: unknown[]): void {
    console.log(`[${new Date().toISOString()}] [${this.context}] INFO: ${message}`, ...args);
  }

  warn(message: string, ...args: unknown[]): void {
    console.warn(`[${new Date().toISOString()}] [${this.context}] WARN: ${message}`, ...args);
  }

  error(message: string, error?: unknown): void {
    console.error(`[${new Date().toISOString()}] [${this.context}] ERROR: ${message}`, error);
  }
}

export const logger = new Logger('App');
