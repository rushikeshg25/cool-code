import { spawn, exec } from 'child_process';
import { promisify } from 'util';
import path from 'path';

const execAsync = promisify(exec);

export interface ShellOptions {
  command: string;
  directory?: string;
  description?: string;
  timeout?: number;
}

export interface ShellResult {
  stdout: string;
  stderr: string;
  exitCode: number | null;
  success: boolean;
  error?: string;
}

export async function execCommand(options: ShellOptions): Promise<ShellResult> {
  const { command, directory, timeout = 30000 } = options;

  try {
    const { stdout, stderr } = await execAsync(command, {
      cwd: directory ? path.resolve(directory) : process.cwd(),
      timeout,
      maxBuffer: 1024 * 1024 * 10,
    });

    return {
      stdout: stdout.toString(),
      stderr: stderr.toString(),
      exitCode: 0,
      success: true,
    };
  } catch (error) {
    const execError = error as any;
    return {
      stdout: execError.stdout?.toString() || '',
      stderr: execError.stderr?.toString() || '',
      exitCode: execError.code || null,
      success: false,
      error: execError.message,
    };
  }
}

export async function spawnCommand(
  options: ShellOptions,
  onOutput?: (data: string) => void
): Promise<ShellResult> {
  const { command, directory } = options;

  return new Promise((resolve) => {
    const timeout = options.timeout || 60000;
    const child = spawn('bash', ['-c', command], {
      cwd: directory ? path.resolve(directory) : process.cwd(),
      stdio: ['ignore', 'pipe', 'pipe'],
    });

    let stdout = '';
    let stderr = '';
    const maxBuffer = 1024 * 1024 * 10; // 10MB limit

    const timer = setTimeout(() => {
      child.kill();
      resolve({
        stdout,
        stderr: stderr + '\n[Error] Command timed out',
        exitCode: null,
        success: false,
        error: `Command timed out after ${timeout}ms`,
      });
    }, timeout);

    child.stdout?.on('data', (data) => {
      const output = data.toString();
      if (stdout.length + output.length < maxBuffer) {
        stdout += output;
      }
      if (onOutput) {
        onOutput(output);
      }
    });

    child.stderr?.on('data', (data) => {
      const output = data.toString();
      if (stderr.length + output.length < maxBuffer) {
        stderr += output;
      }
      if (onOutput) {
        onOutput(output);
      }
    });

    child.on('close', (code) => {
      clearTimeout(timer);
      resolve({
        stdout,
        stderr,
        exitCode: code,
        success: code === 0,
        error: code !== 0 ? `Command exited with code ${code}` : undefined,
      });
    });

    child.on('error', (error) => {
      resolve({
        stdout,
        stderr,
        exitCode: null,
        success: false,
        error: error.message,
      });
    });
  });
}
