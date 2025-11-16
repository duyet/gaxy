# Security Policy

## Reporting a Vulnerability

If you discover a security vulnerability in Gaxy, please report it by emailing the maintainer directly or creating a private security advisory on GitHub.

**Please do not create public GitHub issues for security vulnerabilities.**

## Security Features

Gaxy implements multiple layers of security to protect against common web application vulnerabilities:

### 1. SSRF (Server-Side Request Forgery) Protection

**Vulnerability:** Uncontrolled user input in network requests could allow attackers to access internal services, cloud metadata endpoints, or perform unauthorized actions.

**Protection Implemented:**

Our SSRF protection uses a **defense-in-depth strategy** with multiple validation layers:

1. **URI Sanitization** (`pkg/proxy/validation.go`)
   - Validates that user input contains only path + query components
   - Rejects full URLs with schemes (http://, https://, file://, etc.)
   - Blocks protocol-relative URLs (//)
   - Prevents directory traversal (..)
   - Ensures paths start with /

2. **Path Whitelisting**
   - Only allows known Google Analytics/Tag Manager endpoints
   - Explicit allowlist prevents access to unauthorized paths
   - See `isAllowedPath()` in validation.go for full list

3. **Secure Request Construction** (`pkg/proxy/service.go`)
   - **Critical:** Scheme and host are set from TRUSTED configuration, NOT user input
   - Only validated path and query are derived from user input
   - This architecture ensures users cannot control WHERE requests go
   - Users can only select WHICH Google Analytics endpoint is accessed

**Architecture:**
```
User Input (path) → Validation → Whitelist Check → Safe Request
                                                          ↓
Config (scheme/host) ──────────────────────────→  https://www.google-analytics.com/[validated-path]
```

**Example Attack Vectors Blocked:**
```
❌ http://169.254.169.254/latest/meta-data/  (AWS metadata)
❌ http://metadata.google.internal/          (GCP metadata)
❌ http://127.0.0.1:8080/admin               (Localhost)
❌ //attacker.com/steal                      (Protocol-relative)
❌ file:///etc/passwd                        (File access)
```

**Allowed Paths:**
```
✅ /analytics.js
✅ /ga.js
✅ /gtag/js?id=G-12345
✅ /collect?v=1&tid=UA-12345
✅ /batch
```

### 2. Rate Limiting

**Protection:** Per-IP token bucket rate limiting prevents abuse and DDoS attacks.

**Configuration:**
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPS=100      # Requests per second per IP
RATE_LIMIT_BURST=200    # Burst allowance
```

### 3. Security Headers

**Protection:** Security headers prevent common client-side attacks.

**Headers Set:**
- `X-Frame-Options: SAMEORIGIN` - Prevents clickjacking
- `X-Content-Type-Options: nosniff` - Prevents MIME sniffing
- `X-XSS-Protection: 1; mode=block` - XSS protection
- `Referrer-Policy: strict-origin-when-cross-origin` - Privacy
- `X-Powered-By: gaxy` - Server identification

### 4. Input Validation

**Protection:** All configuration is validated at startup to prevent misconfigurations.

**Validations:**
- URL format validation for upstream origins
- Timeout bounds checking
- Connection pool limits verification
- Configuration type safety

### 5. Container Security

**Protection:** Docker container runs with minimal privileges.

**Features:**
- Distroless base image (minimal attack surface)
- Non-root user execution
- No shell or package managers in final image
- Static binary compilation
- Multi-stage build (no build tools in runtime)

### 6. Dependency Management

**Protection:** Regular dependency updates and security scanning.

**Practices:**
- Pin dependency versions in go.mod
- Regular dependency updates
- Automated security scanning (GitHub Dependabot)
- Minimal dependency tree

## Security Best Practices for Deployment

### 1. HTTPS Only
Always run Gaxy behind HTTPS in production:
```yaml
# Use a reverse proxy like nginx or a load balancer
location /analytics {
    proxy_pass http://gaxy:3000;
    proxy_set_header X-Forwarded-Proto https;
}
```

### 2. Restrict CORS Origins
Don't use wildcard CORS in production:
```bash
# Bad
CORS_ALLOW_ORIGINS=*

# Good
CORS_ALLOW_ORIGINS=https://yourdomain.com
```

### 3. Enable Rate Limiting
Protect against abuse:
```bash
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPS=100
RATE_LIMIT_BURST=200
```

### 4. Monitor Metrics
Watch for security-related events:
```promql
# Rate limit violations
rate(gaxy_rate_limit_dropped_total[5m])

# Error rates
rate(gaxy_requests_total{status=~"4..|5.."}[5m])

# Upstream errors
rate(gaxy_upstream_errors_total[5m])
```

### 5. Review Logs
Monitor for suspicious activity:
```json
{
  "level": "WARN",
  "message": "Request URI validation failed",
  "original_uri": "http://evil.com/steal"
}
```

### 6. Network Isolation
Run Gaxy in isolated network segments:
- Use Kubernetes NetworkPolicies
- Restrict egress to Google Analytics domains only
- Block access to cloud metadata endpoints
- Use private subnets

### 7. Resource Limits
Set container resource limits:
```yaml
resources:
  limits:
    memory: 256Mi
    cpu: 500m
  requests:
    memory: 128Mi
    cpu: 250m
```

### 8. Secrets Management
Don't expose sensitive configuration:
```bash
# Use Kubernetes secrets, AWS Secrets Manager, etc.
# Never commit .env files to version control
```

## Security Checklist for Production

- [ ] HTTPS enabled and enforced
- [ ] Rate limiting configured
- [ ] CORS origins restricted (not `*`)
- [ ] Security headers enabled
- [ ] Container running as non-root
- [ ] Resource limits configured
- [ ] Monitoring and alerting set up
- [ ] Logs aggregated and monitored
- [ ] Network policies in place
- [ ] Dependencies up to date
- [ ] Secrets properly managed
- [ ] Backup and recovery tested

## Vulnerability Disclosure Timeline

When a security vulnerability is reported:

1. **Day 0:** Vulnerability reported privately
2. **Day 1-3:** Vulnerability confirmed and assessed
3. **Day 3-7:** Fix developed and tested
4. **Day 7-14:** Fix released in patch version
5. **Day 14+:** Public disclosure with CVE (if applicable)

## Security Updates

Security updates are released as patch versions (e.g., 1.0.1) and are clearly marked in the changelog.

**To update:**
```bash
# Docker
docker pull ghcr.io/duyet/gaxy:latest

# Go
go get -u github.com/duyet/gaxy

# Helm
helm upgrade gaxy duyet/gaxy
```

## Compliance

Gaxy is designed to help you maintain compliance with:
- GDPR (by proxying and owning your analytics data)
- CCPA (California Consumer Privacy Act)
- ePrivacy Directive (Cookie Law)

## Additional Resources

- [OWASP Top 10](https://owasp.org/www-project-top-ten/)
- [CWE-918: SSRF](https://cwe.mitre.org/data/definitions/918.html)
- [Docker Security Best Practices](https://docs.docker.com/develop/security-best-practices/)
- [Kubernetes Security](https://kubernetes.io/docs/concepts/security/)

---

**Last Updated:** 2025-11-16
**Maintainer:** Duyet (github.com/duyet)
