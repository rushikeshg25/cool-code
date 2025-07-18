import { LanguageModelV1, streamText } from "ai";
import { google } from "@ai-sdk/google";

export class LLM {
  private model: LanguageModelV1;

  constructor(model: string) {
    this.model = google(`models/${model}`);
  }

  async StreamResponse(prompt: string, onChunk?: (chunk: string) => void) {
    const { textStream } = await streamText({
      model: this.model,
      prompt: prompt,
    });
    
    let fullResponse = '';
    for await (const textPart of textStream) {
      fullResponse += textPart;
      if (onChunk) {
        onChunk(textPart);
      } else {
        process.stdout.write(textPart);
      }
    }
    
    return fullResponse;
  }
}
