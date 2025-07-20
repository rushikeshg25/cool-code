import { BASE_PROMPT, EXAMPLES, TOOL_SELECTION_PROMPT } from './prompts';
import { toolRegistery } from './tools/tool-registery';
import { getFolderStructure } from './utils';

export interface Message {
  role: 'user' | 'model';
  content: string; //user's query for user and LLMresponse for model with whatever the toolCall is
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
    this.gitIgnoreChecker = gitIgnoreChecker;
  }

  addResponse(LLMresponse: string, toolCall: ToolCall) {
    const message: Message = {
      role: 'model',
      content: `Output of ${JSON.stringify(toolCall)}:\n${LLMresponse}`,
    };
    this.conversations.push(message);
  }

  addUserMessage(query: string) {
    this.conversations.push({
      role: 'user',
      content: query,
    });
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
    }
    return section;
  }

  private buildToolInfoSection(): string {
    const toolInfo = JSON.stringify(toolRegistery);
    return `These are your Tools and what they expect:\n${toolInfo} here are some examples:\n${EXAMPLES} and the response I am expecting is like \n${TOOL_SELECTION_PROMPT}`;
  }

  updateProjectCWD(cwd: string) {
    this.projectState.cwd = cwd;
  }

  async updateProjectStateTree() {
    this.projectState.fileTree = await getFolderStructure({
      gitIgnoreChecker: this.gitIgnoreChecker,
      rootDir: this.projectState.rootDir,
    });
  }
}
