package main

import (
	"bytes"
	"fmt"
	"strings"

	pgs "github.com/lyft/protoc-gen-star"
	"github.com/pascaldekloe/name"
	"github.com/tomlinford/droto"
)

func main() {
	pgs.Init(pgs.DebugEnv("DEBUG")).
		RegisterModule(New()).
		// RegisterPostProcessor(&myPostProcessor{}).
		Render()
}

type drotoModule struct {
	*pgs.ModuleBase
}

func New() pgs.Module               { return &drotoModule{&pgs.ModuleBase{}} }
func (m *drotoModule) Name() string { return "droto" }

func (m *drotoModule) Execute(targets map[string]pgs.File,
	packages map[string]pgs.Package) []pgs.Artifact {

	buf := &bytes.Buffer{}

	for _, f := range targets {
		m.printSerializerFile(f, buf)
		buf.Reset()
	}

	return m.Artifacts()
}

func (m *drotoModule) printSerializerFile(f pgs.File, buf *bytes.Buffer) {
	modelSerializers := []*droto.ModelSerializer{}
	ok, err := f.Extension(droto.E_ModelSerializers, &modelSerializers)
	if err != nil {
		panic(err)
	} else if !ok || len(modelSerializers) == 0 {
		return
	}

	for _, serializer := range modelSerializers {
		goPath := strings.Split(serializer.ViewGoPackage, "/")
		packageName := goPath[len(goPath)-1]
		modelsGoPath := strings.Split(serializer.ModelsGoPackage, "/")
		modelsPackageName := modelsGoPath[len(modelsGoPath)-1]
		fmt.Fprintf(buf, "package %s\n\n", packageName)
		fmt.Fprintf(buf, "import %q\n\n", serializer.ModelsGoPackage)
		fmt.Fprintf(buf, "func %sFromModel(v *%s.%s) *%s {\n",
			serializer.ModelName, modelsPackageName, serializer.ModelName,
			serializer.ModelName)
		fmt.Fprintf(buf, "\treturn &%s{\n", serializer.ModelName)
		for _, f := range serializer.FieldsArr {
			f = name.CamelCase(f, true)
			fmt.Fprintf(buf, "\t\t%s: v.%s,\n", f, f)
		}
		buf.WriteString("\t}\n")
		buf.WriteString("}\n")
	}

	// fmt.Fprintln(buf, modelSerializers[0].FieldsArr)

	m.AddGeneratorFile(
		f.InputPath().SetExt("_serializers.gen.go").String(),
		buf.String(),
	)
}
