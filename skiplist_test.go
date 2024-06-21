package cache

import (
	"testing"
)

func Test_SkipList_Simple(t *testing.T) {
	sl := makeSkiplist()

	sl.insert("A", 100)

	if int32(1) != sl.length {
		t.Fatalf("skiplist's length should be 1 now, but %d", sl.length)
	}

	{
		rank := sl.getRank("A", 100) // rank is 1 based

		if int64(1) != rank {
			t.Fatalf("the first node in skiplist should has rank 1, but %d", rank)
		}

		A := sl.getByRank(rank)

		if nil == A {
			t.Fatalf("A should exist in skiplist")
		}

		if A.Member != "A" {
			t.Fatalf("expect: %s, but: %s", "A", A.Member)
		}

		if A.Score != 100 {
			t.Fatalf("expect: %v, but: %v", 100, A.Score)
		}
	}

	sl.insert("B", 100)

	if int32(2) != sl.length {
		t.Fatalf("skiplist's length should be 2 now, but %d", sl.length)
	}

	{
		rank := sl.getRank("B", 100)

		if int64(2) != rank {
			t.Fatalf("the first node in skiplist should has rank 2, but %d", rank)
		}

		B := sl.getByRank(rank)

		if nil == B {
			t.Fatalf("B should exist in skiplist")
		}

		if B.Member != "B" {
			t.Fatalf("expect: %s, but: %s", "B", B.Member)
		}

		if B.Score != 100 {
			t.Fatalf("expect: %v, but: %v", 100, B.Score)
		}
	}

	sl.insert("C", 200)

	if int32(3) != sl.length {
		t.Fatalf("skiplist's length should be 3 now, but %d", sl.length)
	}

	{
		rank := sl.getRank("C", 200)

		if int64(3) != rank {
			t.Fatalf("the first node in skiplist should has rank 3, but %d", rank)
		}

		C := sl.getByRank(rank)

		if nil == C {
			t.Fatalf("C should exist in skiplist")
		}

		if C.Member != "C" {
			t.Fatalf("expect: %s, but: %s", "C", C.Member)
		}

		if C.Score != 200 {
			t.Fatalf("expect: %v, but: %v", 100, C.Score)
		}
	}

	sl.remove("B", 100)

	if int32(2) != sl.length {
		t.Fatalf("skiplist's length should be 2 now, but %d", sl.length)
	}

	{
		{
			rank := sl.getRank("A", 100)

			if int64(1) != rank {
				t.Fatalf("the first node in skiplist should has rank 1, but %d", rank)
			}

			A := sl.getByRank(rank)

			if nil == A {
				t.Fatalf("A should exist in skiplist")
			}

			if A.Member != "A" {
				t.Fatalf("expect: %s, but: %s", "A", A.Member)
			}

			if A.Score != 100 {
				t.Fatalf("expect: %v, but: %v", 100, A.Score)
			}
		}

		{
			rank := sl.getRank("C", 200)

			if int64(2) != rank {
				t.Fatalf("the first node in skiplist should has rank 2, but %d", rank)
			}

			C := sl.getByRank(rank)

			if nil == C {
				t.Fatalf("C should exist in skiplist")
			}

			if C.Member != "C" {
				t.Fatalf("expect: %s, but: %s", "C", C.Member)
			}

			if C.Score != 200 {
				t.Fatalf("expect: %v, but: %v", 100, C.Score)
			}
		}
	}

}

func Test_SkipList_GetRangeByXXX(t *testing.T) {

	sl := makeSkiplist()

	sl.insert("A", 100)
	sl.insert("B", 100)
	sl.insert("C", 200)
	sl.insert("D", 300)
	sl.insert("E", 300)
	sl.insert("F", 400)
	sl.insert("G", 500)

	if int32(7) != sl.length {
		t.Fatalf("skiplist's length should be 7 now, but %d", sl.length)
	}

	{
		members := sl.GetRangeByScore(100, 300) // range is [100, 300), NOT (100, 300) or (100, 300]

		if int(3) != len(members) {
			t.Fatalf("members' length should be 3 now, but %d", len(members))
		}

		if string("A") != members[0] {
			t.Fatalf("member[0] should be A, but %s", members[0])
		}
		if string("B") != members[1] {
			t.Fatalf("member[1] should be B, but %s", members[1])
		}
		if string("C") != members[2] {
			t.Fatalf("member[2] should be C, but %s", members[2])
		}
	}

	{
		members := sl.GetRangeByRank(1, 4) // rank is 1 based, so here range is [1, 4)

		if int(3) != len(members) {
			t.Fatalf("members' length should be 3 now, but %d", len(members))
		}

		if string("A") != members[0] {
			t.Fatalf("member[0] should be A, but %s", members[0])
		}
		if string("B") != members[1] {
			t.Fatalf("member[1] should be B, but %s", members[1])
		}
		if string("C") != members[2] {
			t.Fatalf("member[2] should be C, but %s", members[2])
		}
	}

	{
		members := sl.GetRangeByRank(0, 4) // rank is 1 based, so here range is [1, 4)

		if int(3) != len(members) {
			t.Fatalf("members' length should be 3 now, but %d", len(members))
		}

		if string("A") != members[0] {
			t.Fatalf("member[0] should be A, but %s", members[0])
		}
		if string("B") != members[1] {
			t.Fatalf("member[1] should be B, but %s", members[1])
		}
		if string("C") != members[2] {
			t.Fatalf("member[2] should be C, but %s", members[2])
		}
	}

}
