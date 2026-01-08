package pkggogit

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"

	defs "github.com/alphacodinggroup/ponti-backend/pkg/repo-tools/go-git/v5/defs"
)

var (
	instance  defs.Client
	once      sync.Once
	initError error
)

type client struct {
	repo   *git.Repository
	config defs.Config
}

// newClient crea una nueva instancia del cliente Git.
// Utiliza un patrón singleton para asegurar que solo exista una instancia.
func newClient(config defs.Config) (defs.Client, error) {
	once.Do(func() {
		var repo *git.Repository
		var err error

		// Intentar abrir el repositorio existente
		repo, err = git.PlainOpen(config.GetRepoPath())
		if err != nil {
			// Si el repositorio no existe, clonarlo
			if err == git.ErrRepositoryNotExists {
				repo, err = git.PlainClone(config.GetRepoPath(), false, &git.CloneOptions{
					URL:           config.GetRepoURL(),
					ReferenceName: plumbing.NewBranchReferenceName(config.GetRepoBranch()),
					Progress:      nil,
				})
				if err != nil {
					initError = fmt.Errorf("error cloning repository: %w", err)
					return
				}
			} else {
				initError = fmt.Errorf("error opening repository: %w", err)
				return
			}
		}

		instance = &client{
			config: config,
			repo:   repo,
		}
	})
	return instance, initError
}

// GetRepository retorna el repositorio Git asociado al cliente.
func (gc *client) GetRepository() *git.Repository {
	return gc.repo
}

// PullLatest actualiza el repositorio local con los últimos cambios del remoto.
func (gc *client) PullLatest() error {
	w, err := gc.repo.Worktree()
	if err != nil {
		return fmt.Errorf("error getting worktree: %w", err)
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
	})
	if err != nil {
		if err == git.NoErrAlreadyUpToDate {
			return nil
		}
		return fmt.Errorf("error pulling latest changes: %w", err)
	}
	return nil
}

// GetRepoFiles obtiene los archivos del repositorio que cumplen con el filtro proporcionado.
func (gc *client) GetFiles(files []string, extension string) ([]string, error) {
	var repoFiles []string

	if len(files) == 0 {
		// Solo obtener el worktree si no se proporcionan archivos en `files`
		worktree, err := gc.repo.Worktree()
		if err != nil {
			return nil, fmt.Errorf("error obteniendo el worktree: %w", err)
		}

		// Recorrer el sistema de archivos y filtrar los archivos
		err = filepath.Walk(worktree.Filesystem.Root(), func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return fmt.Errorf("error accediendo al archivo %s: %w", path, err)
			}
			// Si no es un directorio y tiene la extensión deseada, agregar el archivo a la lista.
			if !info.IsDir() && filepath.Ext(path) == extension {
				relPath, err := filepath.Rel(worktree.Filesystem.Root(), path)
				if err != nil {
					return fmt.Errorf("error obteniendo la ruta relativa para %s: %w", path, err)
				}
				repoFiles = append(repoFiles, relPath)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("error recorriendo el árbol de archivos: %w", err)
		}
	} else {
		// Filtrar los archivos proporcionados
		for _, file := range files {
			if filepath.Ext(file) == extension {
				repoFiles = append(repoFiles, file)
			}
		}
	}

	return repoFiles, nil
}

// GetFileAuthor obtiene el autor del último commit que modificó el archivo especificado.
func (gc *client) GetFileAuthor(file string) (string, error) {
	commits, err := gc.repo.Log(&git.LogOptions{FileName: &file})
	if err != nil {
		return "", fmt.Errorf("error getting log for file %s: %w", file, err)
	}

	commit, err := commits.Next()
	if err != nil {
		return "", fmt.Errorf("error getting commit for file %s: %w", file, err)
	}

	return commit.Author.Email, nil
}

// GetCommitID retorna el ID del último commit de un archivo.
func (gc *client) GetCommitID(file string) (string, error) {
	commits, err := gc.repo.Log(&git.LogOptions{FileName: &file})
	if err != nil {
		return "", fmt.Errorf("error getting log for file %s: %w", file, err)
	}

	commit, err := commits.Next()
	if err != nil {
		return "", fmt.Errorf("error getting commit for file %s: %w", file, err)
	}

	return commit.Hash.String(), nil
}

// GetRepo abre y retorna un repositorio en la ruta especificada.
func (gc *client) GetRepo(repoPath string) (*git.Repository, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("error opening repository at %s: %w", repoPath, err)
	}
	return repo, nil
}
