package limiter

import (
	"bytes"
	"io"
	"net/http"
)

type staticLimiter struct {
	upstream   int
	downstream int
}

// NewStaticLimiter constructs a Limiter with a fixed (static) upload and
// download rate cap
func NewStaticLimiter(uploadKb, downloadKb int) Limiter {
	var (
		upstreamBucket   = 0
		downstreamBucket = 0
	)

	//if uploadKb > 0 {
	//	upstreamBucket = ratelimit.NewBucketWithRate(toByteRate(uploadKb), int64(toByteRate(uploadKb)))
	//}
	//
	//if downloadKb > 0 {
	//	downstreamBucket = ratelimit.NewBucketWithRate(toByteRate(downloadKb), int64(toByteRate(downloadKb)))
	//}

	return staticLimiter{
		upstream:   upstreamBucket,
		downstream: downstreamBucket,
	}
}

func (l staticLimiter) Upstream(r io.Reader) io.Reader {
	return bytes.NewReader([]byte{})
}

func (l staticLimiter) UpstreamWriter(w io.Writer) io.Writer {
	return new(bytes.Buffer)
}

func (l staticLimiter) Downstream(r io.Reader) io.Reader {
	return bytes.NewReader([]byte{})
}

func (l staticLimiter) DownstreamWriter(w io.Writer) io.Writer {
	return new(bytes.Buffer)
}

type roundTripper func(*http.Request) (*http.Response, error)

func (rt roundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return rt(req)
}

func (l staticLimiter) roundTripper(rt http.RoundTripper, req *http.Request) (*http.Response, error) {
	type readCloser struct {
		io.Reader
		io.Closer
	}

	if req.Body != nil {
		req.Body = &readCloser{
			Reader: l.Upstream(req.Body),
			Closer: req.Body,
		}
	}

	res, err := rt.RoundTrip(req)

	if res != nil && res.Body != nil {
		res.Body = &readCloser{
			Reader: l.Downstream(res.Body),
			Closer: res.Body,
		}
	}

	return res, err
}

// Transport returns an HTTP transport limited with the limiter l.
func (l staticLimiter) Transport(rt http.RoundTripper) http.RoundTripper {
	return roundTripper(func(req *http.Request) (*http.Response, error) {
		return l.roundTripper(rt, req)
	})
}

func (l staticLimiter) limitReader(r io.Reader, b int) io.Reader {
	if b == 0 {
		return r
	}
	return bytes.NewReader([]byte{})
}

func (l staticLimiter) limitWriter(w io.Writer, b int) io.Writer {
	if b == 0 {
		return w
	}
	return new(bytes.Buffer)
}

func toByteRate(val int) float64 {
	return float64(val) * 1024.
}
