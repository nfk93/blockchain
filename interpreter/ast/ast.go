package ast

import (
	"fmt"
	"github.com/pkg/errors"
)

// TODO: Massive overhaul using type casing instead of naive casting, returning errors where relevant

type Exp interface {
	String() string
}

/* BinOpExp */
type BinOpExp struct {
	Left  Exp
	Op    Oper
	Right Exp
}

func (b BinOpExp) String() string {
	return fmt.Sprintf("BinOpExp(Left: %s, Op: %s, Right: %s)", b.Left.String(),
		operToString(b.Op), b.Right.String())
}

func NewBinOpExp(left, oper, right interface{}) (BinOpExp, error) {
	return BinOpExp{left.(Exp), oper.(Oper), right.(Exp)}, nil
}

/* TODO Exp */
type TodoExp struct{}

func NewTodoExp() (Exp, error) {
	return TodoExp{}, nil
}

func (e TodoExp) String() string {
	return "TODOEXP"
}

/* Simple Type Declaration */
type TypeDecl struct {
	id  string
	typ Type
}

func NewTypeDecl(id string, typ interface{}) (Exp, error) {
	return TypeDecl{id, typ.(Type)}, nil
}

func (e TypeDecl) String() string {
	return fmt.Sprintf("TypeDecl(Id: %s, typ: %s)", e.id, e.typ.String())
}

/* Top Level */
type TopLevel struct {
	Roots []Exp
}

func NewRoot(e interface{}) (Exp, error) {
	switch e.(type) {
	case TypeDecl, EntryExpression, InitExp:
		return TopLevel{[]Exp{e.(Exp)}}, nil
	default:
		ex, _ := fail(fmt.Sprintf("Toplevel error, New root can't be type %T", e))
		return TopLevel{[]Exp{ex.(Exp)}}, nil
	}
}

func (e TopLevel) String() string {
	str := fmt.Sprint("TopLevel([")
	for _, v := range e.Roots {
		str = str + "\n\t" + v.String()
	}
	return str + "\n])"
}

func AppendRoots(e1, e2 interface{}) (Exp, error) {
	switch e2.(type) {
	case TopLevel:
		e2 := e2.(TopLevel)
		return TopLevel{append([]Exp{e1.(Exp)}, e2.Roots...)}, nil
	default:
		return fail(fmt.Sprintf("TopLevel is not a TopLevel expression"))
	}
}

/* New Entry */
type EntryExpression struct {
	Id      string
	Params  Pattern
	Storage Pattern
	Body    Exp
}

func (e EntryExpression) String() string {
	return fmt.Sprintf("EntryExpression(Id: %s, params: %s, storage: %s, body: %s)", e.Id, e.Params.String(),
		e.Storage.String(), e.Body.String())
}

func NewEntryExpression(id string, params, pattern, body interface{}) (Exp, error) {
	return EntryExpression{id, params.(Pattern), pattern.(Pattern), body.(Exp)}, nil
}

/* Pattern */
type Param struct {
	id   string
	anno TypeOption
}

func (p Param) String() string {
	if p.anno.opt {
		return fmt.Sprintf("(%s: %s)", p.id, p.anno.typ.String())
	} else {
		return fmt.Sprintf("%s", p.id)
	}
}

func NewParam(id string) (Param, error) {
	return Param{id, TypeOption{false, IntType{}}}, nil
}

func NewAnnoParam(id string, anno interface{}) (Param, error) {
	return Param{id, TypeOption{true, anno.(Type)}}, nil
}

func AppendParams(par, pars interface{}) ([]Param, error) {
	params := pars.([]Param)
	param := par.(Param)
	return append([]Param{param}, params...), nil
}

func NewParamList(param interface{}) ([]Param, error) {
	return []Param{param.(Param)}, nil
}

type Pattern struct {
	params []Param
}

func NewPattern(params interface{}) (Pattern, error) {
	switch params.(type) {
	case []Param:
		return Pattern{params.([]Param)}, nil
	case Param:
		return Pattern{[]Param{params.(Param)}}, nil
	default:
		return Pattern{}, errors.Errorf("Can't derive pattern")
	}
}

