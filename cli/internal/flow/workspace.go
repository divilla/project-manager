package flow

import (
	"bytes"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"syscall"
)

// TmpDir is the single repository-relative Flow temp root.
const TmpDir = ".mch/tmp"

const (
	sessionFileName     = "session-id"
	inputFileName       = "input.md"
	outputFileName      = "output.md"
	agentOutputFileName = "agent-output.md"
	editorDirectoryName = "editor"
	ideaCreateFileName  = "new-idea.md"
)

// WorkspaceScope selects task-local files beneath one persisted artifact.
type WorkspaceScope string

// Supported workspace scopes. The zero value is the shared artifact workspace.
const (
	WorkspaceArtifact WorkspaceScope = ""
	WorkspaceEditor   WorkspaceScope = editorDirectoryName
)

var uuidPattern = regexp.MustCompile(`^[0-9A-Fa-f]{8}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{4}-[0-9A-Fa-f]{12}$`)

// Workspace resolves one Change- and artifact-scoped runtime directory.
type Workspace struct {
	Root      string
	ChangeRef string
	Artifact  Artifact
	Scope     WorkspaceScope
}

// CreateIdeaWorkspace creates one isolated pre-persistence Idea attempt.
func CreateIdeaWorkspace(root string) (string, string, error) {
	return createIdeaWorkspace(root, rand.Reader)
}

func createIdeaWorkspace(root string, random io.Reader) (string, string, error) {
	if strings.TrimSpace(root) == "" {
		return "", "", fmt.Errorf("repository root is required")
	}
	if err := requireTempRoot(root); err != nil {
		return "", "", err
	}
	for {
		attemptUUID, err := randomUUIDv4(random)
		if err != nil {
			return "", "", fmt.Errorf("generate IdeaCreate attempt UUID: %w", err)
		}
		directory := filepath.Join(root, TmpDir, attemptUUID)
		if err := os.Mkdir(directory, 0o755); err != nil {
			if os.IsExist(err) {
				continue
			}
			return "", "", fmt.Errorf("create IdeaCreate workspace %q: %w", directory, err)
		}
		path := filepath.Join(directory, ideaCreateFileName)
		file, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
		if err != nil {
			_ = os.Remove(directory)
			return "", "", fmt.Errorf("create %s: %w", ideaCreateFileName, err)
		}
		if err := file.Close(); err != nil {
			_ = os.Remove(path)
			_ = os.Remove(directory)
			return "", "", fmt.Errorf("close %s: %w", ideaCreateFileName, err)
		}
		return attemptUUID, path, nil
	}
}

// CleanupIdeaWorkspace removes the attempt file, then its directory when empty.
func CleanupIdeaWorkspace(root, attemptUUID string) error {
	if strings.TrimSpace(root) == "" {
		return fmt.Errorf("repository root is required")
	}
	if !uuidPattern.MatchString(attemptUUID) {
		return fmt.Errorf("IdeaCreate attempt UUID must be a valid UUID")
	}
	directory := filepath.Join(root, TmpDir, attemptUUID)
	path := filepath.Join(directory, ideaCreateFileName)
	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove %s: %w", ideaCreateFileName, err)
	}
	if err := os.Remove(directory); err != nil && !os.IsNotExist(err) && !isDirectoryNotEmptyError(err) {
		return fmt.Errorf("remove IdeaCreate workspace %q: %w", directory, err)
	}
	return nil
}

func randomUUIDv4(random io.Reader) (string, error) {
	value := make([]byte, 16)
	if _, err := io.ReadFull(random, value); err != nil {
		return "", err
	}
	value[6] = value[6]&0x0f | 0x40
	value[8] = value[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", value[0:4], value[4:6], value[6:8], value[8:10], value[10:16]), nil
}

func isDirectoryNotEmptyError(err error) bool {
	return errors.Is(err, syscall.ENOTEMPTY) || errors.Is(err, syscall.EEXIST)
}

// Directory returns the artifact workspace directory.
func (w Workspace) Directory() (string, error) {
	if strings.TrimSpace(w.Root) == "" {
		return "", fmt.Errorf("repository root is required")
	}
	if !uuidPattern.MatchString(w.ChangeRef) {
		return "", fmt.Errorf("flow context ref_uuid must be a valid UUID")
	}
	if !supportedArtifact(w.Artifact) {
		return "", fmt.Errorf("unsupported artifact %q", w.Artifact)
	}
	directory := filepath.Join(w.Root, TmpDir, w.ChangeRef, string(w.Artifact))
	switch w.Scope {
	case WorkspaceArtifact:
		return directory, nil
	case WorkspaceEditor:
		return filepath.Join(directory, editorDirectoryName), nil
	default:
		return "", fmt.Errorf("unsupported workspace scope %q", w.Scope)
	}
}

// SessionPath returns the session-id path.
func (w Workspace) SessionPath() (string, error) { return w.resourcePath(sessionFileName) }

// InputPath returns the immutable Task baseline path.
func (w Workspace) InputPath() (string, error) { return w.resourcePath(inputFileName) }

// OutputPath returns the mutable artifact path.
func (w Workspace) OutputPath() (string, error) { return w.resourcePath(outputFileName) }

// AgentOutputPath returns the Exec agent-output path.
func (w Workspace) AgentOutputPath() (string, error) { return w.resourcePath(agentOutputFileName) }

func (w Workspace) resourcePath(name string) (string, error) {
	directory, err := w.Directory()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, name), nil
}

