package greq

import "net/http"

// Middleware interface for cli request.
type Middleware interface {
	Handle(r *http.Request, next HandleFunc) (*Response, error)
}

// MiddleFunc implements the Middleware interface
type MiddleFunc func(r *http.Request, next HandleFunc) (*Response, error)

// Handle request
func (mf MiddleFunc) Handle(r *http.Request, next HandleFunc) (*Response, error) {
	return mf(r, next)
}

func (h *Client) wrapMiddlewares() {
	h.handler = func(r *http.Request) (*Response, error) {
		rawResp, err := h.doer.Do(r)
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

func (h *Client) wrapMiddleware(m Middleware) {
	next := h.handler

	// wrap handler
	h.handler = func(r *http.Request) (*Response, error) {
		return m.Handle(r, next)
	}
}
