import { spawnSync } from 'child_process';

export function copyToClipboard(text: string): { success: boolean; error?: string } {
  const platform = process.platform;

  if (platform === 'darwin') {
    return runClipboardCommand('pbcopy', [], text);
  }

  if (platform === 'win32') {
    return runClipboardCommand('clip', [], text);
  }

  // linux or other unix
  let result = runClipboardCommand('wl-copy', [], text);
  if (result.success) return result;
  result = runClipboardCommand('xclip', ['-selection', 'clipboard'], text);
  if (result.success) return result;

  return {
    success: false,
    error: 'No clipboard utility found (install wl-copy or xclip).',
  };
}

function runClipboardCommand(
  command: string,
  args: string[],
  text: string
): { success: boolean; error?: string } {
  try {
    const proc = spawnSync(command, args, {
      input: text,
      stdio: ['pipe', 'ignore', 'pipe'],
    });
    if (proc.status === 0) {
      return { success: true };
    }
    return { success: false, error: proc.stderr?.toString() || 'Failed' };
  } catch (error) {
    return { success: false, error: (error as Error).message };
  }
}