func (w Workspace) replaceBaseline(content []byte) error {
	directory, err := w.Directory()
	if err != nil {
		return err
	}
	if err := requireTempRoot(w.Root); err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create artifact workspace %q: %w", directory, err)
	}
	input, _ := w.InputPath()
	output, _ := w.OutputPath()
	if err := os.WriteFile(input, content, 0o644); err != nil {
		return fmt.Errorf("write input.md: %w", err)
	}
	if err := os.WriteFile(output, content, 0o644); err != nil {
		return fmt.Errorf("write output.md: %w", err)
	}
	return nil
}

// publishEditorPreview copies an Editor comparison pair to the shared artifact
// workspace so Preview can render it without exposing the Editor workspace as
// the persisted artifact identity.
func (w Workspace) publishEditorPreview() error {
	if w.Scope != WorkspaceEditor {
		return fmt.Errorf("publish Editor Preview requires the Editor workspace")
	}
	sourceDirectory, err := w.Directory()
	if err != nil {
		return err
	}
	input, err := os.ReadFile(filepath.Join(sourceDirectory, inputFileName))
	if err != nil {
		return fmt.Errorf("read Editor input.md: %w", err)
	}
	output, err := os.ReadFile(filepath.Join(sourceDirectory, outputFileName))
	if err != nil {
		return fmt.Errorf("read Editor output.md: %w", err)
	}
	target := w
	target.Scope = WorkspaceArtifact
	directory, err := target.Directory()
	if err != nil {
		return err
	}
	if err := requireTempRoot(w.Root); err != nil {
		return err
	}
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return fmt.Errorf("create artifact workspace %q: %w", directory, err)
	}
	if err := os.WriteFile(filepath.Join(directory, inputFileName), input, 0o644); err != nil {
		return fmt.Errorf("publish Editor input.md: %w", err)
	}
	if err := os.WriteFile(filepath.Join(directory, outputFileName), output, 0o644); err != nil {
		return fmt.Errorf("publish Editor output.md: %w", err)
	}
	return nil
}

func requireTempRoot(root string) error {
	path := filepath.Join(root, TmpDir)
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("required Flow temp directory %q does not exist", path)
		}
		return fmt.Errorf("inspect required Flow temp directory %q: %w", path, err)
	}
	if !info.IsDir() {
		return fmt.Errorf("required Flow temp path %q is not a directory", path)
	}
	return nil
}

func (w Workspace) canonicalize(expectedBaseline []byte, options DocumentOptions) ([]byte, bool, error) {
	input, err := w.InputPath()
	if err != nil {
		return nil, false, err
	}
	output, err := w.OutputPath()
	if err != nil {
		return nil, false, err
	}
	baseline, err := os.ReadFile(input)
	if err != nil {
		return nil, false, fmt.Errorf("read input.md: %w", err)
	}
	if !bytes.Equal(baseline, expectedBaseline) {
		return nil, false, fmt.Errorf("input.md changed during Task")
	}
	current, err := os.ReadFile(output)
	if err != nil {
		return nil, false, fmt.Errorf("read output.md: %w", err)
	}
	canonical, err := CanonicalizeDocument(current, options)
	if err != nil {
		return nil, false, err
	}
	if err := os.WriteFile(output, canonical.Bytes, 0o644); err != nil {
		return nil, false, fmt.Errorf("write canonical output.md: %w", err)
	}
	return canonical.Bytes, bytes.Equal(baseline, canonical.Bytes), nil
}

func (w Workspace) outputIsEmpty() (bool, error) {
	path, err := w.OutputPath()
	if err != nil {
		return false, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return false, fmt.Errorf("read output.md: %w", err)
	}
	return len(content) == 0, nil
}

func (w Workspace) readSession() (string, error) {
	path, err := w.SessionPath()
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("read session-id: %w", err)
	}
	session := strings.TrimSpace(string(content))
	if !uuidPattern.MatchString(session) {
		return "", fmt.Errorf("session-id is empty or invalid")
	}
	return session, nil
}

func (w Workspace) readAgentOutput(required bool) (string, error) {
	path, err := w.AgentOutputPath()
	if err != nil {
		return "", err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		if !required && os.IsNotExist(err) {
			return "", nil
		}
		return "", fmt.Errorf("read agent-output.md: %w", err)
	}
	return string(content), nil
}
