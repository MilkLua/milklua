package lua

import (
	"fmt"
)

/*
	gopherlua uses modified Lua 5.1.4's opcodes.
	Instruction layout (32 bits, fixed length):

		+---------------------------------------------+
		|0-5(6bits)|6-13(8bits)|14-22(9bits)|23-31(9bits)|
		|==========+==========+===========+===========|
		|  opcode  |    A     |     C     |    B      |  (OpTypeABC)
		|----------+----------+-----------+-----------|
		|  opcode  |    A     |      Bx(unsigned)     |  (OpTypeABx)
		|----------+----------+-----------+-----------|
		|  opcode  |    A     |      sBx(signed)      |  (OpTypeASbx)
		+---------------------------------------------+

	opcode: 6 bits (0 ~ 63)
	A: 8 bits
	B: 9 bits
	C: 9 bits
	Bx: 18 bits (unsigned)
	sBx: 18 bits (signed)

	Opcodes (0~45): total 46
	 0  MOVE       (A B C)   R(A) := R(B)
	 1  MOVEN      (A B C)   R(A) := R(B); followed by R(C) MOVE ops
	 2  LOADK      (A Bx)    R(A) := Kst(Bx)
	 3  LOADBOOL   (A B C)   R(A) := (Bool)B; if (C) pc++
	 4  LOADNIL    (A B)     R(A) := ... := R(B) := nil
	 5  GETUPVAL   (A B)     R(A) := UpValue[B]
	 6  GETGLOBAL  (A Bx)    R(A) := Gbl[Kst(Bx)]
	 7  GETTABLE   (A B C)   R(A) := R(B)[RK(C)]
	 8  GETTABLEKS (A B C)   R(A) := R(B)[RK(C)] ; RK(C) is const string
	 9  SETGLOBAL  (A Bx)    Gbl[Kst(Bx)] := R(A)
	 10 SETUPVAL   (A B)     UpValue[B] := R(A)
	 11 SETTABLE   (A B C)   R(A)[RK(B)] := RK(C)
	 12 SETTABLEKS (A B C)   R(A)[RK(B)] := RK(C) ; RK(B) is const string
	 13 NEWTABLE   (A B C)   R(A) := {} (size=BC)
	 14 SELF       (A B C)   R(A+1) := R(B); R(A) := R(B)[RK(C)]
	 15 ADD        (A B C)   R(A) := RK(B) + RK(C)
	 16 SUB        (A B C)   R(A) := RK(B) - RK(C)
	 17 MUL        (A B C)   R(A) := RK(B) * RK(C)
	 18 DIV        (A B C)   R(A) := RK(B) / RK(C)
	 19 MOD        (A B C)   R(A) := RK(B) % RK(C)
	 20 POW        (A B C)   R(A) := RK(B) ^ RK(C)
	 21 UNM        (A B)     R(A) := -R(B)
	 22 NOT        (A B)     R(A) := not R(B)
	 23 LEN        (A B)     R(A) := length of R(B)
	 24 CONCAT     (A B C)   R(A) := R(B).. ... ..R(C)
	 25 JMP        (sBx)     pc += sBx
	 26 EQ         (A B C)   if ((RK(B) == RK(C)) ~= A) then pc++
	 27 LT         (A B C)   if ((RK(B) <  RK(C)) ~= A) then pc++
	 28 LE         (A B C)   if ((RK(B) <= RK(C)) ~= A) then pc++
	 29 TEST       (A C)     if not (R(A) <=> C) then pc++
	 30 TESTSET    (A B C)   if (R(B) <=> C) then R(A) := R(B) else pc++
	 31 CALL       (A B C)   R(A)...R(A+C-2) := R(A)(R(A+1)...R(A+B-1))
	 32 TAILCALL   (A B C)   return R(A)(R(A+1)...R(A+B-1))
	 33 RETURN     (A B)     return R(A)...R(A+B-2)
	 34 FORLOOP    (A sBx)   R(A)+=R(A+2); if R(A)<?=R(A+1) {pc+=sBx; R(A+3)=R(A)}
	 35 FORPREP    (A sBx)   R(A)-=R(A+2); pc+=sBx
	 36 TFORLOOP   (A C)     R(A+3)...R(A+3+C) := R(A)(R(A+1), R(A+2));
													 if R(A+3)~=nil { pc++; R(A+2)=R(A+3) }
	 37 SETLIST    (A B C)   R(A)[(C-1)*FPF+i] := R(A+i)
	 38 CLOSE      (A)       close all vars up to R(A)
	 39 CLOSURE    (A Bx)    R(A) := closure(KPROTO[Bx], R(A)...R(A+n))
	 40 VARARG     (A B)     R(A)...R(A+B-1) = vararg
	 41 BAND       (A B C)   R(A) := RK(B) & RK(C)
	 42 BOR        (A B C)   R(A) := RK(B) | RK(C)
	 43 SHL        (A B C)   R(A) := RK(B) << RK(C)
	 44 SHR        (A B C)   R(A) := RK(B) >> RK(C)
	 45 NOP        (A B C)   no operation
*/

