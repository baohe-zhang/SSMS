package main 

import (
	"fmt"
)

type MemberList struct {
	Members []*Member
	size int
}

type Member struct {
	TimeStamp uint64
	IP uint32
	State uint8
}

func NewMemberList(capacity int) *MemberList {
	ml := MemberList{}
	ml.Members = make([]*Member, capacity)
	return &ml
}

func (ml *MemberList) Size() int {
	return ml.size
}

func (ml *MemberList) Retrieve(ts uint64, ip uint32) Member{
	idx := ml.Select(ts, ip)
	if idx > -1 {
		return *ml.Members[idx]
	} else {
		panic("[ERROR]: invalid retrieve")
	}
}

func (ml *MemberList) Insert(ts uint64, ip uint32, state uint8) {
	// Resize when needed
	// if ml.size == len(ml.Members) {
	// 	ml.resize(ml.size * 2)
	// }
	// Insert new member
	ml.Members[ml.size] = &Member{ts, ip, state}
	ml.size += 1
}

func (ml *MemberList) Delete(ts uint64, ip uint32) {
	idx := ml.Select(ts, ip)
	if idx > -1 {
		// Replace the delete member with the last member
		ml.Members[idx] = ml.Members[ml.size - 1]
		ml.size -= 1
	} else {
		panic("[ERROR]: invalid delete")
	}
}

func (ml *MemberList) Update(ts uint64, ip uint32, state uint8) {
	idx := ml.Select(ts, ip)
	if idx > -1 {
		ml.Members[idx].State = state
	} else {
		panic("[ERROR]: invalid update")
	}
}

func (ml *MemberList) Select(ts uint64, ip uint32) int {
	for idx := 0; idx < ml.size; idx += 1 {
		if (ml.Members[idx].TimeStamp == ts) && (ml.Members[idx].IP == ip) {
			// Search hit
			return idx
		}
	}
	// Search failed
	return -1
}

// func (ml *MemberList) Resize(capacity int)


// Test client
// func main() {
// 	ml := NewMemberList(10)
// 	ml.Insert(0, 0 ,0)
// 	ml.Insert(1, 2, 3)
// 	ml.Insert(3, 2, 1)
// 	ml.Delete(3, 2)
// 	ml.Insert(4, 5, 5)
// 	ml.Insert(8, 2, 1)
// 	ml.Insert(7, 5, 5)
// 	x := ml.Retrieve(1, 2)
// 	fmt.Printf("state: %d\n", x.State)
// 	fmt.Printf("size: %d\n", ml.Size())
// }








