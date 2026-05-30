export * from './getFolderStructure';
export * from './scanProject';

// Non-fatal debug logging. Stays silent during normal (TUI) operation so it
// does not corrupt the rendered output; set COOLCODE_DEBUG=1 to surface notes.
export function debugLog(context: string, error: unknown): void {
  if (process.env.COOLCODE_DEBUG) {
    console.error(`[coolcode:debug] ${context}: ${getErrorMessage(error)}`);
  }
}

export function getErrorMessage(error: unknown): string {
  if (error instanceof Error) {
    return error.message;
  }
  try {
    return String(error);
  } catch {
    return 'Failed to get error details';
  }
}
