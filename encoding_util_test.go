package vnc2video

import "testing"

func TestSetChanged(t *testing.T) {
	canvas := &VncCanvas{}
	rect := &Rectangle{X: 1, Y: 1, Width: 1024, Height: 64}
	canvas.SetChanged(rect)
	if canvas.Changed["64,0"] == false ||
		canvas.Changed["64,1"] == false ||
		canvas.Changed["64,4"] == false {
		t.Fail()
	}

}
