package sc

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	. "github.com/briansorahan/sc/types"
	. "github.com/briansorahan/sc/ugens"
	"io"
	"io/ioutil"
	"os"
)

const (
	synthdefStart   = "SCgf"
	synthdefVersion = 2
)

var byteOrder = binary.BigEndian

// synthdef defines the structure of synth def data as defined
// in http://doc.sccode.org/Reference/Synth-Definition-File-Format.html
type Synthdef struct {
	// Name is the name of the synthdef
	Name string `json:"name" xml:"Name,attr"`

	// Constants is a list of constants that appear in the synth def
	Constants []float32 `json:"constants" xml:"Constants>Constant"`

	// InitialParamValues is an array of initial values for synth params
	InitialParamValues []float32 `json:"initialParamValues" xml:"InitialParamValues>initialParamValue"`

	// ParamNames contains the names of the synth parameters
	ParamNames []ParamName `json:"paramNames" xml:"ParamNames>ParamName"`

	// Ugens is the list of ugens that appear in the synth def
	Ugens []*ugen `json:"ugens" xml:"Ugens>Ugen"`

	// Variants is the list of variants contained in the synth def
	Variants []*Variant `json:"variants" xml:"Variants>Variant"`

	// seen is an array of ugen nodes that have been added
	// to the synthdef
	seen []UgenNode
}

// Write writes a binary representation of a synthdef to an io.Writer.
// The binary representation written by this method is
// the data that scsynth expects at its /d_recv endpoint.
func (self *Synthdef) Write(w io.Writer) error {
	written, err := w.Write(bytes.NewBufferString(synthdefStart).Bytes())
	if written != len(synthdefStart) {
		return fmt.Errorf("Could not write synthdef")
	}
	if err != nil {
		return err
	}
	// write synthdef version
	err = binary.Write(w, byteOrder, int32(synthdefVersion))
	if err != nil {
		return err
	}
	// write number of synthdefs
	err = binary.Write(w, byteOrder, int16(1))
	if err != nil {
		return err
	}
	// write synthdef name
	name := newPstring(self.Name)
	err = name.Write(w)
	if err != nil {
		return err
	}
	// write number of constants
	err = binary.Write(w, byteOrder, int32(len(self.Constants)))
	if err != nil {
		return err
	}
	// write constant values
	for _, constant := range self.Constants {
		err = binary.Write(w, byteOrder, constant)
		if err != nil {
			return err
		}
	}
	// write number of params
	err = binary.Write(w, byteOrder, int32(len(self.ParamNames)))
	if err != nil {
		return err
	}
	// write initial param values
	// BUG(briansorahan) what happens in sclang when a ugen graph func
	//                   does not provide initial param values? do they
	//                   not appear in the synthdef? default to 0?
	for _, val := range self.InitialParamValues {
		err = binary.Write(w, byteOrder, val)
		if err != nil {
			return err
		}
	}
	// write number of param names
	err = binary.Write(w, byteOrder, int32(len(self.ParamNames)))
	if err != nil {
		return err
	}
	// write param names
	for _, p := range self.ParamNames {
		err = p.Write(w)
		if err != nil {
			return err
		}
	}
	// write number of ugens
	err = binary.Write(w, byteOrder, int32(len(self.Ugens)))
	if err != nil {
		return err
	}
	// write ugens
	for _, u := range self.Ugens {
		err = u.Write(w)
		if err != nil {
			return err
		}
	}
	// write number of variants
	err = binary.Write(w, byteOrder, int16(len(self.Variants)))
	if err != nil {
		return err
	}
	// write variants
	for _, v := range self.Variants {
		err = v.Write(w)
		if err != nil {
			return err
		}
	}
	return nil
}

// WriteJSON writes a json-formatted representation of a
// synthdef to an io.Writer
func (self *Synthdef) WriteJSON(w io.Writer) error {
	enc := json.NewEncoder(w)
	return enc.Encode(self)
}

// Bytes writes a synthdef to a byte array
func (self *Synthdef) Bytes() ([]byte, error) {
	arr := make([]byte, 0)
	buf := bytes.NewBuffer(arr)
	err := self.Write(buf)
	if err != nil {
		return arr, err
	}
	return buf.Bytes(), nil
}

// CompareToFile compares this synthdef to a synthdef
// stored on disk
func (self *Synthdef) CompareToFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	fromDisk, err := ioutil.ReadAll(f)
	if err != nil {
		return false, err
	}
	buf := bytes.NewBuffer(make([]byte, 0))
	err = self.Write(buf)
	if err != nil {
		return false, err
	}
	return compareBytes(buf.Bytes(), fromDisk), nil
}

// compareBytes returns true if two byte arrays
// are identical, false if they are not
func compareBytes(a, b []byte) bool {
	la, lb := len(a), len(b)
	if la != lb {
		fmt.Printf("different lengths a=%d b=%d\n", la, lb)
		return false
	}
	for i, octet := range a {
		if octet != b[i] {
			return false
		}
	}
	return true
}

