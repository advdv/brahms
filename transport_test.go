package brahms

import "testing"

func TestNetCoreTranport(t *testing.T) {
	tr := NewMemNetTransport()

	t.Run("push", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		tr.Push(nil, NID{0x01}, NID{0x02})
	})

	t.Run("pull", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		tr.Pull(nil, nil, NID{0x02})
	})
}
