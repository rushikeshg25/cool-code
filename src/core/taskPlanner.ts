import { LLM } from './llm';
import type { CoolCodeConfig } from './config';

export interface TaskPlanStep {
  title: string;
  detail: string;
}

export interface TaskPlan {
  goal: string;
  steps: TaskPlanStep[];
  assumptions: string[];
  risks: string[];
}

const TASK_PROMPT = (goal: string) => `
You are an expert engineering lead. Produce a concise, actionable plan.
Return ONLY valid JSON with this shape:
{
  "goal": string,
  "steps": [{"title": string, "detail": string}],
  "assumptions": string[],
  "risks": string[]
}

Constraints:
- 4 to 8 steps.
- Each detail should be 1-2 sentences.
- Keep assumptions and risks short.

Goal: ${goal}
`;

export async function createTaskPlan(
  goal: string,
  config: CoolCodeConfig
): Promise<TaskPlan | null> {
  const llm = new LLM({
    model: config.llm.model,
    temperature: config.llm.temperature,
    maxTokens: config.llm.maxTokens,
  });
  const response = await llm.StreamResponse(TASK_PROMPT(goal), () => {});
  const parsed = safeParseJson<TaskPlan>(response);
  if (!parsed) return null;
  if (!parsed.goal || !Array.isArray(parsed.steps)) return null;
  return parsed;
}

function safeParseJson<T>(raw: string): T | null {
  try {
    let cleaned = raw.trim();
    if (cleaned.startsWith('```')) {
      cleaned = cleaned.replace(/^```[a-zA-Z]*\n?/, '').replace(/```$/, '');
    }
    return JSON.parse(cleaned) as T;
  } catch {
    return null;
  }
}
