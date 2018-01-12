package vnc2video

// Button represents a mask of pointer presses/releases.
type Button uint8

//go:generate stringer -type=Button

// All available button mask components.
const (
	BtnLeft Button = 1 << iota
	BtnMiddle
	BtnRight
	BtnFour
	BtnFive
	BtnSix
	BtnSeven
	BtnEight
	BtnNone Button = 0
)

// Mask returns button mask
func Mask(button Button) uint8 {
	return uint8(button)
}
