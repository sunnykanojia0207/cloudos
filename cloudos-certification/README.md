# CloudOS Certification Program

> **Proving that CloudOS can deploy real applications repeatedly, reliably, and without manual intervention.**

This directory contains the formal certification program for CloudOS v0.6. Every supported stack is tested against multiple real-world repositories. A stack is **certified** only when every repository passes the full deployment lifecycle.

## Certification Criteria

For each repository, the following must succeed without manual intervention:

| # | Step | Criterion |
|---|------|-----------|
| 1 | **Clone** | Git repository is cloned successfully |
| 2 | **Detect** | Correct buildpack is automatically detected |
| 3 | **Install** | Dependencies are installed (within venv/isolated env) |
| 4 | **Build** | Application builds successfully (if a build step is required) |
| 5 | **Start** | Application starts and binds to the allocated port |
| 6 | **Health** | Health check returns HTTP 200 within timeout |
| 7 | **URL** | Application endpoint is accessible and returns content |
| 8 | **Logs** | Live logs are streaming install, build, and runtime output |
| 9 | **Timeline** | Workflow timeline shows every step with correct status |
| 10 | **Report** | Deployment report contains commit SHA, detected runtime, artifact info |
| 11 | **Redeploy** | A second deployment succeeds (gracefully stops the first) |
| 12 | **Stop** | Application stops cleanly via Stop |
| 13 | **Cleanup** | Resources are released after Destroy |

## Overall Scoreboard

| Stack | Passed | Total | Progress | Status |
|-------|--------|-------|----------|--------|
| Go | 0 | 5 | ░░░░░░░░░░ 0% | ⏳ Not started |
| Node.js | 0 | 5 | ░░░░░░░░░░ 0% | ⏳ Not started |
| React | 0 | 5 | ░░░░░░░░░░ 0% | ⏳ Not started |
| Next.js | 0 | 5 | ░░░░░░░░░░ 0% | ⏳ Not started |
| Python | 0 | 5 | ░░░░░░░░░░ 0% | ⏳ Not started |
| Laravel | 0 | 5 | ░░░░░░░░░░ 0% | ⏳ Not started |
| Static | 0 | 5 | ░░░░░░░░░░ 0% | ⏳ Not started |
| **Overall** | **0** | **35** | ░░░░░░░░░░ 0% | ⏳ |

## Certified Stacks

*None yet — certification in progress.*

## Failure Tracking

Every failed deployment becomes a certification issue (CERT-NNN). Each issue documents:

- **Stack** and **repository**
- **Problem** observed
- **Root cause** identified
- **Resolution** applied
- **Commit** that fixed it

See `notes.md` in each stack directory for active issues.

## How to Run Certification Tests

```bash
# Run certification for a specific repository
cloudos deploy https://github.com/golang/example

# Verify the deployment
curl http://localhost:31000/health
curl http://localhost:31000/

# Check deployment report
curl http://localhost:31000/api/v1/applications/<app-id>/deployments/1/timeline
```
