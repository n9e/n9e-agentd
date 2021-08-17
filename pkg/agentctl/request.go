package agentctl

/*
import (
	"context"
	"io"
	"time"

	"github.com/yubo/apiserver/pkg/rest"
	"github.com/yubo/golib/scheme"
)

type RequestOptions struct {
	method string
	url    string
	input  interface{}
	output interface{}
	cb     []func(interface{})
	ctx    context.Context
	client *rest.RESTClient
}
type RequestOption interface {
	apply(*RequestOptions)
}

type funcRequestOption struct {
	f func(*RequestOptions)
}

func (p *funcRequestOption) apply(opt *RequestOptions) {
	p.f(opt)
}

func newFuncRequestOption(f func(*RequestOptions)) *funcRequestOption {
	return &funcRequestOption{
		f: f,
	}
}

func WithCallback(cb ...func(interface{})) RequestOption {
	return newFuncRequestOption(func(o *RequestOptions) {
		o.cb = cb
	})
}
func WithContext(ctx context.Context) RequestOption {
	return newFuncRequestOption(func(o *RequestOptions) {
		o.ctx = ctx
	})
}
func WithClient(client *rest.RESTClient) RequestOption {
	return newFuncRequestOption(func(o *RequestOptions) {
		o.client = client
	})
}

// ("GET", "https://example.com/api/v{version}/{model}/{subject}?a=1&b=2", {"subject":"abc", "model": "instance", "version": 1}, nil)
func (p *RequestOptions) do() error {
	req := p.client.Verb(p.method).Prefix(p.url)

	if p.input != nil {
		req = req.VersionedParams(p.input, scheme.ParameterCodec)
	}

	if p.ctx == nil {
		p.ctx = context.Background()
	}

	if w, ok := p.output.(io.Writer); ok {
		b, err := req.Do(p.ctx).Raw()
		if err != nil {
			return err
		}

		if _, err := w.Write(b); err != nil {
			return err
		}
		w.Write([]byte("\n"))
		return nil
	}

	if err := req.Do(p.ctx).Into(p.output); err != nil {
		return err
	}

	for _, fn := range p.cb {
		if fn != nil {
			fn(p.output)
		}
	}

	return nil
}
*/
