# Architecture

This document describes the internal architecture of `req`, including data flow, component interactions, and system design.

## Overview

`req` is structured as a pipeline: **Parse → Plan → Execute → Output**

```mermaid
graph LR
    A[User Input] --> B[Parser]
    B --> C[AST]
    C --> D[Planner]
    D --> E[Execution Plan]
    E --> F[Executor]
    F --> G[HTTP Client]
    G --> H[Response]
    H --> I[Output Formatter]
    I --> J[stdout/stderr]
```

## Component Architecture

```mermaid
graph TB
    subgraph "Command Line"
        CLI[main.go]
    end
    
    subgraph "Parser Package"
        Parser[Parser]
        Tokenizer[Tokenizer]
        AST[AST Builder]
    end
    
    subgraph "Planner Package"
        Planner[Planner]
        Validator[Validator]
        Defaults[Default Applier]
    end
    
    subgraph "Runtime Package"
        Executor[Executor]
        HTTPClient[HTTP Client]
        SessionMgr[Session Manager]
    end
    
    subgraph "Types Package"
        Types[Type Definitions]
    end
    
    CLI --> Parser
    Parser --> Tokenizer
    Tokenizer --> AST
    AST --> Types
    Types --> Planner
    Planner --> Validator
    Planner --> Defaults
    Planner --> Types
    Types --> Executor
    Executor --> HTTPClient
    Executor --> SessionMgr
    Executor --> Types
```

## Request Lifecycle

```mermaid
sequenceDiagram
    participant User
    participant Parser
    participant Planner
    participant Executor
    participant HTTPClient
    participant Server
    
    User->>Parser: Command string
    Parser->>Parser: Tokenize
    Parser->>Parser: Build AST
    Parser->>Planner: Command AST
    Planner->>Planner: Apply defaults
    Planner->>Planner: Validate
    Planner->>Executor: Execution Plan
    Executor->>Executor: Build URL
    Executor->>Executor: Build body
    Executor->>Executor: Set headers
    Executor->>Executor: Apply session
    Executor->>HTTPClient: HTTP Request
    HTTPClient->>Server: HTTP Request
    Server->>HTTPClient: HTTP Response
    HTTPClient->>Executor: Response
    Executor->>Executor: Decompress
    Executor->>Executor: Run expectations
    Executor->>Executor: Format output
    Executor->>User: stdout/stderr
```

## Parser Architecture

The parser uses a two-phase approach: tokenization followed by parsing.

```mermaid
graph TD
    A[Input String] --> B[Tokenize]
    B --> C{Respect Quotes?}
    C -->|Yes| D[Preserve Quoted Strings]
    C -->|No| E[Split on Whitespace]
    D --> F[Token Stream]
    E --> F
    F --> G[Parse Verb]
    G --> H[Parse Target]
    H --> I[Parse Clauses]
    I --> J{AST Complete?}
    J -->|No| I
    J -->|Yes| K[Command AST]
    K --> L[Validation]
    L --> M{Valid?}
    M -->|No| N[Parse Error]
    M -->|Yes| O[Return AST]
```

### Tokenization Flow

```mermaid
graph LR
    A[Input] --> B{In Quotes?}
    B -->|Yes| C[Accumulate]
    B -->|No| D{Equals Sign?}
    D -->|Yes| E[Mark Clause Value]
    D -->|No| F{Whitespace?}
    E --> G[Accumulate Value]
    F -->|Yes| H{New Clause?}
    F -->|No| C
    H -->|Yes| I[Emit Token]
    H -->|No| C
    G --> J[Emit Token]
    I --> K[Token Stream]
    J --> K
```

## Execution Flow

```mermaid
graph TD
    A[Execution Plan] --> B[Build URL]
    B --> C[Build Body]
    C --> D{Body Type?}
    D -->|Multipart| E[Build Multipart]
    D -->|JSON/Text| F[Set Content]
    E --> G[Create Request]
    F --> G
    G --> H[Set Headers]
    H --> I[Set Cookies]
    I --> J[Apply Session]
    J --> K{Authenticate Verb?}
    K -->|Yes| L[Capture Cookies]
    K -->|No| M[Execute Request]
    L --> M
    M --> N{Redirect?}
    N -->|Yes| O{Follow Policy?}
    N -->|No| P[Read Response]
    O -->|Yes| Q[Follow Redirect]
    O -->|No| R[Return Response]
    Q --> M
    P --> S[Decompress]
    S --> T[Run Expectations]
    T --> U{Expectations Pass?}
    U -->|No| V[Exit Code 3]
    U -->|Yes| W[Format Output]
    W --> X[Write stdout/stderr]
```

## Session Management Flow

```mermaid
sequenceDiagram
    participant User
    participant Executor
    participant SessionMgr
    participant FileSystem
    
    Note over User,FileSystem: Authenticate Flow
    User->>Executor: authenticate verb
    Executor->>Executor: Execute request
    Executor->>Executor: Capture Set-Cookie
    Executor->>Executor: Extract access_token
    Executor->>SessionMgr: UpdateSessionFromResponse
    SessionMgr->>FileSystem: Save session file
    FileSystem-->>SessionMgr: Success
    SessionMgr-->>Executor: Session saved
    
    Note over User,FileSystem: Auto-Apply Flow
    User->>Executor: Regular request
    Executor->>Executor: Check for explicit auth
    alt Explicit auth present
        Executor->>Executor: Use explicit auth
    else No explicit auth
        Executor->>SessionMgr: LoadSession
        SessionMgr->>FileSystem: Read session file
        FileSystem-->>SessionMgr: Session data
        SessionMgr-->>Executor: Session
        Executor->>Executor: Apply cookies/auth
    end
```

