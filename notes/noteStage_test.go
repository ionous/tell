package notes

import "testing"

// we always expect a value to intervene stages
func TestPanicLoop(t *testing.T) {
	stage := startStage
	func() {
		defer func() { _ = recover() }()
		stage.set(footerStage)
		t.Fatal("expected start to footer panic")
	}()
	stage = keyStage
	func() {
		defer func() { _ = recover() }()
		stage.set(keyStage)
		t.Fatal("expected key to key panic")
	}()
}
