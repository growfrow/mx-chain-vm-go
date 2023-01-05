package elrondapigenerate

import (
	"fmt"
)

func WriteRustCapiVMHooksPointers(out *eiGenWriter, eiMetadata *EIMetadata) {
	autoGeneratedHeader(out)
	out.WriteString(`
use std::ffi::c_void;

#[repr(C)]
#[derive(Clone)]
#[rustfmt::skip]
pub struct vm_exec_vm_hook_c_func_pointers {`)

	for _, funcMetadata := range eiMetadata.AllFunctions {
		out.WriteString(fmt.Sprintf(
			"\n    pub %s: extern \"C\" fn%s",
			cgoFuncPointerFieldName(funcMetadata),
			writeRustFnDeclarationArguments(
				"context: *mut c_void",
				funcMetadata,
			),
		))

		out.WriteString(",")
	}

	out.WriteString(`
}

impl std::fmt::Debug for vm_exec_vm_hook_c_func_pointers {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(f, "vm_exec_vm_hook_c_func_pointers")
    }
}
`)
}
