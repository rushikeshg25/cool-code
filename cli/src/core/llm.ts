import { LanguageModelV1, streamText } from "ai";
import { google } from "@ai-sdk/google";

export class LLM {
  private model: LanguageModelV1;

  constructor(model: string) {
    this.model = google(`models/${model}`);
  }

  async StreamResponse(prompt: string) {
    const { textStream } = await streamText({
      model: this.model,
      prompt: prompt,
    });
    for await (const textPart of textStream) {
      console.log(textPart);
    }
  }
}
