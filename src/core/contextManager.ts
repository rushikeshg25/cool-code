import { BASE_PROMPT, EXAMPLES, TOOL_SELECTION_PROMPT, MODE_PROMPTS } from './prompts';
import { toolRegistery } from './tools/tool-registery';
import { getFolderStructure } from './utils';
import type { CoolCodeConfig } from './config';
import type { AgentMode } from '../types';
import * as fs from 'fs';
import * as path from 'path';

export interface Message {
  role: 'user' | 'model';
  content: string;
  tokens?: number;
}

interface ProjectStateType {
  rootDir: string;
  cwd: string;
  fileTree: string;
}

interface ToolCall {
  tool: string;
  toolOptions: any;
}

export class ContextManager {
  systemPrompt = BASE_PROMPT;
  private gitIgnoreChecker: (a: string) => boolean | null;
  private conversations: Message[];
  private projectState: ProjectStateType;
  private mode: AgentMode = 'agent';
  private summary: string | null = null;
  private maxTokens = 20000; // Default max tokens for the context window
  private pinnedFiles: Set<string> = new Set();

  constructor(
    cwd: string,
    gitIgnoreChecker: (a: string) => boolean | null,
    config?: CoolCodeConfig
  ) {
    this.projectState = {
      rootDir: cwd,
      cwd,
      fileTree: getFolderStructure({
        gitIgnoreChecker,
        rootDir: cwd,
        maxDepth: config?.features?.fileTreeMaxDepth,
      }),
    };
    this.conversations = [];
    this.gitIgnoreChecker = gitIgnoreChecker;
  }

  addResponse(LLMresponse: string, toolCall: ToolCall) {
    const content = `Output of ${JSON.stringify(toolCall)}:\n${LLMresponse}`;
    const message: Message = {
      role: 'model',
      content,
      tokens: estimateTokens(content),
    };
    this.conversations.push(message);
    this.pruneContext();
  }

  addUserMessage(query: string) {
    this.conversations.push({
      role: 'user',
      content: query,
      tokens: estimateTokens(query),
    });
    this.pruneContext();
  }

  buildPrompt(): string {
    const sections = [];
    sections.push(this.buildSystemSection());
    sections.push(this.buildProjectStateSection());
    sections.push(this.buildToolInfoSection());
    sections.push(this.buildConversationSection());
    return sections.filter(Boolean).join('\n\n');
  }

  private buildSystemSection(): string {
    const modePrompt = MODE_PROMPTS[this.mode];
    return `${this.systemPrompt}\n\n${modePrompt}`;
  }

  private buildProjectStateSection(): string {
    const projectParts = [];
    projectParts.push(`CWD: ${this.projectState.cwd}`);
    projectParts.push(`File Tree: ${this.projectState.fileTree}`);
    
    if (this.pinnedFiles.size > 0) {
      projectParts.push('\n--- Pinned Files Context ---');
      for (const filePath of this.pinnedFiles) {
        try {
          if (fs.existsSync(filePath)) {
            const content = fs.readFileSync(filePath, 'utf-8');
            projectParts.push(`File: ${path.relative(this.projectState.rootDir, filePath)}\n\`\`\`\n${content}\n\`\`\``);
          }
        } catch (err) {
          projectParts.push(`File: ${filePath} (Error reading file)`);
        }
      }
    }
    
    return projectParts.join('\n');
  }

  private buildConversationSection(): string {
    let section = '--- Recent Conversation ---\n';
    if (this.summary) {
      section += `[SUMMARY OF PREVIOUS CONTEXT]: ${this.summary}\n\n`;
    }

    // Include as many recent messages as fit in the token budget
    const recentMessages = [...this.conversations].reverse();
    const includedMessages: Message[] = [];
    let currentTokens = 0;

    for (const msg of recentMessages) {
      const msgTokens = msg.tokens || estimateTokens(msg.content);
      if (currentTokens + msgTokens > this.maxTokens) break;
      includedMessages.unshift(msg);
      currentTokens += msgTokens;
    }

    for (const msg of includedMessages) {
      section += `${msg.role}: ${msg.content}\n`;
    }
    return section;
  }

  private pruneContext() {
    // We don't actually delete messages here anymore, 
    // we just let buildConversationSection handle the sliding window.
    // However, if the list gets too huge (e.g. 100+), we might want to trim it.
    if (this.conversations.length > 100) {
      this.conversations = this.conversations.slice(-50);
    }
  }

  public setSummary(summary: string) {
    this.summary = summary;
  }

  public setMaxTokens(max: number) {
    this.maxTokens = max;
  }

  getStats() {
    const totalTokens = this.conversations.reduce((acc, msg) => acc + (msg.tokens || estimateTokens(msg.content)), 0);
    return {
      messageCount: this.conversations.length,
      fileTreeChars: this.projectState.fileTree.length,
      totalTokens,
    };
  }

  private buildToolInfoSection(): string {
    const toolInfo = JSON.stringify(toolRegistery);
    return `These are your Tools and what they expect:\n${toolInfo} here are some examples:\n${EXAMPLES} and the response I am expecting is like \n${TOOL_SELECTION_PROMPT}`;
  }

  updateProjectCWD(cwd: string) {
    this.projectState.cwd = cwd;
  }

  async updateProjectStateTree(config?: CoolCodeConfig) {
    this.projectState.fileTree = await getFolderStructure({
      gitIgnoreChecker: this.gitIgnoreChecker,
      rootDir: this.projectState.rootDir,
      maxDepth: config?.features?.fileTreeMaxDepth,
    });
  }

  setMode(mode: AgentMode) {
    this.mode = mode;
  }

  pinFile(filePath: string) {
    this.pinnedFiles.add(filePath);
  }

  unpinFile(filePath: string) {
    this.pinnedFiles.delete(filePath);
  }

  getPinnedFiles(): string[] {
    return Array.from(this.pinnedFiles);
  }
}

function estimateTokens(text: string): number {
  // Rough estimation: 4 characters per token
  return Math.ceil(text.length / 4);
}
