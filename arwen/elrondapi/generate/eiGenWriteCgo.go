package elrondapigenerate

import (
	"fmt"
	"os"
)

type cgoWriter struct {
	goPackage string
	cgoPrefix string
}

func cgoType(goType string) string {
	if goType == "int32" {
		return "int32_t"
	}
	if goType == "int64" {
		return "long long"
	}
	return goType
}

func (writer *cgoWriter) cgoFuncName(funcMetadata *EIFunction) string {
	return writer.cgoPrefix + lowerInitial(funcMetadata.Name)
}

func (writer *cgoWriter) cgoImportName(funcMetadata *EIFunction) string {
	return fmt.Sprintf("C.%s", writer.cgoFuncName(funcMetadata))
}

// WriteWasmer1Cgo writes the metadata in the provided file
func WriteWasmer1Cgo(out *os.File, eiMetadata *EIMetadata) {
	writer := &cgoWriter{
		goPackage: "wasmer",
		cgoPrefix: "v1_5_",
	}
	writer.writeHeader(out, eiMetadata)
	writer.writeCgoFunctions(out, eiMetadata)
	writer.writePopulateImports(out, eiMetadata)
	writer.writeGoExports(out, eiMetadata)
}

// WriteWasmer2Cgo writes the metadata in the provided file
func WriteWasmer2Cgo(out *os.File, eiMetadata *EIMetadata) {
	writer := &cgoWriter{
		goPackage: "wasmer2",
		cgoPrefix: "w2_",
	}
	writer.writeHeader(out, eiMetadata)
	writer.writeCgoFunctions(out, eiMetadata)
	writer.writePopulateFuncPointers(out, eiMetadata)
	writer.writeGoExports(out, eiMetadata)
}

func (writer *cgoWriter) writeHeader(out *os.File, _ *EIMetadata) {
	_, _ = out.WriteString(fmt.Sprintf(`package %s

// Code generated by elrondapi generator. DO NOT EDIT.

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// !!!!!!!!!!!!!!!!!!!!!! AUTO-GENERATED FILE !!!!!!!!!!!!!!!!!!!!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

// // Declare the function signatures (see [cgo](https://golang.org/cmd/cgo/)).
//
// #include <stdlib.h>
// typedef int int32_t;
//
`,
		writer.goPackage))
}

func (writer *cgoWriter) writeCgoFunctions(out *os.File, eiMetadata *EIMetadata) {
	for _, funcMetadata := range eiMetadata.AllFunctions {
		_, _ = out.WriteString(fmt.Sprintf("// extern %-9s %s(void* context",
			externResult(funcMetadata.Result),
			writer.cgoFuncName(funcMetadata),
		))
		for _, arg := range funcMetadata.Arguments {
			_, _ = out.WriteString(fmt.Sprintf(", %s %s", cgoType(arg.Type), arg.Name))
		}
		_, _ = out.WriteString(");\n")
	}

	_, _ = out.WriteString("import \"C\"\n\n")
}

func (writer *cgoWriter) writePopulateImports(out *os.File, eiMetadata *EIMetadata) {
	_, _ = out.WriteString(`import (
	"unsafe"
)

// populateWasmerImports populates imports with the ElrondEI API methods
func populateWasmerImports(imports *wasmerImports) error {
	var err error
`)
	for _, funcMetadata := range eiMetadata.AllFunctions {
		_, _ = out.WriteString(fmt.Sprintf("\terr = imports.append(\"%s\", %s, %s)\n",
			lowerInitial(funcMetadata.Name),
			writer.cgoFuncName(funcMetadata),
			writer.cgoImportName(funcMetadata),
		))
		_, _ = out.WriteString("\tif err != nil {\n")
		_, _ = out.WriteString("\t\treturn err\n")
		_, _ = out.WriteString("\t}\n\n")
	}
	_, _ = out.WriteString("\treturn nil\n")
	_, _ = out.WriteString("}\n")
}

func (writer *cgoWriter) writePopulateFuncPointers(out *os.File, eiMetadata *EIMetadata) {
	_, _ = out.WriteString(`// populateCgoFunctionPointers populates imports with the ElrondEI API methods
func populateCgoFunctionPointers() *cWasmerVmHookPointers {
	return &cWasmerVmHookPointers{`)

	for _, funcMetadata := range eiMetadata.AllFunctions {
		_, _ = out.WriteString(fmt.Sprintf("\n\t\t%s: funcPointer(%s)",
			cgoFuncPointerFieldName(funcMetadata),
			writer.cgoFuncName(funcMetadata),
		))
	}
	_, _ = out.WriteString(`
	}
}
`)
}

func (writer *cgoWriter) writeGoExports(out *os.File, eiMetadata *EIMetadata) {
	for _, funcMetadata := range eiMetadata.AllFunctions {
		_, _ = out.WriteString(fmt.Sprintf("\n//export %s\n",
			writer.cgoFuncName(funcMetadata),
		))
		_, _ = out.WriteString(fmt.Sprintf("func %s(context unsafe.Pointer",
			writer.cgoFuncName(funcMetadata),
		))
		for _, arg := range funcMetadata.Arguments {
			_, _ = out.WriteString(fmt.Sprintf(", %s %s", arg.Name, arg.Type))
		}
		_, _ = out.WriteString(")")
		if funcMetadata.Result != nil {
			_, _ = out.WriteString(fmt.Sprintf(" %s", funcMetadata.Result.Type))
		}
		_, _ = out.WriteString(" {\n")
		_, _ = out.WriteString("\tvmHooks := getVMHooksFromContextRawPtr(context)\n")
		_, _ = out.WriteString("\t")
		if funcMetadata.Result != nil {
			_, _ = out.WriteString("return ")
		}
		_, _ = out.WriteString(fmt.Sprintf("vmHooks.%s(",
			upperInitial(funcMetadata.Name),
		))
		for argIndex, arg := range funcMetadata.Arguments {
			if argIndex > 0 {
				_, _ = out.WriteString(", ")
			}
			_, _ = out.WriteString(arg.Name)
		}
		_, _ = out.WriteString(")\n")

		_, _ = out.WriteString("}\n")
	}
}
