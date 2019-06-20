package brahms

import (
	"testing"

	"github.com/advanderveer/go-test"
)

func TestParams(t *testing.T) {
	p1, err := NewParams(0.1, 0.7, 0.2, 100, 10)
	test.Ok(t, err)
	test.Equals(t, 10, p1.αl1())
	test.Equals(t, 70, p1.βl1())
	test.Equals(t, 20, p1.γl1())
	test.Equals(t, 10, p1.l2())
}

func TestParamsFails(t *testing.T) {
	_, err := NewParams(0.1, 0.6, 0.2, 100, 10)
	test.Equals(t, ErrPartsDonAddToOne, err)

	_, err = NewParams(0.1, 0.7, 0.2, 1, 10)
	test.Equals(t, ErrL1AtLeast, err)

	_, err = NewParams(0.1, 0.7, 0.2, 10, 0)
	test.Equals(t, ErrL2AtLeast, err)
}
