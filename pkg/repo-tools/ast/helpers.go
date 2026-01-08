package pkgast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strings"

	defs "github.com/alphacodinggroup/ponti-backend/pkg/repo-tools/ast/defs"
	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// Función auxiliar para parsear un archivo Go y obtener el AST.
func (s *service) parseFile(filePath string, mode parser.Mode) (*ast.File, *token.FileSet, error) {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("error reading file: %w", err)
	}
	node, err := parser.ParseFile(fset, filePath, src, mode)
	if err != nil {
		return nil, nil, fmt.Errorf("error parsing file: %w", err)
	}
	return node, fset, nil
}

// collectNodes es una función genérica para recolectar nodos AST basados en un filtro.
func collectNodes[T any](node ast.Node, filter func(ast.Node) (T, bool)) ([]T, error) {
	var results []T
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}
		if result, ok := filter(n); ok {
			results = append(results, result)
		}
		return true
	})
	return results, nil
}

// getdefs.ParameterInfo es una función auxiliar para obtener información de parámetros.
func getParameterInfo(fl *ast.FieldList) []defs.ParameterInfo {
	var params []defs.ParameterInfo
	if fl == nil {
		return params
	}
	for _, field := range fl.List {
		typeStr := exprToString(field.Type)
		if len(field.Names) == 0 {
			// Parámetro anónimo
			params = append(params, defs.ParameterInfo{
				Name: "",
				Type: typeStr,
			})
		} else {
			for _, name := range field.Names {
				params = append(params, defs.ParameterInfo{
					Name: name.Name,
					Type: typeStr,
				})
			}
		}
	}
	return params
}

// exprToString es una función auxiliar para convertir ast.Expr a string.
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.FuncType:
		return "func"
	case *ast.InterfaceType:
		return "any"
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	default:
		return fmt.Sprintf("%T", t)
	}
}

// getVariableType obtiene el tipo de una variable.
func (s *service) getVariableType(expr ast.Expr, name *ast.Ident, pkg *packages.Package, imdefs map[string]string) string {
	if expr != nil {
		return s.getTypeFromAST(expr, imdefs)
	}
	return pkg.TypesInfo.ObjectOf(name).Type().String()
}

// getKindFromObj devuelve el tipo (struct, interface) de un objeto.
func (s *service) getKindFromObj(obj types.Object) string {
	if obj == nil {
		return "unknown"
	}
	switch obj.Type().Underlying().(type) {
	case *types.Interface:
		return "interface"
	case *types.Struct:
		return "struct"
	}
	return "other"
}

// getTypeFromAST obtiene el tipo desde una expresión AST.
func (s *service) getTypeFromAST(expr ast.Expr, imdefs map[string]string) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.SelectorExpr:
		if pkgIdent, ok := t.X.(*ast.Ident); ok {
			if pkgPath, exists := imdefs[pkgIdent.Name]; exists {
				return fmt.Sprintf("%s.%s", pkgPath, t.Sel.Name)
			}
			return fmt.Sprintf("%s.%s", pkgIdent.Name, t.Sel.Name)
		}
	case *ast.StarExpr:
		return "*" + s.getTypeFromAST(t.X, imdefs)
	case *ast.ArrayType:
		return "[]" + s.getTypeFromAST(t.Elt, imdefs)
	}
	return "unknown"
}

// extractImdefs extrae las importaciones del archivo.
func (s *service) extractImdefs(file *ast.File) map[string]string {
	imdefs := make(map[string]string)
	for _, i := range file.Imports {
		alias := ""
		if i.Name != nil {
			alias = i.Name.Name
		} else {
			path := strings.Trim(i.Path.Value, "\"")
			parts := strings.Split(path, "/")
			alias = parts[len(parts)-1]
		}
		imdefs[alias] = strings.Trim(i.Path.Value, "\"")
	}
	return imdefs
}

// getFileFromPackage obtiene el archivo AST desde el paquete.
func (s *service) getFileFromPackage(pkg *packages.Package, filePath string) (*ast.File, error) {
	for _, file := range pkg.Syntax {
		if pkg.Fset.Position(file.Pos()).Filename == filePath {
			return file, nil
		}
	}
	return nil, fmt.Errorf("file not found in package")
}

// getParentNode es una función auxiliar para obtener el nodo padre de un nodo dado.
func (s *service) getParentNode(root ast.Node, child ast.Node) ast.Node {
	var parent ast.Node
	astutil.Apply(root, func(c *astutil.Cursor) bool {
		if c.Node() == child {
			parent = c.Parent()
			return false
		}
		return true
	}, nil)
	return parent
}
