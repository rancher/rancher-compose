package locks

var lockRequests chan *lockRequest

func init() {
	lockRequests = make(chan *lockRequest)
	go locker()
}

// Interface for unlocking a key that was provided to Lock()
type Unlocker interface {
	Unlock()
}

// Provides an application-wide lock on the provided key.
// If the lock is obtained, it will return an Unlocker.
// If the lock is not successfully obtained, nil will be returned.
func Lock(key string) Unlocker {
	lockRequest := newLockRequest(key, LOCK)
	lockRequests <- lockRequest
	return <-lockRequest.response
}

type operation int

const (
	LOCK operation = iota
	UNLOCK
)

type lockRequest struct {
	key      string
	op       operation
	response chan Unlocker
}

func newLockRequest(key string, op operation) *lockRequest {
	return &lockRequest{
		key:      key,
		op:       op,
		response: make(chan Unlocker, 1),
	}
}

type unlockerImpl struct {
	key string
}

func (u *unlockerImpl) Unlock() {
	lockRequests <- newLockRequest(u.key, UNLOCK)
}

func newUnlocker(key string) *unlockerImpl {
	return &unlockerImpl{key: key}
}

func locker() {
	// note: bool value is meaningless. This is a set
	lockedItems := make(map[string]bool)
	for {
		lockReq := <-lockRequests
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
