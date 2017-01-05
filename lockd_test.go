package lockd_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/lockd"
)

func TestLockd(t *testing.T) {
	t.Run("lockd with string", func(t *testing.T) {

		timeout := 1
		names := "ab"
		a := assert.New(t)
		locker := lockd.NewApp()

		res, err := locker.Lock(time.Duration(timeout)*time.Second, names)
		a.Empty(err)
		a.NotEmpty(res)

	})
	t.Run("lockd with remove key", func(t *testing.T) {
		timeout := 1
		names := "abc"
		a := assert.New(t)
		locker := lockd.NewApp()

		res, err := locker.Lock(time.Duration(timeout)*time.Second, names)
		a.Empty(err)
		a.NotEmpty(res)

		res, err = locker.UnlockKey(names)
		a.Empty(err)
		a.Equal("Unlock key: abc ok", res)

	})

	t.Run("lockd with big goroutine ", func(t *testing.T) {
		timeout := 1
		names := "abcd"
		a := assert.New(t)
		locker := lockd.NewApp()
		var wg sync.WaitGroup
		wg.Add(1000)
		for i := 0; i < 1000; i++ {

			go func() {
				newi := strconv.Itoa(i)
				newname := names + newi

				locker.Lock(time.Duration(timeout)*time.Second, newname)
				wg.Done()
			}()
		}
		wg.Wait()
		res, err := locker.Lock(time.Duration(timeout)*time.Second, "dsafafsafs")
		a.Empty(err)

		a.Equal("dsafafsafs", res)

	})

}
