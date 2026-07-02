# React Certification Notes

## Status: Not started

## Known Issues

*None yet — certification pending.*

## Architecture Notes

- React buildpack produces `ArtifactTypeStatic` (static files).
- After build, files are served via `npx serve -s {outputDir} -l {port}`.
- CRA outputs to `build/` directory; Vite outputs to `dist/` — buildpack detects Vite by checking for `vite` in devDependencies.
- React apps have no runtime server — the start command uses `npx serve` which is fetched on demand.
- Consider bundling `serve` with the kernel or using a Go-based static file server for production.
