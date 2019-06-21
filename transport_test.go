package brahms

import "testing"

func TestNetCoreTranport(t *testing.T) {
	n1 := N("127.0.0.1", 1)
	n2 := N("127.0.0.1", 2)

	tr := NewMemNetTransport()

	t.Run("push", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		tr.Push(nil, n1, n2.Hash())
	})

	t.Run("pull", func(t *testing.T) {
		defer func() {
			if r := recover(); r == nil {
				t.Errorf("The code did not panic")
			}
		}()
		tr.Pull(nil, nil, n2.Hash())
	})
}
