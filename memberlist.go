package main 

import (
	"fmt"
	"math/rand"
)

type MemberList struct {
	Members []*Member
	size int
	curPos int
	shuffleList []int
}

type Member struct {
	TimeStamp uint64
	IP uint32
	State uint8
}

func NewMemberList(capacity int) *MemberList {
	ml := MemberList{}
	ml.Members = make([]*Member, capacity)
	fmt.Printf("[INFO]: Member list created\n")
	return &ml
}

func (ml *MemberList) Size() int {
	return ml.size
}

func (ml *MemberList) Retrieve(ts uint64, ip uint32) *Member {
	idx := ml.Select(ts, ip)
	if idx > -1 {
		return ml.Members[idx]
	} else {
		panic("[ERROR]: Invalid retrieve")
	}
}

func (ml *MemberList) RetrieveByIdx(idx int) *Member {
	if idx < ml.size && idx > -1 {
		return ml.Members[idx]
	} else {
		panic("[ERROR]: Invalid retrieve")
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
	// Log Insert
	fmt.Printf("[INFO]: Insert member ts: %d\n", m.TimeStamp)

	// Prolong the shuffle list
	ml.shuffleList = append(ml.shuffleList, len(ml.shuffleList))
	fmt.Printf("[INFO]: Prolong the length of shuffleList to: %d\n", len(ml.shuffleList))
}

func (ml *MemberList) Delete(ts uint64, ip uint32) {
	idx := ml.Select(ts, ip)
	if idx > -1 {
		// Replace the delete member with the last member
		ml.Members[idx] = ml.Members[ml.size - 1]
		ml.size -= 1
		fmt.Printf("[INFO]: Delete member ts: %d\n", ts)
	} else {
		panic("[ERROR]: Invalid delete")
	}

	// Shorten the shuffle list
	// Find the index of the maximum value in the shuffleList
	maxidx := 0
	for idx := 1; idx < len(ml.shuffleList); idx += 1 {
		if ml.shuffleList[idx] > ml.shuffleList[maxidx] {
			maxidx = idx
		}
	}
	// Delete this maximum value
	ml.shuffleList[maxidx] = ml.shuffleList[len(ml.shuffleList) - 1]
	ml.shuffleList = ml.shuffleList[:len(ml.shuffleList) - 1 ]
	fmt.Printf("[INFO]: Shorten the length of shuffleList to: %d\n", len(ml.shuffleList))
}

func (ml *MemberList) Update(ts uint64, ip uint32, state uint8) {
	idx := ml.Select(ts, ip)
	if idx > -1 {
		ml.Members[idx].State = state
		fmt.Printf("[INFO]: Update member ts: %d to state: %d\n", ts, state)
	} else {
		panic("[ERROR]: Invalid update")
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
	fmt.Printf("------------------------------------------\n")
	fmt.Printf("Size: %d, Capacity: %d\n", ml.size, len(ml.Members))
	for idx := 0; idx < ml.size; idx +=1 {
		m := ml.Members[idx]
		fmt.Printf("idx: %d, TS: %d, IP: %d, ST: %b\n", idx, 
			m.TimeStamp, m.IP, m.State)
	}
	fmt.Printf("------------------------------------------\n")
}

// Return an round-robin random IP address of member
func (ml *MemberList) Shuffle() uint32 {
	// Shuffle the shuffleList when the curPos comes to the end
	if ml.curPos == (len(ml.shuffleList) - 1) {
		ip := ml.Members[ml.shuffleList[ml.curPos]].IP
		ml.curPos = (ml.curPos + 1) % len(ml.shuffleList)
		// Shuffle the shuffleList
		rand.Shuffle(len(ml.shuffleList), func(i, j int) {
			ml.shuffleList[i], ml.shuffleList[j] = ml.shuffleList[j], ml.shuffleList[i]
		})
		fmt.Printf("[INFO]: IP: %d is selected by shuffling\n", ip)
		return ip
	} else {
		ip := ml.Members[ml.shuffleList[ml.curPos]].IP
		ml.curPos = (ml.curPos + 1) % len(ml.shuffleList)
		fmt.Printf("[INFO]: IP: %d is selected by shuffling\n", ip)
		return ip
	}
}


// // Test client
// func main() {
// 	ml := NewMemberList(1)

// 	m1 := Member{1, 1, 1}
// 	m2 := Member{2, 2, 2}
// 	m3 := Member{3, 3, 3}
// 	m4 := Member{4, 4, 4}
// 	m5 := Member{5, 5, 5}
// 	m6 := Member{6, 6, 6}
// 	m7 := Member{7, 7, 7}
// 	m8 := Member{8, 8, 8}
// 	m9 := Member{9, 9, 9}
// 	m10 := Member{10, 10, 10}


// 	// Test insert and delete
// 	ml.Insert(&m1)
// 	ml.Insert(&m2)
// 	ml.Insert(&m3)
// 	ml.Delete(3, 3)
// 	ml.Insert(&m4)
// 	ml.Insert(&m5)
// 	ml.Insert(&m6)

// 	// Test update
// 	x := ml.Retrieve(2, 2)
// 	fmt.Printf("origin state: %d\n", x.State)
// 	ml.Update(2, 2, 4)
// 	x = ml.Retrieve(2, 2)
// 	fmt.Printf("update state: %d\n", x.State)


// 	// Test shuffle for 4 rounds
// 	for i := 0; i < 4; i++ {
// 		for i := 0; i < len(ml.shuffleList); i++ {
// 			ml.Shuffle()
// 		}
// 	}

// 	// Test shuffle during insert and delete
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Insert(&m7)
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Insert(&m8)
// 	ml.Shuffle()
// 	ml.Delete(1,1)
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Insert(&m9)
// 	ml.Insert(&m10)
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// 	ml.Shuffle()
// }








