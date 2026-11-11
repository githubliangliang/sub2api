package httputil

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Run explicitly in a 1GiB/one-CPU container; normal unit tests use small fixtures.
func TestRequestBodyBudgetOneGiBStress(t *testing.T) {
	if os.Getenv("SUB2API_RESOURCE_STRESS") != "1" {
		t.Skip("isolated 1GiB stress runner only")
	}
	max, err := os.ReadFile("/sys/fs/cgroup/memory.max")
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(string(max)) != "1073741824" {
		t.Fatalf("expected 1GiB cgroup, got %s", max)
	}
	budget := NewBodyMemoryBudget(64 << 20)
	releaseWave := make(chan struct{})
	var accepted, rejected atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx, release := WithRequestBodyMemoryBudget(r.Context(), budget)
		defer release()
		body, err := ReadRequestBodyWithPrealloc(r.WithContext(ctx))
		if err != nil {
			rejected.Add(1)
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(503)
			return
		}
		accepted.Add(1)
		<-releaseWave
		runtime.KeepAlive(body)
		w.WriteHeader(204)
	}))
	defer server.Close()
	payload := bytes.Repeat([]byte("x"), 4<<20)
	started := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r, err := server.Client().Post(server.URL, "application/json", bytes.NewReader(payload))
			if err != nil {
				t.Error(err)
				return
			}
			_, _ = io.Copy(io.Discard, r.Body)
			_ = r.Body.Close()
		}()
	}
	deadline := time.Now().Add(10 * time.Second)
	for accepted.Load()+rejected.Load() < 32 && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}
	close(releaseWave)
	wg.Wait()
	if accepted.Load() == 0 || rejected.Load() == 0 {
		t.Fatalf("expected admission and overload: accepted=%d rejected=%d", accepted.Load(), rejected.Load())
	}
	budget.mu.Lock()
	used := budget.used
	budget.mu.Unlock()
	if used != 0 {
		t.Fatalf("leaked %d bytes", used)
	}
	response, err := server.Client().Post(server.URL, "application/json", bytes.NewReader(payload))
	if err != nil {
		t.Fatal(err)
	}
	_ = response.Body.Close()
	if response.StatusCode != 204 {
		t.Fatal("budget did not recover")
	}
	peak, _ := os.ReadFile("/sys/fs/cgroup/memory.peak")
	current, _ := os.ReadFile("/sys/fs/cgroup/memory.current")
	t.Logf("1GiB cgroup, 1 CPU, budget=64MiB, 32 x 4MiB concurrent HTTP bodies: accepted=%d rejected=%d elapsed=%s memory_peak=%s memory_current=%s", accepted.Load(), rejected.Load(), time.Since(started), bytes.TrimSpace(peak), bytes.TrimSpace(current))
	fmt.Fprintln(os.Stdout, "budget recovered; used=0 after wave")
}
