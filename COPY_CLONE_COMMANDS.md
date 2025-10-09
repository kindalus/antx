# Copy and Clone Commands Documentation

This document describes the implementation and usage of the copy and clone commands in the Antbox CLI.

## Overview

The Antbox CLI provides three related commands for duplicating nodes:

- **`cp`** (copy) - Copy a node to a different location with optional title
- **`duplicate`** - Duplicate a node in the same location
- **`clone`** - Alternative name for duplicate functionality

All commands follow the project's established patterns and use the proper Antbox API endpoints.

## Commands

### `cp` (Copy Command)

Copies a node to a different location with an optional new title.

#### Usage
```
cp <source_uuid> <destination_uuid> [new_title]
```

#### Arguments
- `source_uuid` - UUID of the node to copy
- `destination_uuid` - UUID of the destination folder
- `new_title` - Optional new title for the copied node

#### Special Aliases
- `.` - Current node (for source or destination)
- `..` - Parent node (for destination)

#### Examples
```bash
# Copy with auto-generated title
cp abc123 def456

# Copy to current location
cp abc123 . "Local Copy"

# Copy current node to another folder
cp . folder-uuid "Copy of Current"

# Copy to parent directory
cp doc-uuid .. "Moved Up Copy"
```

#### Behavior
- Validates source node exists before copying
- Validates destination is a folder (regular or smart folder)
- Generates default title "Copy of <original_title>" if no title provided
- Shows detailed success message with source, destination, and new node info

#### API Endpoint
Uses `POST /nodes/{uuid}/-/copy` with JSON payload:
```json
{
  "parent": "destination-uuid",
  "title": "new title"
}
```

### `duplicate` Command

Creates a duplicate of a node in the same location (same parent folder).

#### Usage
```
duplicate <uuid>
```

#### Arguments
- `uuid` - UUID of the node to duplicate

#### Special Aliases
- `.` - Current node

#### Examples
```bash
# Duplicate a specific node
duplicate abc123-def456-ghi789

# Duplicate current node
duplicate .
```

#### Behavior
- Creates duplicate in same parent folder as original
- Auto-generates title "Copy of <original_title>"
- Preserves original node properties (mimetype, size, etc.)
- Shows success message with original and new node info

#### API Endpoint
Uses `GET /nodes/{uuid}/-/duplicate`

### `clone` Command

Alternative name for the duplicate command with identical functionality.

#### Usage
```
clone <uuid>
```

#### Arguments
- `uuid` - UUID of the node to clone

#### Special Aliases
- `.` - Current node

#### Examples
```bash
# Clone a specific node
clone abc123-def456-ghi789

# Clone current node
clone .
```

#### Behavior
- Identical to `duplicate` command
- Uses same API endpoint internally
- Provided for user preference (some prefer "clone" terminology)

## Implementation Details

### File Structure
```
antx/cli/
├── copy.go      # cp command implementation
├── duplicate.go # duplicate command implementation
└── clone.go     # clone command implementation
```

### API Integration

Both commands use the Antbox client interface defined in `antbox/antbox.go`:

```go
type Antbox interface {
    CopyNode(uuid, parent, title string) (*Node, error)
    DuplicateNode(uuid string) (*Node, error)
    // ... other methods
}
```

### Error Handling

All commands provide comprehensive error handling:

1. **Source Validation**
   - Checks if source node exists
   - Provides clear error message if not found

2. **Destination Validation** (cp only)
   - Checks if destination exists
   - Validates destination is a folder
   - Rejects non-folder destinations

3. **API Errors**
   - Catches and displays API-level errors
   - Provides context about which operation failed

### Auto-completion Support

All commands support intelligent auto-completion:

- **cp command**:
  - First argument: Any node
  - Second argument: Only folders
  - Third argument: No suggestions (free text title)

- **duplicate/clone commands**:
  - Single argument: Any node

### Project Style Compliance

The commands follow established project patterns:

1. **Command Structure**
   - Implements `Command` interface
   - Provides `GetName()`, `GetDescription()`, `Execute()`, `Suggest()`
   - Registers via `init()` function

2. **Help Format**
   - Comprehensive usage information
   - Description, arguments, examples
   - Special aliases documentation

3. **Output Format**
   - Consistent success/error message styling
   - Detailed operation feedback
   - UUID and title information

4. **Error Messages**
   - Prefixed with "Error:"
   - Descriptive and actionable
   - Consistent formatting

## Testing

### Automated Testing

Comprehensive test suite in `tests/test_copy_clone_commands.go`:

- ✅ Copy operations with various scenarios
- ✅ Duplicate/clone operations
- ✅ Error handling validation
- ✅ API integration verification
- ✅ Edge case coverage

### Manual Testing

Commands can be tested interactively:

1. Start CLI: `./antx <server-url> --credentials`
2. List available nodes: `ls`
3. Test copy: `cp <source> <dest> "Test Copy"`
4. Test duplicate: `duplicate <uuid>`
5. Test clone: `clone <uuid>`
6. Verify results: `ls`

## Command Comparison

| Feature | cp | duplicate | clone |
|---------|----|-----------| ------|
| **Destination** | Different location | Same location | Same location |
| **Title Control** | Optional custom | Auto-generated | Auto-generated |
| **Arguments** | 2-3 | 1 | 1 |
| **API Used** | CopyNode | DuplicateNode | DuplicateNode |
| **Use Case** | Move copy elsewhere | Quick duplicate | Quick duplicate |

## Best Practices

### When to Use Each Command

- **Use `cp`** when you need to:
  - Copy to a different folder
  - Specify a custom title
  - Organize content across locations

- **Use `duplicate`** when you need to:
  - Quick duplicate in same location
  - Preserve original location/context
  - Create backup copy nearby

- **Use `clone`** when you prefer this terminology over "duplicate"

### Tips

1. **Use aliases**: `.` and `..` make commands more convenient
2. **Check destination**: Ensure target folder exists before copying
3. **Title planning**: Consider if auto-generated titles are sufficient
4. **Batch operations**: Use these commands in scripts for bulk operations

## Error Resolution

### Common Issues

1. **"Node not found"**
   - Verify UUID is correct
   - Check if node was deleted
   - Use `ls` to see available nodes

2. **"Destination is not a folder"**
   - Ensure copying to folder UUID, not file UUID
   - Use `stat <uuid>` to check node type

3. **"Permission denied"**
   - Verify authentication credentials
   - Check folder write permissions
   - Ensure API key has sufficient rights

### Debugging Steps

1. Verify source exists: `stat <source_uuid>`
2. Verify destination exists: `stat <dest_uuid>`
3. Check current location: `pwd`
4. List available nodes: `ls`
5. Use verbose mode: `antx --verbose ...`

## API Endpoints Reference

### Copy Endpoint
```
POST /nodes/{uuid}/-/copy
Content-Type: application/json

{
  "parent": "destination-folder-uuid",
  "title": "New title for copied node"
}

Response: Node object with new UUID
```

### Duplicate Endpoint
```
GET /nodes/{uuid}/-/duplicate

Response: Node object with new UUID and auto-generated title
```

## Status

✅ **IMPLEMENTED** - All commands are fully implemented and tested
✅ **TESTED** - Comprehensive test suite validates functionality
✅ **DOCUMENTED** - Complete documentation and examples provided
✅ **PRODUCTION READY** - Commands follow project standards and best practices

The copy and clone commands provide a complete solution for node duplication needs in the Antbox CLI, offering flexibility for both simple duplication and complex cross-folder copying scenarios.
