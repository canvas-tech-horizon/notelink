# Security Policy

## Supported Versions

We release patches for security vulnerabilities for the following versions:

| Version | Supported          |
| ------- | ------------------ |
| 1.x.x   | :white_check_mark: |
| < 1.0   | :x:                |

## Reporting a Vulnerability

We take security vulnerabilities seriously. If you discover a security issue in Notelink, please report it by emailing our security team or creating a private security advisory on GitHub.

**Please do not disclose security vulnerabilities publicly until they have been addressed.**

### Reporting Process

1. **Email**: Send details to the repository maintainers (see GitHub profile)
2. **GitHub Security Advisory**: Use GitHub's private vulnerability reporting feature
3. **Include**:
   - Description of the vulnerability
   - Steps to reproduce
   - Potential impact
   - Suggested fix (if available)

### What to Expect

- **Acknowledgment**: Within 48 hours
- **Status Update**: Within 7 days
- **Fix Timeline**: Critical issues within 30 days, others within 90 days
- **Credit**: We'll acknowledge your contribution in the release notes (unless you prefer to remain anonymous)

## Security Measures

### Current Protections

1. **Input Validation**
   - Recursive validation for nested structures and arrays
   - Type checking for all request parameters
   - JSON schema validation for request bodies

2. **XSS Protection**
   - HTML escaping for all user-provided content
   - JavaScript string escaping for dynamic content
   - Content Security Policy headers recommended

3. **Authentication**
   - JWT token support with Bearer authentication
   - Custom authentication middleware support
   - Secure token storage recommendations

4. **Dependencies**
   - Regular automated vulnerability scanning via GitHub Actions
   - Minimal dependency footprint
   - Use of well-maintained, popular libraries

### Best Practices for Users

1. **JWT Secrets**
   ```go
   // NEVER hardcode secrets in production
   jwtSecret := os.Getenv("JWT_SECRET")
   if jwtSecret == "" {
       log.Fatal("JWT_SECRET environment variable is required")
   }
   ```

2. **HTTPS**
   Always use HTTPS in production. Never transmit tokens over unencrypted connections.

3. **CORS**
   Configure appropriate CORS policies for your API:
   ```go
   app := fiber.New(fiber.Config{
       // Configure CORS appropriately
   })
   ```

4. **Rate Limiting**
   Implement rate limiting to prevent abuse:
   ```go
   import "github.com/gofiber/fiber/v3/middleware/limiter"

   app.Use(limiter.New(limiter.Config{
       Max: 100,
       Expiration: 1 * time.Minute,
   }))
   ```

5. **Input Sanitization**
   While Notelink validates and escapes input, always validate business logic constraints in your handlers.

## Known Security Considerations

### Token Storage

The documentation UI stores JWT tokens in localStorage for convenience during development. For production applications:

- Implement secure token storage (httpOnly cookies recommended)
- Use short-lived access tokens with refresh tokens
- Implement token rotation

### API Testing UI

The built-in API testing interface is designed for development. For production:

- Consider disabling the UI in production environments
- Implement additional authentication for the docs endpoint
- Use environment-based configuration

```go
if os.Getenv("ENVIRONMENT") == "production" {
    // Disable or restrict API docs access
}
```

## Security Checklist for Deployments

- [ ] JWT_SECRET is set via environment variable
- [ ] HTTPS is enforced
- [ ] CORS is properly configured
- [ ] Rate limiting is implemented
- [ ] API documentation access is restricted in production
- [ ] Dependencies are up to date
- [ ] Logging is configured (but doesn't log sensitive data)
- [ ] Error messages don't expose internal details
- [ ] Input validation is applied at application level
- [ ] Database queries use parameterized statements

## Vulnerability Disclosure Timeline

1. **Day 0**: Vulnerability reported
2. **Day 0-2**: Acknowledge receipt
3. **Day 2-7**: Validate and assess severity
4. **Day 7-30**: Develop and test fix (critical issues)
5. **Day 30-90**: Develop and test fix (non-critical issues)
6. **Release**: Publish security advisory and patched version
7. **Post-Release**: Update documentation and notify users

## Security Updates

Security updates are released as patch versions (e.g., 1.0.1, 1.0.2). We recommend:

- Subscribing to GitHub release notifications
- Regularly updating to the latest patch version
- Reading release notes for security-related changes

## Attribution

We appreciate the security research community and will acknowledge contributors who help us improve Notelink's security (unless they prefer to remain anonymous).

## Contact

For security-related questions or concerns, please contact the maintainers through:

- GitHub Issues (for general security questions, not vulnerabilities)
- GitHub Security Advisories (for vulnerability reports)
- Email (see repository for contact information)

---

Last Updated: 2026-02-15
