# antx - A Shell-like CLI for Antbox

`antx` is a command-line interface (CLI) that provides a shell-like experience for interacting with an Antbox server. It allows you to manage files and interact with AI agents directly from your terminal.

## What is Antbox?

Antbox is a flexible content management and automation platform. It provides a hierarchical structure for storing and managing "nodes" (files and folders), along with powerful features for automation and AI integration. Key features of Antbox include:

*   **Node Management:** Create, delete, move, and organize files and folders.
*   **Smart Folders:** Create dynamic folders whose content is determined by filters.
*   **AI Integration:** Interact with AI agents for chat, question answering, and other tasks.
*   **AI Integration:** Upload agents and interact with them from the shell.

## `antx` Manual

### Installation and Connection

To start using `antx`, you need to connect to an Antbox server.

```bash
antx [server_url] --api-key [your_api_key]
```

You can also authenticate using a JWT token, root password, or Lightray browser/device login:

```bash
antx [server_url] --jwt [your_jwt_token]
antx [server_url] --root [root_password]
antx https://vcrm.lightray.cloud --lightray
```

With `--lightray`, the URL is treated as the Lightray URL and `antx` uses `${url}/api` as the Antbox API endpoint. The CLI prints a browser verification link and device code, then waits for approval. The default Lightray OAuth client ID is `terminal-cli`; override it with `--lightray-client-id` if your deployment uses a different device-flow client.

### Basic Commands

Once connected, you can use the following commands to interact with Antbox:

*   **`ls [folder_uuid]`**: List the content of a folder. If no `folder_uuid` is provided, it lists the content of the current folder.
*   **`cd [folder_uuid]`**: Change the current directory to the specified folder.
*   **`pwd`**: Print the current working directory (the current node's path).
*   **`/[agent_uuid] [message]`** or **`/agent_uuid [message]`**: Start an interactive chat session with an AI agent.
*   **`@[agent_uuid] <question>`** or **`@agent_uuid <question>`**: Ask an AI agent for a single answer.
*   **`upload [file_path]`**: Upload a file to the current folder. Use `upload -i agent.json` to create or replace an AI agent from a JSON payload.
*   **`download [node_uuid] [download_path]`**: Download a file from Antbox.
*   **`mkdir [name]`**: Create a new folder in the current folder.
*   **`rm [node_uuid]`**: Remove a file or folder.
*   **`mv [node_uuid] [new_parent_uuid]`**: Move a file or folder to a new location.
*   **`rename [node_uuid] [new_name]`**: Rename a file or folder.
*   **`find [query]`**: Search for nodes based on a query.
*   **`audit [node_uuid]`**: Show the audit history for a node.
*   **`clone [node_uuid]`**: Clone a node in the same location.
*   **`help`**: Display a list of available commands.
*   **`exit`**: Exit the `antx` shell.

### Advanced Usage

`antx` also supports more advanced features of Antbox, such as:

*   **Smart Folders:** Create and manage smart folders using the `mksmart` command.
*   **Agents:** List agents and interact with them using `/` and `@` shortcuts.

For more detailed information on each command, you can use the `help` command within the `antx` shell.
