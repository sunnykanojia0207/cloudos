# Security Policy

## Supported Versions

| Version | Supported          |
| ------- | ------------------ |
| 0.x     | :white_check_mark: |

## Reporting a Vulnerability

We take security seriously. If you discover a security vulnerability in CloudOS, please follow these steps:

1. **Do not** disclose the vulnerability publicly
2. Email the security team at security@cloudos.dev
3. Include a detailed description and steps to reproduce
4. Allow time for a fix before public disclosure

We will acknowledge receipt within 48 hours and aim to release a fix within 14 days.

## Security Considerations

- All secrets and credentials must use CloudOS's built-in secrets management
- Run the kernel on localhost only (no authentication in v0.6)
- Review our [Architecture Decision Records](adr/) for security-related decisions