const opInvalidInstruction = ^uint32(0)

const opSizeCode = 6
const opSizeA = 8
const opSizeB = 9
const opSizeC = 9
const opSizeBx = 18
const opSizesBx = 18

const opMaxArgsA = (1 << opSizeA) - 1
const opMaxArgsB = (1 << opSizeB) - 1
const opMaxArgsC = (1 << opSizeC) - 1
const opMaxArgBx = (1 << opSizeBx) - 1
const opMaxArgSbx = opMaxArgBx >> 1

// opcodes: total 46
const (
	OP_MOVE     int = iota /*   A B       R(A) := R(B)                                      */
	OP_MOVEN               /*   A B       R(A) := R(B); followed by R(C) MOVE ops           */
	OP_LOADK               /*   A Bx      R(A) := Kst(Bx)                                   */
	OP_LOADBOOL            /*   A B C     R(A) := (Bool)B; if (C) pc++                      */
	OP_LOADNIL             /*   A B       R(A) := ... := R(B) := nil                        */
	OP_GETUPVAL            /*   A B       R(A) := UpValue[B]                                */

	OP_GETGLOBAL  /*   A Bx      R(A) := Gbl[Kst(Bx)]                                  */
	OP_GETTABLE   /*   A B C     R(A) := R(B)[RK(C)]                                   */
	OP_GETTABLEKS /*   A B C     R(A) := R(B)[RK(C)] ; RK(C) is constant string        */

	OP_SETGLOBAL  /*   A Bx      Gbl[Kst(Bx)] := R(A)                                  */
	OP_SETUPVAL   /*   A B       UpValue[B] := R(A)                                    */
	OP_SETTABLE   /*   A B C     R(A)[RK(B)] := RK(C)                                  */
	OP_SETTABLEKS /*   A B C     R(A)[RK(B)] := RK(C) ; RK(B) is constant string       */

	OP_NEWTABLE /*   A B C     R(A) := {} (size = BC)                                  */

	OP_SELF /*   A B C     R(A+1) := R(B); R(A) := R(B)[RK(C)]                          */

	OP_ADD /*   A B C     R(A) := RK(B) + RK(C)                                         */
	OP_SUB /*   A B C     R(A) := RK(B) - RK(C)                                         */
	OP_MUL /*   A B C     R(A) := RK(B) * RK(C)                                         */
	OP_DIV /*   A B C     R(A) := RK(B) / RK(C)                                         */
	OP_MOD /*   A B C     R(A) := RK(B) % RK(C)                                         */
	OP_POW /*   A B C     R(A) := RK(B) ^ RK(C)                                         */
	OP_UNM /*   A B       R(A) := -R(B)                                                */
	OP_NOT /*   A B       R(A) := not R(B)                                             */
	OP_LEN /*   A B       R(A) := length of R(B)                                       */

	OP_CONCAT /*   A B C     R(A) := R(B).. ... ..R(C)                                 */

	OP_JMP /*   sBx        pc += sBx                                                   */

	OP_EQ /*   A B C       if ((RK(B) == RK(C)) ~= A) then pc++                        */
	OP_LT /*   A B C       if ((RK(B) <  RK(C)) ~= A) then pc++                        */
	OP_LE /*   A B C       if ((RK(B) <= RK(C)) ~= A) then pc++                        */

	OP_TEST    /*   A C         if not (R(A) <=> C) then pc++                           */
	OP_TESTSET /*   A B C       if (R(B) <=> C) then R(A) := R(B) else pc++             */

	OP_CALL     /*   A B C       R(A) ... R(A+C-2) := R(A)(R(A+1) ... R(A+B-1))          */
	OP_TAILCALL /*   A B C       return R(A)(R(A+1) ... R(A+B-1))                       */
	OP_RETURN   /*   A B         return R(A) ... R(A+B-2) (see note)                    */

	OP_FORLOOP /*   A sBx       R(A)+=R(A+2);
	if R(A) <?= R(A+1) then { pc += sBx; R(A+3) = R(A) } */

	OP_FORPREP /*   A sBx       R(A)-=R(A+2); pc += sBx                                */

	OP_TFORLOOP /*   A C         R(A+3) ... R(A+3+C) := R(A)(R(A+1), R(A+2));
	if R(A+3) ~= nil then { pc++; R(A+2) = R(A+3); } */

	OP_SETLIST /*   A B C       R(A)[(C-1)*FPF + i] := R(A + i) 1 <= i <= B              */

	OP_CLOSE   /*   A           close all variables in the stack up to (>=) R(A)        */
	OP_CLOSURE /*   A Bx        R(A) := closure(KPROTO[Bx], R(A) ... R(A+n))            */

	OP_VARARG /*   A B         R(A) R(A+1) ... R(A+B-1) = vararg                        */

	OP_BAND /*   A B C         R(A) := RK(B) & RK(C)                                   */
	OP_BOR  /*   A B C         R(A) := RK(B) | RK(C)                                   */
	OP_SHL  /*   A B C         R(A) := RK(B) << RK(C)                                  */
	OP_SHR  /*   A B C         R(A) := RK(B) >> RK(C)                                  */

	OP_TYPEASSERT /*   A B C       R(A) := typeassert(R(B), RK(C))                       */

	OP_NOP /* NOP */
)
const opCodeMax = OP_NOP

