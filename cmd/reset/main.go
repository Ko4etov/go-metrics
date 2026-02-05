package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type StructInfo struct {
    PackageName string
    StructName  string
    Fields      []FieldInfo
    FilePath    string
}

type FieldInfo struct {
    Name     string
    Type     string
    IsPtr    bool
    IsSlice  bool
    IsMap    bool
    IsStruct bool
}

func main() {
    rootDir := "."
    
    structsByPackage := make(map[string][]StructInfo)
    
    err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        
        if info.IsDir() {
            base := filepath.Base(path)
            if strings.HasPrefix(base, ".") && base != "." && base != ".." {
                return filepath.SkipDir
            }
            return nil
        }
        
        if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
            return nil
        }
        
        structs, err := parseFile(path)
        if err != nil {
            fmt.Printf("Ошибка при парсинге файла %s: %v\n", path, err)
            return nil
        }
        
        if len(structs) > 0 {
            relPath, _ := filepath.Rel(rootDir, filepath.Dir(path))
            if relPath == "." {
                relPath = ""
            }
            
            for _, s := range structs {
                s.FilePath = path
                structsByPackage[relPath] = append(structsByPackage[relPath], s)
            }
        }
        
        return nil
    })
    
    if err != nil {
        fmt.Printf("Ошибка при обходе директорий: %v\n", err)
        os.Exit(1)
    }
    
    for pkgPath, structs := range structsByPackage {
        if err := generateResetFile(pkgPath, structs); err != nil {
            fmt.Printf("Ошибка при генерации файла для пакета %s: %v\n", pkgPath, err)
        }
    }
}

func parseFile(filename string) ([]StructInfo, error) {
    fset := token.NewFileSet()
    node, err := parser.ParseFile(fset, filename, nil, parser.ParseComments)
    if err != nil {
        return nil, err
    }
    
    var structs []StructInfo
    
    packageName := node.Name.Name
    
    ast.Inspect(node, func(n ast.Node) bool {
        genDecl, ok := n.(*ast.GenDecl)
        if !ok || genDecl.Tok != token.TYPE {
            return true
        }
        
        if genDecl.Doc == nil {
            return true
        }
        
        hasResetComment := false
        for _, comment := range genDecl.Doc.List {
            if strings.Contains(comment.Text, "generate:reset") {
                hasResetComment = true
                break
            }
        }
        
        if !hasResetComment {
            return true
        }
        
        for _, spec := range genDecl.Specs {
            typeSpec, ok := spec.(*ast.TypeSpec)
            if !ok {
                continue
            }
            
            structType, ok := typeSpec.Type.(*ast.StructType)
            if !ok {
                continue
            }
            
            structInfo := StructInfo{
                PackageName: packageName,
                StructName:  typeSpec.Name.Name,
            }
            
            if structType.Fields != nil {
                for _, field := range structType.Fields.List {
                    if len(field.Names) == 0 {
                        continue
                    }
                    
                    for _, name := range field.Names {
                        fieldInfo := FieldInfo{
                            Name: name.Name,
                        }
                        
                        fieldInfo.Type = getTypeString(field.Type)
                        fieldInfo.IsPtr = isPointerType(field.Type)
                        fieldInfo.IsSlice = isSliceType(field.Type)
                        fieldInfo.IsMap = isMapType(field.Type)
                        fieldInfo.IsStruct = isStructType(field.Type)
                        
                        structInfo.Fields = append(structInfo.Fields, fieldInfo)
                    }
                }
            }
            
            structs = append(structs, structInfo)
        }
        
        return true
    })
    
    return structs, nil
}

func getTypeString(expr ast.Expr) string {
    switch t := expr.(type) {
    case *ast.Ident:
        return t.Name
    case *ast.StarExpr:
        return "*" + getTypeString(t.X)
    case *ast.ArrayType:
        if t.Len == nil {
            return "[]" + getTypeString(t.Elt)
        }
        return "[" + getTypeString(t.Len) + "]" + getTypeString(t.Elt)
    case *ast.MapType:
        return "map[" + getTypeString(t.Key) + "]" + getTypeString(t.Value)
    case *ast.ChanType:
        if t.Dir == ast.SEND {
            return "chan<- " + getTypeString(t.Value)
        } else if t.Dir == ast.RECV {
            return "<-chan " + getTypeString(t.Value)
        }
        return "chan " + getTypeString(t.Value)
    case *ast.SelectorExpr:
        return getTypeString(t.X) + "." + t.Sel.Name
    default:
        return fmt.Sprintf("%T", expr)
    }
}

func isPointerType(expr ast.Expr) bool {
    _, ok := expr.(*ast.StarExpr)
    return ok
}

func isSliceType(expr ast.Expr) bool {
    arrType, ok := expr.(*ast.ArrayType)
    if !ok {
        return false
    }
    return arrType.Len == nil
}

func isMapType(expr ast.Expr) bool {
    _, ok := expr.(*ast.MapType)
    return ok
}

func isStructType(expr ast.Expr) bool {
	switch expr.(type) {
    case *ast.ChanType, *ast.MapType, *ast.ArrayType:
        return false
    }

    if ident, ok := expr.(*ast.Ident); ok {
        switch ident.Name {
        case "int", "int8", "int16", "int32", "int64",
             "uint", "uint8", "uint16", "uint32", "uint64",
             "float32", "float64", "complex64", "complex128",
             "byte", "rune", "string", "bool", "error":
            return false
        }
        return true
    }
    
    if star, ok := expr.(*ast.StarExpr); ok {
        return isStructType(star.X)
    }
    
    return false
}

