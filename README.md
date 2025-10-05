# Antbox CLI

A shell-like CLI for Antbox.

## Usage

```
./antx [server url] (--api-key [api-key]|--root [root passwd]|--jwt [jwt])
```

## Commands

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
- `exit`: Exit the CLI.

## Project Overview

For a detailed description of the project structure and its features, please refer to the [Project Overview](PROJECT_OVERVIEW.md).
