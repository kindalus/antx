# HttpError Enhanced Format Examples

This document demonstrates the new enhanced HttpError format that includes both request and response details with improved formatting.

## Enhanced Error Format

The new HttpError format follows this structure:

```
Error: METHOD URL - STATUS_CODE

==> Request
Header1: Value1
Header2: Value2
Body: {pretty-printed JSON or raw content}

Response <==
Header1: Value1
Header2: Value2
Body: {pretty-printed JSON or raw content}
```

## Example 1: GET Request with JSON Error Response

When a GET request fails with a JSON error response:

### HttpError Output:

```
Error: GET http://localhost:7180/v2/nodes/invalid-uuid - 404

==> Request
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Accept: application/json

Response <==
Content-Type: application/json
Date: Mon, 01 Jan 2024 12:00:00 GMT
Content-Length: 32
Body: {
  "error": "Node not found"
}
```

## Example 2: POST Request with Complex JSON Error

When creating a folder fails with detailed validation errors:

### HttpError Output:

```
Error: POST http://localhost:7180/v2/nodes - 400

==> Request
Content-Type: application/json
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
Body: {
  "mimetype": "application/vnd.antbox.folder",
  "parent": "parent-uuid",
  "title": "invalid/folder/name"
}

Response <==
Content-Type: application/json
Server: AntboxAPI/1.0
Date: Mon, 01 Jan 2024 12:00:00 GMT
Body: {
  "error": {
    "message": "Invalid folder name",
    "code": "VALIDATION_ERROR",
    "details": [
      {
        "field": "title",
        "message": "Folder name cannot contain '/' characters"
      }
    ]
  }
}
```

## Example 3: Authentication Error

When login fails:

### HttpError Output:

```
Error: POST http://localhost:7180/v2/login/root - 401

==> Request
Content-Type: text/plain
Body: a665a45920422f9d417e4867efdc4fb8a04a1f3fff1fa07e998e86f7f7a27ae3

Response <==
Content-Type: application/json
WWW-Authenticate: Bearer
Date: Mon, 01 Jan 2024 12:00:00 GMT
Body: {
  "error": "Invalid credentials"
}
```

## Example 4: Server Error with Plain Text Response

When the server returns a non-JSON error:

### HttpError Output:

```
Error: GET http://localhost:7180/v2/nodes - 500

==> Request
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Response <==
Content-Type: text/plain
Date: Mon, 01 Jan 2024 12:00:00 GMT
Body: Internal server error: Database connection failed
```

## Example 5: Network Timeout with Array Response

When listing nodes fails with validation errors:

### HttpError Output:

```
Error: GET http://localhost:7180/v2/nodes?parent=invalid-parent - 422

==> Request
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

Response <==
Content-Type: application/json
Date: Mon, 01 Jan 2024 12:00:00 GMT
Body: [
  {
    "field": "parent",
    "message": "Parent UUID is not valid"
  },
  {
    "field": "parent",
    "message": "Parent node does not exist"
  }
]
```

## Key Features

### 1. **Clear Error Header**

- Shows HTTP method, full URL, and status code in one line
- Format: `Error: METHOD URL - STATUS_CODE`

### 2. **Complete Request Information**

- All request headers are displayed
- Request body is included and pretty-printed if JSON
- Helps debug what was actually sent to the server

### 3. **Complete Response Information**

- All response headers are displayed
- Response body is included and pretty-printed if JSON
- Shows exactly what the server returned

### 4. **Clean Request JSON**

- Empty fields are automatically omitted from request bodies using `omitempty` tags
- Only populated fields are included in the JSON payload
- Reduces request size and improves readability

### 5. **JSON Pretty Printing**

- Both request and response JSON are automatically formatted
- Nested structures are clearly visible with proper indentation
- Arrays and objects are both handled correctly

### 6. **Fallback for Non-JSON**

- Plain text, HTML, or other content types are displayed as-is
- No formatting is applied to non-JSON content

## Benefits

- **Comprehensive Debugging**: See both sides of the HTTP conversation
- **Better Error Analysis**: Request and response headers help identify issues
- **Improved Readability**: Pretty-printed JSON makes complex structures easy to read
- **Clean Request Bodies**: Only necessary fields are sent, reducing payload size
- **Complete Context**: Full URL and all headers provide complete debugging context
- **Consistent Format**: All API errors follow the same structured format