func generateResetFile(pkgPath string, structs []StructInfo) error {
    var targetDir string
    if pkgPath == "" {
        targetDir = "."
    } else {
        targetDir = pkgPath
    }
    
    targetFile := filepath.Join(targetDir, "reset.gen.go")
    
    const resetTemplate = `// Code generated by reset tool; DO NOT EDIT.

package {{.PackageName}}

{{range .Structs}}
func ({{.ReceiverName}} *{{.StructName}}) Reset() {
    if {{.ReceiverName}} == nil {
        return
    }
    {{range .Fields}}
    // Поле: {{.Name}} ({{.Type}})
    {{if .IsPtr}}
    if {{.ReceiverName}}.{{.Name}} != nil {
        {{if .IsStruct}}
        if resetter, ok := interface{}({{.ReceiverName}}.{{.Name}}).(interface{ Reset() }); ok {
            resetter.Reset()
        } else {
            *{{.ReceiverName}}.{{.Name}} = {{.BaseType}}{}
        }
        {{else}}
        *{{.ReceiverName}}.{{.Name}} = {{.BaseType}}{}
        {{end}}
    }
    {{else if .IsSlice}}
    {{.ReceiverName}}.{{.Name}} = {{.ReceiverName}}.{{.Name}}[:0]
    {{else if .IsMap}}
    clear({{.ReceiverName}}.{{.Name}})
    {{else if .IsStruct}}
    if resetter, ok := interface{}({{.ReceiverName}}.{{.Name}}).(interface{ Reset() }); ok {
        resetter.Reset()
    } else {
        {{.ReceiverName}}.{{.Name}} = {{.Type}}{}
    }
    {{else}}
    {{.ReceiverName}}.{{.Name}} = {{.ZeroValue}}
    {{end}}
    {{end}}
}
{{end}}`

    type TemplateField struct {
        Name         string
        Type         string
        BaseType     string
        IsPtr        bool
        IsSlice      bool
        IsMap        bool
        IsStruct     bool
        ReceiverName string
        ZeroValue    string // Добавляем поле для нулевого значения
    }
    
    type TemplateStruct struct {
        StructName   string
        ReceiverName string
        Fields       []TemplateField
    }
    
    type TemplateData struct {
        PackageName string
        Structs     []TemplateStruct
    }
    
    if len(structs) == 0 {
        return nil
    }
    
    var templateStructs []TemplateStruct
    for _, s := range structs {
        receiverName := strings.ToLower(string(s.StructName[0]))
        if receiverName == "" {
            receiverName = "rs"
        }
        
        var templateFields []TemplateField
        for _, field := range s.Fields {
            // Определяем нулевое значение для типа
            zeroValue := getZeroValue(field.Type)
            
            tf := TemplateField{
                Name:         field.Name,
                Type:         field.Type,
                BaseType:     strings.TrimPrefix(field.Type, "*"),
                IsPtr:        field.IsPtr,
                IsSlice:      field.IsSlice,
                IsMap:        field.IsMap,
                IsStruct:     field.IsStruct,
                ReceiverName: receiverName,
                ZeroValue:    zeroValue,
            }
            templateFields = append(templateFields, tf)
        }
        
        templateStructs = append(templateStructs, TemplateStruct{
            StructName:   s.StructName,
            ReceiverName: receiverName,
            Fields:       templateFields,
        })
    }
    
    data := TemplateData{
        PackageName: structs[0].PackageName,
        Structs:     templateStructs,
    }
    
    tmpl, err := template.New("reset").Parse(resetTemplate)
    if err != nil {
        return fmt.Errorf("ошибка создания шаблона: %v", err)
    }
    
    var buf bytes.Buffer
    if err := tmpl.Execute(&buf, data); err != nil {
        return fmt.Errorf("ошибка выполнения шаблона: %v", err)
    }
    
    if err := os.MkdirAll(targetDir, 0755); err != nil {
        return fmt.Errorf("ошибка создания директории: %v", err)
    }
    
    file, err := os.Create(targetFile)
    if err != nil {
        return fmt.Errorf("ошибка создания файла: %v", err)
    }
    defer file.Close()
    
    _, err = io.Copy(file, &buf)
    if err != nil {
        return fmt.Errorf("ошибка записи файла: %v", err)
    }
    
    fmt.Printf("Сгенерирован файл: %s\n", targetFile)
    return nil
}

func getZeroValue(typeStr string) string {
    // Для указателей
    if strings.HasPrefix(typeStr, "*") {
        return "nil"
    }
    
    // Для базовых типов
    switch typeStr {
    case "int", "int8", "int16", "int32", "int64",
         "uint", "uint8", "uint16", "uint32", "uint64",
         "float32", "float64", "complex64", "complex128":
        return "0"
    case "string":
        return "\"\""
    case "bool":
        return "false"
    case "byte", "rune":
        return "0"
    case "error":
        return "nil"
    default:
        // Проверяем специальные типы
        if strings.HasPrefix(typeStr, "[]") {
            return "nil" // слайсы
        }
        if strings.HasPrefix(typeStr, "map[") {
            return "nil" // мапы
        }
        if strings.HasPrefix(typeStr, "chan ") {
            return "nil" // каналы
        }
        if strings.HasPrefix(typeStr, "chan") && len(typeStr) > 4 {
            return "nil" // каналы без пробела: chan bool
        }
        if strings.Contains(typeStr, "func(") {
            return "nil" // функции
        }
        // Для остальных типов (структуры, интерфейсы, типы из других пакетов)
        return typeStr + "{}"
    }
}