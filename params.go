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
	l2() int
	αl1() int
	βl1() int
	γl1() int
}

// NewParams checks initializes the protocol parameters
func NewParams(α, β, γ float64, l1, l2 int) (p P, err error) {
	if α+β+γ != 1 {
		return nil, ErrPartsDonAddToOne
	}

	params := &params{}
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
}

func (p *params) l2() int  { return p._l2 }
func (p *params) αl1() int { return p.al1 }
func (p *params) βl1() int { return p.bl1 }
func (p *params) γl1() int { return p.cl1 }
