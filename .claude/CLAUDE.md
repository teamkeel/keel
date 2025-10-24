# Keel Codebase Architecture Guide

## Overview

Keel is an all-in-one backend platform that generates production-grade infrastructure from a single schema file. The system combines a schema parser, runtime engine, code generator, and CLI into an integrated platform supporting local development and cloud deployment.

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        Keel CLI (Go)                        │
│  cmd/keel - Main entry point and command orchestration      │
└─────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        │                     │                     │
   ┌────▼────┐         ┌─────▼─────┐        ┌─────▼──────┐
   │ Schema  │         │  Runtime  │        │  Code Gen  │
   │ Parser  │         │  Engine   │        │  & Build   │
   └─────────┘         └───────────┘        └────────────┘
        │                     │                     │
    (Go/Participle)  (HTTP Handlers/APIs)  (Node.js SDKs)
        │                     │                     │
        └─────────────────────┼─────────────────────┘
                              │
                    ┌─────────▼──────────┐
                    │ Proto Schema (.pb) │
                    │ Central Spec       │
                    └──────────────────┘
                              │
        ┌─────────────────────┼──────────────────────┐
        │                     │                      │
   ┌────▼────┐         ┌─────▼────┐        ┌───────▼────┐
   │Database │         │Functions │        │ Testing    │
   │(Postgres)         │Runtime   │        │Framework   │
   └─────────┘         │(Node.js) │        │(Vitest)    │
                       └──────────┘        └────────────┘
