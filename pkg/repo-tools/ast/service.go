package pkgast

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/callgraph"
	"golang.org/x/tools/go/callgraph/static"
	"golang.org/x/tools/go/packages"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"

	defs "github.com/alphacodinggroup/ponti-backend/pkg/repo-tools/ast/defs"
)

var (
	instance  defs.Service
	once      sync.Once
	initError error
)

type service struct {
	config defs.Config
}

// newService crea una nueva instancia del servicio AST.
func newService(config defs.Config) (defs.Service, error) {
	once.Do(func() {
		err := config.Validate()
		if err != nil {
			initError = err
			return
		}

		instance = &service{
			config: config,
		}
	})
	return instance, initError
}

// ReadImdefs analiza un archivo Go y devuelve los imdefs encontrados.
func (s *service) ReadImports(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var imdefs []string
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		imdefs = append(imdefs, importPath)
	}

	return imdefs, nil
}

// ReadFunctions analiza un archivo Go y devuelve las funciones encontradas.
func (s *service) ReadFunctions(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) (string, bool) {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv == nil {
			return fn.Name.Name, true
		}
		return "", false
	}

	return collectNodes(node, filter)
}

// ReadMethods analiza un archivo Go y devuelve los métodos encontrados.
func (s *service) ReadMethods(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) (string, bool) {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv != nil {
			return fn.Name.Name, true
		}
		return "", false
	}

	return collectNodes(node, filter)
}

// ReadStructs analiza un archivo Go y devuelve las structs encontradas.
func (s *service) ReadStructs(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) (string, bool) {
		if ts, ok := n.(*ast.TypeSpec); ok {
			if _, isStruct := ts.Type.(*ast.StructType); isStruct {
				return ts.Name.Name, true
			}
		}
		return "", false
	}

	return collectNodes(node, filter)
}

// ReadInterfaces analiza un archivo Go y devuelve las interfaces encontradas.
func (s *service) ReadInterfaces(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) (string, bool) {
		if ts, ok := n.(*ast.TypeSpec); ok {
			if _, isInterface := ts.Type.(*ast.InterfaceType); isInterface {
				return ts.Name.Name, true
			}
		}
		return "", false
	}

	return collectNodes(node, filter)
}

// ReadConstants analiza un archivo Go y devuelve las constantes encontradas.
func (s *service) ReadConstants(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) ([]string, bool) {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.CONST {
			var constants []string
			for _, spec := range genDecl.Specs {
				if valSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valSpec.Names {
						constants = append(constants, name.Name)
					}
				}
			}
			return constants, len(constants) > 0
		}
		return nil, false
	}

	constantLists, err := collectNodes(node, filter)
	if err != nil {
		return nil, err
	}

	// Aplanar la lista de listas
	var constants []string
	for _, list := range constantLists {
		constants = append(constants, list...)
	}

	return constants, nil
}

// ReadVariables analiza un archivo Go y devuelve las variables globales encontradas.
func (s *service) ReadVariables(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) ([]string, bool) {
		if genDecl, ok := n.(*ast.GenDecl); ok && genDecl.Tok == token.VAR {
			var variables []string
			for _, spec := range genDecl.Specs {
				if valSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valSpec.Names {
						variables = append(variables, name.Name)
					}
				}
			}
			return variables, len(variables) > 0
		}
		return nil, false
	}

	variableLists, err := collectNodes(node, filter)
	if err != nil {
		return nil, err
	}

	// Aplanar la lista de listas
	var variables []string
	for _, list := range variableLists {
		variables = append(variables, list...)
	}

	return variables, nil
}

// ReadTypeAliases analiza un archivo Go y devuelve los type aliases encontrados.
func (s *service) ReadTypeAliases(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) (string, bool) {
		if ts, ok := n.(*ast.TypeSpec); ok && ts.Assign != 0 {
			return ts.Name.Name, true
		}
		return "", false
	}

	return collectNodes(node, filter)
}

// ReadPackageName analiza un archivo Go y devuelve el nombre del paquete.
func (s *service) ReadPackageName(filePath string) (string, error) {
	node, _, err := s.parseFile(filePath, parser.PackageClauseOnly)
	if err != nil {
		return "", err
	}

	return node.Name.Name, nil
}

