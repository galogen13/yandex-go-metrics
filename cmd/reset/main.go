package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/galogen13/yandex-go-metrics/internal/logger"
	"go.uber.org/zap"
)

const resetGenerateComment = "// generate:reset"

const templateStr = `
// Code generated automaticaly by reset utility; DO NOT EDIT.

package {{.Name}}

{{range .StructsInfo}}
func (v *{{.Name}}) Reset() {
	{{range .Fields}}
    	{{.ResetCode}}
	{{end}}
} 
{{end}}
`

type Field struct {
	Name      string
	ResetCode string
}

type StructInfo struct {
	Name          string
	Package       string
	FilePath      string
	OutputPath    string
	HasUsersReset bool // Флаг, есть ли уже пользовательский метод Reset
	Fields        []Field
}

type packageInfo struct {
	Name        string
	StructsInfo []StructInfo
}

func main() {

	if err := logger.Initialize("debug"); err != nil {
		log.Fatalf("Cannot Initialize logger: %v\n", err)
	}
	defer logger.Log.Sync()

	rootDir, err := findProjectRoot()
	if err != nil {
		log.Fatalf("Cannot found project root directory: %v\n", err)
	}

	// Сканируем все пакеты и находим структуры с комментарием // generate:reset
	structs, err := scanPackages(rootDir)
	if err != nil {
		log.Fatalf("Cannot scan packages: %v\n", err)
	}

	if len(structs) == 0 {
		logger.Log.Info("No structs with comment", zap.String("comment", resetGenerateComment))
		return
	}

	structsByPackage := groupByPackage(structs)

	for pkgPath, pkgStructs := range structsByPackage {
		if err := generateResetFile(pkgPath, pkgStructs); err != nil {
			logger.Log.Error("Cannot generate resetter for package", zap.String("package path", pkgPath), zap.Error(err))
		}
	}

	logger.Log.Info("Generation completed")
}

// findProjectRoot ищет корневую директорию проекта (где находится go.mod)
func findProjectRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("go.mod не найден")
		}
		dir = parent
	}
}

