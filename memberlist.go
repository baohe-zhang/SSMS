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

func (ml *MemberList) Retrieve(ts uint64, ip uint32) Member {
	idx := ml.Select(ts, ip)
	if idx > -1 {
		return *ml.Members[idx]
	} else {
		panic("[ERROR]: invalid retrieve")
	}
}

func (ml *MemberList) RetrieveByIdx(idx int) Member {
	if idx < ml.size && idx > -1 {
		return *ml.Members[idx]
	} else {
		panic("[ERROR]: invalid retrieve")
	}
}

func (ml *MemberList) Insert(m *Member) {
	// Resize when needed
	if ml.size == len(ml.Members) {
		ml.Resize(ml.size * 2)
	}
	// Insert new member
	ml.Members[ml.size] = m
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

func (ml *MemberList) Resize(capacity int) {
	members := make([]*Member, capacity)
	// Copy arrays
	for idx := 0; idx < ml.size; idx += 1 {
		members[idx] = ml.Members[idx]
	}
	ml.Members = members
}

func (ml *MemberList) PrintMemberList() {
	fmt.Printf("SIZE: %d, CAPACITY: %d\n", ml.size, len(ml.Members))
	for idx := 0; idx < ml.size; idx +=1 {
		m := ml.Members[idx]
		fmt.Printf("idx: %d, TS: %d, IP: %d, ST: %d\n", idx, 
			m.TimeStamp, m.IP, m.State)
	}
	fmt.Printf("\n")
}


// Test client
// func main() {
// 	ml := NewMemberList(1)

// 	m1 := Member{1, 1, 1}
// 	m2 := Member{2, 2, 2}
// 	m3 := Member{3, 3, 3}
// 	m4 := Member{4, 4, 4}
// 	m5 := Member{5, 5, 5}
// 	m6 := Member{6, 6, 6}


// 	ml.Insert(&m1)
// 	ml.PrintMemberList()
// 	ml.Insert(&m2)
// 	ml.PrintMemberList()
// 	ml.Insert(&m3)
// 	ml.PrintMemberList()
// 	ml.Delete(3, 3)
// 	ml.PrintMemberList()
// 	ml.Insert(&m4)
// 	ml.PrintMemberList()
// 	ml.Insert(&m5)
// 	ml.PrintMemberList()
// 	ml.Insert(&m6)
// 	ml.PrintMemberList()
// 	x := ml.Retrieve(2, 2)
// 	fmt.Printf("state: %d\n", x.State)
// 	ml.Update(2, 2, 4)
// 	x = ml.Retrieve(2, 2)
// 	fmt.Printf("state: %d\n", x.State)
// }








