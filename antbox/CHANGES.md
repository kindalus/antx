# Antbox Go Client Changes

This document outlines the changes made to align the antbox Go client with the OpenAPI specification.

## Breaking Changes

### Agent Operations

#### CreateAgent Method Signature
- **Before**: `CreateAgent(filePath string, metadata AgentCreate) (*Agent, error)`
- **After**: `CreateAgent(agent AgentCreate) (*Agent, error)`
- **Impact**: Method now accepts AgentCreate directly instead of file path + metadata
- **API Endpoint**: Changed from `POST /agents/-/upload` to `POST /agents` with JSON body

#### Agent Schema Changes
- **Removed fields**: `Created`, `Updated`, `Temperature`, `MaxTokens`, `Reasoning`, `UseTools`, `StructuredAnswer`
- **Added fields**: `Group`, `CreatedAt`, `ModifiedAt`
- **Reordered**: `SystemInstructions` is now the first field after UUID

#### AgentCreate Schema Changes
- **Removed fields**: `Temperature`, `MaxTokens`, `Reasoning`, `UseTools`, `StructuredAnswer`
- **Required fields**: Only `SystemInstructions` and `Title` are now required

### Feature Schema Changes

#### Field Name Changes
- **Before**: `ExposeAction` → **After**: `ExposeAsAction`
- **Before**: `ExposeExtension` → **After**: `ExposeAsExtension`

#### Removed Fields
- `RunOnCreates`
- `RunOnUpdates`
- `RunManually`
- `Filters`
- `ExposeAITool`
- `RunAs`
- `GroupsAllowed`
- `ReturnContentType`

#### Added Fields
- `Fid` - File identifier
- `Title` - Feature title
- `Mimetype` - MIME type
- `Parent` - Parent folder UUID
- `Owner` - Owner email
- `Group` - Group UUID
- `Permissions` - Permission settings
- `CreatedAt` - Creation timestamp
- `ModifiedAt` - Last modification timestamp

### API Key Schema Changes

#### Field Name Changes
- **Before**: `Secret` → **After**: `Key`
- **Before**: `Owner` → **After**: `CreatedBy`

#### Added Fields
- `CreatedAt` - Creation timestamp

### User Schema Changes

#### Removed Fields
- `Groups` (array) - Replaced with single `Group` field

#### Added Fields
- `Role` - User role
- `Active` - Account status
- `CreatedAt` - Creation timestamp

#### UserCreate Schema Changes
- **Added**: `Password` field (required)
- **Added**: `Role` field (optional)

#### UserUpdate Schema Changes
- **Added**: `Password` field (optional)
- **Added**: `Role` field (optional)
- **Added**: `Active` field (optional pointer to bool)

### Group Schema Changes

#### Field Name Changes
- **Before**: `Title` → **After**: `Name`

#### Added Fields
- `Description` - Group description
- `CreatedAt` - Creation timestamp
- `ModifiedAt` - Last modification timestamp

#### GroupCreate Schema Changes
- **Before**: `Title` → **After**: `Name`
- **Added**: `Description` field (optional)

#### GroupUpdate Schema Changes
- **Before**: `Title` → **After**: `Name`
- **Added**: `Description` field (optional)

### Template Schema Changes

#### Removed Fields
- `Description`

#### Added Fields
- `Mimetype` - MIME type of template
- `Size` - Size in bytes

### Aspect Schema Changes

#### Added Fields
- `Name` - Aspect name
- `Mimetype` - MIME type
- `Owner` - Owner email
- `Permissions` - Permission settings

#### Removed Fields
- `Filters` - Filter criteria
- `Properties` - Custom properties

#### AspectCreate Schema Changes
- **Added**: `Name` field (required)
- **Added**: `Mimetype` field (required)
- **Added**: `Permissions` field (optional)
- **Removed**: `Filters` field
- **Removed**: `Properties` field

## Migration Guide

### For Agent Creation
```go
// Before
agent, err := client.CreateAgent("/path/to/file", antbox.AgentCreate{
    Title: "My Agent",
    // ... other fields
})

// After
agent, err := client.CreateAgent(antbox.AgentCreate{
    SystemInstructions: "You are a helpful assistant",
    Title: "My Agent",
    // ... other fields
})
```

### For Feature Field Access
```go
// Before
if feature.ExposeAction {
    // ...
}

// After
if feature.ExposeAsAction {
    // ...
}
```

### For API Key Field Access
```go
// Before
fmt.Println("Secret:", apiKey.Secret)
fmt.Println("Owner:", apiKey.Owner)

// After
fmt.Println("Key:", apiKey.Key)
fmt.Println("Created By:", apiKey.CreatedBy)
fmt.Println("Created At:", apiKey.CreatedAt)
```

### For Group Field Access
```go
// Before
fmt.Println("Title:", group.Title)

// After
fmt.Println("Name:", group.Name)
fmt.Println("Description:", group.Description)
```

## Compatibility Notes

- All changes maintain the same HTTP client behavior
- JSON marshaling/unmarshaling automatically handles the new field names
- Tests have been updated to reflect the new schemas
- Mock clients in tests have been updated to match new interfaces

## OpenAPI Specification Compliance

These changes bring the Go client into full compliance with the OpenAPI 3.1.0 specification version 2.0.0, ensuring consistent data structures and API interactions across all client implementations.
