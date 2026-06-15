package gopm

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gomazing/goscript/pkg/buildout"
)

// SetupOptions captures the topology and shape requested for gopm setup.
type SetupOptions struct {
	ProjectDir   string
	ProjectName  string
	Mode         string
	Type         string
	Entrypoint   string
	ManifestName string
	Force        bool
}

func parseSetupArgs(args []string) (SetupOptions, error) {
	opts := SetupOptions{
		Mode: "cs",
		Type: "app",
	}

	for i := 0; i < len(args); i++ {
		arg := strings.TrimSpace(args[i])
		if arg == "" {
			continue
		}

		switch arg {
		case "--cs":
			opts.Mode = "cs"
		case "--sw":
			opts.Mode = "sw"
		case "--force":
			opts.Force = true
		case "--mode":
			i++
			if i >= len(args) {
				return SetupOptions{}, fmt.Errorf("missing value for --mode")
			}
			opts.Mode = strings.ToLower(strings.TrimSpace(args[i]))
		case "--type", "--template":
			i++
			if i >= len(args) {
				return SetupOptions{}, fmt.Errorf("missing value for %s", arg)
			}

			projectType, err := normalizeProjectType(args[i])
			if err != nil {
				return SetupOptions{}, err
			}
			opts.Type = projectType
		case "--entrypoint":
			i++
			if i >= len(args) {
				return SetupOptions{}, fmt.Errorf("missing value for --entrypoint")
			}
			opts.Entrypoint = strings.TrimSpace(args[i])
		case "--manifest":
			i++
			if i >= len(args) {
				return SetupOptions{}, fmt.Errorf("missing value for --manifest")
			}
			opts.ManifestName = strings.TrimSpace(args[i])
		default:
			if strings.HasPrefix(arg, "--") {
				return SetupOptions{}, fmt.Errorf("unknown setup flag %q", arg)
			}
			if opts.ProjectDir != "" {
				return SetupOptions{}, fmt.Errorf("unexpected extra argument %q", arg)
			}
			opts.ProjectDir = arg
		}
	}

	if opts.Mode != "cs" && opts.Mode != "sw" {
		return SetupOptions{}, fmt.Errorf("mode must be either cs or sw")
	}

	if opts.ProjectDir == "" {
		opts.ProjectDir = "."
	}

	projectDir, err := filepath.Abs(opts.ProjectDir)
	if err != nil {
		return SetupOptions{}, fmt.Errorf("resolve project path: %w", err)
	}
	opts.ProjectDir = projectDir

	opts.ProjectName = sanitizeName(filepath.Base(projectDir))
	if opts.ProjectName == "" {
		opts.ProjectName = "goscript-app"
	}

	if opts.Entrypoint == "" {
		opts.Entrypoint = "./cmd/server"
	}

	if opts.ManifestName == "" {
		opts.ManifestName = opts.ProjectName
	}

	return opts, nil
}

func normalizeProjectType(raw string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "", "app":
		return "app", nil
	case "web", "website":
		return "website", nil
	case "erp":
		return "erp", nil
	default:
		return "", fmt.Errorf("project type must be website, app, or erp")
	}
}

func sanitizeName(raw string) string {
	raw = strings.TrimSpace(strings.ToLower(raw))
	if raw == "" {
		return ""
	}

	var b strings.Builder
	lastDash := false
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
			lastDash = false
		case r >= '0' && r <= '9':
			b.WriteRune(r)
			lastDash = false
		default:
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	return strings.Trim(b.String(), "-")
}