// ReadSynthdef reads a synthdef from an io.Reader
func ReadSynthdef(r io.Reader) (*Synthdef, error) {
	// read the type
	startLen := len(synthdefStart)
	start := make([]byte, startLen)
	read, er := r.Read(start)
	if er != nil {
		return nil, er
	}
	if read != startLen {
		return nil, fmt.Errorf("Only read %d bytes of synthdef file", read)
	}
	actual := bytes.NewBuffer(start).String()
	if actual != synthdefStart {
		return nil, fmt.Errorf("synthdef started with %s instead of %s", actual, synthdefStart)
	}
	// read version
	var version int32
	er = binary.Read(r, byteOrder, &version)
	if er != nil {
		return nil, er
	}
	if version != synthdefVersion {
		return nil, fmt.Errorf("bad synthdef version %d", version)
	}
	// read number of synth defs
	var numDefs int16
	er = binary.Read(r, byteOrder, &numDefs)
	if er != nil {
		return nil, er
	}
	if numDefs != 1 {
		return nil, fmt.Errorf("multiple synthdefs not supported")
	}
	// read synthdef name
	defName, er := readPstring(r)
	if er != nil {
		return nil, er
	}
	// read number of constants
	var numConstants int32
	er = binary.Read(r, byteOrder, &numConstants)
	if er != nil {
		return nil, er
	}
	// read constants
	constants := make([]float32, numConstants)
	for i := 0; i < int(numConstants); i++ {
		er = binary.Read(r, byteOrder, &constants[i])
		if er != nil {
			return nil, er
		}
	}
	// read number of parameters
	var numParams int32
	er = binary.Read(r, byteOrder, &numParams)
	if er != nil {
		return nil, er
	}
	// read initial parameter values
	initialValues := make([]float32, numParams)
	for i := 0; i < int(numParams); i++ {
		er = binary.Read(r, byteOrder, &initialValues[i])
		if er != nil {
			return nil, er
		}
	}
	// read number of parameter names
	var numParamNames int32
	er = binary.Read(r, byteOrder, &numParamNames)
	if er != nil {
		return nil, er
	}
	// read param names
	paramNames := make([]ParamName, numParamNames)
	for i := 0; int32(i) < numParamNames; i++ {
		pn, er := readParamName(r)
		if er != nil {
			return nil, er
		}
		paramNames[i] = *pn
	}
	// read number of ugens
	var numUgens int32
	er = binary.Read(r, byteOrder, &numUgens)
	if er != nil {
		return nil, er
	}
	// read ugens
	ugens := make([]*ugen, numUgens)
	for i := 0; int32(i) < numUgens; i++ {
		ugen, er := readugen(r)
		if er != nil {
			return nil, er
		}
		ugens[i] = ugen
	}
	// read number of variants
	var numVariants int16
	er = binary.Read(r, byteOrder, &numVariants)
	if er != nil {
		return nil, er
	}
	// read variants
	variants := make([]*Variant, numVariants)
	for i := 0; int16(i) < numVariants; i++ {
		v, er := readVariant(r, numParams)
		if er != nil {
			return nil, er
		}
		variants[i] = v
	}
	// return a new synthdef
	synthDef := Synthdef{
		defName.String(),
		constants,
		initialValues,
		paramNames,
		ugens,
		variants,
		make([]UgenNode, 0),
	}
	return &synthDef, nil
}

// AddParams will do nothing if there are no synthdef params.
// If there are synthdef params it will
// (1) Add their default values to initialParamValues
// (2) Add their names/indices to paramNames
// (3) Add a Control ugen as the first ugen
func (self *Synthdef) AddParams(p *Params) {
	// HACK convert Params to an interface type
	paramList := p.List()
	numParams := len(paramList)
	self.InitialParamValues = make([]float32, numParams)
	self.ParamNames = make([]ParamName, numParams)
	for i, param := range paramList {
		self.InitialParamValues[i] = param.GetInitialValue()
		self.ParamNames[i] = ParamName{param.Name(), param.Index()}
	}
	if numParams > 0 {
		ctl := p.Control()
		self.seen = append(self.seen, ctl)
		control := []*ugen{cloneUgen(ctl)}
		self.Ugens = append(control, self.Ugens...)
	}
}

// addUgen adds a UgenNode to a synthdef and returns
// the ugen that was added, and the position in the
// ugens array
func (self *Synthdef) addUgen(u UgenNode) (*ugen, int) {
	for i, un := range self.seen {
		if un == u {
			return self.Ugens[i], i
		}
	}
	self.seen = append(self.seen, u)
	idx := len(self.Ugens)
	ugen := cloneUgen(u)
	self.Ugens = append(self.Ugens, ugen)
	return ugen, idx
}

