package nested

type Point struct {
	X float32
	Y float32
}

type Rectangle struct {
	TopLeft Point
	BottomRight Point
	Color uint32
}

type Scene struct {
	Name string
	MainRect Rectangle
	Count uint32
}
