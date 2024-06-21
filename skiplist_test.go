package cache

import "testing"

func Test_Simple(t *testing.T) {
	sl := makeSkiplist()

	sl.insert("A", 100)

	if int32(1) != sl.length {
		t.Fatalf("skiplist's length should be 1 now, but %d", sl.length)
	}

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
