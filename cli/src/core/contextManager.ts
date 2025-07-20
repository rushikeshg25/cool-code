import { BASE_PROMPT, EXAMPLES, TOOL_SELECTION_PROMPT } from "./prompts";
import { toolRegistery } from "./tools/tool-registery";
import { getFolderStructure } from "./utils";

export interface Message {
  role: "user" | "model";
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
  systemPrompt = BASE_PROMPT;
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
    this.currentQuery = "";
    this.gitIgnoreChecker = gitIgnoreChecker;
  }

  addResponse(message: Message) {
    this.conversations.push();
  }

  // Add a user message to the conversation
  addUserMessage(query: string) {
    this.conversations.push({
      role: "user",
      content: query,
    });
    this.currentQuery = query;
  }

  // Add a tool result to the conversation
  addToolResult(toolCall: any, result: any) {
    this.conversations.push({
      role: "model",
      content: result.result?.DisplayResult || result.error || "",
      metadata: {
        toolCalls: [toolCall],
      },
    });
  }

  buildPrompt(): string {
    const sections = [];
    sections.push(this.buildSystemSection());
    sections.push(this.buildProjectStateSection());
    sections.push(this.buildConversationSection());
    sections.push(this.buildToolInfoSection());
    return sections.filter(Boolean).join("\n\n");
  }

  private buildSystemSection(): string {
    return this.systemPrompt;
  }

  private buildProjectStateSection(): string {
    const projectParts = [];
    projectParts.push(`CWD: ${this.projectState.cwd}`);
    projectParts.push(`File Tree: ${this.projectState.fileTree}`);
    return projectParts.join("\n");
  }

  private buildConversationSection(): string {
    const recentMessages = this.conversations.slice(-10);

    let section = "--- Recent Conversation ---\n";
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

  private buildToolInfoSection(): string {
    const toolInfo = toolRegistery.map((tool) => tool.name).join("\n");
    return `These are your Tools and what they expect:\n${toolInfo} here are some examples:\n${EXAMPLES} and the response I am expecting is like \n${TOOL_SELECTION_PROMPT}`;
  }

  updateProjectCWD(cwd: string) {
    this.projectState.cwd = cwd;
  }

  updateCurrentQuery(query: string) {
    this.conversations.push({
      role: "user",
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
