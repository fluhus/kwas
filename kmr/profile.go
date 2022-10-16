package kmr

import (
	"fmt"
	"io"

	"github.com/fluhus/biostuff/sequtil"
	"github.com/fluhus/gostuff/bnry"
	"github.com/fluhus/gostuff/gnum"
)

const (
	maxReadLen = 100
)

type Profile [maxReadLen*2 + K][4]int64
type ProfileSampleCounts [maxReadLen*2 + K]int64

func (p *Profile) Add(other *Profile) {
	for i := range p {
		for j := range p[i] {
			p[i][j] += other[i][j]
		}
	}
}

// FillForward is deprecated.
//
// Deprecated: use Fill.
func (p *Profile) FillForward(seq []byte) {
	if len(seq) > len(p)/2 {
		panic(fmt.Sprintf("fill forward: seq is too long: %d, want at most %d",
			len(seq), len(p)/2))
	}
	for i := range seq {
		if seq[i] == 'N' {
			continue
		}
		p[len(p)/2+i][sequtil.Ntoi(seq[i])]++
	}
}

// FillBackward is deprecated.
//
// Deprecated: use Fill.
func (p *Profile) FillBackward(seq []byte) {
	if len(seq) > len(p)/2 {
		panic(fmt.Sprintf("fill backward: seq is too long: %d, want at most %d",
			len(seq), len(p)/2))
	}
	for i := range seq {
		if seq[i] == 'N' {
			continue
		}
		p[len(p)/2-len(seq)+i][sequtil.Ntoi(seq[i])]++
	}
}

func (p *Profile) Fill(seq []byte, kmerPos int) {
	offset := maxReadLen - kmerPos
	for i := range seq {
		if seq[i] == 'N' {
			continue
		}
		p[i+offset][sequtil.Ntoi(seq[i])]++
	}
}

func (p *Profile) SingleSampleCount() ProfileSampleCounts {
	var c ProfileSampleCounts
	for i := range c {
		if gnum.Sum(p[i][:]) > 0 {
			c[i] = 1
		}
	}
	return c
}

func (p *Profile) flatten() []int64 {
	result := make([]int64, 0, len(p)*len(p[0]))
	for i := range p {
		for j := range p[i] {
			result = append(result, p[i][j])
		}
	}
	return result
}

func (p *Profile) unflatten(a []int64) {
	for i := range p {
		for j := range p[i] {
			p[i][j] = a[0]
			a = a[1:]
		}
	}
}

func (c *ProfileSampleCounts) Add(other *ProfileSampleCounts) {
	for i := range c {
		c[i] += other[i]
	}
}

type ProfileSet[T comparable] map[T]*Profile

func (s ProfileSet[T]) Get(t T) *Profile {
	bl := s[t]
	if bl == nil {
		bl = &Profile{}
		s[t] = bl
	}
	return bl
}

func (s ProfileSet[T]) Add(other ProfileSet[T]) {
	for k, v := range other {
		s.Get(k).Add(v)
	}
}

type ProfileTuple struct {
	Kmer FullKmer
	P    Profile
	C    ProfileSampleCounts
}

func (t *ProfileTuple) GetKmer() []byte {
	return t.Kmer[:]
}

func (t *ProfileTuple) Encode(w io.Writer) error {
	return bnry.Write(w, t.Kmer[:], t.P.flatten(), t.C[:])

}

func (t *ProfileTuple) Decode(r io.ByteReader) error {
	var kmer []byte
	var p, c []int64
	if err := bnry.Read(r, &kmer, &p, &c); err != nil {
		return err
	}
	if len(kmer) != len(t.Kmer) {
		return fmt.Errorf("unexpected kmer length: %d, want %d",
			len(kmer), len(t.Kmer))
	}
	if len(p) != len(t.P)*len(t.P[0]) {
		return fmt.Errorf("unexpected profile length: %d, want %d",
			len(p), len(t.P))
	}
	if len(c) != len(t.C) {
		return fmt.Errorf("unexpected counts length: %d, want %d",
			len(c), len(t.C))
	}
	copy(t.Kmer[:], kmer)
	copy(t.C[:], c)
	t.P.unflatten(p)
	return nil
}

func (t *ProfileTuple) Add(other Tuple) {
	t.P.Add(&other.(*ProfileTuple).P)
	t.C.Add(&other.(*ProfileTuple).C)
}

func (t *ProfileTuple) Copy() Tuple {
	result := &ProfileTuple{}
	*result = *t
	return result
}
