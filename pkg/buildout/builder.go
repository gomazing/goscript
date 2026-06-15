package buildout

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// Builder executes build-out exports from a manifest.
type Builder struct {
	GoBin   string
	DistDir string
	DryRun  bool
	Stdout  io.Writer
	Stderr  io.Writer
}

// NewBuilder creates a builder with sensible defaults.
func NewBuilder() *Builder {
	return &Builder{
		GoBin:   "go",
		DistDir: "dist",
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	}
}

// Build exports the requested target from a manifest.
func (b *Builder) Build(manifestPath string, target Target) (BuildResult, error) {
	manifest, err := LoadManifest(manifestPath)
	if err != nil {
		return BuildResult{}, err
	}

	moduleRoot, err := findModuleRoot(filepath.Dir(manifestPath))
	if err != nil {
		return BuildResult{}, err
	}

	distDir := b.DistDir
	if distDir == "" {
		distDir = "dist"
	}
	if !filepath.IsAbs(distDir) {
		distDir = filepath.Join(moduleRoot, distDir)
	}

	artifactRoot := manifest.PlannedOutputDir(distDir)
	if err := os.MkdirAll(artifactRoot, 0o755); err != nil {
		return BuildResult{}, err
	}

	slice, err := manifest.ResolveSlice(moduleRoot)
	if err != nil {
		return BuildResult{}, err
	}
	slice.ManifestPath = manifestPath
	slice.ModuleRoot = moduleRoot

	plan := manifest.Plan(manifestPath, target, distDir)
	plan.Slice = slice
	plan.Notes = append(plan.Notes, "inspect export-slice.hyper for the resolved file set")
	if err := writeHyperFile(filepath.Join(artifactRoot, "build-plan.hyper"), plan); err != nil {
		return BuildResult{}, err
	}
	if err := writeHyperFile(filepath.Join(artifactRoot, "pack.normalized.hyper"), manifest); err != nil {
		return BuildResult{}, err
	}
	slicePath := filepath.Join(artifactRoot, "export-slice.hyper")
	if err := writeHyperFile(slicePath, slice); err != nil {
		return BuildResult{}, err
	}

	switch target {
	case TargetEXE:
		return b.buildExecutable(manifestPath, manifest, target, moduleRoot, artifactRoot, slicePath)
	case TargetGOE:
		return b.buildGOE(manifestPath, manifest, target, moduleRoot, artifactRoot, slicePath, plan, slice)
	case TargetAPK, TargetIPA, TargetDMG:
		return b.scaffoldPlatformBundle(manifestPath, manifest, target, artifactRoot, slicePath, slice)
	default:
		return BuildResult{}, fmt.Errorf("unsupported target %q", target)
	}
}

func (b *Builder) buildExecutable(manifestPath string, manifest Manifest, target Target, moduleRoot, artifactRoot, slicePath string) (BuildResult, error) {
	outputPath := filepath.Join(artifactRoot, manifest.BinaryName())
	if b.DryRun {
		return BuildResult{
			ManifestPath: manifestPath,
			Manifest:     manifest,
			Target:       target,
			BuildTarget:  manifest.BuildTarget(),
			Status:       "dry-run",
			Message:      "dry run complete; executable not built",
			OutputPath:   outputPath,
			SlicePath:    slicePath,
			Artifacts: []string{
				filepath.Join(artifactRoot, "build-plan.hyper"),
				filepath.Join(artifactRoot, "pack.normalized.hyper"),
				slicePath,
			},
		}, nil
	}

	if err := b.runGoBuild(moduleRoot, manifest.BuildTarget(), outputPath); err != nil {
		return BuildResult{}, err
	}

	return BuildResult{
		ManifestPath: manifestPath,
		Manifest:     manifest,
		Target:       target,
		BuildTarget:  manifest.BuildTarget(),
		Status:       "built",
		Message:      "executable built successfully",
		OutputPath:   outputPath,
		SlicePath:    slicePath,
		Artifacts: []string{
			outputPath,
			filepath.Join(artifactRoot, "build-plan.hyper"),
			filepath.Join(artifactRoot, "pack.normalized.hyper"),
			slicePath,
		},
	}, nil
}

func (b *Builder) buildGOE(manifestPath string, manifest Manifest, target Target, moduleRoot, artifactRoot, slicePath string, plan BuildPlan, slice ExportSlice) (BuildResult, error) {
	tempDir, err := os.MkdirTemp("", "bo-goe-*")
	if err != nil {
		return BuildResult{}, err
	}
	defer os.RemoveAll(tempDir)

	binaryPath := filepath.Join(tempDir, manifest.BinaryName())
	if runtime.GOOS != "windows" && strings.HasSuffix(strings.ToLower(binaryPath), ".exe") {
		binaryPath = strings.TrimSuffix(binaryPath, ".exe")
	}

	if b.DryRun {
		return BuildResult{
			ManifestPath: manifestPath,
			Manifest:     manifest,
			Target:       target,
			BuildTarget:  manifest.BuildTarget(),
			Status:       "dry-run",
			Message:      "dry run complete; GOE bundle not built",
			OutputPath:   binaryPath,
			SlicePath:    slicePath,
		}, nil
	}

	if err := b.runGoBuild(moduleRoot, manifest.BuildTarget(), binaryPath); err != nil {
		return BuildResult{}, err
	}

	bundlePath := filepath.Join(artifactRoot, manifest.Output+".goe")
	if err := b.writeGOEBundle(bundlePath, manifestPath, manifest, binaryPath, plan, slice); err != nil {
		return BuildResult{}, err
	}

	return BuildResult{
		ManifestPath: manifestPath,
		Manifest:     manifest,
		Target:       target,
		BuildTarget:  manifest.BuildTarget(),
		Status:       "built",
		Message:      "goe bundle created successfully",
		OutputPath:   binaryPath,
		BundlePath:   bundlePath,
		SlicePath:    slicePath,
		Artifacts: []string{
			bundlePath,
			filepath.Join(artifactRoot, "build-plan.hyper"),
			filepath.Join(artifactRoot, "pack.normalized.hyper"),
			slicePath,
		},
	}, nil
}

