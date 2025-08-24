package model

import "time"

// NormalizeTick guarantees UTC timestamps and sane defaults.
func NormalizeTick(t Tick) Tick {
	if !t.Ts.IsZero() {
		t.Ts = t.Ts.UTC()
	} else {
		t.Ts = time.Now().UTC()
	}

	return t
}
