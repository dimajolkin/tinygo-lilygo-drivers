// Package tdeck provides the LilyGo T-Deck trackball driver (GPIO pulses on roll, OK button).
//
// Example: cursor from ReadMotion, click from Read().OK
//
//	tb := tdeck.NewTrackballDefault()
//	x, y := int16(160), int16(120)
//	for {
//	    dx, dy := tb.ReadMotion()
//	    s := tb.Read()
//	    x += int16(dx) * 4
//	    y += int16(dy) * 4
//	    if s.OK {
//	        // center button pressed
//	    }
//	    time.Sleep(10 * time.Millisecond)
//	}
package tdeck

import "machine"

type TrackballState struct {
	Left  bool
	Up    bool
	Right bool
	Down  bool
	OK    bool
}

type Trackball struct {
	left  machine.Pin
	up    machine.Pin
	right machine.Pin
	down  machine.Pin
	ok    machine.Pin
	last  [4]bool
}

func NewTrackball(left, up, right, down, ok machine.Pin) *Trackball {
	t := &Trackball{left: left, up: up, right: right, down: down, ok: ok}
	for _, p := range []machine.Pin{left, up, right, down, ok} {
		if p != machine.NoPin {
			p.Configure(machine.PinConfig{Mode: machine.PinInputPullup})
		}
	}
	return t
}

func NewTrackballDefault() *Trackball {
	return NewTrackball(
		machine.Pin(PinLeft),
		machine.Pin(PinUp),
		machine.Pin(PinRight),
		machine.Pin(PinDown),
		machine.Pin(PinOK),
	)
}

func (t *Trackball) pressed(p machine.Pin) bool {
	if p == machine.NoPin {
		return false
	}
	return !p.Get()
}

func (t *Trackball) Read() TrackballState {
	return TrackballState{
		Left:  t.pressed(t.left),
		Up:    t.pressed(t.up),
		Right: t.pressed(t.right),
		Down:  t.pressed(t.down),
		OK:    t.pressed(t.ok),
	}
}

func (t *Trackball) ReadKey() (byte, error) {
	s := t.Read()
	if s.OK {
		return 'O', nil
	}
	if s.Left {
		return 'L', nil
	}
	if s.Up {
		return 'U', nil
	}
	if s.Right {
		return 'R', nil
	}
	if s.Down {
		return 'D', nil
	}
	return 0, nil
}

// ReadMotion returns delta since last call (official firmware style: each pin
// state change = one step). Roll the trackball to get dx, dy.
func (t *Trackball) ReadMotion() (dx, dy int) {
	pins := []machine.Pin{t.right, t.up, t.left, t.down}
	for i := 0; i < 4; i++ {
		if pins[i] == machine.NoPin {
			continue
		}
		level := pins[i].Get()
		if level != t.last[i] {
			t.last[i] = level
			switch i {
			case 0:
				dx++
			case 1:
				dy--
			case 2:
				dx--
			case 3:
				dy++
			}
		}
	}
	return dx, dy
}