// CountStatements analiza un archivo Go y cuenta el número de declaraciones (statements).
func (s *service) CountStatements(filePath string) (int, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return 0, err
	}

	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		if _, ok := n.(ast.Stmt); ok {
			count++
		}
		return true
	})

	return count, nil
}

// ReadComments analiza un archivo Go y devuelve los comentarios encontrados.
func (s *service) ReadComments(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var comments []string
	for _, commentGroup := range node.Comments {
		comments = append(comments, commentGroup.Text())
	}

	return comments, nil
}

// ReadMethodsInfo analiza un archivo Go y devuelve información detallada de los métodos encontrados.
func (s *service) ReadMethodsInfo(filePath string) ([]defs.MethodInfo, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) (defs.MethodInfo, bool) {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv != nil {
			methodInfo := defs.MethodInfo{
				Name: fn.Name.Name,
			}
			if len(fn.Recv.List) > 0 {
				methodInfo.Receiver = exprToString(fn.Recv.List[0].Type)
			}
			methodInfo.InputParams = getParameterInfo(fn.Type.Params)
			methodInfo.OutputParams = getParameterInfo(fn.Type.Results)
			return methodInfo, true
		}
		return defs.MethodInfo{}, false
	}

	return collectNodes(node, filter)
}

// ReadFunctionsInfo analiza un archivo Go y devuelve información detallada de las funciones encontradas.
func (s *service) ReadFunctionsInfo(filePath string) ([]defs.FunctionInfo, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	filter := func(n ast.Node) (defs.FunctionInfo, bool) {
		if fn, ok := n.(*ast.FuncDecl); ok && fn.Recv == nil {
			funcInfo := defs.FunctionInfo{
				Name:         fn.Name.Name,
				InputParams:  getParameterInfo(fn.Type.Params),
				OutputParams: getParameterInfo(fn.Type.Results),
			}
			return funcInfo, true
		}
		return defs.FunctionInfo{}, false
	}

	return collectNodes(node, filter)
}

// ReadVariablesDetailed analiza un archivo Go y devuelve información detallada de las variables.
func (s *service) ReadVariablesDetailed(filePath string) ([]defs.VariableInfo, error) {
	cfg := &packages.Config{Mode: packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo, Dir: ""}
	pkgs, err := packages.Load(cfg, fmt.Sprintf("file=%s", filePath))
	if err != nil || len(pkgs) == 0 {
		return nil, fmt.Errorf("error loading package: %w", err)
	}

	var results []defs.VariableInfo
	pkg := pkgs[0]
	file, err := s.getFileFromPackage(pkg, filePath)
	if err != nil {
		return nil, err
	}

	imdefs := s.extractImdefs(file)

	ast.Inspect(file, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.GenDecl:
			s.processVariables(decl, pkg, imdefs, &results)
		case *ast.FuncDecl:
			s.processFunctionAssignments(decl, pkg, &results)
		}
		return true
	})

	return results, nil
}

// processVariables procesa las variables globales en el archivo.
func (s *service) processVariables(decl *ast.GenDecl, pkg *packages.Package, imdefs map[string]string, results *[]defs.VariableInfo) {
	for _, spec := range decl.Specs {
		if vspec, ok := spec.(*ast.ValueSpec); ok {
			for _, name := range vspec.Names {
				varType := s.getVariableType(vspec.Type, name, pkg, imdefs)
				isGlobal := true
				kind := s.getKindFromObj(pkg.TypesInfo.ObjectOf(name))

				pos := pkg.Fset.Position(name.Pos())
				*results = append(*results, defs.VariableInfo{
					Name:     name.Name,
					Type:     varType,
					Position: pos,
					IsGlobal: isGlobal,
					Kind:     kind,
				})
			}
		}
	}
}

