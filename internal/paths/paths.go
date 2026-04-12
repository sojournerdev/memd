package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName         = "memd"
	EnvHome         = "MEMD_HOME"
	EnvXDGStateHome = "XDG_STATE_HOME"

	dirPerm = 0o700

	dbFileName   = "memd.db"
	blobsDirName = "blobs"
)

// Paths describes the on-disk layout memd uses for local state.
// The fields are derived from a single state directory and are expected to
// remain consistent with that layout.
type Paths struct {
	StateDir string
	DBPath   string
	BlobsDir string
}

// Resolve determines the memd state directory from environment/OS defaults
// and returns a normalized Paths value for that location.
func Resolve() (Paths, error) {
	stateDir, err := stateDir()
	if err != nil {
		return Paths{}, err
	}
	return newPaths(stateDir)
}

// Ensure validates p and creates the state and blobs directories if needed.
func (p Paths) Ensure() error {
	if err := p.validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(p.StateDir, dirPerm); err != nil {
		return fmt.Errorf("paths: mkdir state dir: %w", err)
	}
	if err := os.MkdirAll(p.BlobsDir, dirPerm); err != nil {
		return fmt.Errorf("paths: mkdir blobs dir: %w", err)
	}
	return nil
}

// ValidateReadWrite validates p, confirms the state directory exists and is a
// directory, and verifies write access by creating and removing a temp file.
func (p Paths) ValidateReadWrite() error {
	if err := p.validate(); err != nil {
		return err
	}

	fi, err := os.Stat(p.StateDir)
	if err != nil {
		return fmt.Errorf("paths: stat state dir: %w", err)
	}
	if !fi.IsDir() {
		return fmt.Errorf("paths: state dir is not a directory: %s", p.StateDir)
	}

	f, err := os.CreateTemp(p.StateDir, ".memd-writecheck-*")
	if err != nil {
		return fmt.Errorf("paths: write check failed: %w", err)
	}
	name := f.Name()
	if err := f.Close(); err != nil {
		_ = os.Remove(name)
		return fmt.Errorf("paths: write check close temp file: %w", err)
	}
	if err := os.Remove(name); err != nil {
		return fmt.Errorf("paths: write check remove temp file: %w", err)
	}
	return nil
}

func newPaths(stateDir string) (Paths, error) {
	if stateDir == "" {
		return Paths{}, errors.New("paths: empty StateDir")
	}
	stateDir = filepath.Clean(stateDir)
	p := Paths{
		StateDir: stateDir,
		DBPath:   filepath.Join(stateDir, dbFileName),
		BlobsDir: filepath.Join(stateDir, blobsDirName),
	}
	if err := p.validate(); err != nil {
		return Paths{}, err
	}
	return p, nil
}

func (p Paths) validate() error {
	if p.StateDir == "" {
		return errors.New("paths: empty StateDir")
	}
	stateDir := filepath.Clean(p.StateDir)

	if p.DBPath == "" {
		return errors.New("paths: empty DBPath")
	}
	wantDB := filepath.Join(stateDir, dbFileName)
	if filepath.Clean(p.DBPath) != wantDB {
		return fmt.Errorf("paths: DBPath must be %s", wantDB)
	}

	if p.BlobsDir == "" {
		return errors.New("paths: empty BlobsDir")
	}
	wantBlobs := filepath.Join(stateDir, blobsDirName)
	if filepath.Clean(p.BlobsDir) != wantBlobs {
		return fmt.Errorf("paths: BlobsDir must be %s", wantBlobs)
	}
	return nil
}

func stateDir() (string, error) {
	if v := os.Getenv(EnvHome); v != "" {
		return v, nil
	}
	if v := os.Getenv(EnvXDGStateHome); v != "" {
		return filepath.Join(v, AppName), nil
	}

	switch runtime.GOOS {
	case "darwin", "windows":
		base, err := os.UserConfigDir()
		if err != nil {
			return "", fmt.Errorf("paths: user config dir: %w", err)
		}
		return filepath.Join(base, AppName), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("paths: home dir: %w", err)
		}
		return filepath.Join(home, ".local", "state", AppName), nil
	}
}
