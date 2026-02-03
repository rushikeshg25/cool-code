export const toolRegistery = [
  {
    name: 'edit_file',
    description: `Replaces text within a file. By default, replaces a single occurrence, but can replace multiple occurrences when \`expected_replacements\` is specified. This tool requires providing significant context around the change to ensure precise targeting. Always use the read_file tool to examine the file's current content before attempting a text replacement.

    The user has the ability to modify the \`newString\` content. If modified, this will be stated in the response.

Expectation for required toolOptions:
1. \`filePath\` MUST be an absolute path; otherwise an error will be thrown.
2. \`oldString\` MUST be the exact literal text to replace (including all whitespace, indentation, newlines, and surrounding code etc.).
3. \`newString\` MUST be the exact literal text to replace \`oldString\` with (also including all whitespace, indentation, newlines, and surrounding code etc.). Ensure the resulting code is correct and idiomatic.
4. NEVER escape \`oldString\` or \`newString\`, that would break the exact literal text requirement.
**Important:** If ANY of the above are not satisfied, the tool will fail. CRITICAL for \`oldString\`: Must uniquely identify the single instance to change. Include at least 3 lines of context BEFORE and AFTER the target text, matching whitespace and indentation precisely. If this string matches multiple locations, or does not match exactly, the tool will fail.
**Multiple replacements:** Set \`expected_replacements\` to the number of occurrences you want to replace. The tool will replace ALL occurrences that match \`oldString\` exactly. Ensure the number of replacements matches your expectation.`,
    toolOptions: {
      filePath: {
        description:
          "The absolute path to the file to modify. Must start with '/'.",
        type: String,
      },
      oldString: {
        description:
          'The exact literal text to replace, preferably unescaped. For single replacements (default), include at least 3 lines of context BEFORE and AFTER the target text, matching whitespace and indentation precisely. For multiple replacements, specify expected_replacements tool_Option. If this string is not the exact literal text (i.e. you escaped it) or does not match exactly, the tool will fail.',
        type: String,
      },
      newString: {
        description:
          'The exact literal text to replace `oldString` with, preferably unescaped. Provide the EXACT text. Ensure the resulting code is correct and idiomatic.',
        type: String,
      },
      expected_replacements: {
        type: Number,
        description:
          'Number of replacements expected. Defaults to 1 if not specified. Use when you want to replace multiple occurrences.',
        minimum: 1,
      },
    },
    required: ['filePath', 'oldString', 'newString'],
    type: Object,
  },
  {
    name: 'read_file',
    description:
      'Reads and returns the content of a specified file from the local filesystem. Handles text, images (PNG, JPG, GIF, WEBP, SVG, BMP), and PDF files. For text files, it can read specific line ranges.',
    toolOptions: {
      absolutePath: {
        description:
          "The absolute path to the file to read (e.g., '/home/user/project/file.txt'). Relative paths are not supported. You must provide an absolute path.",
        type: String,
      },
      startLine: {
        description:
          "Optional: For text files, the 0-based line number to start reading from. Requires 'endLine' to be set. Use for paginating through large files.",
        type: Number,
      },
      endLine: {
        description:
          "Optional: For text files, Line number till which to read. Use with 'Endline' to paginate through large files.",
        type: Number,
      },
    },
    required: ['absolutePath'],
    type: Object,
  },
  {
    name: 'shell_command',
    description: `This tool executes a given shell command as \`bash -c <command>\`. Command can start background processes using \`&\`. Command is executed as a subprocess that leads its own process group. Command process group can be terminated as \`kill -- -PGID\` or signaled as \`kill -s SIGNAL -- -PGID\`.

The following information is returned:

Command: Executed command.
Directory: Directory (relative to project root) where command was executed, or \`(root)\`.
Stdout: Output on stdout stream. Can be \`(empty)\` or partial on error and for any unwaited background processes.
Stderr: Output on stderr stream. Can be \`(empty)\` or partial on error and for any unwaited background processes.
Error: Error or \`(none)\` if no error was reported for the subprocess.
Exit Code: Exit code or \`(none)\` if terminated by signal.
Signal: Signal number or \`(none)\` if no signal was received.
Background PIDs: List of background processes started or \`(none)\`.
Process Group PGID: Process group started or \`(none)\``,
    toolOptions: {
      command: {
        type: String,
        description: 'Exact bash command to execute as `bash -c <command>`',
      },
      description: {
        type: String,
        description:
          'Brief description of the command for the user. Be specific and concise. Ideally a single sentence. Can be up to 3 sentences for clarity. No line breaks.',
      },
      directory: {
        type: String,
        description:
          '(OPTIONAL) Directory to run the command in, if not the project root directory. Must be relative to the project root directory and must already exist.',
      },
    },
    required: ['command'],
    type: Object,
  },
  {
    name: 'glob',
    description:
      'Efficiently finds files matching specific glob patterns (e.g., `src/**/*.ts`, `**/*.md`), returning absolute paths sorted by modification time (newest first). Ideal for quickly locating files based on their name or path structure, especially in large codebases.',
    toolOptions: {
      pattern: {
        description:
          "The glob pattern to match against (e.g., '**/*.py', 'docs/*.md').",
        type: String,
      },
    },
    required: ['pattern'],
    type: Object,
  },

  {
    name: 'grep',
    description:
      'Searches for a regular expression pattern within the content of files in a specified directory (or current working directory). Can filter files by a glob pattern. Returns the lines containing matches, along with their file paths and line numbers.',
    toolOptions: {
      pattern: {
        description:
          "The regular expression (regex) pattern to search for within file contents (e.g., 'function\\s+myFunction', 'import\\s+\\{.*\\}\\s+from\\s+.*').",
        type: String,
      },
      path: {
        description:
          'Optional: The absolute path to the directory to search within. If omitted, searches the current working directory.',
        type: String,
      },
      include: {
        description:
          "Optional: A glob pattern to filter which files are searched (e.g., '*.js', '*.{ts,tsx}', 'src/**'). If omitted, searches all files (respecting potential global ignores).",
        type: String,
      },
    },
    required: ['pattern'],
    type: Object,
  },
  {
    name: 'new_file',
    description:
      'Creates a new file with whatever the name specified in the absolute filePath and the content',
    toolOptions: {
      filePath: {
        description: `The Abosolute Path with the filename like if you want to create a file called a.ts in src then it should be /src/a.ts`,
        type: String,
      },
      content: {
        description: `The Content to be stored in this new file. Please format the content`,
        type: String,
      },
    },
    required: ['filePath', 'content'],
    type: Object,
  },
  {
    name: 'list_recent_files',
    description:
      'Lists recently modified files in the project, sorted by most recent first.',
    toolOptions: {
      limit: {
        description: 'Maximum number of files to return. Defaults to 20.',
        type: Number,
      },
      include: {
        description:
          'Optional glob to include files (e.g., "src/**/*.ts"). Defaults to all files.',
        type: String,
      },
      exclude: {
        description:
          'Optional glob to exclude files (e.g., "**/*.test.ts").',
        type: String,
      },
    },
    required: [],
    type: Object,
  },
  {
    name: 'project_summary',
    description:
      'Summarizes the project (entrypoints, frameworks, scripts, languages).',
    toolOptions: {},
    required: [],
    type: Object,
  },
  {
    name: 'open_file_at',
    description:
      'Reads a file or a specific line range. Line numbers are 1-based.',
    toolOptions: {
      absolutePath: {
        description: 'Absolute path to the file to read.',
        type: String,
      },
      startLine: {
        description: 'Start line (1-based). Use with endLine.',
        type: Number,
      },
      endLine: {
        description: 'End line (1-based). Use with startLine.',
        type: Number,
      },
    },
    required: ['absolutePath'],
    type: Object,
  },
  {
    name: 'run_tests',
    description:
      'Runs project tests. Uses package.json test script if command is not provided.',
    toolOptions: {
      command: {
        description: 'Optional command to run tests.',
        type: String,
      },
    },
    required: [],
    type: Object,
  },
  {
    name: 'lint_fix',
    description:
      'Runs lint/format with auto-fix. Uses lint:fix, lint, or format scripts if available.',
    toolOptions: {
      command: {
        description: 'Optional command to run lint/format.',
        type: String,
      },
    },
    required: [],
    type: Object,
  },
  {
    name: 'format_file',
    description: 'Formats a file with prettier.',
    toolOptions: {
      absolutePath: {
        description: 'Absolute path to the file to format.',
        type: String,
      },
    },
    required: ['absolutePath'],
    type: Object,
  },
  {
    name: 'git_status',
    description: 'Shows git status summary.',
    toolOptions: {},
    required: [],
    type: Object,
  },
  {
    name: 'git_diff',
    description:
      'Shows git diff. Optionally specify a file path and/or staged diff.',
    toolOptions: {
      filePath: {
        description: 'Absolute path to file to diff (optional).',
        type: String,
      },
      staged: {
        description: 'If true, show staged diff.',
        type: Boolean,
      },
    },
    required: [],
    type: Object,
  },
  {
    name: 'git_commit',
    description:
      'Stages files and creates a git commit.',
    toolOptions: {
      message: {
        description: 'Commit message.',
        type: String,
      },
      all: {
        description: 'If true, stage all changes.',
        type: Boolean,
      },
      files: {
        description: 'Array of absolute file paths to stage.',
        type: Array,
      },
    },
    required: ['message'],
    type: Object,
  },
  {
    name: 'find_symbol',
    description:
      'Searches for a symbol or pattern using ripgrep.',
    toolOptions: {
      pattern: {
        description: 'Regex or string pattern to search for.',
        type: String,
      },
      include: {
        description: 'Optional glob filter for files.',
        type: String,
      },
      path: {
        description: 'Optional path to search (absolute).',
        type: String,
      },
    },
    required: ['pattern'],
    type: Object,
  },
  {
    name: 'replace_in_files',
    description:
      'Replaces text across files. Supports dry run and regex mode.',
    toolOptions: {
      pattern: {
        description: 'Text or regex pattern to replace.',
        type: String,
      },
      replacement: {
        description: 'Replacement text.',
        type: String,
      },
      include: {
        description: 'Optional glob to include files.',
        type: String,
      },
      exclude: {
        description: 'Optional glob to exclude files.',
        type: String,
      },
      useRegex: {
        description: 'If true, treat pattern as regex.',
        type: Boolean,
      },
      dryRun: {
        description: 'If true, do not write changes.',
        type: Boolean,
      },
    },
    required: ['pattern', 'replacement'],
    type: Object,
  },
  {
    name: 'rename_file',
    description: 'Renames or moves a file.',
    toolOptions: {
      fromPath: {
        description: 'Absolute path to the source file.',
        type: String,
      },
      toPath: {
        description: 'Absolute path to the destination.',
        type: String,
      },
      overwrite: {
        description: 'If true, overwrite destination if it exists.',
        type: Boolean,
      },
    },
    required: ['fromPath', 'toPath'],
    type: Object,
  },
  {
    name: 'new_module',
    description:
      'Scaffolds a new module folder with index export.',
    toolOptions: {
      moduleName: {
        description: 'Module name (folder name).',
        type: String,
      },
      baseDir: {
        description: 'Base directory relative to project root (default: src).',
        type: String,
      },
      exportFromRootIndex: {
        description: 'If true, export from baseDir/index.ts.',
        type: Boolean,
      },
    },
    required: ['moduleName'],
    type: Object,
  },
  {
    name: 'add_script',
    description:
      'Adds a script to package.json.',
    toolOptions: {
      name: {
        description: 'Script name.',
        type: String,
      },
      command: {
        description: 'Script command.',
        type: String,
      },
      overwrite: {
        description: 'If true, overwrite existing script.',
        type: Boolean,
      },
    },
    required: ['name', 'command'],
    type: Object,
  },
  {
    name: 'generate_readme_section',
    description:
      'Appends a section to README.md.',
    toolOptions: {
      title: {
        description: 'Section title.',
        type: String,
      },
      bullets: {
        description: 'Optional bullet list.',
        type: Array,
      },
      content: {
        description: 'Optional raw content.',
        type: String,
      },
    },
    required: ['title'],
    type: Object,
  },
];
