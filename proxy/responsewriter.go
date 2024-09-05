package proxy

import "net/http"

type NopResponseWriter struct {
	HeaderMap http.Header
	StatusCode int
}
func NewNopResponseWriter() *NopResponseWriter  {
	return &NopResponseWriter{
		HeaderMap:  make(http.Header),
	}
}
func (w *NopResponseWriter) Header() http.Header  {
	return w.HeaderMap
}
func (w *NopResponseWriter) Write(src []byte) (int, error)  {
	return len(src),nil
}
func (w *NopResponseWriter) WriteHeader(statusCode int)  {
	w.StatusCode = statusCode
}