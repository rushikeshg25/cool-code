import chalk from "chalk";

export class StreamingSpinner {
  private interval: NodeJS.Timeout | null = null;
  private spinnerChars = ["⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"];
  private currentFrame = 0;
  private statusText = "";
  private isActive = false;

  start(initialText: string = "Processing...") {
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
    this.statusText = text;
    if (this.isActive) {
      this.showSpinner();
    }
  }

  private showSpinner() {
    // Clear current line and show spinner
    process.stdout.write("\r\x1b[K"); // Clear line
    process.stdout.write(
      chalk.cyan(`${this.spinnerChars[this.currentFrame]} ${this.statusText}`)
    );
  }

  stop() {
    if (this.interval) {
      clearInterval(this.interval);
      this.interval = null;
    }
    this.isActive = false;
    // Clear the spinner line
    process.stdout.write("\r\x1b[K");
  }

  succeed(text?: string) {
    this.stop();
    console.log(chalk.green(`✅ ${text || "Success!"}`));
  }

  fail(text?: string) {
    this.stop();
    console.log(chalk.red(`❌ ${text || "Failed!"}`));
  }
}
