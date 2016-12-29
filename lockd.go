package lockd

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"sort"
	"sync"
	"sync/atomic"
	"time"
)

const timeFormat string = "2006-01-02 15:04:05"

var errLockTimeout = errors.New("lock timeout")

//App is ...
type App struct {
	m            sync.Mutex
	httpListener net.Listener
	//	keyLockerGroup *keyLockerGroup
	locks         map[uint64]*lockInfo
	lockstore     map[string]*lockerDetail
	lockIDCounter uint32
	locksMutex    sync.Mutex
	locksMutex2   sync.Mutex
}
type lockerDetail struct {
	duration time.Duration
	keyname  string
	lock     sync.Mutex
	ref      int
}

type lockInfo struct {
	id         uint64
	names      []string
	createTime time.Time
}
type lockInfos []*lockInfo

// type keyLockerGroup struct {
// 	set []*refLockSet
// }

//NewApp is ...
func NewApp() *App {
	a := new(App)
	a.locks = make(map[uint64]*lockInfo, 1024)
	a.lockstore = make(map[string]*lockerDetail)
	return a
}

// genLockID will gen the lockid
func (a *App) genLockID() uint64 {
	//todo, optimize later
	id := uint64(time.Now().Unix())
	c := uint64(atomic.AddUint32(&a.lockIDCounter, 1))
	return id<<32 | c
}

//NewLocker is ..
func NewLocker(id uint64, names []string) *lockInfo {
	l := new(lockInfo)
	l.id = id
	l.names = names
	l.createTime = time.Now()
	return l
}

//GetLockInfos is ...
func (a *App) GetLockInfos() []byte {
	var buf bytes.Buffer
	keyLocks := make(lockInfos, 0, 1024)
	a.locksMutex.Lock()
	for _, l := range a.locks {
		keyLocks = append(keyLocks, l)
	}
	a.locksMutex.Unlock()
	//sort.Sort(keyLocks)
	buf.WriteString("key lock:\n")
	for _, l := range keyLocks {
		buf.WriteString(fmt.Sprintf("%d %v\t%s\n", l.id, l.names, l.createTime.Format(timeFormat)))
	}
	return buf.Bytes()

}

func removeDuplicatedItems(keys ...string) []string {
	if len(keys) <= 1 {
		return keys
	}

	m := make(map[string]struct{}, len(keys))

	p := make([]string, 0, len(keys))
	for _, key := range keys {
		if _, ok := m[key]; !ok {
			m[key] = struct{}{}
			p = append(p, key)
		}

	}

	return p
}

//LockTimeout is ...
func (a *App) LockTimeout(idinfo chan uint64, errs chan error, timeout time.Duration, names []string) {
	var (
		ctx context.Context
		//cancel context.CancelFunc
	)

	ctx, _ = context.WithTimeout(context.Background(), timeout)

	goon := a.checkItem(names...)
	if goon {
		fmt.Println("wrong input")

		errs <- fmt.Errorf("the input keys have the key locked and unlock at the same time")
		return
	}
	lockdone1 := make(chan bool, 1)

	waitover1 := make(chan bool, 1)
	a.getItem(ctx, lockdone1, waitover1, timeout, names...)

	select {
	case <-lockdone1:
		id := a.genLockID()
		l := NewLocker(id, names)
		a.locksMutex2.Lock()
		a.locks[id] = l
		a.locksMutex2.Unlock()
		fmt.Println(id)
		idinfo <- id
		select {
		case <-ctx.Done():
			a.Unlock(id)
		}

	case <-waitover1:
		errs <- errLockTimeout

	}

}

func (a *App) getItem(ctx context.Context, lockdone1 chan bool, waitover1 chan bool, timeout time.Duration, names ...string) {

	ctx, _ = context.WithTimeout(ctx, timeout)
	for _, key := range names {
		//fmt.Println(key)

		a.getRes(key, timeout)
		//the channel get the lock is sucessfully get or not
		lockdone := make(chan bool, 1)

		waitover := make(chan bool, 1)

		go a.LockWitchTimer(ctx, key, lockdone, waitover, timeout)
		select {
		case <-lockdone:

		case <-waitover:
			waitover1 <- true
			return
		}

	}

	lockdone1 <- true
}

func (a *App) getRes(key string, timeout time.Duration) *lockerDetail {

	v, ok := a.lockstore[key]
	if ok {
		v.ref++
		return v
	} else {

		res := &lockerDetail{
			duration: time.Duration(timeout) * time.Second,
			keyname:  key,
			ref:      1,
		}
		a.lockstore[key] = res
		return res
	}

}

//LockWitchTimer is ...
func (a *App) LockWitchTimer(ctx context.Context, key string, lockdone chan bool, waitover chan bool, timeout time.Duration) {
	done := make(chan bool, 1)
	//	res := a.getRes(key, timeout)
	go func() {
		a.locksMutex2.Lock()
		val := a.lockstore[key]
		a.locksMutex2.Unlock()

		val.lock.Lock()

		done <- true
	}()

	select {
	case <-ctx.Done():
		fmt.Println("yoyoyo")
		waitover <- false

	case <-done:
		fmt.Println("accccc")
		lockdone <- true

	}

}

//check wheter the keys are already locked or not
func (a *App) checkItem(names ...string) bool {
	var unlockcount int
	var lockcount int
	//remove duplicated items
	names = removeDuplicatedItems(names...)
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

//Unlock is ...
func (a *App) Unlock(id uint64) error {
	if id == 0 {
		return fmt.Errorf("empty lock names")
	}
	a.locksMutex.Lock()
	l, _ := a.locks[id]
	delete(a.locks, id)
	a.locksMutex.Unlock()

	keys := removeDuplicatedItems(l.names...)
	for _, key := range keys {
		a.lockstore[key].lock.Unlock()
	}
	a.locksMutex.Lock()
	for _, key := range l.names {
		a.lockstore[key].ref--
		if a.lockstore[key].ref <= 0 {
			delete(a.lockstore, key)
		}
	}
	a.locksMutex.Unlock()
	return nil

}
