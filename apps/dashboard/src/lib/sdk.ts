import { CloudOSClient } from '@cloudos/sdk';

/**
 * Singleton SDK instance.
 *
 * In the Vite dev server, `/api/*` is proxied to the Go backend at localhost:8080.
 * The SDK client methods already include the /api/v1/ path prefix, so the base
 * URL is empty (same origin).
 *
 * In production, set VITE_CLOUDOS_API_URL to the full API origin
 * (e.g. "https://api.cloudos.io").
 */
export const cloudos = new CloudOSClient(
  import.meta.env.VITE_CLOUDOS_API_URL ?? '',
);
