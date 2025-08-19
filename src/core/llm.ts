import { LanguageModelV1, streamText } from "ai";
import { google } from "@ai-sdk/google";
import chalk from "chalk";

export class LLM {
  private model: LanguageModelV1;

  constructor(model: string) {
    // Validate API key before initializing the model
    if (!process.env.GOOGLE_GENERATIVE_AI_API_KEY) {
      console.error(chalk.red("❌ Missing API Key!"));
      console.error("");
      console.error(chalk.yellow("Please set your Google AI API key:"));
      console.error("");
      console.error(
        chalk.cyan("export GOOGLE_GENERATIVE_AI_API_KEY=your_api_key_here")
      );
      console.error("");
      console.error(
        chalk.blue(
          "Get your API key at: https://aistudio.google.com/app/apikey"
        )
      );
      process.exit(1);
    }

    this.model = google(`models/${model}`);
  }

  async StreamResponse(prompt: string, onChunk?: (chunk: string) => void) {
    try {
      const { textStream } = await streamText({
        model: this.model,
        prompt: prompt,
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
        console.error(chalk.red("\n❌ Authentication Error!"));
        console.error(
          chalk.yellow("Please check your Google AI API key is valid.")
        );
        console.error(
          chalk.blue(
            "Get your API key at: https://aistudio.google.com/app/apikey"
          )
        );
        process.exit(1);
      }
      throw error;
    }
  }
}
