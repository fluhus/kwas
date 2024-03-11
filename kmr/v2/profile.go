// ProfileTuple logic.

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

type ProfileData struct {
	P Profile
	C ProfileSampleCounts
}

type ProfileHandler struct{}

// ProfileTuple holds a kmer and a distribution of bases around it.
type ProfileTuple = Tuple[ProfileHandler, *ProfileData]

func (h ProfileHandler) encode(p *ProfileData, w *bnry.Writer) error {
	return w.Write(p.P.flatten(), p.C[:])
}

func (h ProfileHandler) decode(p **ProfileData, r io.ByteReader) error {
	var pp, c []int64
	if err := bnry.Read(r, &pp, &c); err != nil {
		return err
	}
	if len(pp) != len((*p).P)*len((*p).P[0]) {
		return fmt.Errorf("unexpected profile length: %d, want %d",
			len(pp), len((*p).P))
	}
	if len(c) != len((*p).C) {
		return fmt.Errorf("unexpected counts length: %d, want %d",
			len(c), len((*p).C))
	}
	copy((*p).C[:], c)
	(*p).P.unflatten(pp)
	return nil
}

func (h ProfileHandler) merge(a, b *ProfileData) *ProfileData {
	p := a
	p.P.Add(&b.P)
	p.C.Add(&b.C)
	return p
}

func (h ProfileHandler) clone(p *ProfileData) *ProfileData {
	pp := &ProfileData{}
	*pp = *p
	return pp
}

func (h ProfileHandler) new() *ProfileData {
	return &ProfileData{}
}
