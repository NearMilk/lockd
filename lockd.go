package lockd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"
)

const timeFormat string = "2006-01-02 15:04:05"

var errLockTimeout = errors.New("lock timeout\n")

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
	workers    []*worker //prepare for broadcast
	lock2      sync.Mutex
}

//prepare for broadcast
type worker struct {
	source chan interface{}
}

type lockstores []*lockerDetail

//NewApp is ...
func NewApp() *App {
	a := new(App)

	a.lockstore = make(map[string]*lockerDetail)
	return a
}

//NewLocker is ..
func NewLocker(key string) *lockerDetail {
	l := new(lockerDetail)

	l.keyname = key
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
func (a *App) Lock(timeout time.Duration, key string) (string, error) {
	var (
		ctx    context.Context
		cancel context.CancelFunc
	)

	ctx, cancel = context.WithTimeout(context.Background(), timeout)
	a.getRes(key, timeout)
	idinfo := make(chan string, 1)
	errs := make(chan error, 1)
	a.locksMutex.Lock()
	num := len(a.lockstore[key].workers) - 1

	a.locksMutex.Unlock()
	go func() {
		a.LockTimeout(ctx, idinfo, errs, timeout, key)
	}()
	select {

	case id := <-idinfo:

		return id, nil
	case err := <-errs:

		if err != nil && err != errLockTimeout {
			return "", err
		} else if err == errLockTimeout {
			return "", errLockTimeout
		}
		return "", err
	case <-a.lockstore[key].workers[num].source:

		newerr := fmt.Errorf("bilibili\n")

		cancel()
		return "", newerr

	}

}

//LockTimeout is ...
func (a *App) LockTimeout(ctx context.Context, idinfo chan string, errs chan error, timeout time.Duration, key string) {

	res := a.getItem(ctx, timeout, key)

	if res {
		a.locksMutex2.Lock()
		a.lockstore[key].createTime = time.Now()
		a.lockstore[key].keyname = key
		a.locksMutex2.Unlock()
		idinfo <- key
		select {
		case <-ctx.Done():

			_, ok := a.lockstore[key]
			if ok {
				go func() { a.UnlockKey(key) }()
			} else {
				return
			}

		}
	} else {
		errs <- errLockTimeout
	}

}
func (a *App) getItem(ctx context.Context, timeout time.Duration, key string) bool {

	ctx, _ = context.WithTimeout(ctx, timeout)

	//the channel get the lock is sucessfully get or not
	lockdone := make(chan bool, 1)

	waitover := make(chan bool, 1)

	go a.LockWitchTimer(ctx, key, lockdone, waitover, timeout)
	select {
	case <-lockdone:
		return true
	case <-waitover:
		return false
	}

}

func (a *App) getRes(key string, timeout time.Duration) {

	w := &worker{}                          //prepare for broadcast
	w.source = make(chan interface{}, 1024) //prepare for broadcast

	v, ok := a.lockstore[key]
	if ok {
		v.ref++
		//v.workers = append(v.workers, myworker)
		a.lockstore[key].workers = append(a.lockstore[key].workers, w) //prepare for broadcast

	} else {
		res := &lockerDetail{

			keyname: key,
			ref:     1,
		}
		a.lockstore[key] = res
		a.lockstore[key].workers = append(a.lockstore[key].workers, w) //prepare for broadcast

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
		a.locksMutex2.Lock()

		a.locksMutex2.Unlock()
		waitover <- false

	case <-done:

		lockdone <- true

	}

}

//UnlockKey is ...
// func (a *App) UnlockKey(key string) error {
// 	if key == "" {
// 		return fmt.Errorf("empty lock key")
// 	}

// 	fmt.Println("Unlock", key)
// 	a.lockstore[key].lock.Unlock()

// 	a.locksMutex.Lock()

// 	//a.lockstore[key].ref--

// 	if a.lockstore[key].ref <= 0 {
// 		delete(a.lockstore, key)

// 	}

// 	a.locksMutex.Unlock()
// 	return nil

// }

//UnlockKey is ...
func (a *App) UnlockKey(key string) error {

	if key == "" {
		return fmt.Errorf("The key is empty!\n")
	}
	_, ok := a.lockstore[key]
	if ok {

		defer delete(a.lockstore, key)

		for kl := range a.lockstore[key].workers {

			a.lockstore[key].workers[kl].source <- 1
			a.lockstore[key].ref--

		}
	} else {
		return fmt.Errorf("The key does not exist\n")
	}

	return nil

}
