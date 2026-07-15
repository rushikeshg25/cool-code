import chalk from 'chalk';
import { c, glyph } from './theme';

// Output surface the processor drives during a turn. Implemented by the console
// StreamingSpinner (default / CLI / quiet) and by the Ink app at runtime.
export interface StatusReporter {
  start(initialText?: string): void;
  updateText(text: string): void;
  succeed(text?: string): void;
  fail(text?: string): void;
  stop(): void;
}

export class StreamingSpinner implements StatusReporter {
  private enabled: boolean;
  private interval: NodeJS.Timeout | null = null;
  private spinnerChars = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
  private currentFrame = 0;
  private statusText = '';
  private isActive = false;
  private startedAt = 0;

  constructor(enabled: boolean = true) {
    this.enabled = enabled;
  }

  start(initialText: string = 'Working…') {
    if (!this.enabled) return;
    this.statusText = initialText;
    this.isActive = true;
    this.startedAt = Date.now();
    this.showSpinner();

    this.interval = setInterval(() => {
      if (this.isActive) {
        this.currentFrame = (this.currentFrame + 1) % this.spinnerChars.length;
        this.showSpinner();
      }
    }, 80);
  }

  updateText(text: string) {
    if (!this.enabled) return;
    this.statusText = text;
    if (this.isActive) {
      this.showSpinner();
    }
  }

  private showSpinner() {
    if (!this.enabled) return;
    // Clear current line and show spinner
    process.stdout.write('\r\x1b[K');
    const elapsed = Math.floor((Date.now() - this.startedAt) / 1000);
    process.stdout.write(
      c.accent(`${this.spinnerChars[this.currentFrame]} `) +
        chalk.dim(this.statusText) +
        chalk.dim(elapsed > 0 ? ` (${elapsed}s)` : '')
    );
  }

  stop() {
    if (!this.enabled) return;
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
    }
    this.isActive = false;
    // Clear the spinner line
    process.stdout.write('\r\x1b[K');
  }

  succeed(text?: string) {
    if (!this.enabled) return;
    this.stop();
    console.log(c.success(`${glyph.ok} ${text || 'Done'}`));
  }

  fail(text?: string) {
    if (!this.enabled) return;
    this.stop();
    console.log(c.error(`${glyph.fail} ${text || 'Failed'}`));
  }
}