func NewEmptyPattern() (Pattern, error) {
	return Pattern{}, nil
}

func (p Pattern) String() string {
	if len(p.params) == 0 {
		return "Pattern()"
	} else {
		res := "Pattern("
		var par Param
		params := p.params
		for len(params) > 1 {
			par, params = params[0], params[1:]
			res = res + fmt.Sprintf("%s, ", par.String())
		}
		par = params[0]
		res = res + fmt.Sprintf("%s)", par.String())
		return res
	}
}

/* KeyLit */
type KeyLit struct {
	key string
}

func (k KeyLit) String() string {
	return fmt.Sprintf("KeyLit(key: %s)", k.key)
}

func NewKeyLit(key []byte) (Exp, error) {
	actualkey := string(key)[3:]
	if checkKey(key) {
		return KeyLit{actualkey}, nil
	} else {
		err := "key is not valid"
		return ErrorExpression{err}, errors.Errorf(err)
	}
}

func checkKey(key []byte) bool {
	// TODO: check if this is an actual base 58 key, i.e. has the Right characters and such
	return true
}

/* BoolLit */
type BoolLit struct {
	val bool
}

func (b BoolLit) String() string {
	return fmt.Sprintf("BoolLit(val: %t)", b.val)
}

func NewBoolLit(val bool) (Exp, error) {
	return BoolLit{val}, nil
}

/* IntLit */
type IntLit struct {
	val int64
}

func (i IntLit) String() string {
	return fmt.Sprintf("IntLit(val: %d)", i.val)
}

func NewIntLit(val int64) (Exp, error) {
	return IntLit{val}, nil
}

/* FloatLit */
type FloatLit struct {
	val float64
}

func (f FloatLit) String() string {
	return fmt.Sprintf("FloatLit(val: %e)", f.val)
}

func NewFloatLit(val float64) (Exp, error) {
	return FloatLit{val}, nil
}

/* KoinType Lit */
type KoinLit struct {
	val int64
}

func (k KoinLit) String() string {
	return fmt.Sprintf("KoinLit(val: %d)", k.val)
}

func NewKoinLit(koins int64) (Exp, error) {
	if koins < 0 {
		err := "koin literal can't have negative value"
		return ErrorExpression{err}, nil
	} else {
		return KoinLit{koins}, nil
	}
}

/* StringType Lit */
type StringLit struct {
	val string
}

func (s StringLit) String() string {
	return fmt.Sprintf("StringLit(val: %s)", s.val)
}

func NewStringLit(str string) (Exp, error) {
	return StringLit{str}, nil
}

/* Unit Lit */
type UnitLit struct{}

func (u UnitLit) String() string {
	return fmt.Sprintf("UnitLit")
}

func NewUnitLit() (Exp, error) {
	return UnitLit{}, nil
}

/* StructLit */
type StructLit struct {
	Ids  []string
	Vals []Exp
}

func (e StructLit) String() string {
	res := "StructLit("
	var id string
	var exp Exp
	idlist := e.Ids
	elist := e.Vals
	for len(idlist) > 0 {
		id, exp, idlist, elist = idlist[0], elist[0], idlist[1:], elist[1:]
		res = res + fmt.Sprintf("\n\t%s = %s;", id, exp.String())
	}
	res = res + fmt.Sprintf("\n)")
	return res
}

func NewStructLit(id string, exp interface{}) (Exp, error) {
	return StructLit{[]string{id}, []Exp{exp.(Exp)}}, nil
}

func AppendStructLit(struc interface{}, id string, exp interface{}) (Exp, error) {
	s := struc.(StructLit)
	return StructLit{append(s.Ids, id), append(s.Vals, exp.(Exp))}, nil
}

/* ListType Lit TODO: Refactor to use head and tail instead? */
type ListLit struct {
	List []Exp
}