// processFunctionAssignments procesa las variables locales dentro de funciones.
func (s *service) processFunctionAssignments(
	funcDecl *ast.FuncDecl,
	pkg *packages.Package,
	results *[]defs.VariableInfo,
) {
	fset := pkg.Fset
	if fset == nil {
		fset = token.NewFileSet()
	}

	ast.Inspect(funcDecl.Body, func(n ast.Node) bool {
		if assignStmt, ok := n.(*ast.AssignStmt); ok {
			for _, lhs := range assignStmt.Lhs {
				if ident, ok := lhs.(*ast.Ident); ok {
					obj := pkg.TypesInfo.ObjectOf(ident)
					if obj != nil {
						kind := s.getKindFromObj(obj)
						varType := obj.Type().String()
						isGlobal := false
						pos := fset.Position(ident.Pos())
						*results = append(*results, defs.VariableInfo{
							Name:     ident.Name,
							Type:     varType,
							Position: pos,
							IsGlobal: isGlobal,
							Kind:     kind,
						})
					}
				}
			}
		}
		return true
	})
}

// BuildCallGraph construye el grafo de llamadas entre funciones.
func (s *service) BuildCallGraph(filePath string) (*callgraph.Graph, error) {
	cfg := &packages.Config{
		Mode: packages.LoadAllSyntax, // Cargar la sintaxis completa
	}
	pkgs, err := packages.Load(cfg, filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading packages: %w", err)
	}

	// Construir el programa SSA
	prog, ssaPkgs := ssautil.AllPackages(pkgs, ssa.BuilderMode(0))
	if len(ssaPkgs) == 0 {
		return nil, fmt.Errorf("no SSA packages found")
	}
	prog.Build()

	// Crear el grafo de llamadas estático
	cg := static.CallGraph(prog)

	return cg, nil
}

// FindUnusedVariables encuentra variables no utilizadas en el código.
func (s *service) FindUnusedVariables(filePath string) ([]string, error) {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesInfo}
	pkgs, err := packages.Load(cfg, filePath)
	if err != nil {
		return nil, fmt.Errorf("error loading package: %w", err)
	}
	unused := []string{}
	for _, pkg := range pkgs {
		for ident, obj := range pkg.TypesInfo.Defs {
			if obj != nil && !obj.Exported() {
				if _, used := pkg.TypesInfo.Uses[ident]; !used {
					unused = append(unused, ident.Name)
				}
			}
		}
	}
	return unused, nil
}

// CalculateCyclomaticComplexity calcula la complejidad ciclomática de las funciones.
func (s *service) CalculateCyclomaticComplexity(filePath string) (map[string]int, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	complexities := make(map[string]int)

	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			complexity := 1
			ast.Inspect(fn.Body, func(n ast.Node) bool {
				switch n.(type) {
				case *ast.IfStmt, *ast.ForStmt, *ast.RangeStmt, *ast.CaseClause, *ast.CommClause:
					complexity++
				case *ast.BinaryExpr:
					if be, ok := n.(*ast.BinaryExpr); ok {
						if be.Op == token.LAND || be.Op == token.LOR {
							complexity++
						}
					}
				}
				return true
			})
			complexities[fn.Name.Name] = complexity
		}
		return true
	})

	return complexities, nil
}

// ExtractDocComments extrae los comentarios de documentación.
func (s *service) ExtractDocComments(filePath string) (map[string]string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	docs := make(map[string]string)

	for _, decl := range node.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Doc != nil {
				docs[d.Name.Name] = d.Doc.Text()
			}
		case *ast.GenDecl:
			if d.Doc != nil {
				for _, spec := range d.Specs {
					if ts, ok := spec.(*ast.TypeSpec); ok {
						docs[ts.Name.Name] = d.Doc.Text()
					}
				}
			}
		}
	}

	return docs, nil
}

// DetectCodeSmells analiza el código para encontrar code smells comunes.
func (s *service) DetectCodeSmells(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	smells := []string{}

	ast.Inspect(node, func(n ast.Node) bool {
		// Detectar structs con muchos campos
		if ts, ok := n.(*ast.TypeSpec); ok {
			if st, ok := ts.Type.(*ast.StructType); ok {
				if len(st.Fields.List) > 5 {
					smells = append(smells, fmt.Sprintf("Struct %s has too many fields (%d)", ts.Name.Name, len(st.Fields.List)))
				}
			}
		}
		// Detectar funciones con muchos parámetros
		if fn, ok := n.(*ast.FuncDecl); ok {
			if fn.Type.Params != nil && len(fn.Type.Params.List) > 3 {
				smells = append(smells, fmt.Sprintf("Function %s has too many parameters (%d)", fn.Name.Name, len(fn.Type.Params.List)))
			}
		}
		return true
	})

	return smells, nil
}

