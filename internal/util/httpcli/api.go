package httpcli

import "github.com/go-resty/resty/v2"

type IHttp interface {
	Get(url string, query map[string]string) (*resty.Response, error)
	Post(url string, body interface{}) (*resty.Response, error)
	Patch(url string, body interface{}) (*resty.Response, error)
}
