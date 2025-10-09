# antx - A Shell-like CLI for Antbox

`antx` is a command-line interface (CLI) that provides a shell-like experience for interacting with an Antbox server. It allows you to manage files, execute actions, and interact with AI agents directly from your terminal.

## What is Antbox?

Antbox is a flexible content management and automation platform. It provides a hierarchical structure for storing and managing "nodes" (files and folders), along with powerful features for automation and AI integration. Key features of Antbox include:

*   **Node Management:** Create, delete, move, and organize files and folders.
*   **Smart Folders:** Create dynamic folders whose content is determined by filters.
*   **AI Integration:** Interact with AI agents for chat, question answering, and other tasks.
*   **Automation:** Execute custom actions and extensions to automate workflows.
*   **Extensibility:** Upload and manage custom features, agents, and aspects.

## `antx` Manual

### Installation and Connection

To start using `antx`, you need to connect to an Antbox server.

```bash
antx [server_url] --api-key [your_api_key]
```

You can also authenticate using a JWT token or root password:

```bash
antx [server_url] --jwt [your_jwt_token]
antx [server_url] --root [root_password]
```

### Basic Commands

Once connected, you can use the following commands to interact with Antbox:

*   **`ls [folder_uuid]`**: List the content of a folder. If no `folder_uuid` is provided, it lists the content of the current folder.
*   **`cd [folder_uuid]`**: Change the current directory to the specified folder.
*   **`pwd`**: Print the current working directory (the current node's path).
*   **`chat [agent_uuid] [message]`**: Start an interactive chat session with an AI agent.
*   **`upload [file_path]`**: Upload a file to the current folder.
*   **`download [node_uuid] [download_path]`**: Download a file from Antbox.
*   **`mkdir [name]`**: Create a new folder in the current folder.
*   **`rm [node_uuid]`**: Remove a file or folder.
*   **`mv [node_uuid] [new_parent_uuid]`**: Move a file or folder to a new location.
*   **`rename [node_uuid] [new_name]`**: Rename a file or folder.
*   **`find [query]`**: Search for nodes based on a query.
*   **`run [action_uuid] [node_uuid]`**: Run an action on a specific node.
*   **`help`**: Display a list of available commands.
*   **`exit`**: Exit the `antx` shell.

### Advanced Usage

`antx` also supports more advanced features of Antbox, such as:

*   **Smart Folders:** Create and manage smart folders using the `mksmart` command.
*   **Agents:** List and interact with AI agents.
*   **Actions and Extensions:** List and execute custom actions and extensions.
*   **Templates:** List and manage templates.

For more detailed information on each command, you can use the `help` command within the `antx` shell.
