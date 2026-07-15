package project

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"time"
)

// Scan is a summary of the project's shape.
type Scan struct {
	RootDir     string   `json:"rootDir"`
	Timestamp   string   `json:"timestamp"`
	Entrypoints []string `json:"entrypoints"`
	Frameworks  []string `json:"frameworks"`
	Scripts     []string `json:"scripts"`
	Languages   []string `json:"languages"`
	HasTsConfig bool     `json:"hasTsConfig"`
	HasReadme   bool     `json:"hasReadme"`
	IsNextJS    bool     `json:"isNextJS"`
	IsAstro     bool     `json:"isAstro"`
}

var commonEntrypoints = []string{
	"src/index.ts", "src/index.js", "src/main.ts", "src/main.js",
	"src/app.ts", "src/app.js", "src/server.ts", "src/server.js",
	"index.ts", "index.js", "app.ts", "app.js", "server.ts", "server.js",
	"main.go", "cmd/main.go",
}

var frameworkMap = map[string][]string{
	"NextJS":   {"next"},
	"React":    {"react", "react-dom"},
	"Vue":      {"vue"},
	"Svelte":   {"svelte"},
	"Express":  {"express"},
	"Fastify":  {"fastify"},
	"NestJS":   {"@nestjs/core"},
	"Prisma":   {"prisma", "@prisma/client"},
	"Tailwind": {"tailwindcss"},
	"Vite":     {"vite"},
}

var extLangMap = map[string]string{
	".ts": "TypeScript", ".tsx": "TypeScript (TSX)", ".js": "JavaScript",
	".jsx": "JavaScript (JSX)", ".py": "Python", ".go": "Go", ".rs": "Rust",
	".java": "Java", ".rb": "Ruby", ".php": "PHP", ".cs": "C#", ".cpp": "C++",
	".c": "C", ".h": "C/C++ Header", ".md": "Markdown", ".json": "JSON",
	".yml": "YAML", ".yaml": "YAML", ".sh": "Shell", ".sql": "SQL",
	".html": "HTML", ".css": "CSS",
}

// ScanProject inspects rootDir for entrypoints, frameworks, scripts, languages.
func ScanProject(rootDir string) Scan {
	var entrypoints []string
	for _, rel := range commonEntrypoints {
		if _, err := os.Stat(filepath.Join(rootDir, rel)); err == nil {
			entrypoints = append(entrypoints, rel)
		}
	}

	var scripts, frameworks []string
	if pkg := readPackageJSON(rootDir); pkg != nil {
		for name := range pkg.Scripts {
			scripts = append(scripts, name)
		}
		sort.Strings(scripts)
		deps := map[string]bool{}
		for d := range pkg.Dependencies {
			deps[d] = true
		}
		for d := range pkg.DevDependencies {
			deps[d] = true
		}
		frameworks = detectFrameworks(deps)
	}

	exists := func(name string) bool {
		_, err := os.Stat(filepath.Join(rootDir, name))
		return err == nil
	}

	frameworkSet := map[string]bool{}
	for _, f := range frameworks {
		frameworkSet[f] = true
	}

	return Scan{
		RootDir:     rootDir,
		Timestamp:   time.Now().UTC().Format(time.RFC3339),
		Entrypoints: entrypoints,
		Frameworks:  frameworks,
		Scripts:     scripts,
		Languages:   detectLanguages(rootDir),
		HasTsConfig: exists("tsconfig.json"),
		HasReadme:   exists("README.md"),
		IsNextJS:    frameworkSet["NextJS"] || exists("next.config.js") || exists("next.config.mjs"),
		IsAstro:     frameworkSet["Astro"] || exists("astro.config.mjs"),
	}
}

type packageJSON struct {
	Scripts         map[string]string `json:"scripts"`
	Dependencies    map[string]string `json:"dependencies"`
	DevDependencies map[string]string `json:"devDependencies"`
}

func readPackageJSON(rootDir string) *packageJSON {
	raw, err := os.ReadFile(filepath.Join(rootDir, "package.json"))
	if err != nil {
		return nil
	}
	var pkg packageJSON
	if err := json.Unmarshal(raw, &pkg); err != nil {
		return nil
	}
	return &pkg
}

func detectFrameworks(deps map[string]bool) []string {
	var found []string
	for framework, keywords := range frameworkMap {
		for _, k := range keywords {
			if deps[k] {
				found = append(found, framework)
				break
			}
		}
	}
	sort.Strings(found)
	return found
}

func detectLanguages(rootDir string) []string {
	counts := map[string]int{}
	_ = filepath.WalkDir(rootDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if d.Name() == "node_modules" || d.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		if lang, ok := extLangMap[filepath.Ext(path)]; ok {
			counts[lang]++
		}
		return nil
	})
	langs := make([]string, 0, len(counts))
	for l := range counts {
		langs = append(langs, l)
	}
	sort.Slice(langs, func(i, j int) bool {
		if counts[langs[i]] != counts[langs[j]] {
			return counts[langs[i]] > counts[langs[j]]
		}
		return langs[i] < langs[j]
	})
	return langs
}
