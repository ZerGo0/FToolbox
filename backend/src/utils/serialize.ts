/**
 * Converts Date objects to Unix timestamps (seconds) for API responses
 */
export function serializeDates<T>(obj: T): T {
  if (obj === null || obj === undefined) {
    return obj;
  }

  if (obj instanceof Date) {
    return Math.floor(obj.getTime() / 1000) as T;
  }

  if (Array.isArray(obj)) {
    return obj.map((item) => serializeDates(item)) as T;
  }

  if (typeof obj === 'object') {
    const result: Record<string, unknown> = {};
    for (const key in obj) {
      if (Object.prototype.hasOwnProperty.call(obj, key)) {
        result[key] = serializeDates(obj[key]);
      }
    }
    return result as T;
  }

  return obj;
}
