package httpt_test

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/advanderveer/brahms"
	httpt "github.com/advanderveer/brahms/transport/http"
	"github.com/advanderveer/go-test"
)

type mockBrahms struct {
	inactive bool
	pushes   []brahms.Node
}

func (b *mockBrahms) IsActive() bool                { return !b.inactive }
func (b *mockBrahms) ReceiveNode(other brahms.Node) { b.pushes = append(b.pushes, other) }
func (b *mockBrahms) ReadView() brahms.View         { return brahms.NewView(brahms.N("127.0.0.1", 8080)) }

func TestPushPullProbeEmit(t *testing.T) {
	b := &mockBrahms{}
	h := httpt.NewHandler(b, 0, time.Millisecond*100)
	s := httptest.NewServer(h)
	defer s.Close()

	t.Run("404 not found", func(t *testing.T) {
		r, err := http.Post(s.URL, "", nil)
		test.Ok(t, err)
		defer r.Body.Close()
		test.Equals(t, http.StatusNotFound, r.StatusCode)
	})

	t.Run("probe active", func(t *testing.T) {
		f := func() *httpt.MsgProbeResp {
			r, err := http.Post(s.URL+"/probe", "", nil)
			test.Ok(t, err)
			defer r.Body.Close()
			test.Equals(t, http.StatusOK, r.StatusCode)
			data, _ := ioutil.ReadAll(r.Body)
			probe := new(httpt.MsgProbeResp)
			err = json.Unmarshal(data, &probe)
			test.Ok(t, err)
			return probe
		}

		test.Equals(t, true, f().Active)
		b.inactive = true
		test.Equals(t, false, f().Active)
		b.inactive = false
	})

	t.Run("push", func(t *testing.T) {
		r, err := http.Post(s.URL+"/push", "", nil)
		test.Ok(t, err)
		test.Equals(t, http.StatusBadRequest, r.StatusCode)

		t.Run("valid push", func(t *testing.T) {
			r, err := http.Post(s.URL+"/push", "", strings.NewReader(`{"ip": "127.0.0.1", "port": 11000}`))
			test.Ok(t, err)
			test.Equals(t, http.StatusOK, r.StatusCode)

			test.Equals(t, 1, len(b.pushes))
			test.Equals(t, net.ParseIP("127.0.0.1"), b.pushes[0].IP)
			test.Equals(t, uint16(11000), b.pushes[0].Port)
		})
	})

	t.Run("pull", func(t *testing.T) {
		r, err := http.Post(s.URL+"/pull", "", nil)
		test.Ok(t, err)
		defer r.Body.Close()
		test.Equals(t, http.StatusOK, r.StatusCode)
		data, _ := ioutil.ReadAll(r.Body)

		var resp httpt.MsgPullResp
		test.Ok(t, json.Unmarshal(data, &resp))
		test.Equals(t, 1, len(resp))
		test.Equals(t, "127.0.0.1", resp[0].IP.String())
		test.Equals(t, uint16(8080), resp[0].Port)
	})

	t.Run("emit", func(t *testing.T) {

		//zer length data is not allowed
		r, err := http.Post(s.URL+"/emit", "", strings.NewReader(`{"data": ""}`))
		test.Ok(t, err)
		test.Equals(t, http.StatusBadRequest, r.StatusCode)

		//no-one receiving messages so should return too busy
		r, err = http.Post(s.URL+"/emit", "", strings.NewReader(`{"data": "`+base64.StdEncoding.EncodeToString([]byte("foo"))+`"}`))
		test.Ok(t, err)
		test.Equals(t, http.StatusTooManyRequests, r.StatusCode)

		done := make(chan struct{})
		go func() {
			msg := <-h.C
			test.Equals(t, []byte("foo"), msg)
			close(done)
		}()

		r, err = http.Post(s.URL+"/emit", "", strings.NewReader(`{"data": "`+base64.StdEncoding.EncodeToString([]byte("foo"))+`"}`))
		test.Ok(t, err)
		test.Equals(t, http.StatusOK, r.StatusCode)
		<-done
	})
}

type encoderFunc func(v interface{}) error

func (f encoderFunc) Encode(v interface{}) error { return f(v) }

func TestEncodingErrors(t *testing.T) {
	b := &mockBrahms{}
	s := httptest.NewServer(httpt.NewHandlerWithEncoding(b, 0, time.Second,
		func(w io.Writer) httpt.Encoder {
			return encoderFunc(func(v interface{}) error {
				return errors.New("foo")
			})
		},
		func(r io.Reader) httpt.Decoder { return json.NewDecoder(r) },
	))

	defer s.Close()

	t.Run("pull", func(t *testing.T) {
		r, err := http.Post(s.URL+"/pull", "", nil)
		test.Ok(t, err)
		test.Equals(t, http.StatusInternalServerError, r.StatusCode)
	})

	t.Run("probe", func(t *testing.T) {
		r, err := http.Post(s.URL+"/probe", "", nil)
		test.Ok(t, err)
		test.Equals(t, http.StatusInternalServerError, r.StatusCode)
	})

	t.Run("emit", func(t *testing.T) {
		r, err := http.Post(s.URL+"/emit", "", nil)
		test.Ok(t, err)
		test.Equals(t, http.StatusBadRequest, r.StatusCode)
	})
}