func (l ListLit) String() string {
	if len(l.List) == 0 {
		return fmt.Sprintf("ListLit(List: [])")
	} else {
		res := "ListLit(List: ["
		var e Exp
		list := l.List
		for len(list) > 1 {
			e, list = list[0], list[1:]
			res = res + fmt.Sprintf("%s, ", e.String())
		}
		e = list[0]
		res = res + fmt.Sprintf("%s])", e.String())
		return res
	}
}

func NewListLit(exp interface{}) (Exp, error) {
	e := exp.(Exp)
	return ListLit{[]Exp{e}}, nil
}

func NewEmptyList() (Exp, error) {
	return ListLit{[]Exp{}}, nil
}

func AppendList(exp1, exp2 interface{}) (Exp, error) {
	lst1 := exp1.(ListLit)
	lst2 := exp2.(Exp)
	return ListLit{append(lst1.List, lst2)}, nil
}

/* ListConcat */
type ListConcat struct {
	Exp  Exp
	List Exp
}

func (l ListConcat) String() string {
	return fmt.Sprintf("ListConcat(Exp: %s, List: %s)", l.Exp.String(), l.List.String())
}

func NewListConcat(exp, list interface{}) (Exp, error) {
	return ListConcat{exp.(Exp), list.(Exp)}, nil
}

/* CallExp */
type CallExp struct {
	Exp1 Exp
	Exp2 Exp
}

func (e CallExp) String() string {
	return fmt.Sprintf("CallExp(Exp1: %s, Exp2: %s)", e.Exp1.String(), e.Exp2.String())
}

func NewCallExp(exp1, exp2 interface{}) (Exp, error) {
	return CallExp{exp1.(Exp), exp2.(Exp)}, nil
}

/* LetExp */
type LetExp struct {
	Patt   Pattern
	DefExp Exp
	InExp  Exp
}

func (l LetExp) String() string {
	return fmt.Sprintf("LetExp(Patt: %s, DefExp: %s, InExp: %s)", l.Patt.String(),
		l.DefExp.String(), l.InExp.String())
}

func NewLetExp(patt, def, in interface{}) (Exp, error) {
	return LetExp{patt.(Pattern), def.(Exp), in.(Exp)}, nil
}

/* AnnoExp */
type AnnoExp struct {
	Exp  Exp
	Anno Type
}

func (a AnnoExp) String() string {
	return fmt.Sprintf("AnnoExp(Exp: %s, Anno: %s)", a.Exp.String(), a.Anno.String())
}

func NewAnnoExp(exp, typ interface{}) (Exp, error) {
	return AnnoExp{exp.(Exp), typ.(Type)}, nil
}

/* TupleExp */
type TupleExp struct {
	Exps []Exp
}

func (t TupleExp) String() string {
	res := "TupleExp("
	var e Exp
	exps := t.Exps
	for len(exps) > 1 {
		e, exps = exps[0], exps[1:]
		res += e.String() + ", "
	}
	return res + exps[0].String() + ")"
}

func NewTupleExp(exp1, exp2 interface{}) (Exp, error) {
	return TupleExp{[]Exp{exp1.(Exp), exp2.(Exp)}}, nil
}

func AddTupleEntry(exp, tuple interface{}) (Exp, error) {
	exps := tuple.(TupleExp).Exps
	return TupleExp{append([]Exp{exp.(Exp)}, exps...)}, nil
}

/* VarExp */
type VarExp struct {
	Id string
}

func (v VarExp) String() string {
	return fmt.Sprintf("VarExp(Id: %s)", v.Id)
}

func NewVarExp(id string) (Exp, error) {
	return VarExp{id}, nil
}

/* ExpSeq */
type ExpSeq struct {
	Left  Exp
	Right Exp
}

func (e ExpSeq) String() string {
	return fmt.Sprintf("ExpSeq(Left: %s, Right: %s)", e.Left.String(), e.Right.String())
}

func NewExpSeq(exp1, exp2 interface{}) (Exp, error) {
	return ExpSeq{exp1.(Exp), exp2.(Exp)}, nil
}

/* IfThenElseExp */
type IfThenElseExp struct {
	If   Exp
	Then Exp
	Else Exp
}

