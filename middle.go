package hireq

import "net/http"

// Middleware interface for client request.
type Middleware interface {
	Handle(r *http.Request, next HandleFunc) (*Response, error)
}

// MiddleFunc implements the Middleware interface
type MiddleFunc func(r *http.Request, next HandleFunc) (*Response, error)

// Handle request
func (mf MiddleFunc) Handle(r *http.Request, next HandleFunc) (*Response, error) {
	return mf(r, next)
}

func (h *HiReq) wrapMiddlewares() {
	h.handler = func(r *http.Request) (*Response, error) {
		rawResp, err := h.client.Do(r)
		if err != nil {
			return nil, err
		}

		return &Response{
			Response: rawResp,
			// with decoder
			decoder: h.respDecoder,
		}, nil
	}

	for _, m := range h.middles {
		h.wrapMiddleware(m)
	}
}

func (h *HiReq) wrapMiddleware(m Middleware) {
	next := h.handler

	// wrap handler
	h.handler = func(r *http.Request) (*Response, error) {
		return m.Handle(r, next)
	}
}
