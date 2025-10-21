package inmemory_test

import (
	"context"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/thisiscetin/podpredict/internal/model"
	"github.com/thisiscetin/podpredict/internal/store"
	"github.com/thisiscetin/podpredict/internal/store/inmemory"
)

// mkPred creates a deterministic, distinguishable prediction for tests.
func mkPred(i int) store.Prediction {
	return store.Prediction{
		// Derive a stable UUID from the index 'i' so repeated calls with the same
		// argument produce identical IDs (critical for equality assertions).
		ID:        uuid.NewSHA1(uuid.NameSpaceOID, []byte(strconv.Itoa(i))).String(),
		Timestamp: time.Unix(1_700_000_000+int64(i), 0).UTC(),
		Input: model.Features{
			GMV:           float64(100*i + 1),
			Users:         float64(10*i + 2),
			MarketingCost: float64(i%3 + 3),
		},
		FEPods: i + 1,
		BEPods: (i + 1) * 2,
	}
}

func TestNewStore_IsEmpty(t *testing.T) {
	s := inmemory.NewStore()

	got, err := s.List(context.Background())
	require.NoError(t, err)
	assert.Len(t, got, 0, "new store should list zero predictions")
}

func TestAppendAndList_OrderAndCopyIsolation(t *testing.T) {
	s := inmemory.NewStore()

	// Append three distinct predictions (check insertion order is preserved)
	p1 := mkPred(0)
	p2 := mkPred(1)
	p3 := mkPred(2)

	require.NoError(t, s.Append(context.Background(), p1))
	require.NoError(t, s.Append(context.Background(), p2))
	require.NoError(t, s.Append(context.Background(), p3))

	// First list call
	got, err := s.List(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 3)

	assert.Equal(t, p1, got[0])
	assert.Equal(t, p2, got[1])
	assert.Equal(t, p3, got[2])

	// Mutate the returned slice (should NOT affect internal state)
	got[0] = mkPred(999)

	// List again and confirm original values are intact
	got2, err := s.List(context.Background())
	require.NoError(t, err)
	require.Len(t, got2, 3)

	assert.Equal(t, p1, got2[0], "internal slice must be protected from external mutation")
	assert.Equal(t, p2, got2[1])
	assert.Equal(t, p3, got2[2])
}

func TestConcurrentAppend_IsRaceSafeAndCounts(t *testing.T) {
	s := inmemory.NewStore()
	ctx := context.Background()

	const N = 200
	var wg sync.WaitGroup
	wg.Add(N)

	for i := 0; i < N; i++ {
		i := i
		go func() {
			defer wg.Done()
			_ = s.Append(ctx, mkPred(i))
		}()
	}
	wg.Wait()

	got, err := s.List(ctx)
	require.NoError(t, err)
	assert.Len(t, got, N, "store should contain all appended predictions")

	// (Optional) spot check a few items exist (order not guaranteed under concurrency)
	// We’ll just ensure at least one known element is present.
	want := mkPred(42)
	found := false
	for _, r := range got {
		if r == want {
			found = true
			break
		}
	}
	assert.True(t, found, "expected to find a previously appended record")
}

func TestList_ReturnsIndependentSlicesEachCall(t *testing.T) {
	s := inmemory.NewStore()
	ctx := context.Background()

	require.NoError(t, s.Append(ctx, mkPred(0)))
	require.NoError(t, s.Append(ctx, mkPred(1)))

	a, err := s.List(ctx)
	require.NoError(t, err)
	b, err := s.List(ctx)
	require.NoError(t, err)

	require.Len(t, a, 2)
	require.Len(t, b, 2)

	// mutate one result; the other should remain unaffected
	a[1] = mkPred(777)

	c, err := s.List(ctx)
	require.NoError(t, err)
	require.Len(t, c, 2)

	assert.Equal(t, mkPred(1), c[1], "fresh List() result must not be influenced by prior callers' mutations")
}
