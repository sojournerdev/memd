package paths

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

const (
	AppName = "memd"
	EnvHome = "MEMD_HOME"

	dirPerm = 0o700
)

type Paths struct {
	StateDir string
	DBPath   string
	BlobsDir string
}

func Resolve() (Paths, error) {
	stateDir, err := stateDir()
	if err != nil {
		return Paths{}, err
	}
	stateDir = filepath.Clean(stateDir)

	return Paths{
		StateDir: stateDir,
		DBPath:   filepath.Join(stateDir, "memd.db"),
		BlobsDir: filepath.Join(stateDir, "blobs"),
	}, nil
}

func (p Paths) Ensure() error {
	if p.StateDir == "" {
		return errors.New("paths: empty StateDir")
	}
	if err := os.MkdirAll(p.StateDir, dirPerm); err != nil {
		return fmt.Errorf("paths: mkdir state dir: %w", err)
	}
	if err := os.MkdirAll(p.BlobsDir, dirPerm); err != nil {
		return fmt.Errorf("paths: mkdir blobs dir: %w", err)
	}
	return nil
}

func (p Paths) ValidateReadWrite() error {
	if p.StateDir == "" {
		return errors.New("paths: empty StateDir")
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
	_ = f.Close()
	_ = os.Remove(name)
	return nil
}

func stateDir() (string, error) {
	if v := os.Getenv(EnvHome); v != "" {
		return v, nil
	}
	if v := os.Getenv("XDG_STATE_HOME"); v != "" {
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