// FindImplementationsOfInterface encuentra todos los tipos que implementan una interfaz dada.
func (s *service) FindImplementationsOfInterface(filePath, interfaceName string) ([]string, error) {
	cfg := &packages.Config{Mode: packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports | packages.NeedSyntax}
	pkgs, err := packages.Load(cfg, filePath)
	if err != nil || len(pkgs) == 0 {
		return nil, fmt.Errorf("error loading package: %w", err)
	}
	pkg := pkgs[0]

	// Buscar la interfaz en el paquete
	var interfaceType types.Type
	for _, obj := range pkg.TypesInfo.Defs {
		if obj != nil && obj.Name() == interfaceName {
			if iface, ok := obj.Type().Underlying().(*types.Interface); ok {
				interfaceType = iface
				break
			}
		}
	}
	if interfaceType == nil {
		return nil, fmt.Errorf("interface %s not found", interfaceName)
	}

	// Encontrar todos los tipos que implementan la interfaz
	var implementations []string
	for _, objName := range pkg.Types.Scope().Names() {
		tObj := pkg.Types.Scope().Lookup(objName)
		if tObj == nil || !tObj.Exported() {
			continue
		}
		if _, ok := tObj.(*types.TypeName); !ok {
			continue
		}
		if types.Implements(tObj.Type(), interfaceType.Underlying().(*types.Interface)) {
			implementations = append(implementations, tObj.Name())
		}
	}
	return implementations, nil
}

// AnalyzePackageDependencies analiza las dependencias entre paquetes en un directorio dado.
func (s *service) AnalyzePackageDependencies(dirPath string) (map[string][]string, error) {
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedImports, Dir: dirPath}
	pkgs, err := packages.Load(cfg, "./...")
	if err != nil {
		return nil, fmt.Errorf("error loading packages: %w", err)
	}

	dependencies := make(map[string][]string)
	for _, pkg := range pkgs {
		var deps []string
		for importPath := range pkg.Imports {
			deps = append(deps, importPath)
		}
		dependencies[pkg.PkgPath] = deps
	}
	return dependencies, nil
}

// ExtractStringLiterals extrae todos los literales de cadena del código.
func (s *service) ExtractStringLiterals(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var strings []string
	ast.Inspect(node, func(n ast.Node) bool {
		if bl, ok := n.(*ast.BasicLit); ok && bl.Kind == token.STRING {
			strValue, err := strconv.Unquote(bl.Value)
			if err == nil {
				strings = append(strings, strValue)
			}
		}
		return true
	})
	return strings, nil
}

// DetectErrorHandlingPatterns analiza el manejo de errores en el código.
func (s *service) DetectErrorHandlingPatterns(filePath string) ([]string, error) {
	node, fset, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var unhandledErrors []string
	ast.Inspect(node, func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			// Obtener el tipo de retorno
			// Aquí podríamos utilizar información de tipos para verificar si retorna error
			// Por simplicidad, asumiremos que todas las funciones podrían retornar error
			// y verificaremos si el error es manejado
			parent := s.getParentNode(node, ce)
			if _, ok := parent.(*ast.ExprStmt); ok {
				pos := fset.Position(ce.Pos())
				unhandledErrors = append(unhandledErrors, fmt.Sprintf("Possible unhandled error at %s", pos))
			}
		}
		return true
	})
	return unhandledErrors, nil
}

// IdentifyReflectionAndUnsafeUsage detecta el uso de los paquetes reflect y unsafe.
func (s *service) IdentifyReflectionAndUnsafeUsage(filePath string) (bool, bool, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return false, false, err
	}

	var usesReflect, usesUnsafe bool
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		if importPath == "reflect" {
			usesReflect = true
		}
		if importPath == "unsafe" {
			usesUnsafe = true
		}
	}
	return usesReflect, usesUnsafe, nil
}

// AnalyzeGoroutineUsage analiza el uso de gorutinas en el código.
func (s *service) AnalyzeGoroutineUsage(filePath string) (int, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return 0, err
	}

	count := 0
	ast.Inspect(node, func(n ast.Node) bool {
		if _, ok := n.(*ast.GoStmt); ok {
			count++
		}
		return true
	})
	return count, nil
}

