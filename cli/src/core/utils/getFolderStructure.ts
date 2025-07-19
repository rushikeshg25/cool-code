import * as fs from 'fs';
import * as path from 'path';

interface getFolderStructureType{
    gitIgnoreChecker: (filePath: string) => boolean | null;
    rootDir:string;
}

interface TreeNode {
  name: string;
  isDirectory: boolean;
  children?: TreeNode[];
}

export function getFolderStructure({gitIgnoreChecker,rootDir}:getFolderStructureType): string {

  if(!gitIgnoreChecker){
    gitIgnoreChecker=(filePath:string)=>{
        return false
    }
  }
  const tree = buildTree(rootDir, gitIgnoreChecker);
  return formatTree(tree, rootDir);
}

function buildTree(dirPath: string, gitIgnoreChecker: (filePath: string) => boolean | null, basePath?: string): TreeNode[] {
  const baseDir = basePath || dirPath;
  const nodes: TreeNode[] = [];

  try {
    const items = fs.readdirSync(dirPath, { withFileTypes: true });
    
    // Sort items: directories first, then files, both alphabetically
    items.sort((a, b) => {
      if (a.isDirectory() && !b.isDirectory()) return -1;
      if (!a.isDirectory() && b.isDirectory()) return 1;
      return a.name.localeCompare(b.name);
    });

    for (const item of items) {
      const fullPath = path.join(dirPath, item.name);
      const relativePath = path.relative(baseDir, fullPath);
      
      // Skip if file/folder should be ignored according to gitignore
      if (gitIgnoreChecker(relativePath)) {
        continue;
      }

      const node: TreeNode = {
        name: item.name,
        isDirectory: item.isDirectory()
      };

      if (item.isDirectory()) {
        // Recursively build children for directories
        node.children = buildTree(fullPath, gitIgnoreChecker, baseDir);
      }

      nodes.push(node);
    }
  } catch (error) {
    // Skip directories that can't be read (permission issues, etc.)
    console.warn(`Warning: Could not read directory ${dirPath}: ${error}`);
  }

  return nodes;
}

function formatTree(nodes: TreeNode[], rootDir: string, prefix: string = '', isLast: boolean = true): string {
  let result = '';
  
  // Add root directory name at the beginning
  if (prefix === '') {
    result += `${path.basename(rootDir)}/\n`;
  }

  for (let i = 0; i < nodes.length; i++) {
    const node = nodes[i];
    const isLastNode = i === nodes.length - 1;
    const currentPrefix = isLastNode ? '└───' : '├───';
    const nextPrefix = prefix + (isLastNode ? '    ' : '│   ');

    // Add current node
    result += `${prefix}${currentPrefix}${node.name}${node.isDirectory ? '/' : ''}\n`;

    // Add children if it's a directory
    if (node.isDirectory && node.children && node.children.length > 0) {
      result += formatTree(node.children, rootDir, nextPrefix, isLastNode);
    }
  }

  return result;
}


// export function getFolderStructureCompact(rootDir: string): string {
//   const gitIgnoreChecker = createGitIgnoreChecker(rootDir);
//   const result: string[] = [];
  
//   function traverse(dirPath: string, depth: number = 0, basePath?: string) {
//     const baseDir = basePath || dirPath;
    
//     try {
//       const items = fs.readdirSync(dirPath, { withFileTypes: true });
      
//       items.sort((a, b) => {
//         if (a.isDirectory() && !b.isDirectory()) return -1;
//         if (!a.isDirectory() && b.isDirectory()) return 1;
//         return a.name.localeCompare(b.name);
//       });

//       for (const item of items) {
//         const fullPath = path.join(dirPath, item.name);
//         const relativePath = path.relative(baseDir, fullPath);
        
//         if (gitIgnoreChecker(relativePath)) {
//           continue;
//         }

//         const indent = '  '.repeat(depth);
//         const marker = item.isDirectory() ? '📁' : '📄';
//         result.push(`${indent}${marker} ${item.name}`);

//         if (item.isDirectory()) {
//           traverse(fullPath, depth + 1, baseDir);
//         }
//       }
//     } catch (error) {
//       console.warn(`Warning: Could not read directory ${dirPath}: ${error}`);
//     }
//   }

//   result.push(`📁 ${path.basename(rootDir)}`);
//   traverse(rootDir, 1);
  
//   return result.join('\n');
// }