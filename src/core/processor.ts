import { StreamingSpinner } from '../ui/spinner';
import { LLM } from './llm';
import { ContextManager } from './contextManager';
import { createGitIgnoreChecker } from './tools/ignoreGitIgnoreFileTool';
import { validateAndRunToolCall } from './tools/toolValidator';
import chalk from 'chalk';
import type { CoolCodeConfig } from './config';
import type { LLMConfig, TaskList, AgentMode } from '../types';
import { renderMarkdown } from '../ui/utils/markdown';
export interface QueryResult {
  query: string;
  response: string;
  suggestions?: string[];
  timestamp: Date;
}

export interface configType {
  LLMConfig: LLMConfig;
  rootDir: string;
  doesExistInGitIgnore: (rootDir: string) => boolean | null;
}

export type SpinnerUpdateCallback = (text: string) => void;

interface ProcessorOptions {
  quiet?: boolean;
  allowDangerous?: boolean;
  mode?: AgentMode;
  confirm?: (message: string) => Promise<boolean>;
  confirmEdit?: (message: string, preview?: string) => Promise<boolean>;
}

const THINKING_MESSAGES = [
  'Thinking out loud...',
  'Crunching the numbers...',
  'Consulting the oracle...',
  'Scanning the matrix...',
  'Analyzing the flux capacitor...',
  'Optimizing neural pathways...',
  'Reading between the lines...',
  'Checking the crystal ball...',
  'Distilling digital wisdom...',
  'Searching the knowledge graph...',
  'Synthesizing a solution...',
  'Debugging the universe...',
];

function getRandomThinkingMessage() {
  return THINKING_MESSAGES[Math.floor(Math.random() * THINKING_MESSAGES.length)];
}

export class Processor {
  public config: configType;
  private LLM: LLM;
  private contextManager: ContextManager;
  private options: ProcessorOptions;
  private allowDangerous: boolean;
  private confirmEdits: boolean;
  private taskList: TaskList | null = null;
  private messageQueue: string[] = [];
  private mode: AgentMode = 'agent';

  constructor(
    rootDir: string,
    config?: CoolCodeConfig,
    options: ProcessorOptions = {}
  ) {
    this.options = options;
    this.allowDangerous =
      options.allowDangerous ?? config?.features?.allowDangerous ?? false;
    this.confirmEdits = config?.features?.confirmEdits ?? false;
    this.config = {
      LLMConfig: {
        model: config?.llm?.model ?? 'gemini-2.5-flash',
        temperature: config?.llm?.temperature,
        maxTokens: config?.llm?.maxTokens,
      },
      rootDir,
      doesExistInGitIgnore: createGitIgnoreChecker(rootDir),
    };
    this.LLM = new LLM(this.config.LLMConfig);
    this.contextManager = new ContextManager(
      rootDir,
      this.config.doesExistInGitIgnore,
      config
    );
    this.contextManager.setMaxTokens(config?.features?.maxContextTokens ?? 20000);
    this.mode = options.mode ?? 'agent';
  }

  isFinalMessage(response: any): boolean {
    return (
      !Array.isArray(response) &&
      typeof response === 'object' &&
      'text' in response
    );
  }

