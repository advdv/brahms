package httpt

import (
	"io"
	"net/http"

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
	b Brahms
	e func(r io.Writer) Encoder
	d func(r io.Reader) Decoder
}

// NewHandler initates a new handler
func NewHandler(b Brahms, enc func(r io.Writer) Encoder, dec func(r io.Reader) Decoder) *Handler {
	return &Handler{b, enc, dec}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/push":
		dec := h.d(r.Body)
		defer r.Body.Close()

		pr := new(MsgPushReq)
		err := dec.Decode(pr)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
			return
		}

		h.b.ReceiveNode(brahms.Node{IP: pr.IP, Port: pr.Port})

	case "/pull":
		enc := h.e(w)
		view := h.b.ReadView()
		resp := make(MsgPullResp, 0, len(view))
		for _, n := range view {
			resp = append(resp, MsgNode{n.IP, n.Port})
		}

		err := enc.Encode(resp)
		if err != nil {
			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}

	case "/probe":
		enc := h.e(w)
		err := enc.Encode(&MsgProbeResp{Active: h.b.IsActive()})
		if err != nil {
			http.Error(w,
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError)
			return
		}

	default:
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}
}
