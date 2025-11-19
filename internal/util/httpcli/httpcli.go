package httpcli

import "github.com/go-resty/resty/v2"

type Client struct {
	o *Option
}

func NewClient(opts ...OptionFunc) *Client {
	o := NewOption()
	o.ApplyOptions(opts...)
	return &Client{o: o}
}

func (c *Client) Get(url string, query map[string]string) (*resty.Response, error) {
	return c.o.client.R().SetQueryParams(query).Get(url)
}

func (c *Client) Post(url string, body interface{}) (*resty.Response, error) {
	return c.o.client.R().SetBody(body).Post(url)
}

func (c *Client) Patch(url string, body interface{}) (*resty.Response, error) {
	return c.o.client.R().SetBody(body).Patch(url)
}