// scanPackages сканирует все пакеты и находит структуры с комментарием // generate:reset
func scanPackages(rootDir string) ([]StructInfo, error) {
	var structs []StructInfo

	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			name := info.Name()
			if strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
		}

		if !strings.HasSuffix(info.Name(), ".go") ||
			strings.HasSuffix(info.Name(), "_test.go") ||
			strings.HasSuffix(info.Name(), ".gen.go") ||
			strings.HasSuffix(info.Name(), ".mock.go") {
			return nil
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
		if err != nil {
			return nil
		}

		packageName := node.Name.Name

		ast.Inspect(node, func(n ast.Node) bool {
			genDecl, ok := n.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.TYPE {
				return true
			}

			for _, spec := range genDecl.Specs {
				if typeSpec, ok := spec.(*ast.TypeSpec); ok {
					if s, isStruct := typeSpec.Type.(*ast.StructType); isStruct {
						if hasGenerateResetComment(genDecl) {
							fields := []Field{}
							for _, field := range s.Fields.List {
								if len(field.Names) == 0 {
									continue
								}
								fields = append(fields,
									Field{
										Name:      field.Names[0].Name,
										ResetCode: getResetCode(field.Names[0].Name, field.Type),
									})

							}

							structInfo := StructInfo{
								Name:       typeSpec.Name.Name,
								Package:    packageName,
								FilePath:   path,
								OutputPath: filepath.Dir(path),
								Fields:     fields,
							}
							structInfo.HasUsersReset = hasResetMethod(node, typeSpec.Name.Name)

							structs = append(structs, structInfo)
						}
					}
				}
			}
			return true
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	return structs, nil
}

// hasGenerateResetComment проверяет наличие комментария // generate:reset над структурой
func hasGenerateResetComment(genDecl *ast.GenDecl) bool {
	if genDecl.Doc == nil {
		return false
	}

	for _, comment := range genDecl.Doc.List {
		if strings.Contains(comment.Text, resetGenerateComment) {
			return true
		}
	}
	return false
}

// hasResetMethod проверяет, существует ли уже метод Reset для структуры
func hasResetMethod(file *ast.File, structName string) bool {
	for _, decl := range file.Decls {
		funcDecl, ok := decl.(*ast.FuncDecl)
		if !ok {
			continue
		}

		if funcDecl.Name.Name != "Reset" {
			continue
		}

		if funcDecl.Recv == nil || len(funcDecl.Recv.List) == 0 {
			continue
		}

		recvType := funcDecl.Recv.List[0].Type

		var typeName string
		switch t := recvType.(type) {
		case *ast.StarExpr:
			if ident, ok := t.X.(*ast.Ident); ok {
				typeName = ident.Name
			}
		case *ast.Ident:
			typeName = t.Name
		}

		if typeName == structName {
			return true
		}
	}
	return false
}

// groupByPackage группирует структуры по директории вывода
func groupByPackage(structs []StructInfo) map[string][]StructInfo {
	result := make(map[string][]StructInfo)
	for _, s := range structs {
		result[s.OutputPath] = append(result[s.OutputPath], s)
	}
	return result
}

var tmpl = template.Must(template.New("reset").Parse(templateStr))

// generateResetFile создает файл reset.gen.go для пакета
func generateResetFile(pkgPath string, structs []StructInfo) error {
	outputFile := filepath.Join(pkgPath, "reset.gen.go")

	packageName := structs[0].Package

	logger.Log.Info("Generating reset methods",
		zap.String("package", packageName),
		zap.String("file dest", outputFile),
		zap.Int("count", len(structs)))

	var buf bytes.Buffer
	err := tmpl.Execute(&buf, packageInfo{Name: packageName, StructsInfo: structs})
	if err != nil {
		return fmt.Errorf("template execution error: %w", err)
	}

	bufFmt, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf("formatting go-file error: %w", err)
	}

	err = os.WriteFile(outputFile, bufFmt, 0644)
	if err != nil {
		return fmt.Errorf("writing file error: %w", err)
	}

	return nil
}

func getResetCode(name string, expr ast.Expr) string {

	switch t := expr.(type) {
	case *ast.Ident:
		return "v." + name + " = " + getZeroValue(t.Name)
	case *ast.StarExpr:
		var buf bytes.Buffer
		buf.WriteString("if v." + name + " != nil {\n")
		buf.WriteString("\t*" + getResetCode(name, t.X) + "\n")
		buf.WriteString("}\n")
		return buf.String()
	case *ast.ArrayType:
		return "v." + name + " = v." + name + "[:0]"
		// if t.Len == nil {
		// 	return "[]" + getTypeName(t.Elt)
		// }
		// // Для массивов с фиксированной длиной
		// return fmt.Sprintf("[%s]%s", t.Len, getTypeName(t.Elt))
	case *ast.SelectorExpr:
		if ident, ok := t.X.(*ast.Ident); ok {
			return "v." + name + " = " + getZeroValue(ident.Name+"."+t.Sel.Name)
		}
		return ""
	case *ast.MapType:
		return "clear(v." + name + ")"
	case *ast.StructType:
		var buf bytes.Buffer
		buf.WriteString("if resetter, ok := v." + name + ".(interface{ Reset() }); ok && v." + name + " != nil {\n")
		buf.WriteString("\tresetter.Reset()\n")
		buf.WriteString("}")
		return buf.String()
	default:
		return "v." + name + " = nil"
	}
}

// getZeroValue возвращает zero value для типа
func getZeroValue(typeName string) string {
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return "0"
	case "float32", "float64":
		return "0.0"
	case "bool":
		return "false"
	case "string":
		return "\"\""
	default:
		// Для сложных типов используем nil или пустую инициализацию
		if strings.HasPrefix(typeName, "[]") ||
			strings.HasPrefix(typeName, "map") ||
			strings.HasPrefix(typeName, "*") ||
			strings.HasPrefix(typeName, "chan") ||
			strings.HasPrefix(typeName, "func") {
			return "nil"
		}
		// Для структур и других типов
		return typeName + "{}"
	}
}
