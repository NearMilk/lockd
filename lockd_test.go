package lockd_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/teambition/lockd"
)

func TestLockd(t *testing.T) {
	t.Run("lockd with string", func(t *testing.T) {

		timeout := 5
		names := "abc"
		a := assert.New(t)
		locker := lockd.NewApp()

		res, err := locker.Lock(time.Duration(timeout)*time.Second, names)
		a.Empty(err)
		a.NotEmpty(res)

	})

}
