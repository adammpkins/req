# Grammar Reference

This document provides a complete specification of the `req` command grammar.

## Command Structure

The basic structure of a `req` command is:

```
req <verb> <url> [clauses...]
```

Where:
- `verb` is one of the action verbs (read, save, send, etc.)
- `url` is the target URL
- `clauses` are optional key=value pairs that modify the request

## EBNF Grammar

```
command          = verb target [clauses]
verb             = "read" | "save" | "send" | "upload" | "watch" | "inspect" | "authenticate" | "session"
target           = url
clauses          = clause { clause }
clause           = using_clause | include_clause | attach_clause | expect_clause | as_clause | to_clause |
                   retry_clause | under_clause | via_clause | follow_clause | insecure_clause | with_clause

using_clause     = "using=" http_method
include_clause   = "include=" include_items
attach_clause    = "attach=" attach_items
expect_clause    = "expect=" expect_checks
as_clause        = "as=" output_format
to_clause        = "to=" path
retry_clause     = "retry=" number
under_clause     = "under=" ( duration | size )
via_clause       = "via=" url
follow_clause    = "follow=" ("smart" | "")
insecure_clause  = "insecure=" ("true" | "false")
with_clause      = "with=" ( string | "@" path | "@-" )

http_method      = "GET" | "POST" | "PUT" | "PATCH" | "DELETE" | "HEAD" | "OPTIONS"
output_format    = "json" | "csv" | "text" | "raw" | "auto"
duration         = number time_unit
size             = number size_unit
time_unit        = "s" | "m" | "h"
size_unit        = "B" | "KB" | "MB" | "GB"
number           = digit { digit }
path             = string
url              = string

include_items    = include_item { ";" include_item }
include_item     = header_item | param_item | cookie_item | basic_item
header_item      = "header:" header_name ":" header_value
param_item       = "param:" param_key "=" param_value
cookie_item      = "cookie:" cookie_key "=" cookie_value
basic_item       = "basic:" username ":" password

attach_items     = attach_item { ";" attach_item }
attach_item      = part_item | boundary_item
part_item        = "part:" part_spec
part_spec        = "name=" name ["," ("file=" "@" path | "value=" value)] ["," "filename=" filename] ["," "type=" mime_type]
boundary_item    = "boundary:" token

expect_checks    = expect_check { "," expect_check }
expect_check     = status_check | header_check | contains_check | jsonpath_check | matches_check
status_check     = "status:" number
header_check     = "header:" header_name "=" header_value
contains_check   = "contains:" string
jsonpath_check   = "jsonpath:" jsonpath_expr
matches_check    = "matches:" regex_pattern
```

## Tokenization Rules

The parser tokenizes input respecting:

1. **Quoted strings**: Single or double quotes preserve whitespace and special characters
2. **Clause values**: Everything after `=` is treated as a single token until the next clause
3. **Shell expansion**: Environment variables are expanded by the shell before parsing
4. **No re-splitting**: The parser does not re-split on spaces within clause values

### Quoting

- Values containing semicolons (`;`) must be quoted
- Values containing spaces should be quoted
- Backslash escapes are allowed inside quotes for the quote character and backslash
- Single quotes preserve everything literally
- Double quotes allow shell variable expansion

Example:
```bash
# Unquoted (simple value)
include='header: Accept: application/json'

# Quoted (contains semicolon)
include='header: Accept: application/json; charset=utf-8'

# Quoted (contains spaces)
include='param: q=search query with spaces'
```

## Clause Ordering

Clauses can appear in any order. The following are equivalent:

```bash
req read https://api.example.com/users as=json include='header: Accept: application/json'

req read https://api.example.com/users include='header: Accept: application/json' as=json
```

## Clause Multiplicity

### Singleton Clauses

These clauses can only appear once per command:
- `using=`
- `with=`
- `expect=`
- `as=`
- `to=`
- `retry=`
- `under=`
- `via=`
- `follow=`
- `insecure=`

**Error**: Duplicate singleton clauses result in a parse error.

### Repeatable Clauses

These clauses can appear multiple times:
- `include=` - Multiple include clauses are merged
- `attach=` - Multiple attach clauses are combined

Example:
```bash
req read https://api.example.com/users \
  include='header: Accept: application/json' \
  include='header: X-Trace: 1' \
  include='param: page=1'
```

## include= Grammar

The `include=` clause accepts multiple items separated by semicolons:

```
include_items = include_item { ";" include_item }
```

### Item Types

#### Header Item
```
header_item = "header:" header_name ":" header_value
```

- Format: `header: Name: Value`
- Name and Value are separated by a colon
- Both Name and Value are trimmed of whitespace
- Example: `include='header: Authorization: Bearer token'`

#### Parameter Item
```
param_item = "param:" param_key "=" param_value
```

- Format: `param: key=value`
- Key and value are separated by equals sign
- Example: `include='param: q=search query'`

#### Cookie Item
```
cookie_item = "cookie:" cookie_key "=" cookie_value
```

- Format: `cookie: key=value`
- Key and value are separated by equals sign
- Example: `include='cookie: session=abc123'`

