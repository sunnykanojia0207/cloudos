# Go Certification Notes

## Status: Not started

## Known Issues

*None yet — certification pending.*

## Architecture Notes

- Go buildpack uses `go mod download` for install and `go build -o app .` for build.
- On Windows, `fixupPlatformPlan` adjusts the binary name to `app.exe`.
- Go applications receive `PORT` env var set by the executor. Apps should read `os.Getenv("PORT")` or `os.Getenv("PORT")` with fallback to `DevPort` (8080).
- The `{port}` placeholder is NOT in the default start command (`./app`) — Go apps are expected to read `PORT` env var instead.

## Common Patterns

Apps that will work out of the box:

```go
// Standard library HTTP server reading PORT env var
port := os.Getenv("PORT")
if port == "" {
    port = "8080"
}
http.ListenAndServe(":"+port, nil)
```

Apps using frameworks (Gin, Echo, Fiber) typically read PORT via framework config.
