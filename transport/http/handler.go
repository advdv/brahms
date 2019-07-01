package httpt

import (
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/advanderveer/brahms"
)

// Brahms provides the handler with the state of the algorithm
type Brahms interface {
	IsActive() bool
	ReceiveNode(other brahms.Node)
	ReadView() brahms.View
}

// Encoder is used for encoding messages to the handlers response
type Encoder interface {
	Encode(v interface{}) (err error)
}

// Decoder is used for decoding messages from the request
type Decoder interface {
	Decode(v interface{}) (err error)
}

// Handler handles brahms related messages
type Handler struct {
	C chan []byte

	to     time.Duration
	brahms Brahms
	enc    func(r io.Writer) Encoder
	dec    func(r io.Reader) Decoder
}

// NewHandlerWithEncoding initates a new handler with custom encoding
func NewHandlerWithEncoding(b Brahms, bufn int, to time.Duration, enc func(r io.Writer) Encoder, dec func(r io.Reader) Decoder) *Handler {
	return &Handler{make(chan []byte, bufn), to, b, enc, dec}
}

// NewHandler initates a new handler with default json encoding
func NewHandler(b Brahms, bufn int, to time.Duration) *Handler {
	return NewHandlerWithEncoding(b, bufn, to, func(w io.Writer) Encoder { return json.NewEncoder(w) },
		func(r io.Reader) Decoder { return json.NewDecoder(r) })

}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/push":
		defer r.Body.Close()

		pr := new(MsgPushReq)
		err := h.dec(r.Body).Decode(pr)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		h.brahms.ReceiveNode(brahms.Node{IP: pr.IP, Port: pr.Port})

	case "/pull":
		view := h.brahms.ReadView()
		resp := make(MsgPullResp, 0, len(view))
		for _, n := range view {
			resp = append(resp, MsgNode{n.IP, n.Port})
		}

		err := h.enc(w).Encode(resp)
		if err != nil {
			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}

	case "/probe":
		err := h.enc(w).Encode(&MsgProbeResp{Active: h.brahms.IsActive()})
		if err != nil {
			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}

	case "/emit":
		defer r.Body.Close()
		msg := new(MsgEmitReq)
		err := h.dec(r.Body).Decode(msg)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		if len(msg.Data) < 1 {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		select {
		case h.C <- msg.Data:
		case <-time.After(h.to):
			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}