type opArgMode int

const (
	opArgModeN opArgMode = iota
	opArgModeU
	opArgModeR
	opArgModeK
)

type opType int

const (
	opTypeABC = iota
	opTypeABx
	opTypeASbx
)

type opProp struct {
	Name     string
	IsTest   bool
	SetRegA  bool
	ModeArgB opArgMode
	ModeArgC opArgMode
	Type     opType
}

var opProps = []opProp{
	{"MOVE", false, true, opArgModeR, opArgModeN, opTypeABC},
	{"MOVEN", false, true, opArgModeR, opArgModeN, opTypeABC},
	{"LOADK", false, true, opArgModeK, opArgModeN, opTypeABx},
	{"LOADBOOL", false, true, opArgModeU, opArgModeU, opTypeABC},
	{"LOADNIL", false, true, opArgModeR, opArgModeN, opTypeABC},
	{"GETUPVAL", false, true, opArgModeU, opArgModeN, opTypeABC},
	{"GETGLOBAL", false, true, opArgModeK, opArgModeN, opTypeABx},
	{"GETTABLE", false, true, opArgModeR, opArgModeK, opTypeABC},
	{"GETTABLEKS", false, true, opArgModeR, opArgModeK, opTypeABC},
	{"SETGLOBAL", false, false, opArgModeK, opArgModeN, opTypeABx},
	{"SETUPVAL", false, false, opArgModeU, opArgModeN, opTypeABC},
	{"SETTABLE", false, false, opArgModeK, opArgModeK, opTypeABC},
	{"SETTABLEKS", false, false, opArgModeK, opArgModeK, opTypeABC},
	{"NEWTABLE", false, true, opArgModeU, opArgModeU, opTypeABC},
	{"SELF", false, true, opArgModeR, opArgModeK, opTypeABC},
	{"ADD", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"SUB", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"MUL", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"DIV", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"MOD", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"POW", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"UNM", false, true, opArgModeR, opArgModeN, opTypeABC},
	{"NOT", false, true, opArgModeR, opArgModeN, opTypeABC},
	{"LEN", false, true, opArgModeR, opArgModeN, opTypeABC},
	{"CONCAT", false, true, opArgModeR, opArgModeR, opTypeABC},
	{"JMP", false, false, opArgModeR, opArgModeN, opTypeASbx},
	{"EQ", true, false, opArgModeK, opArgModeK, opTypeABC},
	{"LT", true, false, opArgModeK, opArgModeK, opTypeABC},
	{"LE", true, false, opArgModeK, opArgModeK, opTypeABC},
	{"TEST", true, true, opArgModeR, opArgModeU, opTypeABC},
	{"TESTSET", true, true, opArgModeR, opArgModeU, opTypeABC},
	{"CALL", false, true, opArgModeU, opArgModeU, opTypeABC},
	{"TAILCALL", false, true, opArgModeU, opArgModeU, opTypeABC},
	{"RETURN", false, false, opArgModeU, opArgModeN, opTypeABC},
	{"FORLOOP", false, true, opArgModeR, opArgModeN, opTypeASbx},
	{"FORPREP", false, true, opArgModeR, opArgModeN, opTypeASbx},
	{"TFORLOOP", true, false, opArgModeN, opArgModeU, opTypeABC},
	{"SETLIST", false, false, opArgModeU, opArgModeU, opTypeABC},
	{"CLOSE", false, false, opArgModeN, opArgModeN, opTypeABC},
	{"CLOSURE", false, true, opArgModeU, opArgModeN, opTypeABx},
	{"VARARG", false, true, opArgModeU, opArgModeN, opTypeABC},
	{"BAND", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"BOR", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"SHL", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"SHR", false, true, opArgModeK, opArgModeK, opTypeABC},
	{"TYPEASSERT", false, true, opArgModeR, opArgModeK, opTypeABC},
	{"NOP", false, false, opArgModeR, opArgModeN, opTypeASbx},
}

