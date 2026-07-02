# Static Site Certification Notes

## Status: Not started

## Known Issues

*None yet — certification pending.*

## Architecture Notes

- Static buildpack is the **universal fallback** — it always matches if no other buildpack does.
- Start command: `npx serve -s . -l {port}` — uses `serve` package via npx.
- Static sites have no install or build step.
- Detecting if a repo is "static HTML" vs something else is done by exclusion (all other buildpacks checked first).
- Consider: should we check for `index.html` before matching Static buildpack?
  - Pro: Would prevent false matches on empty repos
  - Con: Many static sites don't have `index.html` in root (they may have `public/index.html`)

## Potential Issues

- `npx serve` downloads the `serve` package on first use — adds latency to first deployment.
- For production, consider bundling a Go-based static file server or using `go-serve` built into the kernel.
- Large static sites may need configurable `Cache-Control` headers.
