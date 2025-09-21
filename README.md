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
- `exit`: Exit the CLI.

## Project Overview

For a detailed description of the project structure and its features, please refer to the [Project Overview](PROJECT_OVERVIEW.md).
