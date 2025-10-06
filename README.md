# Antbox CLI

A shell-like CLI for Antbox.

## Usage

```
./antx [server url] (--api-key [api-key]|--root [root passwd]|--jwt [jwt])
```

## Commands

### File System Operations

- `ls [uuid]`: List the content of a folder. If no uuid is given, it lists the current folder.
- `pwd`: Show the current location.
- `stat <uuid>`: Show node properties.
- `cd [uuid]`: Change the current folder. If no uuid is given, it changes to the root folder.
- `mkdir <name>`: Create a directory in the current folder.
- `rm <uuid>`: Remove a node given the uuid.
- `mv <uuid> <uuid>`: Move a node to another node (folder).
- `rename <uuid> <new name>`: Change the name of a node to the new name.
- `cp <file path> <uuid>`: Upload a file from the local filesystem to a folder given the uuid.
- `get <uuid>`: Download the node with the given uuid to the Downloads folder. The filename will be the same as the title of the node.

### Interactive Chat Commands

- `chat [options] <agent_uuid> [message]`: Chat with a specific agent (always interactive)
  - Options: `-t <temperature>`, `-m <max_tokens>`
  - Always enters interactive mode. If a message is provided, it's sent as the first message
  - Type `exit` or press Ctrl+D to exit interactive sessions
- `rag [options] [message]`: Chat with the RAG agent
  - Options: `-l` (use location context)
  - Interactive mode: If no message is provided, enters an interactive RAG session
  - Type `exit` or press Ctrl+D to exit interactive sessions

### Other Commands

- `reload`: Reload cached data from server
- `status`: Show cached data statistics
- `exit`: Exit the CLI.

### Interactive Sessions

Interactive sessions:

- `chat` is always interactive - starts a continuous conversation with the agent
- `rag` enters interactive mode when no message is provided
- Exit with `exit` command or Ctrl+D
- Both chat and RAG sessions are stateless (no conversation history persistence)

## Project Overview

For a detailed description of the project structure and its features, please refer to the [Project Overview](PROJECT_OVERVIEW.md).
