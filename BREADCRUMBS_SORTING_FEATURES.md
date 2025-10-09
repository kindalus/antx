# Breadcrumbs and Sorting Features Documentation

This document describes the breadcrumbs display and alphabetical sorting features implemented in the Antbox CLI.

## Overview

Two major enhancements have been added to improve the user experience:

1. **Startup Breadcrumbs** - Display current location path when CLI starts
2. **Alphabetical Sorting** - Sort all listings alphabetically with directories first for node listings

## Features

### 1. Startup Breadcrumbs Display

#### Description
When the CLI starts, it automatically displays the current location as a breadcrumb path, making it immediately clear where you are in the folder hierarchy.

#### Behavior
- Displays after initialization completes
- Shows hierarchical path from root to current location
- Uses the same breadcrumbs API as the `pwd` command
- Gracefully falls back to simple location display if breadcrumbs fail

#### Examples
```
Initializing... ✓ Ready

Current location: /
```

```
Initializing... ✓ Ready

Current location: /Documents/Projects/MyProject
```

#### Implementation
- Located in `cli/prompt.go`
- Function: `showStartupBreadcrumbs()`
- Called after CLI initialization in `Start()` function
- Uses `client.GetBreadcrumbs()` API endpoint

### 2. Alphabetical Sorting

#### Description
All CLI listings now display items in alphabetical order to improve usability and consistency. For node listings (`ls`, `find`), directories are shown first, then files, with both groups sorted alphabetically.

#### Affected Commands

##### Node Listings (`ls`, `find`)
- **Directories first**: Regular folders and smart folders
- **Files second**: All other mimetypes
- **Alphabetical within each group**: Sorted by title
- **Case-sensitive sorting**: Uses standard string comparison

##### Other Listings
- **Agents** (`agents` command): Sorted alphabetically by title
- **Extensions** (`extensions` command): Sorted alphabetically by name
- **Actions** (`actions` command): Sorted alphabetically by name

#### Examples

##### Before Sorting (Random Order)
```
 UUID          SIZE  MODIFIED      MIMETYPE                        TITLE
 ------------  ----  ------------  ------------------------------  -----
 file-readme   1K    Jan 05 16:30  text/markdown                   README.md
 smart-recent  0B    Jan 04 08:00  application/vnd.antbox.smart... Recent Items
 folder-proj   0B    Jan 03 09:15  application/vnd.antbox.folder   Projects
 file-config   512B  Jan 06 11:20  application/json                config.json
 folder-docs   0B    Jan 02 10:30  application/vnd.antbox.folder   Documents
```

##### After Sorting (Directories First, Alphabetical)
```
 UUID          SIZE  MODIFIED      MIMETYPE                        TITLE
 ------------  ----  ------------  ------------------------------  -----
 folder-docs   0B    Jan 02 10:30  application/vnd.antbox.folder   Documents
 folder-proj   0B    Jan 03 09:15  application/vnd.antbox.folder   Projects
 smart-recent  0B    Jan 04 08:00  application/vnd.antbox.smart... Recent Items
 file-readme   1K    Jan 05 16:30  text/markdown                   README.md
 file-config   512B  Jan 06 11:20  application/json                config.json
```

##### Agents Sorting
```
Available agents (4):

UUID: agent-bot
  Title: Chat Bot
  Description: General conversation assistant

UUID: agent-coder
  Title: Code Helper
  Description: Assists with programming

UUID: agent-analyst
  Title: Data Analyst
  Description: Analyzes data patterns

UUID: agent-writer
  Title: Writer Assistant
  Description: Helps with writing tasks
```

## Technical Implementation

### File Structure
```
antx/cli/
├── prompt.go     # Startup breadcrumbs display
├── utils.go      # Shared sorting function
├── ls.go         # Node listing with sorting
├── find.go       # Search results with sorting
├── agents.go     # Agents listing with sorting
├── extensions.go # Extensions listing with sorting
└── actions.go    # Actions listing with sorting
```

### Core Functions

#### Breadcrumbs Display
```go
// showStartupBreadcrumbs displays the current location path on startup
func showStartupBreadcrumbs() {
    breadcrumbs, err := client.GetBreadcrumbs(currentNode.UUID)
    if err != nil {
        // Fallback to simple display
        fmt.Printf("\nCurrent location: %s\n", getCurrentFolderName())
        return
    }

    // Build path from breadcrumbs
    var pathParts []string
    for _, node := range breadcrumbs {
        if node.Title != "" {
            pathParts = append(pathParts, node.Title)
        }
    }

    if len(pathParts) == 0 {
        fmt.Println("\nCurrent location: /")
    } else {
        fmt.Printf("\nCurrent location: /%s\n", strings.Join(pathParts, "/"))
    }
}
```

