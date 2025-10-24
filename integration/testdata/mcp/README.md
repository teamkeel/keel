# MCP Integration Tests

This directory contains integration tests for Keel's MCP (Model Context Protocol) server implementation.

## Test Files

### `tests.test.ts`
Manual fetch-based tests that directly test the MCP JSON-RPC protocol. These tests verify:
- Protocol compliance
- Resource and tool operations
- Authentication flow
- Error handling
- Dynamic descriptions

### `sdk-client.test.ts`
Tests using the **official MCP SDK** (`@modelcontextprotocol/sdk`) from Anthropic. These tests verify:
- Compatibility with real MCP clients
- SDK client can connect and initialize
- SDK client can list resources and tools
- SDK client can read resources
- SDK client can call tools
- Error handling through SDK
- Authentication with Bearer tokens

## Running Tests

### All MCP Tests
```bash
# From integration directory
go test -v -run "mcp"
```

### Just SDK Client Tests
The SDK client tests will run as part of the integration test suite when the Go integration tests execute.

## Dependencies

- `@modelcontextprotocol/sdk` - Official MCP SDK from Anthropic
- `@teamkeel/testing-runtime` - Keel testing utilities
- `vitest` - Test runner

## Test Architecture

```
┌─────────────────────────────────────────────┐
│  Go Integration Test Runner                 │
│  (integration/integration_test.go)          │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  Keel Runtime (Go)                          │
│  - HTTP Server                              │
│  - MCP Handler at /api/mcp                  │
└──────────────────┬──────────────────────────┘
                   │
                   ▼
┌─────────────────────────────────────────────┐
│  TypeScript Tests (Vitest)                  │
│                                             │
│  ┌─────────────┐      ┌─────────────────┐ │
│  │ Manual Tests│      │ SDK Client Tests│ │
│  │ (fetch)     │      │ (@mcp/sdk)      │ │
│  └─────────────┘      └─────────────────┘ │
│                                             │
│  Both test the same MCP endpoints          │
└─────────────────────────────────────────────┘
```

## What Makes These Tests Valuable

### Manual Tests (`tests.test.ts`)
- ✅ Test the raw protocol format
- ✅ Verify JSON-RPC 2.0 compliance
- ✅ Check exact response structures
- ✅ Test edge cases and error conditions

### SDK Client Tests (`sdk-client.test.ts`)
- ✅ Prove real-world compatibility
- ✅ Test with official Anthropic SDK
- ✅ Verify spec compliance
- ✅ Demonstrate actual client usage
- ✅ Catch protocol incompatibilities

## Authentication Testing

Both test suites include authentication scenarios:

1. **No Auth** - Public actions work without tokens
2. **Bearer Token** - Actions requiring auth accept Bearer tokens
3. **Auth Instructions** - Server provides auth endpoint information in `initialize`

Example auth flow in tests:
```typescript
// 1. Get token from auth API
const response = await fetch('/auth/token', {
  method: 'POST',
  body: JSON.stringify({
    grant_type: 'password',
    username: 'user@example.com',
    password: 'password'
  })
});
const { access_token } = await response.json();

// 2. Use with MCP
const transport = new HttpTransport(mcpUrl, {
  Authorization: `Bearer ${access_token}`
});
```

## Custom HTTP Transport

The SDK client tests include a custom `HttpTransport` class that bridges the MCP SDK (designed for stdio/SSE) to work with Keel's HTTP POST endpoint. This demonstrates how MCP clients can adapt to different transport mechanisms.

## Future Enhancements

- [ ] Test with SSE transport (when implemented)
- [ ] Test OAuth flow end-to-end
- [ ] Test with MCP Inspector tool
- [ ] Performance benchmarks
- [ ] Concurrent request testing
