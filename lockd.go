package lockd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"time"
)

const timeFormat string = "2006-01-02 15:04:05"

var errLockTimeout = errors.New("lock timeout")

//App is ...
type App struct {
	m            sync.Mutex
	httpListener net.Listener

	lockstore map[string]*lockerDetail

	locksMutex  sync.Mutex
	locksMutex2 sync.Mutex
}
type lockerDetail struct {
	keyname    string
	lock       sync.Mutex
	ref        int
	createTime time.Time
}

type lockstores []*lockerDetail

//NewApp is ...
func NewApp() *App {
	a := new(App)

	a.lockstore = make(map[string]*lockerDetail)
	return a
}

//NewLocker is ..
func NewLocker(names string) *lockerDetail {
	l := new(lockerDetail)

	l.keyname = names
	l.createTime = time.Now()
	return l
}

//GetLockInfos is ...
func (a *App) GetLockInfos() []byte {
	var buf bytes.Buffer
	keyLocks := make(lockstores, 0, 1024)
	a.locksMutex.Lock()
	for _, l := range a.lockstore {
		keyLocks = append(keyLocks, l)
	}
	a.locksMutex.Unlock()

	buf.WriteString("key lock:\n")
	for _, l := range keyLocks {
		buf.WriteString(fmt.Sprintf("%s %v\t\n", l.keyname, l.createTime.Format(timeFormat)))
	}
	return buf.Bytes()

}

// Lock is...
func (a *App) Lock(timeout time.Duration, names string) (string, error) {

	idinfo := make(chan string, 1)
	errs := make(chan error, 1)
	//chmsg := make(chan string, 1)
	go func() {
		a.LockTimeout(idinfo, errs, timeout, names)
	}()
	select {
	//for broadcast
	// case <-chmsg:
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte("aaaaa"))
	// 	return
	case id := <-idinfo:

		return id, nil
	case err := <-errs:

		if err != nil && err != errLockTimeout {
			return "", err
		} else if err == errLockTimeout {
			return "", errLockTimeout
		}
	}
	return "", nil
}

//LockTimeout is ...
func (a *App) LockTimeout(idinfo chan string, errs chan error, timeout time.Duration, names string) {
	var (
		ctx context.Context
		//cancel context.CancelFunc
	)

	ctx, _ = context.WithTimeout(context.Background(), timeout)

	goon := a.checkItem(names)
	if goon {
		fmt.Println("wrong input")

		errs <- fmt.Errorf("the input keys have the key locked and unlock at the same time")
		return
	}

	res := a.getItem(ctx, timeout, names)

	if res {
		a.locksMutex2.Lock()
		a.lockstore[names].createTime = time.Now()
		a.lockstore[names].keyname = names
		a.locksMutex2.Unlock()
		idinfo <- names
		select {
		case <-ctx.Done():
			a.UnlockKey(names)
		}
	} else {
		errs <- errLockTimeout
	}

}
func (a *App) getItem(ctx context.Context, timeout time.Duration, names string) bool {

	ctx, _ = context.WithTimeout(ctx, timeout)

	a.getRes(names, timeout)
	//the channel get the lock is sucessfully get or not
	lockdone := make(chan bool, 1)

	waitover := make(chan bool, 1)

	go a.LockWitchTimer(ctx, names, lockdone, waitover, timeout)
	select {
	case <-lockdone:
		return true
	case <-waitover:

		return false
	}

}

func (a *App) getRes(key string, timeout time.Duration) {

	v, ok := a.lockstore[key]
	if ok {
		v.ref++

	} else {
		res := &lockerDetail{

			keyname: key,
			ref:     1,
		}
		a.lockstore[key] = res

	}

}

//LockWitchTimer is ...
func (a *App) LockWitchTimer(ctx context.Context, key string, lockdone chan bool, waitover chan bool, timeout time.Duration) {
	done := make(chan bool, 1)

	go func() {
		a.locksMutex2.Lock()
		val := a.lockstore[key]
		a.locksMutex2.Unlock()

		val.lock.Lock()

		done <- true
	}()

	select {
	case <-ctx.Done():

		waitover <- false

	case <-done:

		lockdone <- true

	}

}

//check wheter the keys are already locked or not
func (a *App) checkItem(names ...string) bool {
	var unlockcount int
	var lockcount int

	sort.Strings(names)

	a.locksMutex.Lock()
	//check wheter the key is locked already or not
	for _, key := range names {

		if _, ok := a.lockstore[key]; ok {
			fmt.Println("1")
			lockcount++
		} else {
			unlockcount++
			fmt.Println("222")
		}

	}
	a.locksMutex.Unlock()

	if lockcount >= 1 && unlockcount >= 1 {
		return true
	}
	return false

}

//UnlockKey is ...
func (a *App) UnlockKey(key string) error {
	if key == "" {
		return fmt.Errorf("empty lock names")
	}

	fmt.Println("baba", a.lockstore[key].ref)
	a.lockstore[key].lock.Unlock()

	a.locksMutex.Lock()

	a.lockstore[key].ref--
	if a.lockstore[key].ref <= 0 {
		delete(a.lockstore, key)
	}

	a.locksMutex.Unlock()
	return nil

}