func (pm *PackageManager) setupProject(opts SetupOptions) (string, error) {
	dirs := []string{
		opts.ProjectDir,
		filepath.Join(opts.ProjectDir, "base"),
		filepath.Join(opts.ProjectDir, "base", "config"),
		filepath.Join(opts.ProjectDir, "base", "policies"),
		filepath.Join(opts.ProjectDir, "agents"),
		filepath.Join(opts.ProjectDir, "app"),
		filepath.Join(opts.ProjectDir, "app", "modules"),
		filepath.Join(opts.ProjectDir, "app", "pages"),
		filepath.Join(opts.ProjectDir, "app", "components"),
		filepath.Join(opts.ProjectDir, "app", "services"),
		filepath.Join(opts.ProjectDir, "app", "routes"),
		filepath.Join(opts.ProjectDir, "app", "assets"),
		filepath.Join(opts.ProjectDir, "core"),
		filepath.Join(opts.ProjectDir, "tests"),
		filepath.Join(opts.ProjectDir, "docs"),
		filepath.Join(opts.ProjectDir, "deploy"),
		filepath.Join(opts.ProjectDir, "packs"),
		filepath.Join(opts.ProjectDir, "cmd"),
		filepath.Join(opts.ProjectDir, "cmd", "server"),
	}

	if opts.Mode == "cs" {
		dirs = append(dirs,
			filepath.Join(opts.ProjectDir, "app", "api"),
			filepath.Join(opts.ProjectDir, "app", "controllers"),
			filepath.Join(opts.ProjectDir, "app", "views"),
		)
	} else {
		dirs = append(dirs,
			filepath.Join(opts.ProjectDir, "app", "topology"),
			filepath.Join(opts.ProjectDir, "app", "sync"),
			filepath.Join(opts.ProjectDir, "app", "swarm-policies"),
			filepath.Join(opts.ProjectDir, "app", "trust"),
		)
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return "", fmt.Errorf("create %s: %w", dir, err)
		}
	}

	if err := writeStarterFiles(opts); err != nil {
		return "", err
	}

	if err := writeSetupStub(filepath.Join(opts.ProjectDir, "base", "README.md"), baseReadmeStub(opts), opts.Force); err != nil {
		return "", err
	}

	if err := writeSetupStub(filepath.Join(opts.ProjectDir, "agents", "README.md"), agentsReadmeStub(), opts.Force); err != nil {
		return "", err
	}

	projectManifest := NewProjectManifest(opts.ProjectName)
	projectManifest.Mode = opts.Mode
	projectManifest.Type = opts.Type
	projectManifest.Main = opts.Entrypoint
	projectManifest.Description = fmt.Sprintf("%s scaffold generated by gopm setup", opts.Type)
	projectManifest.Metadata["projectType"] = opts.Type
	projectManifest.Metadata["setupMode"] = opts.Mode

	projectManifestPath := filepath.Join(opts.ProjectDir, "gopm.hyper")
	if err := writeProjectManifestFile(projectManifestPath, projectManifest, opts.Force); err != nil {
		return "", err
	}

	manifest := buildout.Manifest{
		Name:        opts.ManifestName,
		Mode:        opts.Mode,
		Output:      opts.ManifestName,
		Module:      ".",
		Entrypoint:  opts.Entrypoint,
		BaseDir:     "base",
		AgentsDir:   "agents",
		Description: fmt.Sprintf("%s scaffold generated by gopm setup", opts.Type),
		Pages:       defaultPages(opts.Type),
		Metadata: map[string]string{
			"projectType": opts.Type,
			"setupMode":   opts.Mode,
		},
	}

	manifestPath := filepath.Join(opts.ProjectDir, "packs", opts.ManifestName+".pack")
	if err := writePackFile(manifestPath, manifest, opts.Force); err != nil {
		return "", err
	}

	return manifestPath, nil
}

func defaultPages(projectType string) []string {
	switch projectType {
	case "website":
		return []string{"/"}
	case "erp":
		return []string{"/", "/dashboard", "/modules"}
	default:
		return []string{"/"}
	}
}

func writeStarterFiles(opts SetupOptions) error {
	modulePath := starterModulePath(opts.ProjectName)

	if err := writeSetupStub(filepath.Join(opts.ProjectDir, "go.mod"), starterGoMod(opts.ProjectDir, modulePath), opts.Force); err != nil {
		return err
	}

	if err := writeSetupStub(filepath.Join(opts.ProjectDir, "README.md"), starterReadme(opts, modulePath), opts.Force); err != nil {
		return err
	}

	if err := writeSetupStub(filepath.Join(opts.ProjectDir, "cmd", "server", "main.go"), starterServerMain(opts, modulePath), opts.Force); err != nil {
		return err
	}

	if err := writeSetupStub(filepath.Join(opts.ProjectDir, "app", "pages", "home.gsx"), starterHomeGSX(opts), opts.Force); err != nil {
		return err
	}

	if err := writeSetupStub(filepath.Join(opts.ProjectDir, "app", "pages", "home.go"), starterHomeGo(opts), opts.Force); err != nil {
		return err
	}

	return nil
}