func opGetOpCode(inst uint32) int {
	return int(inst >> 26)
}

func opSetOpCode(inst *uint32, opcode int) {
	*inst = (*inst & 0x3ffffff) | uint32(opcode<<26)
}

func opGetArgA(inst uint32) int {
	return int(inst>>18) & 0xff
}

func opSetArgA(inst *uint32, arg int) {
	*inst = (*inst & 0xfc03ffff) | uint32((arg&0xff)<<18)
}

func opGetArgB(inst uint32) int {
	return int(inst & 0x1ff)
}

func opSetArgB(inst *uint32, arg int) {
	*inst = (*inst & 0xfffffe00) | uint32(arg&0x1ff)
}

func opGetArgC(inst uint32) int {
	return int(inst>>9) & 0x1ff
}

func opSetArgC(inst *uint32, arg int) {
	*inst = (*inst & 0xfffc01ff) | uint32((arg&0x1ff)<<9)
}

func opGetArgBx(inst uint32) int {
	return int(inst & 0x3ffff)
}

func opSetArgBx(inst *uint32, arg int) {
	*inst = (*inst & 0xfffc0000) | uint32(arg&0x3ffff)
}

func opGetArgSbx(inst uint32) int {
	return opGetArgBx(inst) - opMaxArgSbx
}

func opSetArgSbx(inst *uint32, arg int) {
	opSetArgBx(inst, arg+opMaxArgSbx)
}

func opCreateABC(op int, a int, b int, c int) uint32 {
	var inst uint32 = 0
	opSetOpCode(&inst, op)
	opSetArgA(&inst, a)
	opSetArgB(&inst, b)
	opSetArgC(&inst, c)
	return inst
}

func opCreateABx(op int, a int, bx int) uint32 {
	var inst uint32 = 0
	opSetOpCode(&inst, op)
	opSetArgA(&inst, a)
	opSetArgBx(&inst, bx)
	return inst
}

func opCreateASbx(op int, a int, sbx int) uint32 {
	var inst uint32 = 0
	opSetOpCode(&inst, op)
	opSetArgA(&inst, a)
	opSetArgSbx(&inst, sbx)
	return inst
}

const opBitRk = 1 << (opSizeB - 1)
const opMaxIndexRk = opBitRk - 1