#### Node Sorting Function
```go
// sortNodesForListing sorts nodes with directories first, then files, both alphabetically by title
func sortNodesForListing(nodes []antbox.Node) []antbox.Node {
    if len(nodes) == 0 {
        return nodes
    }

    // Separate directories and files
    var directories []antbox.Node
    var files []antbox.Node

    for _, node := range nodes {
        isFolder := node.Mimetype == "application/vnd.antbox.folder" ||
            node.Mimetype == "application/vnd.antbox.smartfolder"

        if isFolder {
            directories = append(directories, node)
        } else {
            files = append(files, node)
        }
    }

    // Sort directories alphabetically by title
    sort.Slice(directories, func(i, j int) bool {
        return directories[i].Title < directories[j].Title
    })

    // Sort files alphabetically by title
    sort.Slice(files, func(i, j int) bool {
        return files[i].Title < files[j].Title
    })

    // Combine directories first, then files
    result := make([]antbox.Node, 0, len(nodes))
    result = append(result, directories...)
    result = append(result, files...)

    return result
}
```

### API Integration

#### Breadcrumbs Endpoint
- **Endpoint**: `GET /nodes/{uuid}/-/breadcrumbs`
- **Returns**: Array of Node objects representing the path from root to the specified node
- **Usage**: Called once on startup to display current location

#### Sorting Implementation
- **Client-side sorting**: All sorting is performed in the CLI after receiving data from the server
- **No API changes required**: Uses existing list endpoints
- **Performance**: Minimal impact as sorting is done on typically small datasets

## Benefits

### User Experience Improvements
1. **Immediate Orientation**: Users know where they are as soon as CLI starts
2. **Predictable Ordering**: Consistent alphabetical sorting across all listings
3. **Better Navigation**: Directories grouped together and easy to find
4. **Professional Feel**: Matches expectations from other CLI tools

### Developer Benefits
1. **Consistent Interface**: Same sorting behavior across all commands
2. **Maintainable Code**: Shared sorting function reduces duplication
3. **Extensible**: Easy to add sorting to new listing commands

## Configuration

### No Configuration Required
Both features are enabled by default and require no user configuration:
- Breadcrumbs display automatically on startup
- Sorting is applied to all relevant commands

### Fallback Behavior
- **Breadcrumbs**: Falls back to simple folder name if API call fails
- **Sorting**: Gracefully handles empty lists and missing fields

## Testing

### Automated Testing
Comprehensive test suite validates both features:

```bash
cd tests && go run test_breadcrumbs_sorting.go
```

#### Test Coverage
- ✅ Breadcrumbs display for various folder depths
- ✅ Node sorting with mixed directories and files
- ✅ Agents alphabetical sorting by title
- ✅ Extensions alphabetical sorting by name
- ✅ Actions alphabetical sorting by name
- ✅ Edge cases (empty lists, single items)

### Manual Testing

#### Test Breadcrumbs
1. Start CLI: `./antx <server-url> --credentials`
2. Observe breadcrumbs display after "Ready"
3. Navigate: `cd <folder-uuid>`
4. Restart CLI and verify correct location shown

#### Test Sorting
1. Run listing commands: `ls`, `agents`, `extensions`, `actions`
2. Verify directories appear first in `ls` output
3. Verify alphabetical ordering within each group
4. Test with various folder contents

## Troubleshooting

### Breadcrumbs Issues
**Problem**: Breadcrumbs not displaying or showing error
- **Cause**: Network connectivity or API authentication issues
- **Solution**: Check server connection and credentials
- **Fallback**: Simple folder name will be displayed

**Problem**: Incorrect breadcrumbs path
- **Cause**: Corrupted saved state or server data inconsistency
- **Solution**: Delete `~/.antx` file to reset state

### Sorting Issues
**Problem**: Items not sorting correctly
- **Cause**: Unexpected data format or empty titles
- **Investigation**: Check item titles for special characters
- **Solution**: Sorting handles empty titles gracefully

**Problem**: Directories mixed with files
- **Cause**: Unexpected mimetype values
- **Investigation**: Check node mimetypes in server data
- **Expected**: `application/vnd.antbox.folder` or `application/vnd.antbox.smartfolder`

## Performance Considerations

### Breadcrumbs
- **API Call**: One additional API call on startup
- **Impact**: Minimal - only called once per session
- **Caching**: Breadcrumbs are not cached (updated on each navigation)

### Sorting
- **Client-side**: All sorting performed locally
- **Complexity**: O(n log n) where n is number of items
- **Memory**: Minimal additional memory usage
- **Impact**: Negligible for typical folder sizes (< 1000 items)

## Future Enhancements

### Potential Improvements
1. **Case-insensitive sorting**: Option for case-insensitive alphabetical sorting
2. **Sort configuration**: User preference for sorting method (alphabetical, size, date)
3. **Breadcrumbs caching**: Cache breadcrumbs to reduce API calls
4. **Custom separators**: Configurable path separator in breadcrumbs

### Backward Compatibility
- All changes are additive and maintain full backward compatibility
- No breaking changes to existing CLI behavior
- No impact on server-side functionality

## Status

✅ **IMPLEMENTED** - Both features fully implemented and tested
✅ **TESTED** - Comprehensive automated and manual testing completed
✅ **DOCUMENTED** - Complete documentation and examples provided
✅ **PRODUCTION READY** - Features ready for production deployment

The breadcrumbs and sorting features significantly enhance the Antbox CLI user experience by providing immediate location awareness and consistent, predictable ordering of all listings.
