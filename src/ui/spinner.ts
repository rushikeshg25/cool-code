import chalk from 'chalk';

export class StreamingSpinner {
  private enabled: boolean;
  private interval: NodeJS.Timeout | null = null;
  private spinnerChars = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
  private colors = [chalk.cyan, chalk.blue, chalk.magenta, chalk.green];
  private currentFrame = 0;
  private statusText = '';
  private isActive = false;

  constructor(enabled: boolean = true) {
    this.enabled = enabled;
  }

  start(initialText: string = 'Processing...') {
    if (!this.enabled) return;
    this.statusText = initialText;
    this.isActive = true;
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
    const color = this.colors[this.currentFrame % this.colors.length];
    process.stdout.write(
      color(`${this.spinnerChars[this.currentFrame]} `) + chalk.gray(this.statusText)
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
    console.log(chalk.green(`✅ ${text || 'Success!'}`));
  }

  fail(text?: string) {
    if (!this.enabled) return;
    this.stop();
    console.log(chalk.red(`❌ ${text || 'Failed!'}`));
  }
}