  async processQuery(query: string): Promise<string | null> {
    this.contextManager.addUserMessage(query);

    const streamingSpinner = new StreamingSpinner(!this.options.quiet);
    if (!this.options.quiet) {
      streamingSpinner.start(getRandomThinkingMessage());
    }

    let finalText: string | null = null;

    while (true) {
      const prompt = this.contextManager.buildPrompt();
      const response = await this.LLM.StreamResponse(prompt, () => {});
      // console.log('[RAW RESPONSE]', response);
      let toolCalls: any;
      try {
        const cleanResponse = extractJsonFromResponse(response);
        toolCalls = JSON.parse(cleanResponse);

        // Handle mixed responses (object with text and tools) or simple final messages
        if (typeof toolCalls === 'object' && !Array.isArray(toolCalls)) {
          if ('text' in toolCalls) {
            const formattedText = renderMarkdown(toolCalls.text);
            if (this.options.quiet) {
              process.stdout.write(`${formattedText}\n`);
            } else {
              streamingSpinner.succeed(formattedText);
            }
            finalText = toolCalls.text;
            
            // If it has NO tool_calls, or tool_calls is empty, we are done
            if (!toolCalls.tool_calls || !Array.isArray(toolCalls.tool_calls) || toolCalls.tool_calls.length === 0) {
              break; 
            }
            // Otherwise, continue to process tool_calls
            toolCalls = toolCalls.tool_calls;
          }
        }

        if (typeof toolCalls === 'string') {
          toolCalls = JSON.parse(toolCalls);
        }
      } catch (error) {
        if (!this.options.quiet) {
          console.error(chalk.red('\n[PARSE ERROR]'), error instanceof Error ? error.message : String(error));
          // console.log('[RAW RESPONSE WAS]', response);
        }
        streamingSpinner.succeed('Processing turn complete.');
        break;
      }

      // Handle Task List updates from the model (internal)
      if (Array.isArray(toolCalls)) {
        const taskTool = toolCalls.find(t => t.tool === 'update_task_list');
        if (taskTool) {
          this.taskList = taskTool.toolOptions;
          // Filter out the internal task tool so it doesn't get processed by validateAndRunToolCall
          toolCalls = toolCalls.filter(t => t.tool !== 'update_task_list');
        }
      }

      for (const toolCall of toolCalls) {
        // console.log('[DEBUG] toolCall:', toolCall, typeof toolCall);
        if (toolCall && typeof toolCall === 'object' && 'tool' in toolCall) {

          // ASK mode: Block all mutating tools
          if (this.mode === 'ask') {
            const mutatingTools = ['edit_file', 'new_file', 'shell_command', 'rename_file', 'replace_in_files', 'delete_file'];
            if (mutatingTools.includes(toolCall.tool)) {
              this.contextManager.addResponse(
                `[ASK MODE] Cannot execute tool '${toolCall.tool}' in Ask mode. Only reading and answering questions is allowed.`,
                toolCall
              );
              continue;
            }
          }

          if (!this.allowDangerous) {
            const dangerReason = getDangerReason(toolCall);
            const editPreview = buildEditPreview(toolCall);
            if (dangerReason || editPreview) {
              if (!this.options.quiet) {
                streamingSpinner.stop();
              }
              if (dangerReason) {
                const confirmDanger = await this.confirmIfDangerous(
                  toolCall,
                  dangerReason
                );
                if (!confirmDanger) {
                  const message =
                    'User declined to run a potentially dangerous tool.';
                  this.contextManager.addResponse(message, toolCall);
                  continue;
                }
              }
              if (editPreview && this.confirmEdits) {
                const confirmEdit = await this.confirmIfEdit(
                  toolCall,
                  editPreview
                );
                if (!confirmEdit) {
                  const message = 'User declined to apply edits.';
                  this.contextManager.addResponse(message, toolCall);
                  continue;
                }
              }
              if (!this.options.quiet) {
                streamingSpinner.start(getRandomThinkingMessage());
              }
            }
          }
          if (!this.options.quiet) {
            streamingSpinner.updateText(
              toolCall.description || 'Thinking out loud...'
            );
          }
          try {
            // console.log(toolCall);
            const result = await validateAndRunToolCall(
              toolCall,
              this.config,
              this.config.rootDir
            );
            this.contextManager.addResponse(
              result.result?.LLMresult as string,
              toolCall
            );
          } catch (err) {
            if (!this.options.quiet) {
              streamingSpinner.updateText(
                '[AGENT ERROR] ' +
                  (err instanceof Error ? err.message : String(err))
              );
            }
            console.error('[AGENT ERROR]', err);
          }
        }

        // After each tool call, check if there are background messages enqueued
        if (this.messageQueue.length > 0) {
          const combined = this.messageQueue.join('\n');
          this.messageQueue = [];
          this.contextManager.addUserMessage(
            `[SYSTEM: USER INTERRUPTED WITH NEW MESSAGE]\n${combined}`
          );
          // break out of the current toolCalls loop to let the LLM see the new message in the next turn
          break; 
        }
      }
    }
    // Perform periodic cleanup/summarization after the response is complete
    await this.summarizeContext();
    return finalText;
  }

  private async confirmIfDangerous(
    toolCall: any,
    dangerReason: string
  ): Promise<boolean> {
    if (this.options.quiet || !this.options.confirm) {
      return false;
    }
    const prompt = `Allow potentially dangerous action (${dangerReason})?`;
    return this.options.confirm(prompt);
  }

  private async confirmIfEdit(
    toolCall: any,
    preview: string
  ): Promise<boolean> {
    if (this.options.quiet || !this.options.confirmEdit) {
      return false;
    }
    return this.options.confirmEdit('Apply this edit?', preview);
  }

  getStatus() {
    const stats = this.contextManager.getStats();
    return {
      model: this.config.LLMConfig.model,
      ...stats,
      mode: this.mode,
    };
  }

  getContextPreview() {
    const prompt = this.contextManager.buildPrompt();
    return {
      prompt,
      stats: this.contextManager.getStats(),
    };
  }

  public enqueueMessage(message: string) {
    this.messageQueue.push(message);
  }

  public setTaskList(taskList: TaskList) {
    this.taskList = taskList;
  }

  public getTaskList() {
    return this.taskList;
  }

  public getMode() {
    return this.mode;
  }

  public setMode(mode: AgentMode) {
    this.mode = mode;
    this.contextManager.setMode(mode);
  }

  public pinFile(filePath: string) {
    this.contextManager.pinFile(filePath);
  }

  public unpinFile(filePath: string) {
    this.contextManager.unpinFile(filePath);
  }

  public getPinnedFiles(): string[] {
    return this.contextManager.getPinnedFiles();
  }

