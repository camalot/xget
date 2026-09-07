package cli

import (
	"testing"
)

// fakeProgress records Update/Stop calls instead of animating a real spinner,
// so tests can assert what a refresh loop reported without a terminal attached.
type fakeProgress struct {
	messages []string
	stopped  int
}

func (f *fakeProgress) Update(message string) {
	f.messages = append(f.messages, message)
}

func (f *fakeProgress) Stop() {
	f.stopped++
}

func stubProgress(t *testing.T) *fakeProgress {
	t.Helper()
	original := newProgress
	fake := &fakeProgress{}
	newProgress = func() progress { return fake }
	t.Cleanup(func() { newProgress = original })
	return fake
}

func TestRefreshInstalledStoreReportsProgressPerPackage(t *testing.T) {
	fake := stubProgress(t)
	useTempInstalledStore(t,
		samplePackage("bschaatsbergen/cidr", "v2.2.0"),
		samplePackage("nektos/act", "v0.2.89"),
	)
	stubRefresh(t, map[string]string{
		"bschaatsbergen/cidr": "v2.3.0",
		"nektos/act":          "v0.2.89",
	})

	if _, err := runCLI(t, "list", "--installed"); err != nil {
		t.Fatal(err)
	}

	want := []string{checkingMessage("bschaatsbergen/cidr"), checkingMessage("nektos/act")}
	if len(fake.messages) != len(want) {
		t.Fatalf("messages = %v, want %v", fake.messages, want)
	}
	for i, message := range want {
		if fake.messages[i] != message {
			t.Fatalf("messages[%d] = %q, want %q", i, fake.messages[i], message)
		}
	}
	if fake.stopped != 1 {
		t.Fatalf("stopped = %d, want 1", fake.stopped)
	}
}

func TestRefreshNamedPackagesReportsProgressPerLocation(t *testing.T) {
	fake := stubProgress(t)
	storePath := useTempInstalledStore(t, samplePackage("nektos/act", "v0.2.89"))
	stubRefresh(t, map[string]string{"nektos/act": "v0.3.0"})
	stubEngine(t, storePath, nil)

	if _, err := runCLI(t, "upgrade", "nektos/act"); err != nil {
		t.Fatal(err)
	}

	want := []string{checkingMessage("nektos/act")}
	if len(fake.messages) != len(want) {
		t.Fatalf("messages = %v, want %v", fake.messages, want)
	}
	if fake.messages[0] != want[0] {
		t.Fatalf("messages[0] = %q, want %q", fake.messages[0], want[0])
	}
	if fake.stopped != 1 {
		t.Fatalf("stopped = %d, want 1", fake.stopped)
	}
}
