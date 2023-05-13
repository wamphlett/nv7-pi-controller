package controller

type State struct {
	Speed  int
	Theme  string
	Colour string
}

type theme struct {
	name        string
	colourIndex int
	colours     []string
}

func (t *theme) nextColour() {
	t.colourIndex = (t.colourIndex + 1) % len(t.colours)
}
