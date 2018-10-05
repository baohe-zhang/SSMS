package main

import (
	"errors"
	"fmt"
	"math/rand"
	"time"
)

// TTL Map, type as map[uint64]*Update
type TtlCache struct {
	TtlList []*Update
	Pointer int
	RandGen *rand.Rand
}

// Return a new TTL Map
func NewTtlCache() *TtlCache {
	// Source for genearting random number
	randSource := rand.NewSource(time.Now().UnixNano())
	randGen := rand.New(randSource)
	ttllist := make([]*Update, 0)

	fmt.Printf("[INFO]: TTL cache created\n")
	ttlcache := TtlCache{ttllist, 0, randGen}
	return &ttlcache
}

// Set the update packet in TTL Cache
func (tc *TtlCache) Set(val *Update) {
	if val.TTL == 0 {
		fmt.Printf("[INFO]: TTL cache cannot set for ttl=0 %d\n", val.UpdateID)
		return
	}
	tc.TtlList = append(tc.TtlList, val)
	fmt.Printf("[INFO]: TTL cache add a new\n")
}

// Get one entry each time in TTL Cache
func (tc *TtlCache) Get() (*Update, error) {
	if len(tc.TtlList) == 0 {
		fmt.Printf("[INFO]: TTL cache empty\n")
		return nil, errors.New("Empty TTL List, cannot Get()")
	}
	cur := tc.TtlList[tc.Pointer]
	cur.TTL -= 1
	if cur.TTL <= 0 {
		fmt.Printf("[INFO]: TTL cache expired %d\n", cur.UpdateID)
		// Delete this entry
		copy(tc.TtlList[tc.Pointer:], tc.TtlList[tc.Pointer+1:])
		tc.TtlList[len(tc.TtlList)-1] = nil
		tc.TtlList = tc.TtlList[:len(tc.TtlList)-1]
	}
	if len(tc.TtlList) != 0 {
		tc.Pointer = (tc.Pointer + 1) % len(tc.TtlList)
	} else {
		tc.Pointer = 0
	}
	return cur, nil
}

/*func main() {*/
//tc := NewTtlCache()
//u1 := Update{0, 3}
//tc.Set(&u1)
//tc.Set(&Update{0, 3})
//fmt.Println(len(tc.TtlList))
//u, err := tc.Get()
//fmt.Println(len(tc.TtlList), u.UpdateID)
//u, err = tc.Get()
//fmt.Println(len(tc.TtlList), u.UpdateID)
//u, err = tc.Get()
//fmt.Println(len(tc.TtlList), u.UpdateID)
//if err != nil {
//fmt.Println("ERR ", err)
//}
//u, err = tc.Get()
//fmt.Println(len(tc.TtlList), u.UpdateID)
//if err != nil {
//fmt.Println("ERR ", err)
//}
/*}*/
