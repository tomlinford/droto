package main

import (
	"bytes"
	"fmt"
	"io"
	"strings"

	"github.com/huandu/go-sqlbuilder"
	pgs "github.com/lyft/protoc-gen-star"
	"github.com/tomlinford/droto"
)

func main() {
	pgs.Init(pgs.DebugEnv("DEBUG")).
		RegisterModule(New()).
		// RegisterPostProcessor(&myPostProcessor{}).
		Render()
}

type sqlModule struct {
	*pgs.ModuleBase
}

func New() pgs.Module { return &sqlModule{&pgs.ModuleBase{}} }

func (m *sqlModule) Name() string { return "sql" }

func (m *sqlModule) Execute(targets map[string]pgs.File,
	packages map[string]pgs.Package) []pgs.Artifact {

	buf := &bytes.Buffer{}

	for _, f := range targets {
		m.printFile(f, buf)
	}

	return m.Artifacts()
}

func (p *sqlModule) printFile(f pgs.File, buf *bytes.Buffer) {
	p.Push(f.Name().String())
	defer p.Pop()

	buf.Reset()

	for _, m := range f.AllMessages() {
		b := sqlbuilder.CreateTable(m.Name().LowerCamelCase().String())
		for _, f := range m.Fields() {
			// prefix := []string{}
			// if i < len(m.Fields()) {
			// 	prefix = []string{"\n"}
			// }
			// suffix := []string{}
			// // if i < len(m.Fields())-1 {
			// // 	suffix = []string{"\n"}
			// // }
			typ := ""
			// b.Define(append(append(prefix, defs...), suffix...)...)
			suffix := []string{}
			var options droto.FieldOptions
			if ok, err := f.Extension(droto.E_FieldOptions, &options); err != nil {
				p.Fail(err)
			} else if ok {
				if !options.Nullable {
					suffix = append(suffix, "not null")
				}
			}
			// switch f.Type().String() {
			protoType := f.Type().ProtoType()
			wkt := pgs.UnknownWKT
			if f.Type().Embed() != nil {
				wkt = f.Type().Embed().WellKnownType()
			}
			if protoType == pgs.StringT || wkt == pgs.StringValueWKT {
				var options droto.CharFieldOptions
				if ok, err := f.Extension(droto.E_CharFieldOptions, &options); err != nil {
					p.Fail(err)
				} else if !ok {
					p.Fail("must set char_field_options")
				}
				typ = fmt.Sprintf("varchar(%d)", options.MaxLength)
			} else if protoType == pgs.Int64T {
				typ = "bigint"
			}
			// if f.Name().String() == "about_me" {
			// 	panic(fmt.Sprint(options))
			// }
			// options, ok := proto.GetExtension(
			// 	f.Descriptor(), droto.E_FieldOptions).(*droto.FieldOptions)
			// if !ok {
			// 	p.Fail("field options not set")
			// }
			defs := []string{"\n   ", f.Name().String(), typ}
			b.Define(append(defs, suffix...)...)
		}
		buf.WriteString(b.String())
		buf.WriteString(";\n")
	}
	// v := initPrintVisitor(buf, "")
	// p.CheckErr(pgs.Walk(v, f), "unable to print AST tree")

	out := buf.String()

	// if ok, _ := p.Parameters().Bool("log_tree"); ok {
	// 	p.Logf("Proto Tree:\n%s", out)
	// }

	p.AddGeneratorFile(
		f.InputPath().SetExt("_init.sql").String(),
		out,
	)
}

const (
	startNodePrefix = "┳ "
	subNodePrefix   = "┃"
	leafNodePrefix  = "┣"
	leafNodeSpacer  = "━ "
)

type PrinterVisitor struct {
	pgs.Visitor
	prefix string
	w      io.Writer
}

func initPrintVisitor(w io.Writer, prefix string) pgs.Visitor {
	v := PrinterVisitor{
		prefix: prefix,
		w:      w,
	}
	v.Visitor = pgs.PassThroughVisitor(&v)
	return v
}

func (v PrinterVisitor) leafPrefix() string {
	if strings.HasSuffix(v.prefix, subNodePrefix) {
		return strings.TrimSuffix(v.prefix, subNodePrefix) + leafNodePrefix
	}
	return v.prefix
}

func (v PrinterVisitor) writeSubNode(str string) pgs.Visitor {
	fmt.Fprintf(v.w, "%s%s%s\n", v.leafPrefix(), startNodePrefix, str)
	return initPrintVisitor(v.w, fmt.Sprintf("%s%v", v.prefix, subNodePrefix))
}

func (v PrinterVisitor) writeLeaf(str string) {
	fmt.Fprintf(v.w, "%s%s%s\n", v.leafPrefix(), leafNodeSpacer, str)
}

func (v PrinterVisitor) VisitFile(f pgs.File) (pgs.Visitor, error) {
	return v.writeSubNode("File: " + f.Name().String()), nil
}

func (v PrinterVisitor) VisitMessage(m pgs.Message) (pgs.Visitor, error) {
	return v.writeSubNode("Message: " + m.Name().String()), nil
}

func (v PrinterVisitor) VisitEnum(e pgs.Enum) (pgs.Visitor, error) {
	return v.writeSubNode("Enum: " + e.Name().String()), nil
}

func (v PrinterVisitor) VisitService(s pgs.Service) (pgs.Visitor, error) {
	return v.writeSubNode("Service: " + s.Name().String()), nil
}

func (v PrinterVisitor) VisitEnumValue(ev pgs.EnumValue) (pgs.Visitor, error) {
	v.writeLeaf(ev.Name().String())
	return nil, nil
}

func (v PrinterVisitor) VisitField(f pgs.Field) (pgs.Visitor, error) {
	v.writeLeaf(f.Name().String())
	return nil, nil
}

func (v PrinterVisitor) VisitMethod(m pgs.Method) (pgs.Visitor, error) {
	v.writeLeaf(m.Name().String())
	return nil, nil
}
