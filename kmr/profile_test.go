package kmr

import (
	"bytes"
	"reflect"
	"testing"

	"golang.org/x/exp/slices"
)

func TestProfileAdd(t *testing.T) {
	p1 := &Profile{
		3: [4]int64{1, 2, 3, 4},
		4: [4]int64{7, 5, 8, 6},
	}
	p2 := &Profile{
		3: [4]int64{20, 30, 40, 50},
		5: [4]int64{11, 7, 3, 6},
	}
	want := &Profile{
		3: [4]int64{21, 32, 43, 54},
		4: [4]int64{7, 5, 8, 6},
		5: [4]int64{11, 7, 3, 6},
	}

	p1.Add(p2)

	if !reflect.DeepEqual(p1, want) {
		t.Fatalf("Add(...)=%v, want %v", p1, want)
	}
}

func TestProfileFillForward(t *testing.T) {
	t.SkipNow()
	p := &Profile{}
	want := &Profile{
		100: [4]int64{2, 0, 0, 0},
		101: [4]int64{1, 1, 0, 0},
		102: [4]int64{1, 1, 0, 0},
		103: [4]int64{0, 2, 0, 0},
		104: [4]int64{0, 0, 1, 0},
		105: [4]int64{0, 0, 0, 1},
	}
	p.FillForward([]byte("AACCGT"))
	p.FillForward([]byte("ACAC"))

	if !reflect.DeepEqual(p, want) {
		t.Fatalf("FillForward(...)=%v, want %v", p, want)
	}
}

func TestProfileFill(t *testing.T) {
	p := &Profile{}
	want := &Profile{
		98:  [4]int64{1, 0, 1, 0},
		99:  [4]int64{0, 1, 0, 1},
		100: [4]int64{2, 0, 0, 0},
		101: [4]int64{0, 2, 0, 0},
		102: [4]int64{0, 0, 2, 0},
		103: [4]int64{0, 0, 0, 2},
		104: [4]int64{2, 0, 0, 0},
		105: [4]int64{0, 0, 2, 0},
		106: [4]int64{0, 0, 0, 2},
		107: [4]int64{2, 0, 0, 0},
		108: [4]int64{0, 0, 2, 0},
		109: [4]int64{2, 0, 0, 0},
		110: [4]int64{0, 0, 0, 2},
		111: [4]int64{2, 0, 0, 0},
		112: [4]int64{0, 0, 2, 0},
		113: [4]int64{0, 0, 2, 0},
		114: [4]int64{0, 0, 0, 2},
		115: [4]int64{0, 0, 2, 0},
		116: [4]int64{0, 2, 0, 0},
		117: [4]int64{2, 0, 0, 0},
		118: [4]int64{0, 0, 2, 0},
		119: [4]int64{0, 2, 0, 0},
		120: [4]int64{0, 0, 2, 0},
		121: [4]int64{2, 0, 0, 0},
		122: [4]int64{0, 0, 0, 2},
		123: [4]int64{0, 0, 2, 0},
		124: [4]int64{2, 0, 0, 0},
		125: [4]int64{2, 0, 0, 0},
		126: [4]int64{0, 2, 0, 0},
		127: [4]int64{0, 2, 0, 0},
		128: [4]int64{0, 0, 2, 0},
		129: [4]int64{0, 0, 0, 2},
		130: [4]int64{0, 2, 0, 0},
		131: [4]int64{0, 0, 2, 0},
		132: [4]int64{2, 0, 0, 0},
		133: [4]int64{0, 2, 0, 0},
		134: [4]int64{1, 0, 0, 1},
	}
	p.Fill([]byte("GTACGTAGTAGATAGGTGCAGCGATGAACCGTCGACA"), 2)
	p.Fill([]byte("ACACGTAGTAGATAGGTGCAGCGATGAACCGTCGACT"), 2)

	if !reflect.DeepEqual(p, want) {
		for i := range p {
			if !reflect.DeepEqual(p[i], want[i]) {
				t.Errorf("FillForward(...)[%d]=%v, want %v", i, p[i], want[i])
			}
		}
	}
}

func TestProfileFillBackward(t *testing.T) {
	t.SkipNow()
	p := &Profile{}
	want := &Profile{
		94: [4]int64{1, 0, 0, 0},
		95: [4]int64{1, 0, 0, 0},
		96: [4]int64{1, 1, 0, 0},
		97: [4]int64{0, 2, 0, 0},
		98: [4]int64{1, 0, 1, 0},
		99: [4]int64{0, 1, 0, 1},
	}
	p.FillBackward([]byte("AACCGT"))
	p.FillBackward([]byte("ACAC"))

	if !reflect.DeepEqual(p, want) {
		t.Fatalf("FillBackward(...)=%v, want %v", p, want)
	}
}

func TestProfileSampleCount(t *testing.T) {
	input := Profile{
		10: [4]int64{1},
		11: [4]int64{10},
		12: [4]int64{100},
	}
	want := ProfileSampleCounts{10: 1, 11: 1, 12: 1}
	got := input.SingleSampleCount()
	if !slices.Equal(want[:], got[:]) {
		t.Fatalf("SingleSampleCount()=%v, want %v", got, want)
	}
}

func TestProfileTupleEncode(t *testing.T) {
	input := &ProfileTuple{
		Kmer: FullKmer{1: 2, 3: 4, 5: 6},
		P: Profile{
			10: [4]int64{20, 40, 50, 33},
			12: [4]int64{11111, 12, 0, 0},
		},
		C: ProfileSampleCounts{10: 13, 12: 100},
	}
	want := &ProfileTuple{}
	*want = *input

	buf := bytes.NewBuffer(nil)
	if err := input.Encode(buf); err != nil {
		t.Fatalf("Encode() failed: %v", err)
	}

	got := &ProfileTuple{}
	if err := got.Decode(buf); err != nil {
		t.Fatalf("Decode() failed: %v", err)
	}

	if !reflect.DeepEqual(input, want) {
		t.Fatalf("Encode() modified input: %v, want %v",
			input, want)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Decode()=%v, want %v", got, want)
	}
}
