import { LanguageModelV1, streamText, generateText } from 'ai';
import { google } from '@ai-sdk/google';
import dotenv from 'dotenv';

export class LLM {
  private model: LanguageModelV1;

  constructor(model: string) {
    this.model = google(`models/${model}`);
    console.log(process.env.GOOGLE_GENERATIVE_AI_API_KEY);
  }

  async generateTextAns(prompt: string) {
    const { text } = await generateText({
      model: this.model,
      prompt,
    });
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
