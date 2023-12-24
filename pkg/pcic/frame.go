package pcic

type Frame struct {
	Chunks []Chunk
	size   int
}

func (f *Frame) Size() int {
	return f.size
}
