# Next.js Certification Notes

## Status: Not started

## Known Issues

### CERT-014 (Fixed: 80a84f5)
- **Problem:** `next start` ignored `PORT` env var
- **Root cause:** Next.js reads port from `-p` flag, not `PORT` env var
- **Resolution:** Changed default start command from `npm start` to `npx next start -p {port}`
- **Commit:** 80a84f5

## Architecture Notes

- Next.js uses `next start` for production mode, not `npm run dev`.
- The user's `scripts.start` in `package.json` is preferred if it handles port correctly.
- Default fallback is now `npx next start -p {port}` (unlike Node.js which reads `PORT`).
- `NEXT_TELEMETRY_DISABLED=1` env var is set to disable telemetry.
- The build produces a `.next` directory containing the compiled application.
