import { getFolderStructure } from './utils';

export interface Message {
  role: 'user' | 'model';
  content: string;
  metadata?: {
    toolCalls?: ToolCall[];
  };
}

interface ProjectStateType {
  rootDir: string;
  cwd: string;
  fileTree: string;
}

interface ToolCall {
  tool: string;
  parameters: any;
}

export class ContextManager {
  systemPrompt = ``;
  private gitIgnoreChecker: (a: string) => boolean | null;
  private conversations: Message[];
  private projectState: ProjectStateType;
  private currentQuery: string;
  constructor(cwd: string, gitIgnoreChecker: (a: string) => boolean | null) {
    this.projectState = {
      rootDir: cwd,
      cwd,
      fileTree: getFolderStructure({
        gitIgnoreChecker,
        rootDir: cwd,
      }),
    };
    this.conversations = [];
    this.currentQuery = '';
    this.gitIgnoreChecker = gitIgnoreChecker;
  }

  addResponse(message: Message) {
    this.conversations.push(message);
  }

  buildPrompt(): string {
    const sections = [];
    sections.push(this.buildSystemSection());
    sections.push(this.buildProjectStateSection());
    sections.push(this.buildConversationSection());
    return sections.filter(Boolean).join('\n\n');
  }

  private buildSystemSection(): string {
    return this.systemPrompt;
  }

  private buildProjectStateSection(): string {
    const projectParts = [];
    projectParts.push(`CWD: ${this.projectState.cwd}`);
    projectParts.push(`File Tree: ${this.projectState.fileTree}`);
    return projectParts.join('\n');
  }

  private buildConversationSection(): string {
    const recentMessages = this.conversations.slice(-10);

    let section = '--- Recent Conversation ---\n';
    for (const msg of recentMessages) {
      section += `${msg.role}: ${msg.content}\n`;

      if (msg.metadata?.toolCalls) {
        for (const tool of msg.metadata.toolCalls.slice(-2)) {
          section += `Tool: ${tool}\n`;
        }
      }
    }
    return section;
  }

  updateProjectCWD(cwd: string) {
    this.projectState.cwd = cwd;
  }

  updateCurrentQuery(query: string) {
    this.conversations.push({
      role: 'user',
      content: query,
    });
    this.currentQuery = query;
  }

  async updateProjectStateTree() {
    this.projectState.fileTree = await getFolderStructure({
      gitIgnoreChecker: this.gitIgnoreChecker,
      rootDir: this.projectState.rootDir,
    });
  }
}
