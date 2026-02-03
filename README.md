# Cool-Code: CLI Coding AI Agent

An intelligent command-line interface that leverages AI to help you interact with codebases and work on them similar to Gemini CLI and Claude Code.

## Overview

Cool-Code is a powerful tool that combines the capabilities of large language models with a comprehensive set of development tools to provide an interactive development experience. Simply describe what you want to accomplish, and the AI agent will understand your intent and execute the necessary operations.

Demo of it spinning up a Node Express server with Prisma:

https://github.com/user-attachments/assets/b1b59602-f118-4bbe-8b38-e4be0f39119f

## Features

- **Agent Modes**: Switch between Planning, Agent, and Ask modes for different interaction styles.
- **Task Tracking**: Real-time checklists for complex, multi-step operations.
- **Input Queuing**: Send new instructions while the agent is still processing a previous request.
- **Natural Language Processing**: Interact with your codebase using plain English.
- **Intelligent Code Analysis**: Understands your project structure and coding patterns without needing vector databases.
- **File Operations**: Read, create, edit, and search files with AI assistance.
- **Shell Command Execution**: Run system commands safely through the AI agent.
- **Real-time Streaming**: Get live feedback as the AI processes your requests.
- **Context-Aware**: Maintains conversation history and project context automatically.

## Setup Guide

### Prerequisites

- Node.js
- A Google AI API key for Gemini

### Quick Install (Recommended)

Install globally from npm:

```bash
npm install -g cool-code
```

Set your Google AI API key:

```bash
export GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here
```

### Development Setup

1. Clone the repository:

```bash
git clone https://github.com/rushikeshg25/cool-code.git
cd cool-code
```

2. Install dependencies:

```bash
npm install
```

3. Set up environment variables:

```bash
cp .env.example .env
```

Edit `.env` and add your Google AI API key:

```
GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here
```

4. Build the project:

```bash
npm run build
```

5. Link for local development:

```bash
npm link
```

## Usage

### Starting the CLI

Navigate to your project's directory and run:

```bash
cool-code
```

On startup, the CLI will ask to initialize in the current directory before building context.

### Command Line Options

```bash
# Auto-initialize in current dir
cool-code --yes

# Exit without initializing
cool-code --no-init

# Reduce UI output for automation
cool-code --quiet

# Allow dangerous actions without prompts
cool-code --allow-dangerous

# Copy final response to clipboard
cool-code --copy
```

### Agent Modes

Cool-Code supports three distinct modes of operation:

1. **Planning Mode**: Analyzes your request and generates a Task List (TODOs) without touching your code. Useful for architecting changes before execution.
2. **Agent Mode**: The default mode. Executes tasks autonomously, applying code changes and running commands.
3. **Ask Mode**: Read-only mode for questions and explanations. Mutating tools (like shell or write_file) are blocked.

Switch modes mid-session using the `:mode` command:
- `:mode planning`
- `:mode agent`
- `:mode ask`

### Advanced CLI Commands

```bash
# Project summary and framework detection
cool-code scan

# Force refresh the scan cache
cool-code scan --refresh

# JSON output
cool-code scan --json

# Task planning
cool-code task "Add user authentication and password reset"

# Config management
cool-code config get llm.model
cool-code config set llm.model "gemini-2.5-flash"
cool-code config set llm.maxTokens 2048
```

### Interactive Commands

Inside the CLI prompt:

- `:help` Show interactive commands
- `:mode` Show or switch current mode
- `:context` Preview the prompt context
- `:clear` Clear the screen
- `:exit` or `:quit` Exit

### Non-blocking Input

You can type and send new messages even while the agent is processing tool calls. These messages will be queued and processed immediately after the current step is completed, allowing you to pivot or provide feedback mid-task.

## Architecture

### Core Components

#### 1. Entry Point (`src/index.ts`)
The main entry point that initializes the CLI using Commander.js and starts the interactive session.

#### 2. UI Layer (`src/ui/`)
- **Landing (`landing.ts`)**: Displays a compact, professional welcome screen.
- **Query Handler (`query.ts`)**: Manages the main input loop, status bar, and task rendering.
- **Spinner (`spinner.ts`)**: Provides visual feedback with synchronized status messages.

#### 3. Core Engine (`src/core/`)
- **Processor (`processor.ts`)**: Main orchestrator handling query turns, tool execution, and message queuing.
- **LLM (`llm.ts`)**: Manages communication with Google's Gemini models.
- **Context Manager (`contextManager.ts`)**: Maintains project state, file trees, and mode-specific prompt building.
- **Prompts (`prompts.ts`)**: Contains high-level system logic, mode definitions, and tool instructions.

#### 4. Tool System (`src/core/tools/`)
- **File Operations**: `readFileTool`, `editTool`, `newFileTool`, `deleteFileTool`.
- **Search & Discovery**: `globTool`, `grepTool`.
- **System Operations**: `shellTool`, `ignoreGitIgnoreFileTool`.

## Configuration

### Environment Variables
- `GOOGLE_GENERATIVE_AI_API_KEY`: Your Gemini API key.

### Project Config (.coolcode.json)

```json
{
  "llm": {
    "model": "gemini-2.5-flash",
    "temperature": 0.2,
    "maxTokens": 2048
  },
  "features": {
    "scanCache": true,
    "fileTreeMaxDepth": 4,
    "allowDangerous": false,
    "confirmEdits": false
  },
  "guardrails": {
    "blockReadPatterns": [
      ".env",
      "*.pem",
      "*.key",
      "id_rsa"
    ]
  }
}
```

## Safety Features

- **Path Validation**: Operations are restricted to the project root.
- **Git Integration**: Automatically respects .gitignore patterns.
- **Dangerous Action Prompts**: Requires explicit confirmation for potentially destructive shell commands.
- **Edit Confirmations**: Configurable prompts for code edits.
- **Ask Mode**: Strict read-only enforcement for safer exploration.

## Project Structure

```
cool-code/
├── src/
│   ├── core/                 # Core engine
│   │   ├── tools/           # Tool implementations
│   │   ├── utils/           # Utility functions
│   │   ├── contextManager.ts
│   │   ├── llm.ts
│   │   ├── processor.ts
│   │   └── prompts.ts
│   ├── types/               # TypeScript definitions
│   ├── ui/                  # UI components
│   └── index.ts            # Entry point
├── dist/                   # Compiled output
├── package.json
└── README.md
```
