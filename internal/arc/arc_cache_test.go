package arc

import (
	"github.com/stretchr/testify/require"
	"go.uber.org/atomic"
	"sync"
	"testing"
)

type fakeArchive struct {
	path       string
	closeCount atomic.Int32
}

var _ Archive = &fakeArchive{}

func (f *fakeArchive) Path() string {
	return f.path
}

func (f *fakeArchive) GetFiles() []File {
	panic("implement me")
}

func (f *fakeArchive) ReadFileAt(_ int64) ([]byte, error) {
	panic("implement me")
}

func (f *fakeArchive) GetFile(_ string) (*File, error) {
	panic("implement me")
}

func (f *fakeArchive) Close() error {
	f.closeCount.Add(1)
	return nil
}

func TestArchiveCache_GetSet(t *testing.T) {
	var actual Archive
	ac := NewArchiveCache(1)
	actual = ac.Get("k1")
	require.Nil(t, actual)

	f1 := &fakeArchive{path: "f1"}
	require.True(t, ac.Set("k1", f1))
	actual = ac.Get("k1")
	require.NotNil(t, actual)
	_ = actual.Close()
	// k1 not evicted yet, Close should do nothing
	require.Equal(t, int32(0), f1.closeCount.Load())

	// add k2, should evict k1
	f2 := &fakeArchive{path: "f2"}
	require.True(t, ac.Set("k2", f2))
	actual = ac.Get("k1")
	require.Nil(t, actual)
	require.Equal(t, int32(1), f1.closeCount.Load())

	// can not overwrite k2
	f3 := &fakeArchive{path: "f3"}
	require.False(t, ac.Set("k2", f3))

	// race Close
	const testCount = 1000
	var wg sync.WaitGroup
	wg.Add(testCount)
	for i := 0; i < testCount; i++ {
		go func() {
			defer wg.Done()
			a := ac.Get("k2")
			if a != nil { // could happen after ac.Remove
				_ = a.Close()
			}
		}()
	}

	ac.Remove("k2")
	wg.Wait()
	require.Equal(t, int32(1), f2.closeCount.Load())
}
