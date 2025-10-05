# Project Overview

This document provides an overview of the antx project, outlining its purpose, structure, features, and future direction.

## Project Description

The antx project is a command-line interface (CLI) tool designed to streamline the management and operation of Antbox deployments. It provides a comprehensive shell-like interface for interacting with the Antbox file system, including file manipulation, folder navigation, uploads, downloads, and node management operations. Its modular architecture ensures that each feature is encapsulated and maintainable for future enhancements.

## Codebase Structure

The project is structured as follows:

- **CLI Core (`cli/`):**
  Contains the primary logic for command parsing, execution, and interactive prompt functionality with auto-completion and suggestions.

- **API Client (`antbox/`):**
  Comprehensive HTTP client for interacting with the Antbox REST API, including authentication, file operations, and enhanced error handling with detailed HTTP request/response logging.

- **Command Interface (`cmd/`):**
  Entry point and command-line argument parsing using Cobra framework for clean CLI argument handling.

- **Project Directories:**
  The source code is located in the `antx/` directory, with clear separation between API client, CLI interface, and command handling modules.

## Key Features

- **Complete File System Operations:**
  - `ls` - List folder contents with detailed information
  - `cd` - Navigate between folders using UUIDs or names
  - `pwd` - Show current location
  - `mkdir` - Create new directories
  - `rm` - Remove nodes (files and folders)
  - `mv` - Move nodes between locations
  - `rename` - Rename files and folders
  - `cp` - Copy nodes to different locations
  - `duplicate` - Duplicate nodes in the same location

- **File Transfer Operations:**
  - `upload` - Upload local files to Antbox folders
  - `get` - Download Antbox nodes to local Downloads folder

- **Node Information:**
  - `stat` - Display detailed node properties and metadata

- **Action and Extension Operations:**
  - `run` - Execute actions on nodes with optional parameters
  - `call` - Execute extensions with optional parameters

- **Template Operations:**
  - `template` - Download templates to Downloads folder

- **Enhanced User Experience:**
  - Interactive auto-completion with smart suggestions
  - Folder-only suggestions for navigation commands
  - UUID-based operations with user-friendly names
  - Visual indicators (📁) for different node types

- **Robust Error Handling:**
  - Detailed HTTP error messages with request/response information
  - Pretty-printed JSON for better readability
  - Complete debugging context for API failures

- **Clean JSON Operations:**
  - Request bodies with `omitempty` tags to exclude empty fields
  - Efficient network usage with minimal payloads

## How to Use

The antx provides a shell-like interactive interface. Start the CLI with connection parameters:

```
./antx [server url] (--api-key [api-key]|--root [root passwd]|--jwt [jwt])
```

Once connected, you can use familiar commands:

- `ls` - List current folder contents
- `cd folder-name` - Navigate to folders (with auto-completion)
- `mkdir new-folder` - Create directories
- `cp /local/file.txt destination-folder` - Upload files
- `get file-uuid` - Download files to ~/Downloads
- `rm node-uuid` - Delete files and folders
- `mv source-uuid destination-uuid` - Move nodes
- `rename node-uuid "new name"` - Rename nodes
- `cp source-uuid destination-uuid` - Copy nodes to different locations
- `duplicate node-uuid` - Duplicate nodes in the same location
- `run action-uuid node-uuid param=value` - Execute actions on nodes
- `call extension-uuid param=value` - Execute extensions
- `template template-uuid` - Download templates

All commands support intelligent auto-completion with suggestions appearing after typing 2+ characters.

## Enhanced Auto-completion Features

- **Smart Action/Extension Discovery**: `run` and `call` commands dynamically list available actions and extensions
- **Agent Suggestions**: `chat` and `answer` commands suggest available AI agents
- **Node UUID Completion**: Most commands provide filtered node suggestions based on context
- **Folder-Only Filtering**: Navigation commands like `cd` and `cp` only suggest appropriate folder targets

## Architecture Highlights

- **HTTP Client with Advanced Error Handling:**
  Custom `HttpError` type that captures complete request/response details for debugging

- **Smart Command Completion:**
  Context-aware suggestions that understand command requirements (e.g., folders-only for `cd`)

- **UUID-Based Operations:**
  Efficient backend operations using UUIDs while presenting user-friendly names in the interface

- **Multipart File Uploads:**
  Robust file upload handling with proper content-type detection

- **Stream-Based Downloads:**
  Efficient file downloads with automatic directory creation

## Contribution and Future Enhancements

Contributions are welcome! The project adheres to a modular design which means new features or improvements can be easily integrated. Future enhancements include:

- Additional file operations (copy, symbolic links, permissions management)
- Batch operations for multiple files
- Search and filtering capabilities
- Progress bars for large file transfers
- Configuration management and profiles

## Additional Resources

- [Project Overview](PROJECT_OVERVIEW.md) – You are here!
- [README](README.md) – Contains setup instructions and detailed usage information.

---

A link to this overview has been added at the bottom of the README file for easy access. Enjoy exploring the antx project!