// addConstant adds a constant to a synthdef and returns
// the index in the constants array where that constant is
// located.
// It ensures that constants are not added twice by returning the
// position in the constants array of the existing constant if
// you try to add a duplicate.
func (self *Synthdef) addConstant(c C) int {
	for i, f := range self.Constants {
		if f == float32(c) {
			return i
		}
	}
	l := len(self.Constants)
	self.Constants = append(self.Constants, float32(c))
	return l
}

// flatten
func (self *Synthdef) flatten(root UgenNode, params *Params) {
	self.AddParams(params)
	// get a topologically sorted ugens list
	ugenNodes := topsort(root)

	for _, ugenNode := range ugenNodes {
		// add ugen to synthdef
		ugen, _ := self.addUgen(ugenNode)
		// add inputs to synthdef and to ugen
		inputs := ugenNode.Inputs()

		var in *input

		for _, input := range inputs {
			switch v := input.(type) {
			case UgenNode:
				_, idx := self.addUgen(v)
				// will we ever need to use a different output index? [bps]
				in = newInput(int32(idx), 0)
				break
			case C:
				idx := self.addConstant(v)
				in = newInput(-1, int32(idx))
				break
			case *Param:
				idx := v.Index()
				in = newInput(0, idx)
				break
			case MultiInput:
				mins := v.InputArray()
				for _, min := range mins {
					switch x := min.(type) {
					case UgenNode:
						_, idx := self.addUgen(x)
						// will we ever need to use a different output index? [bps]
						in = newInput(int32(idx), 0)
						break
					case C:
						idx := self.addConstant(x)
						in = newInput(-1, int32(idx))
						break
					case *Param:
						idx := x.Index()
						in = newInput(0, idx)
						break
					}
					ugen.AppendInput(in)
				}
				continue
			// default:
			// 	fmt.Printf("unrecognized input type: %v\n", v)
			}
			ugen.AppendInput(in)
		}
	}
}

func newsynthdef(name string) *Synthdef {
	def := Synthdef{
		name,
		make([]float32, 0),
		make([]float32, 0),
		make([]ParamName, 0),
		make([]*ugen, 0),
		make([]*Variant, 0),
		make([]UgenNode, 0), // seen
	}
	return &def
}

// NewSynthdef creates a synthdef by traversing a ugen graph
func NewSynthdef(name string, graphFunc UgenGraphFunc) *Synthdef {
	def := newsynthdef(name)
	// It would be nice to parse synthdef params from function arguments
	// with the reflect package, but see
	// https://groups.google.com/forum/#!topic/golang-nuts/nM_ZhL7fuGc
	// for discussion of the (im)possibility of getting function argument
	// names at runtime.
	// Since this is not possible, what we need to do is let users add
	// synthdef params anywhere in their UgenGraphFunc using the Params interface.
	// Then in order to correctly map the values passed when creating
	// a synth node they have to be passed in the same order
	// they were created in the UgenGraphFunc.
	params := NewParams()
	root := graphFunc(params)
	def.flatten(root, params)
	return def
}

//                                 Out
//                                  |
//                              +-------+
//                              |       |
//                              0       AllpassC
//                                         |
//                          +--------+--------+--------+
//                          |        |        |        |
//               BinaryOpUgen      0.01     XLine     0.2
//                  |                         |
//              +--------+          +------+-------+-------+
//              |        |          |      |       |       |
//      WhiteNoise      0.1      0.0001   0.01     20      0
//
// constants: [0.1, 0.0001, 0.01, 20, 0, 0.2]
//
// ugens: [WhiteNoise, BinaryOpUgen, XLine, AllpassC, Out]

// topsort performs a depth-first-search of a ugen tree
func topsort(root UgenNode) []UgenNode {
	stack := newStack()
	topsortr(root, stack)
	n := stack.Size()
	ugens := make([]UgenNode, n)
	i := 0
	for v := stack.Pop(); v != nil; v = stack.Pop() {
		ugens[i] = v.(UgenNode)
		i = i + 1
	}
	return ugens
}

// topsortr performs a depth-first-search of a ugen tree
func topsortr(root UgenNode, stack *stack) {
	stack.Push(root)
	inputs := root.Inputs()
	n := len(inputs)
	if root.Name() == "BinaryOpUGen" {
		for _, input := range inputs {
			switch v := input.(type) {
			case UgenNode:
				topsortr(v, stack)
				break
			case MultiInput:
				mins := v.InputArray()
				for j := len(mins)-1; j >= 0; j-- {
					min := mins[j]
					switch w := min.(type) {
					case UgenNode:
						topsortr(w, stack)
						break
					}
				}
				break
			}
		}
	} else {
		for i := n - 1; i >= 0; i-- {
			input := inputs[i]
			switch v := input.(type) {
			case UgenNode:
				topsortr(v, stack)
				break
			case MultiInput:
				mins := v.InputArray()
				for j := len(mins)-1; j >= 0; j-- {
					min := mins[j]
					switch w := min.(type) {
					case UgenNode:
						topsortr(w, stack)
						break
					}
				}
				break
			}
		}
	}
}
