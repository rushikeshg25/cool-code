import { LanguageModelV1, streamText } from "ai";
import { google } from "@ai-sdk/google";
import { openai } from "@ai-sdk/openai";
import { anthropic } from "@ai-sdk/anthropic";
import chalk from "chalk";
import type { LLMConfig, LLMProvider } from "../types";

interface ProviderInfo {
  envKey: string;
  keyUrl: string;
  build: (model: string) => LanguageModelV1;
}

const PROVIDERS: Record<LLMProvider, ProviderInfo> = {
  google: {
    envKey: "GOOGLE_GENERATIVE_AI_API_KEY",
    keyUrl: "https://aistudio.google.com/app/apikey",
    build: (model) => google(`models/${model}`),
  },
  openai: {
    envKey: "OPENAI_API_KEY",
    keyUrl: "https://platform.openai.com/api-keys",
    build: (model) => openai(model),
  },
  anthropic: {
    envKey: "ANTHROPIC_API_KEY",
    keyUrl: "https://console.anthropic.com/settings/keys",
    build: (model) => anthropic(model),
  },
};

// Infers the provider from a model id when not set explicitly, so users can
// just set `llm.model` and supply the matching API key.
export function inferProvider(model: string): LLMProvider {
  const m = model.toLowerCase();
  if (m.includes("gpt") || m.startsWith("o1") || m.startsWith("o3")) return "openai";
  if (m.includes("claude")) return "anthropic";
  return "google";
}

export function resolveProvider(config: LLMConfig): LLMProvider {
  return config.provider ?? inferProvider(config.model);
}

export class LLM {
  private model: LanguageModelV1;
  private config: LLMConfig;
  private provider: LLMProvider;

  constructor(config: LLMConfig) {
    this.config = config;
    this.provider = resolveProvider(config);
    const info = PROVIDERS[this.provider];

    if (!process.env[info.envKey]) {
      console.error(chalk.red.bold(`Missing API key for ${this.provider}`));
      console.error("");
      console.error(chalk.dim(`Set your ${this.provider} API key:`));
      console.error(chalk.cyan(`  export ${info.envKey}=your_api_key_here`));
      console.error("");
      console.error(chalk.dim(`Get a key at ${info.keyUrl}`));
      process.exit(1);
    }

    this.model = info.build(config.model);
  }

  async StreamResponse(prompt: string, onChunk?: (chunk: string) => void) {
    try {
      const { textStream } = await streamText({
        model: this.model,
        prompt: prompt,
        temperature: this.config.temperature,
        maxTokens: this.config.maxTokens,
      });

      let fullResponse = "";
      for await (const textPart of textStream) {
        fullResponse += textPart;
        if (onChunk) {
          onChunk(textPart);
        } else {
          process.stdout.write(textPart);
        }
      }

      return fullResponse;
    } catch (error: any) {
      if (
        error.message?.includes("API key") ||
        error.message?.includes("authentication")
      ) {
        const info = PROVIDERS[this.provider];
        console.error(chalk.red.bold("\nAuthentication error"));
        console.error(
          chalk.dim(`Check that your ${this.provider} API key is valid.`)
        );
        console.error(chalk.dim(`Get a key at ${info.keyUrl}`));
        process.exit(1);
      }
      throw error;
    }
  }
}