func starterModulePath(projectName string) string {
	projectName = sanitizeName(projectName)
	if projectName == "" {
		projectName = "goscript-app"
	}
	return "example.com/" + projectName
}

func starterGoMod(projectDir, modulePath string) string {
	replacePath := starterReplacePath(projectDir)

	var b strings.Builder
	fmt.Fprintf(&b, "module %s\n\n", modulePath)
	b.WriteString("go 1.21\n\n")
	b.WriteString("require github.com/gomazing/goscript v0.0.0\n")
	if replacePath != "" {
		fmt.Fprintf(&b, "\nreplace github.com/gomazing/goscript => %s\n", replacePath)
	}
	return b.String()
}

func starterReplacePath(projectDir string) string {
	sourceDir, err := os.Getwd()
	if err != nil {
		return ""
	}

	rel, err := filepath.Rel(projectDir, sourceDir)
	if err != nil {
		return ""
	}

	rel = filepath.ToSlash(strings.TrimSpace(rel))
	if rel == "" {
		return ""
	}
	return rel
}

func starterReadme(opts SetupOptions, modulePath string) string {
	return fmt.Sprintf(`# %s

GoScript starter scaffold for a framework-ready language runtime.

## Included batteries

- module path: %s
- %s mode scaffold
- %s type manifest
- starter home page in GoScript source form
- generated Go runtime page for immediate execution
- HTTP server with a sample API route
- TALK endpoint example for bidirectional transport

## Run

1. Review gopm.hyper if you want to adjust the project contract.
2. Run the starter server:

   go run ./cmd/server

## Notes

- The starter uses a local replace entry for the GoScript runtime so the scaffold can run from the same checkout.
- The browser is one target, not the target.
- This scaffold is batteries-included, but it still leaves composable frameworks to the ecosystem.
`, opts.ProjectName, modulePath, strings.ToUpper(opts.Mode), opts.Type)
}

func starterHomeGSX(opts SetupOptions) string {
	title := "Built for the AI Era"
	if strings.TrimSpace(opts.ProjectName) != "" {
		title = fmt.Sprintf("%s - Built for the AI Era", strings.TrimSpace(opts.ProjectName))
	}

	return fmt.Sprintf(`package pages

import (
	"github.com/gomazing/goscript/pkg/goscript"
)

func Home(props goscript.Props) string {
	return <main class="shell">
		<section class="hero">
			<p class="eyebrow">GoScript</p>
			<h1>%s</h1>
			<p>Framework-ready language batteries included for fast vibe coding.</p>
		</section>
		<section class="grid">
			<article class="card">
				<h2>Go FAST</h2>
				<p>Lower allocations and keep hot paths lean.</p>
			</article>
			<article class="card">
				<h2>Go PAINT</h2>
				<p>Compose spatial interfaces with canvas-first primitives.</p>
			</article>
			<article class="card">
				<h2>Go IRT</h2>
				<p>Ship realtime sync and live collaboration by default.</p>
			</article>
			<article class="card">
				<h2>Go TALK</h2>
				<p>Support bidirectional transport for APIs, text, media, and sensors.</p>
			</article>
		</section>
	</main>
}
`, title)
}

