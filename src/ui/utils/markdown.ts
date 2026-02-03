import chalk from 'chalk';

export function renderMarkdown(text: string): string {
  if (!text) return '';

  let rendered = text;

  // 1. Handle Fenced Code Blocks first to avoid styling within them
  const codeBlocks: string[] = [];
  rendered = rendered.replace(/```(\w*)\n([\s\S]*?)```/g, (match, lang, code) => {
    const placeholder = `__CODE_BLOCK_${codeBlocks.length}__`;
    const header = lang ? chalk.bgWhite.black(` ${lang.toUpperCase()} `) : chalk.bgWhite.black(' CODE ');
    const borderedCode = code
      .split('\n')
      .map((line: string) => `${chalk.gray('│')} ${line}`)
      .join('\n');
    
    codeBlocks.push(`\n${header}\n${chalk.gray('┌────────────────────────────────────────────────────────────')}\n${borderedCode}\n${chalk.gray('└────────────────────────────────────────────────────────────')}\n`);
    return placeholder;
  });

  // 2. Bold (**text** or __text__)
  rendered = rendered.replace(/(\*\*|__)(.*?)\1/g, (_, __, content) => chalk.bold(content));

  // 3. Italic (*text* or _text_)
  rendered = rendered.replace(/(\*|_)(.*?)\1/g, (_, __, content) => chalk.italic(content));

  // 4. Inline Code (`code`)
  rendered = rendered.replace(/`(.*?)`/g, (_, content) => chalk.bgGray.white(` ${content} `));

  // 5. Links ([text](url))
  rendered = rendered.replace(/\[(.*?)\]\((.*?)\)/g, (_, label, url) => `${chalk.cyan(label)} ${chalk.gray(`(${url})`)}`);

  // 6. Restore Code Blocks
  codeBlocks.forEach((block, i) => {
    rendered = rendered.replace(`__CODE_BLOCK_${i}__`, block);
  });

  return rendered;
}
