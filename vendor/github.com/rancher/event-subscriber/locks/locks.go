package locks

var lockRequests = make(chan *lockRequest)

func init() {
	go locker()
}

type Locker interface {
	Lock() Unlocker
}

// Unlocker Interface for unlocking a key that was provided to Lock()
type Unlocker interface {
	Unlock()
}

type nopLocker struct{}

var theNopLocker = &nopLocker{}

// NopLocker returns a Locker which always works but does nothing: it's used to effectively skip locking.
func NopLocker() Locker {
	return theNopLocker
}

// Lock method of *nopLocker always succeeds but does not actually do anything.
func (nl *nopLocker) Lock() Unlocker {
	return nl
}

// Unlock method of *nopLocker does not actually do anything. It is here to satisfy the interface.
func (nl *nopLocker) Unlock() {
	return
}

type keyLocker struct {
	key interface{}
}

// KeyLocker provides an application-wide locker on the specified key.
func KeyLocker(key interface{}) Locker {
	return &keyLocker{key: key}
}

// Lock method works like this: if the lock is obtained, it will return an Unlocker.
// If the lock is not successfully obtained, nil will be returned.
func (kl *keyLocker) Lock() Unlocker {
	lockRequest := newLockRequest(kl.key, LOCK)
	lockRequests <- lockRequest
	return <-lockRequest.response
}

// Lock function provides a backwards compatible API for application-wide locking on a key
func Lock(key interface{}) Unlocker {
	return KeyLocker(key).Lock()
}

type operation int

const (
	LOCK operation = iota
	UNLOCK
)

type lockRequest struct {
	key      interface{}
	op       operation
	response chan Unlocker
}

func newLockRequest(key interface{}, op operation) *lockRequest {
	return &lockRequest{
		key:      key,
		op:       op,
		response: make(chan Unlocker, 1),
	}
}

type unlockerImpl struct {
	key interface{}
}

func (u *unlockerImpl) Unlock() {
	lockRequests <- newLockRequest(u.key, UNLOCK)
}

func newUnlocker(key interface{}) *unlockerImpl {
	return &unlockerImpl{key: key}
}

func locker() {
	// note: bool value is meaningless. This is a set
	lockedItems := make(map[interface{}]bool)
	for lockReq := range lockRequests {
		switch lockReq.op {
		case LOCK:
			if _, locked := lockedItems[lockReq.key]; locked {
				lockReq.response <- nil
			} else {
				lockedItems[lockReq.key] = true
				lockReq.response <- newUnlocker(lockReq.key)
			}
		case UNLOCK:
			delete(lockedItems, lockReq.key)
		}
	}
}