func (b *Builder) scaffoldPlatformBundle(manifestPath string, manifest Manifest, target Target, artifactRoot, slicePath string, slice ExportSlice) (BuildResult, error) {
	platformRoot := filepath.Join(artifactRoot, strings.ToLower(string(target)))
	if err := os.MkdirAll(platformRoot, 0o755); err != nil {
		return BuildResult{}, err
	}

	readme := fmt.Sprintf(
		"BO generated a %s scaffold for %s.\n\n"+
			"Manifest: %s\n"+
			"Build target: %s\n\n"+
			"This artifact is a build contract, not a finalized mobile/desktop packager yet.\n",
		strings.ToUpper(string(target)),
		manifest.Name,
		manifestPath,
		manifest.BuildTarget(),
	)

	if err := os.WriteFile(filepath.Join(platformRoot, "README.txt"), []byte(readme), 0o644); err != nil {
		return BuildResult{}, err
	}
	if err := writeHyperFile(filepath.Join(platformRoot, "pack.hyper"), manifest); err != nil {
		return BuildResult{}, err
	}
	if err := writeHyperFile(filepath.Join(platformRoot, "slice.hyper"), slice); err != nil {
		return BuildResult{}, err
	}

	return BuildResult{
		ManifestPath: manifestPath,
		Manifest:     manifest,
		Target:       target,
		BuildTarget:  manifest.BuildTarget(),
		Status:       "scaffolded",
		Message:      fmt.Sprintf("%s scaffold generated; native packaging can be layered on later", strings.ToUpper(string(target))),
		OutputPath:   platformRoot,
		SlicePath:    slicePath,
		Artifacts: []string{
			platformRoot,
			filepath.Join(artifactRoot, "build-plan.hyper"),
			filepath.Join(artifactRoot, "pack.normalized.hyper"),
			slicePath,
		},
	}, nil
}

func (b *Builder) runGoBuild(workDir, buildTarget, outputPath string) error {
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		return err
	}

	args := []string{"build", "-trimpath", "-o", outputPath, "-ldflags", "-s -w", buildTarget}
	cmd := exec.Command(b.GoBin, args...)
	cmd.Dir = workDir
	cmd.Stdout = b.Stdout
	cmd.Stderr = b.Stderr
	cmd.Env = os.Environ()

	return cmd.Run()
}

func (b *Builder) writeGOEBundle(bundlePath, manifestPath string, manifest Manifest, binaryPath string, plan BuildPlan, slice ExportSlice) error {
	if err := os.MkdirAll(filepath.Dir(bundlePath), 0o755); err != nil {
		return err
	}

	file, err := os.Create(bundlePath)
	if err != nil {
		return err
	}
	defer file.Close()

	zipWriter := zip.NewWriter(file)
	defer zipWriter.Close()

	if err := addFileToZip(zipWriter, "pack.hyper", mustMarshalHyper(manifest)); err != nil {
		return err
	}

	plan.Notes = append(plan.Notes, "portable GOE bundle with embedded binary and manifest")
	if err := addFileToZip(zipWriter, "build-plan.hyper", mustMarshalHyper(plan)); err != nil {
		return err
	}
	if err := addFileToZip(zipWriter, "slice.hyper", mustMarshalHyper(slice)); err != nil {
		return err
	}

	if err := addBinaryToZip(zipWriter, filepath.Join("bin", manifest.BinaryName()), binaryPath); err != nil {
		return err
	}

	runtimeNote := fmt.Sprintf(
		"GOE bundle for %s\n\n"+
			"Source manifest: %s\n"+
			"Target: %s\n"+
			"Build target: %s\n"+
			"Created: %s\n\n"+
			"This package is a portable app bundle format. The Go runtime is already compiled into the binary itself.\n"+
			"Future runtime loaders can add sandboxing, launch metadata, or device-specific adapters.\n",
		manifest.Name,
		manifestPath,
		TargetGOE,
		manifest.BuildTarget(),
		time.Now().UTC().Format(time.RFC3339),
	)
	if err := addFileToZip(zipWriter, "README.txt", []byte(runtimeNote)); err != nil {
		return err
	}

	return nil
}

func addFileToZip(zipWriter *zip.Writer, name string, content []byte) error {
	writer, err := zipWriter.Create(name)
	if err != nil {
		return err
	}

	_, err = writer.Write(content)
	return err
}

func addBinaryToZip(zipWriter *zip.Writer, name, path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer, err := zipWriter.Create(name)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
