package brahms

import (
	"errors"
	"strconv"
)

var (
	minL1 = 2
	minL2 = 1

	//ErrPartsDonAddToOne is returned when the partials don't add up to 1
	ErrPartsDonAddToOne = errors.New("α, β, γ don't add together to 1.0")

	//ErrL1AtLeast is returns if l1 was too low
	ErrL1AtLeast = errors.New("l1 must be at least " + strconv.Itoa(minL1))

	//ErrL2AtLeast is returned when l2 is too low
	ErrL2AtLeast = errors.New("l2 must be at least 1" + strconv.Itoa(minL2))
)

// P offers parameter values to the algorithm
type P interface {
	L2() int

	L1α() int
	L1β() int
	L1γ() int
	VN() int
}

// NewParams checks initializes the protocol parameters
func NewParams(α, β, γ float64, l1, l2, vn int) (p P, err error) {
	if α+β+γ != 1 {
		return nil, ErrPartsDonAddToOne
	}

	params := &params{vn: vn}
	if l1 < minL1 {
		return nil, ErrL1AtLeast
	}

	fl1 := float64(l1)
	params.al1 = int(fl1 * α)
	params.bl1 = int(fl1 * β)
	params.cl1 = int(fl1 * γ)

	if l2 < minL2 {
		return nil, ErrL2AtLeast
	}

	params._l2 = l2
	return params, nil
}

type params struct {
	_l2 int
	al1 int
	bl1 int
	cl1 int
	vn  int
}

func (p *params) L2() int  { return p._l2 }
func (p *params) L1α() int { return p.al1 }
func (p *params) L1β() int { return p.bl1 }
func (p *params) L1γ() int { return p.cl1 }
func (p *params) VN() int  { return p.vn }
