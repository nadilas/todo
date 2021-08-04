package testutils

import "testing"

// Scenario inspiration from:
//   https://levelup.gitconnected.com/an-alternative-approach-to-bdd-in-go-776bbbc24be9
func Scenario(t *testing.T, name string, steps ...func(t *testing.T)) {
	t.Run(name, func(t *testing.T) {
		for _, step := range steps {
			step(t)
		}
	})
}

func step(fn func(t *testing.T)) func(t *testing.T) {
	return fn
}

var Given = step
var When = step
var Then = step
var And = step
