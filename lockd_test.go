package lockd_test

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/lockd"
)

func Testlockd(t *testing.T) {
	t.Run("lockd", func(t *testing.T) {
		idinfo := make(chan uint64, 1)
		errs := make(chan error, 1)

		timeout := 5
		names := strings.Split("a,b,c", ",")
		a := assert.New(t)
		locker := lockd.NewApp()
		go func() {
			locker.LockTimeout(idinfo, errs, time.Duration(timeout)*time.Second, names)
		}()

		id := <-idinfo
		a.NotEmpty(id)
		a.Empty(errs)

	})
}
