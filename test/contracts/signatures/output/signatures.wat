(module
  (type (;0;) (func))
  (type (;1;) (func (result i32)))
  (type (;2;) (func (param i32)))
  (type (;3;) (func (param i32 i32) (result i32)))
  (func (;0;) (type 0))
  (func (;1;) (type 1) (result i32)
    i32.const 0)
  (func (;2;) (type 2) (param i32))
  (func (;3;) (type 3) (param i32 i32) (result i32)
    i32.const 0)
  (table (;0;) 1 1 funcref)
  (memory (;0;) 2)
  (global (;0;) (mut i32) (i32.const 66560))
  (export "memory" (memory 0))
  (export "goodFunction" (func 0))
  (export "wrongReturn" (func 1))
  (export "wrongParams" (func 2))
  (export "wrongParamsAndReturn" (func 3)))