#### Basic Auth Item
```
basic_item = "basic:" username ":" password
```

- Format: `basic: username:password`
- Username and password are separated by a colon
- Automatically encoded as `Authorization: Basic <base64>` header
- Example: `include='basic: user:pass'`

### Merging Rules

When multiple `include=` clauses or items are present:

- **Headers**: Last value wins (except for multi-valued headers which keep all values)
- **Params**: Repeated keys become repeated query parameters in insertion order
- **Cookies**: Last value wins per cookie name
- **Basic Auth**: Sets Authorization header, overrides any existing Authorization header

### Examples

```bash
# Single item
include='header: Accept: application/json'

# Multiple items in one clause
include='header: Accept: application/json; param: q=search; cookie: session=abc'

# Multiple include clauses
include='header: Accept: application/json'
include='param: page=1'
include='param: page=2'  # Results in ?page=1&page=2

# With Basic Auth
include='basic: user:pass; header: Accept: application/json'
```

## attach= Grammar

The `attach=` clause accepts multiple parts separated by semicolons:

```
attach_items = attach_item { ";" attach_item }
```

### Part Specification

```
part_item = "part:" part_spec
part_spec = "name=" name ["," ("file=" "@" path | "value=" value)] ["," "filename=" filename] ["," "type=" mime_type]
```

Required:
- `name=` - The form field name

Exactly one of:
- `file=@path` - Path to file (must exist)
- `value=...` - Text value

Optional:
- `filename=` - Filename for file parts
- `type=` - MIME type for the part

### Boundary Specification

```
boundary_item = "boundary:" token
```

- Optional explicit boundary token
- If not specified, a boundary is automatically generated

### Examples

```bash
# Single file part
attach='part: name=file, file=@./avatar.png'

# File with filename and type
attach='part: name=avatar, file=@./me.png, filename=avatar.png, type=image/png'

# Text part
attach='part: name=meta, value={"name":"adam"}'

# Multiple parts
attach='part: name=file, file=@./a.png; part: name=meta, value=xyz'

# Explicit boundary
attach='boundary: ----WebKitFormBoundary; part: name=file, file=@test.png'
```

### Validation Rules

- `name` is required
- Exactly one of `file=` or `value=` is required
- File paths must exist at execution time
- If both `file=` and `value=` are provided, it's an error

## with= Grammar

The `with=` clause specifies the request body:

```
with_clause = "with=" ( string | "@" path | "@-" )
```

### Body Modes

1. **Inline text/JSON**: `with='{"name":"Alice"}'`
2. **File**: `with=@file.json` (reads from file)
3. **Stdin**: `with=@-` (reads from stdin)

### Content-Type Inference

- If Content-Type header is not set and inline value begins with `{` or `[`, infer `application/json`
- A one-line note is printed to stderr when inference occurs
- Explicit Content-Type header always overrides inference

Example:
```bash
# JSON inference occurs
req send https://api.example.com/users with='{"name":"Alice"}'
# stderr: Inferred Content-Type: application/json

# Explicit Content-Type overrides
req send https://api.example.com/users \
  include='header: Content-Type: application/xml' \
  with='{"name":"Alice"}'
# No inference note, uses application/xml
```

## expect= Grammar

The `expect=` clause specifies response assertions:

```
expect_checks = expect_check { "," expect_check }
```

### Check Types

#### Status Check
```
status_check = "status:" number
```
- Example: `expect=status:200`

#### Header Check
```
header_check = "header:" header_name "=" header_value
```
- Example: `expect=status:200, header:Content-Type=application/json`

#### Contains Check
```
contains_check = "contains:" string
```
- Example: `expect=contains:"success"`

#### JSONPath Check
```
jsonpath_check = "jsonpath:" jsonpath_expr
```
- Example: `expect=jsonpath:"$.items[0].id"`

#### Matches Check
```
matches_check = "matches:" regex_pattern
```
- Example: `expect=matches:"^OK\\b"`

### Exit Codes

- **0**: All checks pass
- **3**: Request succeeded but an expectation failed

All checks must pass. If any check fails, exit code 3 is returned.

## Error Cases

### Parse Errors (Exit Code 5)

- Unknown clause key
- Duplicate singleton clauses
- Unknown include item tag
- Malformed header item (missing Name: Value)
- Param or cookie missing equals
- Basic item missing colon
- Unquoted semicolon in include item
- Attach part missing name
- Attach part missing both file and value
- Attach part with both file and value
- Invalid URL
- File not found for `with=` or `attach=`

### Execution Errors (Exit Code 4)

- Network errors
- TLS errors (when `insecure=false`)
- Timeout exceeded
- Size limit exceeded

### Expectation Errors (Exit Code 3)

- Status code mismatch
- Header value mismatch
- Content does not contain expected text
- JSONPath does not match
- Regex does not match

## See Also

- [Verbs Reference](VERBS.md) - Detailed verb documentation
- [Clauses Reference](CLAUSES.md) - Complete clause reference
- [Error Handling](ERRORS.md) - Error codes and troubleshooting

