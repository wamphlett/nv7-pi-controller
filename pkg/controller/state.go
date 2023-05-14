package controller

// State defines the state of the controller and is used
// when publishing events
type State struct {
	Speed  int
	Theme  string
	Colour string
}

// theme defines a configured theme
type theme struct {
	name        string
	colourIndex int
	colours     []string
}

// nextColour increments the colour on the theme
func (t *theme) nextColour() {
	t.colourIndex = (t.colourIndex + 1) % len(t.colours)
}