func starterHomeGo(opts SetupOptions) string {
	return `package pages

import (
	"strings"

	"github.com/gomazing/goscript/pkg/goscript"
)

func Home(props goscript.Props) string {
	title := "Built for the AI Era"
	if value, ok := props["title"].(string); ok && strings.TrimSpace(value) != "" {
		title = value
	}

	return goscript.CreateElement("main", goscript.Props{"class": "shell"},
		goscript.CreateElement("section", goscript.Props{"class": "hero"},
			goscript.CreateElement("p", goscript.Props{"class": "eyebrow"}, "GoScript"),
			goscript.CreateElement("h1", nil, title),
			goscript.CreateElement("p", nil, "Framework-ready language batteries included for fast vibe coding."),
		),
		goscript.CreateElement("section", goscript.Props{"class": "grid"},
			goscript.CreateElement("article", goscript.Props{"class": "card"},
				goscript.CreateElement("h2", nil, "Go FAST"),
				goscript.CreateElement("p", nil, "Lower allocations and keep hot paths lean."),
			),
			goscript.CreateElement("article", goscript.Props{"class": "card"},
				goscript.CreateElement("h2", nil, "Go PAINT"),
				goscript.CreateElement("p", nil, "Compose spatial interfaces with canvas-first primitives."),
			),
			goscript.CreateElement("article", goscript.Props{"class": "card"},
				goscript.CreateElement("h2", nil, "Go IRT"),
				goscript.CreateElement("p", nil, "Ship realtime sync and live collaboration by default."),
			),
			goscript.CreateElement("article", goscript.Props{"class": "card"},
				goscript.CreateElement("h2", nil, "Go TALK"),
				goscript.CreateElement("p", nil, "Support bidirectional transport for APIs, text, media, and sensors."),
			),
		),
	)
}
`
}

func starterServerMain(opts SetupOptions, modulePath string) string {
	var b strings.Builder

	fmt.Fprintf(&b, "package main\n\n")
	b.WriteString("import (\n")
	b.WriteString("\t\"context\"\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString("\t\"log\"\n")
	b.WriteString("\t\"net/http\"\n")
	b.WriteString("\t\"strings\"\n\n")
	fmt.Fprintf(&b, "\t%q\n", modulePath+"/app/pages")
	b.WriteString("\t\"github.com/gomazing/goscript/pkg/goscript\"\n")
	b.WriteString(")\n\n")

	b.WriteString("func main() {\n")
	fmt.Fprintf(&b, "\tapp := goscript.NewApp(%q, \"0.1.0\")\n", opts.ProjectName)
	b.WriteString("\tapp.DefaultMeta[\"theme\"] = \"midnight\"\n")
	b.WriteString("\tapp.Styles = []string{\n")
	styles := []string{
		"body { margin: 0; font-family: Inter, Arial, sans-serif; background: #0f172a; color: #e2e8f0; }",
		".shell { max-width: 1120px; margin: 0 auto; padding: 64px 24px; }",
		".hero { background: linear-gradient(135deg, #0ea5e9, #8b5cf6); padding: 32px; border-radius: 24px; }",
		".grid { display: grid; gap: 20px; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); margin-top: 24px; }",
		".card { background: rgba(15, 23, 42, 0.72); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 20px; padding: 24px; }",
		".eyebrow { text-transform: uppercase; letter-spacing: 0.18em; font-size: 0.75rem; opacity: 0.8; }",
	}
	for _, style := range styles {
		fmt.Fprintf(&b, "\t\t%q,\n", style)
	}
	b.WriteString("\t}\n\n")

	b.WriteString("\thome := goscript.FunctionalComponent(func(props goscript.Props, children ...interface{}) string {\n")
	b.WriteString("\t\treturn pages.Home(props)\n")
	b.WriteString("\t})\n\n")

	fmt.Fprintf(&b, "\tif err := app.RegisterPage(goscript.Page{\n")
	b.WriteString("\t\tPath:        \"/\",\n")
	fmt.Fprintf(&b, "\t\tTitle:       %q,\n", opts.ProjectName)
	b.WriteString("\t\tDescription: \"GoScript starter scaffold for the AI era.\",\n")
	b.WriteString("\t\tComponent:   home,\n")
	b.WriteString("\t\tHydrate:     true,\n")
	b.WriteString("\t\tMeta: map[string]string{\n")
	b.WriteString("\t\t\t\"section\": \"home\",\n")
	b.WriteString("\t\t},\n")
	b.WriteString("\t}); err != nil {\n")
	b.WriteString("\t\tlog.Fatal(err)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\tapp.GET(\"/api/hello\", func(w http.ResponseWriter, r *http.Request, params map[string]string) {\n")
	b.WriteString("\t\tw.Header().Set(\"Content-Type\", \"text/plain; charset=utf-8\")\n")
	b.WriteString("\t\t_, _ = fmt.Fprintln(w, \"Hello from GoScript\")\n")
	b.WriteString("\t})\n\n")

	b.WriteString("\tif err := app.RegisterTalkEndpoint(goscript.TalkEndpoint{\n")
	b.WriteString("\t\tContract: goscript.TalkContract{\n")
	b.WriteString("\t\t\tName:        \"hello\",\n")
	b.WriteString("\t\t\tPath:        \"/talk/hello\",\n")
	b.WriteString("\t\t\tMethod:      http.MethodPost,\n")
	b.WriteString("\t\t\tDescription: \"Starter TALK endpoint\",\n")
	b.WriteString("\t\t\tProfile: goscript.TalkProfile{\n")
	b.WriteString("\t\t\t\tText:         true,\n")
	b.WriteString("\t\t\t\tData:         true,\n")
	b.WriteString("\t\t\t\tBidirectional: true,\n")
	b.WriteString("\t\t\t\tMultiThreaded: true,\n")
	b.WriteString("\t\t\t},\n")
	b.WriteString("\t\t},\n")
	b.WriteString("\t\tHandler: func(ctx context.Context, req goscript.TalkRequest) (goscript.TalkResponse, error) {\n")
	b.WriteString("\t\t\treply := \"hello from TALK\"\n")
	b.WriteString("\t\t\tif value, ok := req.Query[\"name\"]; ok && strings.TrimSpace(value) != \"\" {\n")
	b.WriteString("\t\t\t\treply = \"hello, \" + strings.TrimSpace(value)\n")
	b.WriteString("\t\t\t}\n")
	b.WriteString("\t\t\treturn goscript.TalkResponse{\n")
	b.WriteString("\t\t\t\tStatus: http.StatusOK,\n")
	b.WriteString("\t\t\t\tFrame: goscript.TalkFrame{\n")
	b.WriteString("\t\t\t\t\tKind:      goscript.TalkFrameKindText,\n")
	b.WriteString("\t\t\t\t\tDirection: goscript.TalkDirectionOutbound,\n")
	b.WriteString("\t\t\t\t\tPayload:   reply,\n")
	b.WriteString("\t\t\t\t},\n")
	b.WriteString("\t\t\t}, nil\n")
	b.WriteString("\t\t},\n")
	b.WriteString("\t}); err != nil {\n")
	b.WriteString("\t\tlog.Fatal(err)\n")
	b.WriteString("\t}\n\n")

	b.WriteString("\tport := 8080\n")
	b.WriteString("\tfmt.Printf(\"GoScript starter running on http://localhost:%d\\n\", port)\n")
	b.WriteString("\tlog.Fatal(http.ListenAndServe(fmt.Sprintf(\":%d\", port), app))\n")
	b.WriteString("}\n")

	return b.String()
}