func opIsK(value int) bool {
	return bool((value & opBitRk) != 0)
}

/*
func opIndexK(value int) int {
	return value & ^opBitRk
}
*/

func opRkAsk(value int) int {
	return value | opBitRk
}

func opToString(inst uint32) string {
	op := opGetOpCode(inst)
	if op > opCodeMax {
		return ""
	}
	prop := &(opProps[op])

	arga := opGetArgA(inst)
	argb := opGetArgB(inst)
	argc := opGetArgC(inst)
	argbx := opGetArgBx(inst)
	argsbx := opGetArgSbx(inst)

	buf := ""
	switch prop.Type {
	case opTypeABC:
		buf = fmt.Sprintf("%s      |  %d, %d, %d", prop.Name, arga, argb, argc)
	case opTypeABx:
		buf = fmt.Sprintf("%s      |  %d, %d", prop.Name, arga, argbx)
	case opTypeASbx:
		buf = fmt.Sprintf("%s      |  %d, %d", prop.Name, arga, argsbx)
	}

	switch op {
	case OP_MOVE:
		buf += fmt.Sprintf("; R(%v) := R(%v)", arga, argb)
	case OP_MOVEN:
		buf += fmt.Sprintf("; R(%v) := R(%v); followed by %v MOVE ops", arga, argb, argc)
	case OP_LOADK:
		buf += fmt.Sprintf("; R(%v) := Kst(%v)", arga, argbx)
	case OP_LOADBOOL:
		buf += fmt.Sprintf("; R(%v) := (Bool)%v; if (%v) pc++", arga, argb, argc)
	case OP_LOADNIL:
		buf += fmt.Sprintf("; R(%v) := ... := R(%v) := nil", arga, argb)
	case OP_GETUPVAL:
		buf += fmt.Sprintf("; R(%v) := UpValue[%v]", arga, argb)
	case OP_GETGLOBAL:
		buf += fmt.Sprintf("; R(%v) := Gbl[Kst(%v)]", arga, argbx)
	case OP_GETTABLE:
		buf += fmt.Sprintf("; R(%v) := R(%v)[RK(%v)]", arga, argb, argc)
	case OP_GETTABLEKS:
		buf += fmt.Sprintf("; R(%v) := R(%v)[RK(%v)] ; RK(%v) is constant string", arga, argb, argc, argc)
	case OP_SETGLOBAL:
		buf += fmt.Sprintf("; Gbl[Kst(%v)] := R(%v)", argbx, arga)
	case OP_SETUPVAL:
		buf += fmt.Sprintf("; UpValue[%v] := R(%v)", argb, arga)
	case OP_SETTABLE:
		buf += fmt.Sprintf("; R(%v)[RK(%v)] := RK(%v)", arga, argb, argc)
	case OP_SETTABLEKS:
		buf += fmt.Sprintf("; R(%v)[RK(%v)] := RK(%v) ; RK(%v) is constant string", arga, argb, argc, argb)
	case OP_NEWTABLE:
		buf += fmt.Sprintf("; R(%v) := {} (size = BC)", arga)
	case OP_SELF:
		buf += fmt.Sprintf("; R(%v+1) := R(%v); R(%v) := R(%v)[RK(%v)]", arga, argb, arga, argb, argc)
	case OP_ADD:
		buf += fmt.Sprintf("; R(%v) := RK(%v) + RK(%v)", arga, argb, argc)
	case OP_SUB:
		buf += fmt.Sprintf("; R(%v) := RK(%v) - RK(%v)", arga, argb, argc)
	case OP_MUL:
		buf += fmt.Sprintf("; R(%v) := RK(%v) * RK(%v)", arga, argb, argc)
	case OP_DIV:
		buf += fmt.Sprintf("; R(%v) := RK(%v) / RK(%v)", arga, argb, argc)
	case OP_MOD:
		buf += fmt.Sprintf("; R(%v) := RK(%v) %% RK(%v)", arga, argb, argc)
	case OP_POW:
		buf += fmt.Sprintf("; R(%v) := RK(%v) ^ RK(%v)", arga, argb, argc)
	case OP_UNM:
		buf += fmt.Sprintf("; R(%v) := -R(%v)", arga, argb)
	case OP_NOT:
		buf += fmt.Sprintf("; R(%v) := not R(%v)", arga, argb)
	case OP_LEN:
		buf += fmt.Sprintf("; R(%v) := length of R(%v)", arga, argb)
	case OP_CONCAT:
		buf += fmt.Sprintf("; R(%v) := R(%v).. ... ..R(%v)", arga, argb, argc)
	case OP_JMP:
		buf += fmt.Sprintf("; pc+=%v", argsbx)
	case OP_EQ:
		buf += fmt.Sprintf("; if ((RK(%v) == RK(%v)) ~= %v) then pc++", argb, argc, arga)
	case OP_LT:
		buf += fmt.Sprintf("; if ((RK(%v) <  RK(%v)) ~= %v) then pc++", argb, argc, arga)
	case OP_LE:
		buf += fmt.Sprintf("; if ((RK(%v) <= RK(%v)) ~= %v) then pc++", argb, argc, arga)
	case OP_TEST:
		buf += fmt.Sprintf("; if not (R(%v) <=> %v) then pc++", arga, argc)
	case OP_TESTSET:
		buf += fmt.Sprintf("; if (R(%v) <=> %v) then R(%v) := R(%v) else pc++", argb, argc, arga, argb)
	case OP_CALL:
		buf += fmt.Sprintf("; R(%v) ... R(%v+%v-2) := R(%v)(R(%v+1) ... R(%v+%v-1))",
			arga,
			arga,
			argc,
			arga,
			arga,
			arga,
			argb,
		)
	case OP_TAILCALL:
		buf += fmt.Sprintf("; return R(%v)(R(%v+1) ... R(%v+%v-1))",
			arga,
			arga,
			arga,
			argb,
		)
	case OP_RETURN:
		buf += fmt.Sprintf("; return R(%v) ... R(%v+%v-2)", arga, arga, argb)
	case OP_FORLOOP:
		buf += fmt.Sprintf("; R(%v)+=R(%v+2); if R(%v) <?= R(%v+1) then { pc+=%v; R(%v+3)=R(%v) }", arga, arga, arga, arga, argsbx, arga, arga)
	case OP_FORPREP:
		buf += fmt.Sprintf("; R(%v)-=R(%v+2); pc+=%v", arga, arga, argsbx)
	case OP_TFORLOOP:
		buf += fmt.Sprintf("; R(%v+3) ... R(%v+3+%v) := R(%v)(R(%v+1) R(%v+2)); if R(%v+3) ~= nil then { pc++; R(%v+2)=R(%v+3); }", arga, arga, argc, arga, arga, arga, arga, arga, arga)
	case OP_SETLIST:
		buf += fmt.Sprintf("; R(%v)[(%v-1)*FPF+i] := R(%v+i) 1 <= i <= %v", arga, argc, arga, argb)
	case OP_CLOSE:
		buf += fmt.Sprintf("; close all variables in the stack up to (>=) R(%v)", arga)
	case OP_CLOSURE:
		buf += fmt.Sprintf("; R(%v) := closure(KPROTO[%v] R(%v) ... R(%v+n))", arga, argbx, arga, arga)
	case OP_VARARG:
		buf += fmt.Sprintf(";  R(%v) R(%v+1) ... R(%v+%v-1) = vararg", arga, arga, arga, argb)
	case OP_BAND:
		buf += fmt.Sprintf("; R(%v) := RK(%v) & RK(%v)", arga, argb, argc)
	case OP_BOR:
		buf += fmt.Sprintf("; R(%v) := RK(%v) | RK(%v)", arga, argb, argc)
	case OP_SHL:
		buf += fmt.Sprintf("; R(%v) := RK(%v) << RK(%v)", arga, argb, argc)
	case OP_SHR:
		buf += fmt.Sprintf("; R(%v) := RK(%v) >> RK(%v)", arga, argb, argc)
	case OP_TYPEASSERT:
		buf += fmt.Sprintf("; R(%v) := typeassert(R(%v), RK(%v))", arga, argb, argc)
	case OP_NOP:
		/* nothing to do */
	}
	return buf
}