## Redirect Handling

```mermaid
graph TD
    A[Response Received] --> B{Status Code?}
    B -->|2xx| C[Success]
    B -->|3xx| D{Verb Type?}
    D -->|read/save| E[Follow Redirect]
    D -->|send/upload| F{follow=smart?}
    F -->|Yes| G{Status Code?}
    F -->|No| H[Don't Follow]
    G -->|307/308| E
    G -->|301/302/303| I[Advisory Message]
    I --> H
    E --> J{Redirect Count < 5?}
    J -->|Yes| K[Create New Request]
    J -->|No| L[Error: Too Many Redirects]
    K --> M[Execute Request]
    M --> A
```

## Error Handling Flow

```mermaid
graph TD
    A[Operation] --> B{Error?}
    B -->|No| C[Success]
    B -->|Yes| D{Error Type?}
    D -->|Parse Error| E[Exit Code 5]
    D -->|Validation Error| E
    D -->|Network Error| F[Exit Code 4]
    D -->|Timeout| F
    D -->|Expectation Failed| G[Exit Code 3]
    D -->|Execution Error| H{Error Code?}
    H -->|3| G
    H -->|4| F
    H -->|5| E
    E --> I[Print Error Message]
    F --> I
    G --> I
    I --> J[Exit with Code]
```

## Include Clause Processing

```mermaid
graph TD
    A[include= Clause] --> B[Parse Items]
    B --> C{Split by Semicolon}
    C --> D[Parse Each Item]
    D --> E{Item Type?}
    E -->|header:| F[Extract Name: Value]
    E -->|param:| G[Extract key=value]
    E -->|cookie:| H[Extract key=value]
    E -->|basic:| I[Extract username:password]
    F --> J[Merge Headers]
    G --> K[Merge Params]
    H --> L[Merge Cookies]
    I --> M[Encode Basic Auth]
    J --> N[Add to Plan]
    K --> N
    L --> N
    M --> N
```

## Data Structures

### Command AST

```mermaid
classDiagram
    class Command {
        +Verb verb
        +Target target
        +Clause[] clauses
        +string sessionSubcommand
    }
    
    class Target {
        +string URL
    }
    
    class Clause {
        <<interface>>
    }
    
    class IncludeClause {
        +IncludeItem[] items
    }
    
    class IncludeItem {
        +string type
        +string name
        +string value
    }
    
    Command --> Target
    Command --> Clause
    IncludeClause --> IncludeItem
    Clause <|-- IncludeClause
```

### Execution Plan

```mermaid
classDiagram
    class ExecutionPlan {
        +Verb verb
        +string method
        +string URL
        +map[string]string headers
        +map[string]string queryParams
        +map[string]string cookies
        +BodyPlan body
        +OutputPlan output
        +RetryPlan retry
        +time.Duration timeout
        +int64 sizeLimit
        +string proxy
        +bool insecure
        +string follow
        +ExpectCheck[] expect
    }
    
    class BodyPlan {
        +string type
        +string content
        +string filePath
        +AttachPart[] attachParts
        +string boundary
    }
    
    class OutputPlan {
        +string format
        +string destination
        +string pick
    }
    
    ExecutionPlan --> BodyPlan
    ExecutionPlan --> OutputPlan
```

## Component Responsibilities

### Parser (`internal/parser`)

- **Tokenization**: Converts input string into tokens
- **Parsing**: Builds Abstract Syntax Tree (AST)
- **Validation**: Basic syntax validation
- **Error Reporting**: Provides position and suggestion information

### Planner (`internal/planner`)

- **Default Application**: Applies verb-specific defaults
- **Validation**: Validates method-verb compatibility
- **Plan Generation**: Creates execution plan from AST
- **Clause Processing**: Merges and processes all clauses

### Executor (`internal/runtime`)

- **Request Building**: Constructs HTTP request from plan
- **Session Management**: Applies and captures sessions
- **Redirect Handling**: Implements redirect policies
- **Response Processing**: Decompression, formatting, expectations
- **Error Handling**: Maps errors to exit codes

### Session Manager (`internal/session`)

- **Storage**: Manages session file storage
- **Security**: Enforces file permissions
- **Retrieval**: Loads sessions for auto-application
- **Updates**: Captures and stores session data

## File Structure

```
req/
├── cmd/req/          # Main entry point
├── internal/
│   ├── parser/      # Command parsing
│   ├── planner/     # Execution planning
│   ├── runtime/     # Request execution
│   ├── session/     # Session management
│   ├── types/       # Type definitions
│   └── grammar/     # Grammar definitions
├── tests/           # Test suite
└── docs/            # Documentation
```

## See Also

- [Grammar Reference](GRAMMAR.md) - Grammar specification
- [Error Handling](ERRORS.md) - Error flow details
- [Session Management](SESSIONS.md) - Session architecture
- [Contributing](CONTRIBUTING.md) - Development guide

