import ora, { Ora } from "ora";
import chalk from "chalk";

export class DynamicSpinner {
  private spinner: Ora;

  constructor(initialText: string = "Processing...") {
    this.spinner = ora({
      text: chalk.cyan(initialText),
      spinner: "dots12",
      color: "cyan",
    }).start();
  }

  updateText(text: string) {
    this.spinner.text = chalk.cyan(text);
  }

  succeed(text?: string) {
    this.spinner.succeed(chalk.green(text || "Success!"));
  }

  fail(text?: string) {
    this.spinner.fail(chalk.red(text || "Failed!"));
  }

  warn(text?: string) {
    this.spinner.warn(chalk.yellow(text || "Warning!"));
  }

  info(text?: string) {
    this.spinner.info(chalk.blue(text || "Info"));
  }

  stop() {
    this.spinner.stop();
  }
}