func (e IfThenElseExp) String() string {
	return fmt.Sprintf("IfThenElseExp(If: %s, Then: %s, Else: %s)", e.If.String(),
		e.Then.String(), e.Else.String())
}

func NewIfThenElseExp(if_, then, else_ interface{}) (Exp, error) {
	return IfThenElseExp{if_.(Exp), then.(Exp), else_.(Exp)}, nil
}

/* IfThenExp */
type IfThenExp struct {
	If   Exp
	Then Exp
}

func (e IfThenExp) String() string {
	return fmt.Sprintf("IfThenExp(If: %s, Then: %s)", e.If.String(), e.Then.String())
}

func NewIfThenExp(if_, then interface{}) (Exp, error) {
	return IfThenExp{if_.(Exp), then.(Exp)}, nil
}

/* ModuleLookupExp */
type ModuleLookupExp struct {
	ModId   string
	FieldId string
}

func (e ModuleLookupExp) String() string {
	return fmt.Sprintf("ModuleLookupExp(ModId: %s, FieldId: %s)", e.ModId, e.FieldId)
}

func NewModuleLookupExp(mod, field string) (Exp, error) {
	// TODO check module existance?
	return ModuleLookupExp{mod, field}, nil
}

/* LookupExp */
type LookupExp struct {
	PathIds []string
	LeafId  string
}

func (e LookupExp) String() string {
	res := "LookupExp(PathIds: ["
	var s string
	idpath := e.PathIds
	for len(idpath) > 1 {
		s, idpath = idpath[0], idpath[1:]
		res += s + ", "
	}
	return res + idpath[0] + "], LeafId: " + e.LeafId + ")"
}

func NewLookupExp(path interface{}, leaf string) (Exp, error) {
	return LookupExp{path.([]string), leaf}, nil
}

func LookupPathRoot(id string) []string {
	return []string{id}
}

func AddPathElement(list interface{}, id string) []string {
	return append(list.([]string), id)
}

/* UpdateStructExp */
type UpdateStructExp struct {
	PathIds []string
	LeafId  string
	Exp     Exp
}

func (e UpdateStructExp) String() string {
	res := "UpdateStructExp(PathIds: ["
	var s string
	idpath := e.PathIds
	for len(idpath) > 1 {
		s, idpath = idpath[0], idpath[1:]
		res += s + ", "
	}
	return res + idpath[0] + "], LeafId: " + e.LeafId + ", Exp: " + e.Exp.String() + ")"
}

func NewUpdateStructExp(path interface{}, leafid string, exp interface{}) (Exp, error) {
	return UpdateStructExp{path.([]string), leafid, exp.(Exp)}, nil
}

/* InitExp */
type InitExp struct {
	Exp Exp
}

func (e InitExp) String() string {
	return fmt.Sprintf("InitExp(Exp: %s)", e.Exp.String())
}

func NewInitExp(id string, exp interface{}) (Exp, error) {
	if id != "storage" {
		return (ErrorExpression{"Inits must initialize storage only"}),
			errors.Errorf("init is not initializing storage, but %s", id)
	}
	return InitExp{exp.(Exp)}, nil
}

/* ExpList */
func NewExpList(exp interface{}) []Exp {
	return []Exp{exp.(Exp)}
}

func AddExpToList(exp, list interface{}) []Exp {
	e := exp.(Exp)
	l := list.([]Exp)
	return prependList(e, l)
}

// ---------------------------

func notImplemented() *notImplementedError { return &notImplementedError{} }

type notImplementedError struct{}

func (e *notImplementedError) Error() string {
	return fmt.Sprint("Not Implemented!")
}

type ErrorExpression struct {
	errormsg string
}

func (e ErrorExpression) String() string {
	return fmt.Sprintf("ErrorExpression(\"%s\")", e.errormsg)
}
func fail(str string) (Exp, error) {
	return ErrorExpression{str}, nil
}

func prependList(exp Exp, list []Exp) []Exp {
	return append([]Exp{exp}, list...)
}