  private async summarizeContext() {
    const stats = this.contextManager.getStats();
    // Only summarize if context is significantly large
    if (stats.messageCount < 20) return;

    const prompt = `Please summarize the key technical objectives and progress made in the following conversation so far in under 200 words. Focus on specific code changes and design decisions. Avoid introductory filler.\n\n${this.contextManager.buildPrompt()}`;
    
    try {
      const summary = await this.LLM.StreamResponse(prompt, () => {});
      if (summary) {
        this.contextManager.setSummary(summary);
      }
    } catch (err) {
      console.error('[SUMMARIZATION ERROR]', err);
    }
  }
}

function getDangerReason(toolCall: any): string | null {
  if (!toolCall || typeof toolCall !== 'object') return null;
  if (toolCall.tool === 'shell_command') {
    const cmd = String(toolCall.toolOptions?.command ?? '');
    const risky = [
      /\brm\b.*(-rf|-fr)\b/i,
      /\bsudo\b/i,
      /\bmkfs\b/i,
      /\bdd\b/i,
      /\bshutdown\b|\breboot\b|\bpoweroff\b/i,
      /\bkill\s+-9\b/i,
      /\bgit\s+reset\s+--hard\b/i,
      /\bgit\s+clean\b/i,
      /\bgit\s+push\b.*--force\b/i,
      /\bcurl\b.*\s*\|\s*(bash|sh)\b/i,
      /\bwget\b.*\s*\|\s*(bash|sh)\b/i,
    ];
    if (risky.some((r) => r.test(cmd))) {
      return 'shell command';
    }
  }
  if (toolCall.tool === 'replace_in_files') {
    const dryRun = toolCall.toolOptions?.dryRun;
    if (dryRun === false) {
      return 'bulk replace (write)';
    }
  }
  if (toolCall.tool === 'rename_file') {
    if (toolCall.toolOptions?.overwrite) {
      return 'file overwrite rename';
    }
  }
  return null;
}

function buildEditPreview(toolCall: any): string | null {
  if (!toolCall || typeof toolCall !== 'object') return null;
  if (toolCall.tool === 'edit_file') {
    return [
      `File: ${toolCall.toolOptions?.filePath ?? ''}`,
      '--- old ---',
      truncateText(String(toolCall.toolOptions?.oldString ?? ''), 400),
      '--- new ---',
      truncateText(String(toolCall.toolOptions?.newString ?? ''), 400),
    ].join('\n');
  }
  if (toolCall.tool === 'new_file') {
    return [
      `File: ${toolCall.toolOptions?.filePath ?? ''}`,
      '--- content ---',
      truncateText(String(toolCall.toolOptions?.content ?? ''), 400),
    ].join('\n');
  }
  if (toolCall.tool === 'replace_in_files') {
    if (toolCall.toolOptions?.dryRun === false) {
      return [
        `Pattern: ${toolCall.toolOptions?.pattern ?? ''}`,
        `Replacement: ${toolCall.toolOptions?.replacement ?? ''}`,
        `Include: ${toolCall.toolOptions?.include ?? 'all'}`,
        `Exclude: ${toolCall.toolOptions?.exclude ?? 'none'}`,
        `Regex: ${toolCall.toolOptions?.useRegex ? 'yes' : 'no'}`,
      ].join('\n');
    }
  }
  return null;
}

function estimateTokens(text: string): number {
  return Math.ceil(text.length / 4);
}

function extractJsonFromResponse(response: string): string {
  const trimmed = response.trim();
  if (!trimmed) {
    return JSON.stringify({ text: 'I apologize, but I encountered an issue generating a response. Please try again or rephrase your request.' });
  }

  // 1. Try to find fenced code blocks
  const codeBlockRegex = /```(?:json)?\s*([\s\S]*?)\s*```/i;
  const match = trimmed.match(codeBlockRegex);
  if (match) {
    const extracted = match[1].trim();
    if (isLikelyJson(extracted)) return extracted;
  }

  // 2. Try to find the start of an array or object
  // Find the first [ or {
  const firstOpen = trimmed.search(/[\[\{]/);
  if (firstOpen !== -1) {
    const lastClose = Math.max(trimmed.lastIndexOf(']'), trimmed.lastIndexOf('}'));
    if (lastClose > firstOpen) {
      const extracted = trimmed.substring(firstOpen, lastClose + 1).trim();
      if (isLikelyJson(extracted)) return extracted;
    }
  }

  // 3. Last fallback: if it's not JSON, wrap it as text
  if (isLikelyJson(trimmed)) return trimmed;
  
  return JSON.stringify({ 
    text: trimmed,
    tool_calls: [] 
  });
}

function isLikelyJson(text: string): boolean {
  const t = text.trim();
  return (t.startsWith('{') && t.endsWith('}')) || (t.startsWith('[') && t.endsWith(']'));
}

function truncateText(text: string, max: number) {
  if (text.length <= max) return text;
  return text.slice(0, max) + '\n... (truncated)';
}
