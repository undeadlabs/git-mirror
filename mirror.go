package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
)

func mirror(cfg config, r repo) error {
	workspaceDir := path.Join(cfg.BasePath, r.Name)
	if _, err := os.Stat(workspaceDir); err == nil {
		// Directory exists, update.
		cmd := exec.Command("git", "fetch", "-p", "origin")
		cmd.Dir = workspaceDir
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to fetch origin in %s: %s", workspaceDir, err)
		}
	} else if os.IsNotExist(err) {
		// Clone
		parent := path.Dir(workspaceDir)
		if err = os.MkdirAll(parent, 0755); err != nil {
			return fmt.Errorf("failed to create parent directory for cloning %s, %s", workspaceDir, err)
		}
		cmd := exec.Command("git", "clone", "--mirror", r.Origin, workspaceDir)
		cmd.Dir = parent
		if err = cmd.Run(); err != nil {
			return fmt.Errorf("failed to clone %s: %s", r.Origin, err)
		}
		// git remote set-url --push origin
		if r.Mirror != "" {
			cmd := exec.Command("git", "remote", "set-url", "--push", "origin", r.Mirror)
			cmd.Dir = workspaceDir
			if err = cmd.Run(); err != nil {
				return fmt.Errorf("failed to set mirror url %s: %s", r.Origin, err)
			}
		}
	} else {
		return fmt.Errorf("failed to stat %s, %s", workspaceDir, err)
	}

	// Push to mirror
	if r.Mirror != "" {
		cmd := exec.Command("git", "push", "--mirror", "--quiet")
		cmd.Dir = workspaceDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("errors pushing to mirror %s: %s", workspaceDir, err)
		}
	}

	if !cfg.NoServe {
		cmd := exec.Command("git", "update-server-info")
		cmd.Dir = workspaceDir
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("failed to update-server-info for %s, %s", workspaceDir, err)
		}
	}
	return nil
}