```

## Core Components

### 1. Schema System (`schema/` directory)

**Purpose**: Parse and validate Keel schema files (.keel) and convert them to proto.Schema

**Key Components**:
- **Parser** (`schema/parser/`): Uses Participle v2 lexer to parse .keel files into AST
- **Reader** (`schema/reader/`): Reads .keel files from the filesystem
- **Validation** (`schema/validation/`): Validates parsed schemas against rules
- **Builder** (`schema/schema.go`): Orchestrates parsing, validation, and proto generation
- **MakeProto** (`schema/makeproto.go`): Transforms AST into proto.Schema

**Flow**:
```
.keel files → Reader → Parser (AST) → Validation → Builder → proto.Schema
```

**Key Classes**:
- `Builder`: Main orchestrator for schema construction
- Proto messages: Model, API, Action, Task, Job, Flow, etc.

### 2. Proto Schema (`proto/` directory)

**Purpose**: Define the canonical in-memory representation of a Keel application

**Key Components**:
- **schema.proto**: Protocol Buffer definition for Schema structure
- **schema.pb.go**: Generated protobuf Go code
- **query.go**: Query helper methods for proto.Schema

**Central Concepts**:
- Schema contains: Models, APIs, Actions, Jobs, Subscribers, Events, Routes, Flows, Tasks
- Models: Database entities with fields and actions
- Actions: Operations on models (CRUD, custom functions)
- Jobs: Scheduled or manual background tasks
- Flows: Orchestrated workflows
- Tasks: Event-driven queue items

### 3. Runtime Engine (`runtime/` directory)

**Purpose**: Execute Keel applications at runtime, handling HTTP requests and business logic

**Architecture**:
```
HTTP Request
    │
    └─→ NewHttpHandler (runtime.go)
          │
          ├─→ /topics/* → NewTasksHandler (event queue)
          ├─→ /flows/*  → NewFlowsHandler (workflow execution)
          ├─→ /auth/*   → NewAuthHandler (authentication/OAuth)
          └─→ others    → NewApiHandler (actions)
```

**Key Packages**:
- **runtime.go**: Main HTTP handler router
- **apis/graphql/**: GraphQL API implementation
- **apis/jsonrpc/**: JSON-RPC API implementation
- **apis/httpjson/**: REST JSON API implementation
- **apis/authapi/**: Authentication and OAuth handling
- **apis/flowsapi/**: Workflow execution
- **apis/tasksapi/**: Event queue/pub-sub
- **actions/**: Action execution logic (create, read, update, delete, custom)
- **oauth/**: OAuth provider integrations
- **runtimectx/**: Context utilities for request handling

**Request Processing**:
1. Request arrives at NewHttpHandler
2. Routed based on path prefix
3. Handler collects headers into runtime context
4. Action executed with authorization checks
5. Events generated and sent
6. Response returned

### 4. Functions System (`functions/` directory, `packages/functions-runtime/`)

**Purpose**: Interface between runtime and custom function implementations

**Architecture** (IPC via HTTP):
```
Runtime (Go) 
    │
    └─→ FunctionsRuntimeRequest (JSON-RPC)
          │
          └─→ HTTP POST to Node.js Functions Runtime
                │
                └─→ Custom Function Handler
                     │
                     └─→ Database queries via Kysely
                     └─→ HTTP responses
                │
          ←─────FunctionsRuntimeResponse
    │
    └─→ Parse & return result
```

**Key Components**:
- **functions.go**: Defines FunctionType, FunctionErrorCode, Transport protocol
- **@teamkeel/functions-runtime**: Node.js package providing database access, tracing, etc.
- **Custom function handler**: Generated code that calls user's custom functions

**Function Types**: Action, Job, Subscriber, Flow, Route

### 5. Code Generation (`node/`, `codegen/`)

**Purpose**: Generate TypeScript/JavaScript SDKs and project files

**Generation Process** (node/codegen.go):
```
proto.Schema → Node.js Code Generation
                   │
                   ├─→ SDK package (@teamkeel/keel generated)
                   │    ├─ Types (models, actions, enums, messages)
                   │    ├─ API factory
                   │    ├─ Query builders
                   │    └─ Permission helpers
                   │
                   ├─→ Testing package (@teamkeel/testing generated)
                   │    ├─ Test utilities
                   │    ├─ Reset/seed database
                   │    └─ Action executors
                   │
                   └─→ Setup files (package.json, tsconfig.json, etc.)
```

**Key Generators**:
- `writeTableInterface()`: Database table types
- `writeModelInterface()`: Model data types
- `writeFunctionWrapperType()`: Custom function types
- `writeAPIFactory()`: API client factory
- `writeMessages()`: Message types

### 6. Development Server (`cmd/program/`)

**Purpose**: Orchestrate local development experience

**Key Components** (model.go):
- **Model**: Bubbletea TUI application state
- **Update**: Event handling for status changes
- **View**: Terminal UI rendering

**Lifecycle** (commands.go):
```
1. StatusCheckingDependencies: Verify Node.js, Docker, etc.
2. StatusParsePrivateKey: Load authentication key
3. StatusSetupDatabase: Start/prepare database container
4. StatusSetupFunctions: Initialize Node.js functions runtime
5. StatusLoadSchema: Parse and validate .keel files
6. StatusRunMigrations: Apply database migrations
7. StatusUpdateFunctions: Regenerate SDK/functions code
8. StatusStartingFunctions: Start Node.js dev server
9. StatusRunning: Accept HTTP requests
```

**File Watching**: Monitors for schema changes and regenerates/rebuilds

### 7. Database System (`db/`, `migrations/`)

**Purpose**: Manage database schema and connections

**Key Components**:
- **db.go**: Connection pooling and management (uses pgx/v5)
- **gorm.go**: GORM integration for ORM
- **migrations.go**: Schema migration engine
  - Introspection: Analyze current DB state
  - Diffing: Compare with desired schema
  - Generation: Create migration SQL
  - Execution: Apply migrations

**Migration Strategy**:
1. Read proto.Schema to get desired state
2. Query database for actual state
3. Generate SQL diff (ALTER TABLE, CREATE TABLE, etc.)
4. Apply migrations in transactions
5. Run introspection to verify

### 8. Testing Framework (`testing/`, `packages/testing-runtime/`)

**Purpose**: Enable testing of Keel applications with TypeScript/Vitest

**Components**:
- **testing.go**: Go-side test orchestration
  - Builds test environment
  - Starts isolated Node.js functions runtime
  - Handles HTTP requests to test API
- **@teamkeel/testing-runtime**: Node.js package
  - ActionExecutor: Run actions
  - JobExecutor: Run jobs
  - SubscriberExecutor: Run subscribers
  - FlowExecutor: Run flows
  - Reset database utilities
  - Custom test matchers (toHaveError, toHaveAuthorizationError, etc.)

**Test Execution** (testing.go):
```
.test.ts files
    │
    └─→ testing.Run()
          │
          ├─→ deploy.Build() → generate code
          ├─→ Start test API endpoints
          ├─→ Start Node.js dev server
          └─→ vitest run
                │
                └─→ Test code calls actions via HTTP
                     │
                     └─→ Response validated
```

### 9. Deployment System (`deploy/`)

**Purpose**: Build production-ready Lambda functions and configuration

**Build Process** (deploy/build.go):
```
.keel files → schema parsing
    │
    ├─→ Code generation (functions handlers, types)
    ├─→ esbuild: Bundle JavaScript
    ├─→ Create Lambda packages (ZIP)
    │    ├─ Runtime Lambda: HTTP handler + database
    │    └─ Functions Lambda: Custom function executors
    │
    └─→ BuildResult (schema, paths to Lambda packages)
```

**Output**:
- Runtime package: Full runtime engine bundled
- Functions package: Node.js functions with @teamkeel/functions-runtime

### 10. Expressions System (`expressions/`, `runtime/expressions/`)

**Purpose**: Evaluate dynamic expressions in schema (permissions, defaults, computed fields)

**Components**:
- **expressions/parser.go**: Parse expression syntax
- **expressions/resolve/**: Evaluate expressions at runtime
- **expressions/typing/**: Type checking for expressions

**Use Cases**:
- Permission rules: `ctx.identity.email == context.email`
- Default values: `now()`, `generateID()`
- Computed fields: `post.authorName == author.name`

### 11. Configuration System (`config/`)

**Purpose**: Load and validate keelconfig.yaml

**Key Concepts**:
- Environment variables declaration
- Secrets configuration
- Auth provider settings
- Database connection overrides
- API endpoint customization

## Data Flow

### Request Processing Flow

```
Client HTTP Request
    │
    ▼
Runtime HTTP Handler (runtime.go)
    │
    ├─► Identify path (auth/api/flows/topics)
    │
    ▼
Route-Specific Handler
    │
    ├─► Collect headers into context
    ├─► Validate request format
    ├─► Call functions.Transport
    │
    ▼
Functions Runtime (Node.js via HTTP)
    │
    ├─► Execute custom function (if CUSTOM action)
    ├─► Access database via Kysely
    ├─► Generate events
    │
    ▼
Return to Runtime
    │
    ├─► Collect generated events
    ├─► Send events to handlers
    ├─► Format response (JSON/GraphQL)
    │
    ▼
HTTP Response to Client
```

### Schema Build Flow

```
User edits .keel files
    │
    ▼
File watcher detects change
    │
    ▼
schema.Builder.MakeFromDirectory()
    │
    ├─► reader.FromDir() → read all .keel files
    ├─► parser.Parse() → generate ASTs
    ├─► validation.Validate() → check rules
    ├─► Builder.makeProtoModels() → transform AST to proto
    │
    ▼
proto.Schema created
    │
    ├─► Passed to node.Generate() for code gen
    ├─► Passed to migrations for DB changes
    ├─► Passed to runtime for handler setup
    │
    ▼
Application updated
```

## Key Architectural Patterns

### 1. Proto-Centric Design
- All components work from a single canonical proto.Schema
- Reduces round-trip serialization
- Enables type-safe generation

### 2. Language Interop (Go + Node.js)
- Go handles schema parsing, runtime orchestration, CLI
- Node.js handles custom functions, SDK, testing
- Communication via HTTP JSON-RPC and standard protocols

### 3. Separation of Concerns
- **Parsing**: Pure schema compilation
- **Runtime**: HTTP request handling
- **Functions**: Custom business logic execution
- **Code Generation**: TypeScript SDK creation
- **Database**: Schema management

### 4. Context-Based Authorization
- All requests carry context (identity, headers, etc.)
- Permissions evaluated against context
- Early permission checks when possible

### 5. Event-Driven Architecture
- Actions can emit events
- Events trigger subscribers
- Tasks (async jobs) in queue system
- Flows orchestrate workflows

## Technology Stack

**Backend**:
- Go 1.23+ for runtime, CLI, schema parsing
- PostgreSQL for data storage
- Protocol Buffers for schema definition
- GORM for ORM
- Participle v2 for schema parsing
- httprouter for HTTP routing
- OpenTelemetry for tracing

**Frontend/SDK**:
- TypeScript for SDK and custom functions
- Kysely for type-safe database queries
- Vitest for testing
- esbuild for bundling

**Infrastructure**:
- Docker for local database
- AWS Lambda for production functions
- S3 for file storage
- EventBridge/SQS for async processing

## Testing Strategy

- Integration tests: Full schema → actions → verify output
- Unit tests: Individual functions and components
- Testing framework generates @teamkeel/testing package
- Tests run with isolated database per test
- Supports pattern matching for selective test runs

## CLI Commands & Entry Points

- `keel init`: Create new project
- `keel run`: Start development server
- `keel generate`: Generate SDK code
- `keel validate`: Validate schema
- `keel test`: Run tests
- `keel secrets`: Manage secrets
- `keel deploy`: Deploy to platform

## Important Implementation Notes

### Schema Validation
- Happens in `schema/validation/` with 40+ validation rules
- Checks model uniqueness, field types, action compatibility
- Validates permissions, permissions rules syntax
- Ensures referential integrity of relationships

### Permission System
- Rules defined on models and actions
- Evaluated at request time
- Can be evaluated early or lazily
- Context includes identity, headers, request data

### Auto-Generated vs Custom Actions
- Auto actions: CRUD operations generated by Keel
- Custom actions: User-provided functions in `functions/` directory
- Both follow same interface: proto.Action with ActionImplementation flag

### Database Migrations
- Automatic: Schema changes trigger migrations
- No manual SQL needed for most cases
- Introspection-based: Queries DB to understand current state
- Conservative: Preserves data when possible

### Code Generation Targets
- SDK package: Types for frontend consumption
- Testing package: Utilities for test writing
- Function stubs: Scaffolding for custom implementations
- Supporting files: package.json, tsconfig.json, vitest setup

## File Organization

```
/cmd/keel/        - CLI entry point
/cmd/program/     - Dev server TUI orchestration
/cmd/database/    - Database utilities
/schema/          - Schema parsing & validation
/proto/           - Protocol buffer definitions
/runtime/         - HTTP runtime engine
/functions/       - Function call interface
/node/            - Node.js code generation
/codegen/         - Code generation utilities
/db/              - Database connections
/migrations/      - DB schema management
/testing/         - Test framework
/deploy/          - Production build system
/expressions/     - Expression evaluation
/config/          - Configuration loading
/packages/        - Published npm packages
  /keel/          - CLI package
  /functions-runtime/
  /testing-runtime/
  /client-react/
  /client-react-query/
```

## Working in This Codebase

### When Adding a New Feature

1. **Schema Level**: Update proto/schema.proto, regenerate .pb.go
2. **Validation**: Add validation rules in schema/validation/
3. **Runtime**: Implement handler in runtime/
4. **Code Gen**: Update node/codegen.go to generate types
5. **Testing**: Add integration tests in integration/testdata/
6. **CLI**: Add command if user-facing in cmd/

### Key Files to Know

- `schema/schema.go`: Main entry for schema processing
- `runtime/runtime.go`: Main HTTP handler
- `node/codegen.go`: Code generation logic
- `proto/schema.proto`: Central spec
- `cmd/program/model.go`: Dev server lifecycle
- `testing/testing.go`: Test execution

### Debugging

- Enable verbose logging: Set KEEL_LOG_LEVEL=debug
- Trace expressions: Check expressions/resolve/
- Database issues: Check migrations/introspection.go
- Function communication: Check functions/functions.go HTTP transport
