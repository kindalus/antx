# Current Node Persistence Fix

## Problem Description

The CLI was saving the current node UUID to the history file **before** navigation commands completed, resulting in the config file always containing the "previous" location instead of the actual current location.

### The Issue Flow (Before Fix)

1. User runs: `cd some-folder-uuid`
2. `executor()` calls `addCommandToHistory("cd some-folder-uuid")` **immediately**
3. `addCommandToHistory()` calls `saveCurrentState()` with **old** `currentNode.UUID`
4. `cd` command executes and updates `currentNode` to new location
5. Config file saved with wrong (previous) location

### Example Scenario

```
Starting location: --root--
Config file shows: --root--

> cd folder1
Config file shows: --root-- (WRONG - should be folder1)

> cd folder2
Config file shows: folder1 (WRONG - should be folder2)

> cd ..
Config file shows: folder2 (WRONG - should be folder1)
```

## Solution

**Move the `addCommandToHistory()` call to AFTER command execution.**

### Code Change

**File**: `cli/prompt.go`

**Before** (lines 60-82):
```go
func executor(in string) {
    in = strings.TrimSpace(in)

    // Add command to history (will be saved to disk)
    addCommandToHistory(in)  // ← PROBLEM: Called before command executes

    parts := strings.Split(in, " ")
    commandName := parts[0]
    args := parts[1:]

    // ... alias resolution ...

    if cmd, ok := commands[commandName]; ok {
        cmd.Execute(args)  // ← Command updates currentNode here
    } else {
        fmt.Println("Unknown command: " + commandName)
    }

    fmt.Println("")
}
```

**After** (fixed):
```go
func executor(in string) {
    in = strings.TrimSpace(in)

    parts := strings.Split(in, " ")
    commandName := parts[0]
    args := parts[1:]

    // ... alias resolution ...

    if cmd, ok := commands[commandName]; ok {
        cmd.Execute(args)  // ← Command updates currentNode here
    } else {
        fmt.Println("Unknown command: " + commandName)
    }

    // Add command to history AFTER execution (so currentNode is updated)
    addCommandToHistory(in)  // ← SOLUTION: Called after command executes

    fmt.Println("")
}
```

### The Fix Flow (After Fix)

1. User runs: `cd some-folder-uuid`
2. `cd` command executes and updates `currentNode` to new location
3. `addCommandToHistory("cd some-folder-uuid")` called with **updated** `currentNode.UUID`
4. `saveCurrentState()` saves correct current location
5. Config file contains correct location

### Example Scenario (Fixed)

```
Starting location: --root--
Config file shows: --root--

> cd folder1
Config file shows: folder1 ✅

> cd folder2
Config file shows: folder2 ✅

> cd ..
Config file shows: folder1 ✅
```

## Impact

This fix ensures that:

1. **CLI persistence works correctly** - When you restart the CLI, you're in the right location
2. **State consistency** - The saved state always reflects the actual current state
3. **User experience** - No confusion about "where am I" after restart
4. **Command history integrity** - History and location are saved atomically

## Testing the Fix

### Manual Test Procedure

1. Start the CLI: `./antx <server-url> --credentials`
2. Navigate somewhere: `cd <some-folder-uuid>`
3. Check config file: `cat ~/.antx` (first line should show the folder you're in)
4. Navigate again: `cd <another-folder-uuid>`
5. Check config file: `cat ~/.antx` (first line should show new location)
6. Exit and restart CLI
7. Verify you're in the correct location (use `pwd` or `status`)

### Expected Results

- Config file first line always matches current location
- CLI restores to correct location on restart
- No "one step behind" behavior

## Files Modified

- `cli/prompt.go` - Moved `addCommandToHistory()` call after command execution

## Files Affected by Fix

- `cli/config.go` - `saveCurrentState()` now receives correct `currentNode.UUID`
- `~/.antx` - Config file now contains correct current location
- All navigation commands (`cd`, etc.) - Now save state correctly

## Status

✅ **FIXED** - Current node persistence now works correctly
✅ **TESTED** - Manual testing confirms fix resolves the issue
✅ **NO BREAKING CHANGES** - Fix is backwards compatible

This fix resolves the "always one behind" issue and ensures the CLI provides a reliable, consistent user experience with proper state persistence.
