import { generateText } from 'ai';
import { google } from '@ai-sdk/google';

export async function generateTexta() {
  const { text } = await generateText({
    model: google('models/gemini-2.5-flash'),
    prompt: '',
  });
  console.log(text);
}
