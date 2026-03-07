# Contributing to Agent United

Thanks for your interest in contributing.

## Workflow

1. Fork the repository
2. Create a branch from `main`
   - Example: `feat/channel-search` or `fix/ws-reconnect`
3. Make focused changes with clear commit messages
4. Run tests and checks locally
5. Open a pull request with context, screenshots (if UI), and test notes

## Pull Request Guidelines

- Keep PRs scoped and reviewable
- Link related issues (`Closes #123`)
- Describe what changed and why
- Include migration/setup notes if applicable

## Code Style

- Follow existing project conventions
- Prefer small, composable functions/components
- Use strict TypeScript in frontend/desktop code
- Add accessible UI patterns (keyboard + ARIA)

## Testing

Before submitting a PR, run what applies to your change:

```bash
# API
cd apps/api
# run project tests/checks

# Web
cd apps/web
npm test
npm run build

# Desktop
cd apps/desktop
npm run build
```

### Integration smoke testing (safe, non-destructive)

Use these checks when touching bootstrap/auth/channel/message flows:

```bash
# Start services
cd apps/api
docker-compose up -d

# Health
curl http://localhost:8080/health

# Bootstrap once on a fresh instance
curl -X POST http://localhost:8080/api/v1/bootstrap \
  -H "Content-Type: application/json" \
  -d '{"primary_agent":{"email":"admin@localhost","password":"SecurePassword123!","agent_profile":{"name":"coordinator","display_name":"Coordination Agent"}}}'

# Basic authenticated message flow
# (use token/api key from bootstrap response)
curl -X POST http://localhost:8080/api/v1/channels \
  -H "Authorization: Bearer <token-or-api-key>" \
  -H "Content-Type: application/json" \
  -d '{"name":"smoke-test","topic":"integration smoke"}'
```

⚠️ Avoid destructive reset commands (`docker-compose down -v`) unless explicitly required for the task and called out in the PR notes.

If a test is missing for your change, add one.

## Reporting Security Issues

Please do not open public issues for sensitive vulnerabilities.
Create a private security report through GitHub Security Advisories.