func writeSetupStub(path, contents string, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return nil
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check %s: %w", path, err)
		}
	}

	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writePackFile(path string, manifest buildout.Manifest, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("pack already exists: %s (use --force to overwrite)", path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check %s: %w", path, err)
		}
	}

	manifest.Normalize(path)
	if err := manifest.Write(path); err != nil {
		return fmt.Errorf("write pack %s: %w", path, err)
	}
	return nil
}

func writeProjectManifestFile(path string, manifest ProjectManifest, force bool) error {
	if !force {
		if _, err := os.Stat(path); err == nil {
			return fmt.Errorf("project manifest already exists: %s (use --force to overwrite)", path)
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("check %s: %w", path, err)
		}
	}

	if err := manifest.Write(path); err != nil {
		return fmt.Errorf("write project manifest %s: %w", path, err)
	}

	return nil
}

func baseReadmeStub(opts SetupOptions) string {
	return fmt.Sprintf("# Project Base Guidance\n\nThis project is scaffolded in `%s` mode as a `%s` project.\n\nUse this folder for project-local AI guidance that extends the shared GoScript `base/` contract.\n", opts.Mode, opts.Type)
}

func agentsReadmeStub() string {
	return "# Runtime Agents\n\nUse this folder only for autonomous roles that exist inside the application at runtime.\n"
}
