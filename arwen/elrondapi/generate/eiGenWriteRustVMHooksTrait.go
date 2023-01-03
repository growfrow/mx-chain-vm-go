package elrondapigenerate

import (
	"fmt"
	"os"
)

// WriteRustVMHooksTrait autogenerate data in the provided file
func WriteRustVMHooksTrait(out *os.File, eiMetadata *EIMetadata) {
	_, _ = out.WriteString(`// Code generated by elrondapi generator. DO NOT EDIT.

// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
// !!!!!!!!!!!!!!!!!!!!!! AUTO-GENERATED FILE !!!!!!!!!!!!!!!!!!!!!!
// !!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!

pub trait VMHooks: 'static {
`)

	for _, funcMetadata := range eiMetadata.AllFunctions {
		_, _ = out.WriteString(fmt.Sprintf(
			"    fn %s%s;\n",
			snakeCase(funcMetadata.Name),
			writeRustFnDeclarationArguments(
				"&self",
				funcMetadata,
			),
		))
	}

	_, _ = out.WriteString(`}

pub struct VMHooksDefault;

#[allow(unused)]
#[rustfmt::skip]
impl VMHooks for VMHooksDefault {
`)

	for i, funcMetadata := range eiMetadata.AllFunctions {
		if i > 0 {
			_, _ = out.WriteString("\n")
		}

		_, _ = out.WriteString(fmt.Sprintf(
			"    fn %s%s {\n",
			snakeCase(funcMetadata.Name),
			writeRustFnDeclarationArguments(
				"&self",
				funcMetadata,
			),
		))

		_, _ = out.WriteString(fmt.Sprintf(
			"        println!(\"Called: %s\");\n",
			snakeCase(funcMetadata.Name),
		))

		if funcMetadata.Result != nil {
			_, _ = out.WriteString("        0\n")
		}

		_, _ = out.WriteString("    }\n")
	}

	_, _ = out.WriteString("}\n")
}
