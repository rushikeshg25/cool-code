# Cool-Code:Cli coding Ai agent

An intelligent command-line interface that leverages AI to help you interact with codebases and work on them similar to Gemini cli and Claude code.

## 🚀 Overview

Cool-Code is a powerful tool that combines the capabilities of large language models with a comprehensive set of development tools to provide an interactive database and code management experience. Simply describe what you want to accomplish, and the AI agent will understand your intent and execute the necessary operations.

Demo of it spinning up a Node Express server with Prisma:

https://github.com/user-attachments/assets/b1b59602-f118-4bbe-8b38-e4be0f39119f

## ✨ Features

- **Natural Language Processing**: Interact with your codebase using plain English
- **Intelligent Code Analysis**: Understands your project structure and coding patterns. No vector Dbs needed.
- **File Operations**: Read, create, edit and search files with AI assistance
- **Shell Command Execution**: Run system commands through the AI agent
- **Real-time Streaming**: Get live feedback as the AI processes your requests
- **Context-Aware**: Maintains conversation history and project context.

## 🛠 Setup Guide

### Prerequisites

- Node.js (v16 or higher)
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

## 🎯 Usage

### Starting the CLI

Navigate to your project's dir

```bash
cool-code
```

## 🏗 Architecture

### Core Components

#### 1. Entry Point (`src/index.ts`)

The main entry point that initializes the CLI using Commander.js and starts the interactive session.

#### 2. UI Layer (`src/ui/`)

- **Landing (`landing.ts`)**: Displays the welcome screen with ASCII art
- **Query Handler (`query.ts`)**: Manages user input and query processing
- **Spinner (`spinner.ts`)**: Provides visual feedback during processing

#### 3. Core Engine (`src/core/`)

- **Processor (`processor.ts`)**: Main orchestrator that handles query processing
- **LLM (`llm.ts`)**: Manages communication with Google's Gemini AI model
- **Context Manager (`contextManager.ts`)**: Maintains conversation history and project state
- **Prompts (`prompts.ts`)**: Contains system prompts and examples for the AI

#### 4. Tool System (`src/core/tools/`)

A comprehensive set of tools that the AI can use:

- **File Operations**:
  - `readFileTool.ts`: Read file contents
  - `editTool.ts`: Edit existing files
  - `newFileTool.ts`: Create new files

- **Search & Discovery**:
  - `globTool.ts`: Find files using glob patterns
  - `grepTool.ts`: Search file contents using regex

- **System Operations**:
  - `shellTool.ts`: Execute shell commands
  - `ignoreGitIgnoreFileTool.ts`: Handle .gitignore patterns

- **Tool Management**:
  - `tool-registery.ts`: Registry of all available tools
  - `toolValidator.ts`: Validates and executes tool calls

## 🔄 How It Works

### 1. Initialization

```
User starts CLI → Landing screen → Query input prompt
```

### 2. Query Processing Flow

```
User Query → Context Manager → LLM Processing → Tool Selection → Tool Execution → Response
```

### 3. Detailed Flow

1. **User Input**: User enters a natural language query
2. **Context Building**: The Context Manager builds a comprehensive prompt including:
   - System instructions
   - Project file structure
   - Available tools
   - Conversation history
3. **AI Processing**: The LLM (Gemini) processes the prompt and decides which tools to use
4. **Tool Execution**: Selected tools are validated and executed in sequence
5. **Response Generation**: Results are formatted and presented to the user
6. **Context Update**: The conversation history is updated for future queries

### 4. Tool Selection Process

The AI uses a sophisticated tool selection system:

```typescript
// Example tool call structure
[
  {
    tool: 'read_file',
    description: 'Reading current server configuration',
    toolOptions: {
      absolutePath: '/src/server.js',
    },
  },
  {
    tool: 'edit_file',
    description: 'Adding new middleware',
    toolOptions: {
      filePath: '/src/server.js',
      oldString: "const express = require('express');\nconst app = express();",
      newString:
        "const express = require('express');\nconst auth = require('./auth');\nconst app = express();",
    },
  },
];
```

## 🔧 Configuration

### Environment Variables

- `GOOGLE_GENERATIVE_AI_API_KEY`: Your Google AI API key for Gemini

### LLM Configuration

The system uses Google's Gemini 2.5 Flash model by default. You can modify the configuration in `src/core/processor.ts`:

```typescript
this.config = {
  LLMConfig: {
    model: 'gemini-2.5-flash',
  },
  // ... other config
};
```

## 🛡 Safety Features

- **Path Validation**: All file operations use absolute paths to prevent directory traversal
- **Git Integration**: Respects .gitignore patterns to avoid sensitive files
- **Tool Validation**: All tool calls are validated before execution
- **Context Limits**: Conversation history is limited to prevent token overflow
- **Error Handling**: Comprehensive error handling with user-friendly messages

## 📁 Project Structure

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
│   ├── types/               # TypeScript type definitions
│   ├── ui/                  # User interface components
│   └── index.ts            # Entry point
├── dist/                   # Compiled output
├── package.json
├── tsconfig.json
└── README.md
```