// ExtractBuildTags extrae las etiquetas de compilación condicional (build tags) del archivo.
func (s *service) ExtractBuildTags(filePath string) ([]string, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}
	content := string(src)
	var tags []string
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "// +build") || strings.HasPrefix(line, "//go:build") {
			tag := strings.TrimPrefix(line, "// +build")
			tag = strings.TrimPrefix(tag, "//go:build")
			tags = append(tags, strings.TrimSpace(tag))
		} else if line == "" || strings.HasPrefix(line, "//") {
			continue
		} else {
			break
		}
	}
	return tags, nil
}

// DetectDeprecatedFunctions detecta el uso de funciones o métodos marcados como obsoletos (Deprecated).
func (s *service) DetectDeprecatedFunctions(filePath string) ([]string, error) {
	node, fset, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var deprecatedUsages []string

	// Primero, construir un mapa de funciones o métodos deprecados
	deprecatedFuncs := make(map[string]bool)
	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			if decl.Doc != nil {
				for _, comment := range decl.Doc.List {
					if strings.Contains(comment.Text, "Deprecated") {
						deprecatedFuncs[decl.Name.Name] = true
					}
				}
			}
		}
		return true
	})

	// Luego, buscar llamadas a esas funciones
	ast.Inspect(node, func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if ident, ok := ce.Fun.(*ast.Ident); ok {
				if deprecatedFuncs[ident.Name] {
					pos := fset.Position(ce.Pos())
					deprecatedUsages = append(deprecatedUsages, fmt.Sprintf("Deprecated function %s used at %s", ident.Name, pos))
				}
			}
		}
		return true
	})

	return deprecatedUsages, nil
}

// GenerateUMLClassDiagram genera una representación simplificada para diagramas UML.
func (s *service) GenerateUMLClassDiagram(filePath string) (string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return "", err
	}

	var umlStrings []string
	umlStrings = append(umlStrings, "@startuml")
	ast.Inspect(node, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			switch t := ts.Type.(type) {
			case *ast.StructType:
				umlStrings = append(umlStrings, fmt.Sprintf("class %s {", ts.Name.Name))
				for _, field := range t.Fields.List {
					fieldType := exprToString(field.Type)
					if len(field.Names) > 0 {
						for _, name := range field.Names {
							umlStrings = append(umlStrings, fmt.Sprintf("  %s %s", fieldType, name.Name))
						}
					} else {
						umlStrings = append(umlStrings, fmt.Sprintf("  %s", fieldType))
					}
				}
				umlStrings = append(umlStrings, "}")
			case *ast.InterfaceType:
				umlStrings = append(umlStrings, fmt.Sprintf("interface %s {", ts.Name.Name))
				for _, method := range t.Methods.List {
					methodType := exprToString(method.Type)
					if len(method.Names) > 0 {
						for _, name := range method.Names {
							umlStrings = append(umlStrings, fmt.Sprintf("  %s %s", methodType, name.Name))
						}
					}
				}
				umlStrings = append(umlStrings, "}")
			}
		}
		return true
	})
	umlStrings = append(umlStrings, "@enduml")
	return strings.Join(umlStrings, "\n"), nil
}

// IdentifyMagicNumbers encuentra números mágicos en el código.
func (s *service) IdentifyMagicNumbers(filePath string) ([]string, error) {
	node, fset, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var magicNumbers []string
	ast.Inspect(node, func(n ast.Node) bool {
		if bl, ok := n.(*ast.BasicLit); ok && bl.Kind == token.INT {
			// Excluir valores comunes como 0 y 1
			if bl.Value != "0" && bl.Value != "1" {
				pos := fset.Position(bl.Pos())
				magicNumbers = append(magicNumbers, fmt.Sprintf("Magic number %s at %s", bl.Value, pos))
			}
		}
		return true
	})
	return magicNumbers, nil
}

// CheckFormatting verifica si el código está formateado correctamente.
func (s *service) CheckFormatting(filePath string) (bool, error) {
	src, err := os.ReadFile(filePath)
	if err != nil {
		return false, fmt.Errorf("error reading file: %w", err)
	}
	formattedSrc, err := format.Source(src)
	if err != nil {
		return false, fmt.Errorf("error formatting source: %w", err)
	}
	isFormatted := bytes.Equal(src, formattedSrc)
	return isFormatted, nil
}

