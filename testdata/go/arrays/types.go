package arrays

type ArraysOfPrimitives struct {
	U8Array []uint8
	U32Array []uint32
	F64Array []float64
	StrArray []string
	BoolArray []bool
}

type Item struct {
	Id uint32
	Name string
}

type ArraysOfStructs struct {
	Items []Item
	Count uint32
}
