package pkgutils

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// FileNameAndExtension separa el nombre y la extensión de un archivo dado su path.
// Recibe como parámetro una ruta de archivo (`filePath`) y devuelve tres valores:
//   - El nombre del archivo sin extensión
//   - La extensión del archivo
//   - Un posible error
func FileNameAndExtension(filePath string) (string, string, error) {
	fileName := filepath.Base(filePath)

	// Verificar si el archivo comienza con un punto y no tiene una extensión real
	if strings.HasPrefix(fileName, ".") && strings.Count(fileName, ".") == 1 {
		// Caso donde el archivo es algo como ".env", ".gitignore", etc.
		return fileName, "", nil
	}

	// Obtener la extensión del archivo
	fileExtension := filepath.Ext(fileName)

	// Si no tiene extensión (archivo sin punto o solo con nombre), devolvemos error
	if fileExtension == "" {
		return "", "", fmt.Errorf("file %s has no extension", fileName)
	}

	// Eliminar el punto de la extensión
	fileExtension = strings.TrimPrefix(fileExtension, ".")

	// Obtener el nombre del archivo sin la extensión
	fileNameWithoutExt := strings.TrimSuffix(fileName, "."+fileExtension)

	// Si el archivo no tiene un nombre válido, devolver error
	if fileNameWithoutExt == "" {
		return "", "", fmt.Errorf("invalid file name for file %s", fileName)
	}

	return fileNameWithoutExt, fileExtension, nil
}

// IsEnvFile verifica si un archivo es un .env
func IsEnvFile(filePath string) bool {
	return strings.HasSuffix(filePath, ".env")
}

// FilesFinder busca los archivos especificados en fileNames, interpretándolos como
// rutas relativas al directorio raíz del proyecto que está utilizando el pkg.
// Es decir, tienes que indicar en qué subdirectorio buscar con respecto a la raíz del proyecto.
// Retorna un slice con las rutas absolutas de los archivos encontrados.
func FilesFinder(fileNames ...string) ([]string, error) {
	// Obtener el directorio raíz del proyecto que está utilizando el pkg
	rootDir, err := getProjectRootDir()
	if err != nil {
		return nil, fmt.Errorf("error finding project root directory: %w", err)
	}

	var foundFiles []string
	for _, relativePath := range fileNames {
		// Construir la ruta absoluta del archivo
		absPath := filepath.Join(rootDir, relativePath)

		// Verificar si el archivo existe
		if _, err := os.Stat(absPath); os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", absPath)
		} else if err != nil {
			return nil, fmt.Errorf("error accessing file %s: %w", absPath, err)
		}

		// Agregar la ruta del archivo al slice de archivos encontrados
		foundFiles = append(foundFiles, absPath)
	}

	return foundFiles, nil
}

// getProjectRootDir encuentra el directorio raíz del proyecto buscando primero '.git' y luego 'go.mod'.
// Prioriza la detección del directorio raíz basado en el control de versiones (Git).
func getProjectRootDir() (string, error) {
	// Si se define APP_ROOT, úsala
	if pr := os.Getenv("APP_ROOT"); pr != "" {
		return pr, nil
	}

	// Obtener el archivo del código que llama al pkg
	callerPath, err := getCallerPath()
	if err != nil {
		return "", fmt.Errorf("error obtaining caller path: %w", err)
	}

	dir := filepath.Dir(callerPath)
	for {
		// Ruta al directorio .git
		gitDir := filepath.Join(dir, ".git")
		// Ruta al archivo go.mod
		goModPath := filepath.Join(dir, "go.mod")

		// Verifica si existe el directorio .git
		if stat, err := os.Stat(gitDir); err == nil && stat.IsDir() {
			return dir, nil
		}

		// Verifica si existe el archivo go.mod
		if stat, err := os.Stat(goModPath); err == nil && !stat.IsDir() {
			// **Opcional:** Si deseas detener la búsqueda al encontrar el primer go.mod,
			// descomenta la siguiente línea:
			// return dir, nil

			// **Recomendación:** En un monorepo con múltiples go.mod, es mejor continuar buscando .git
			// para asegurarse de identificar la raíz real del monorepo.
		}

		// Mueve al directorio padre.
		parent := filepath.Dir(dir)
		if parent == dir {
			// Se ha llegado al directorio raíz del sistema de archivos sin encontrar .git ni go.mod.
			return "", fmt.Errorf("could not find project root directory from %s", callerPath)
		}
		dir = parent
	}
}

// getCallerPath utiliza runtime.Caller para obtener la ruta del archivo que está llamando al pkg.
// Itera sobre las llamadas de pila para encontrar el archivo que no pertenece al paquete estándar o módulos.
func getCallerPath() (string, error) {
	for skip := 2; skip < 10; skip++ {
		_, file, _, ok := runtime.Caller(skip)
		if !ok {
			break
		}

		// Excluye archivos dentro de módulos de Go y el directorio estándar.
		if !strings.Contains(file, "/pkg/mod/") && !strings.Contains(file, "/go/src/") {
			return file, nil
		}
	}
	return "", fmt.Errorf("could not determine caller path")
}