// ExtractStructTags extrae los tags de los campos de las estructuras.
func (s *service) ExtractStructTags(filePath string) (map[string]map[string]string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	structTags := make(map[string]map[string]string)
	ast.Inspect(node, func(n ast.Node) bool {
		if ts, ok := n.(*ast.TypeSpec); ok {
			if st, ok := ts.Type.(*ast.StructType); ok {
				fields := make(map[string]string)
				for _, field := range st.Fields.List {
					if field.Tag != nil {
						tag := strings.Trim(field.Tag.Value, "`")
						if len(field.Names) > 0 {
							for _, name := range field.Names {
								fields[name.Name] = tag
							}
						} else {
							fields[""] = tag // Campo anónimo
						}
					}
				}
				structTags[ts.Name.Name] = fields
			}
		}
		return true
	})
	return structTags, nil
}

// IdentifyPanicsAndRecovers detecta el uso de panic y recover en el código.
func (s *service) IdentifyPanicsAndRecovers(filePath string) ([]string, error) {
	node, fset, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var panicsAndRecovers []string
	ast.Inspect(node, func(n ast.Node) bool {
		if ce, ok := n.(*ast.CallExpr); ok {
			if ident, ok := ce.Fun.(*ast.Ident); ok {
				if ident.Name == "panic" || ident.Name == "recover" {
					pos := fset.Position(ce.Pos())
					panicsAndRecovers = append(panicsAndRecovers, fmt.Sprintf("%s called at %s", ident.Name, pos))
				}
			}
		}
		return true
	})
	return panicsAndRecovers, nil
}

// AnalyzeTestsAndBenchmarks analiza las funciones de prueba y benchmarks.
func (s *service) AnalyzeTestsAndBenchmarks(filePath string) ([]string, error) {
	node, fset, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var testsAndBenchmarks []string
	ast.Inspect(node, func(n ast.Node) bool {
		if fn, ok := n.(*ast.FuncDecl); ok {
			if strings.HasPrefix(fn.Name.Name, "Test") || strings.HasPrefix(fn.Name.Name, "Benchmark") {
				pos := fset.Position(fn.Pos())
				testsAndBenchmarks = append(testsAndBenchmarks, fmt.Sprintf("%s found at %s", fn.Name.Name, pos))
			}
		}
		return true
	})
	return testsAndBenchmarks, nil
}

// IdentifyGlobalVariables identifica variables globales en el paquete.
func (s *service) IdentifyGlobalVariables(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var globalVars []string
	for _, decl := range node.Decls {
		if genDecl, ok := decl.(*ast.GenDecl); ok && (genDecl.Tok == token.VAR || genDecl.Tok == token.CONST) {
			for _, spec := range genDecl.Specs {
				if valSpec, ok := spec.(*ast.ValueSpec); ok {
					for _, name := range valSpec.Names {
						globalVars = append(globalVars, name.Name)
					}
				}
			}
		}
	}
	return globalVars, nil
}

// AnalyzeThirdPartyPackages analiza el uso de paquetes de terceros.
func (s *service) AnalyzeThirdPartyPackages(filePath string) ([]string, error) {
	node, _, err := s.parseFile(filePath, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var thirdPartyImdefs []string
	for _, imp := range node.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		if !s.isStandardLibrary(importPath) {
			thirdPartyImdefs = append(thirdPartyImdefs, importPath)
		}
	}
	return thirdPartyImdefs, nil
}

// isStandardLibrary verifica si una importación es de la biblioteca estándar.
func (s *service) isStandardLibrary(importPath string) bool {
	// Simples comprobaciones para determinar si es estándar.
	return !strings.Contains(importPath, ".")
}

// ReadDocumentation analiza un archivo Go y devuelve la documentación del paquete.
func (s *service) ReadDocumentation(filePath string) (string, error) {
	node, _, err := s.parseFile(filePath, parser.ParseComments)
	if err != nil {
		return "", err
	}

	if node.Doc != nil {
		return node.Doc.Text(), nil
	}
	return "", nil
}
