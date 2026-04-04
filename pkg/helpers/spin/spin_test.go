package spin

import (
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	s := New("loading %s")
	if s == nil {
		t.Fatal("New() returned nil")
	}
	if len(s.frames) == 0 {
		t.Error("New() frames should not be empty")
	}
	if string(s.frames) != Default {
		t.Errorf("New() frames = %q, want default %q", string(s.frames), Default)
	}
	if s.tpf != 100*time.Millisecond {
		t.Errorf("New() tpf = %v, want 100ms", s.tpf)
	}
}

func TestNew_withOptions(t *testing.T) {
	s := New("loading %s",
		WithFrames(Spin1),
		WithTimePerFrame(200*time.Millisecond),
		WithDoneMessage("done!"),
	)

	if string(s.frames) != Spin1 {
		t.Errorf("WithFrames() frames = %q, want %q", string(s.frames), Spin1)
	}
	if s.tpf != 200*time.Millisecond {
		t.Errorf("WithTimePerFrame() tpf = %v, want 200ms", s.tpf)
	}
	if s.done != "done!" {
		t.Errorf("WithDoneMessage() done = %q, want %q", s.done, "done!")
	}
}

func TestSpinner_Set(t *testing.T) {
	s := New("test %s")
	s.Set(Spin2)
	if string(s.frames) != Spin2 {
		t.Errorf("Set() frames = %q, want %q", string(s.frames), Spin2)
	}
}

func TestSpinner_next(t *testing.T) {
	s := New("test %s", WithFrames(Spin1))
	frames := []rune(Spin1)

	for i := 0; i < len(frames)*2; i++ {
		got := s.next()
		want := string(frames[i%len(frames)])
		if got != want {
			t.Errorf("next() at pos %d = %q, want %q", i, got, want)
		}
	}
}

func TestSpinner_StopWithoutStart(t *testing.T) {
	s := New("test %s")
	if s.Stop() {
		t.Error("Stop() should return false when spinner was not started")
	}
}

func TestSpinner_StartStop(t *testing.T) {
	s := New("test %s", WithTimePerFrame(10*time.Millisecond))

	s.Start()

	// Double start should not panic
	s.Start()

	time.Sleep(30 * time.Millisecond)

	if !s.Stop() {
		t.Error("Stop() should return true when spinner was active")
	}

	// Double stop
	if s.Stop() {
		t.Error("Stop() should return false on second call")
	}
}

func TestSpinner_allFrameSets(t *testing.T) {
	frameSets := map[string]string{
		"Box1":  Box1,
		"Box2":  Box2,
		"Box3":  Box3,
		"Box4":  Box4,
		"Box5":  Box5,
		"Box6":  Box6,
		"Box7":  Box7,
		"Spin1": Spin1,
		"Spin2": Spin2,
		"Spin3": Spin3,
		"Spin4": Spin4,
		"Spin5": Spin5,
		"Spin6": Spin6,
		"Spin7": Spin7,
		"Spin8": Spin8,
	}

	for name, frames := range frameSets {
		t.Run(name, func(t *testing.T) {
			s := New("test %s", WithFrames(frames))
			if len(s.frames) == 0 {
				t.Errorf("%s: frames should not be empty", name)
			}
			// Verify next() cycles through all frames
			got := s.next()
			if got == "" {
				t.Errorf("%s: next() returned empty string", name)
			}
		})
	}
}
