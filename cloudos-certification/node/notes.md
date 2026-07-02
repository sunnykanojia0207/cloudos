# Node.js Certification Notes

## Status: Not started

## Known Issues

*None yet — certification pending.*

## Architecture Notes

- Node.js buildpack reads `scripts.start` from `package.json`; falls back to `npm start`.
- `PORT` env var is set by the executor — Express/Fastify/Koa apps read `process.env.PORT` by convention.
- NestJS requires a build step (`npm run build`) before the start command.
- Install uses `npm install` (not `npm ci`) — this may need to be addressed for production deployments with lockfiles.

## Common Patterns

Apps that will work out of the box:

```js
const port = process.env.PORT || 3000;
app.listen(port, () => console.log(`Listening on ${port}`));
```
