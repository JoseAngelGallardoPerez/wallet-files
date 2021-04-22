package http

import (
	"math"
	"strconv"

	"github.com/gin-gonic/gin"
)

type Error struct {
	Title   string  `json:"title"`
	Details string  `json:"details"`
	Status  int     `json:"status"`
	Code    *string `json:"code"`
}

type List struct {
	HasMore bool        `json:"has_more"`
	Items   interface{} `json:"items"`
}

type Response struct {
	Links    interface{} `json:"links,omitempty"`
	Data     interface{} `json:"data,omitempty"`
	Messages []string    `json:"messages,omitempty"`
	Errors   []*Error    `json:"errors,omitempty"`
}

type Links struct {
	Self  string  `json:"self"`
	Next  *string `json:"next"`
	Prev  *string `json:"prev"`
	First *string `json:"first"`
	Last  *string `json:"last"`
}

func NewResponse() *Response {
	return new(Response)
}

func NewResponseList(items interface{}) (*List, error) {
	result := &List{}
	result.Items = items

	return result, nil
}

func NewResponseWithList(items interface{}) (*Response, error) {
	list, err := NewResponseList(items)
	if nil != err {
		return nil, err
	}

	res := NewResponse()
	res.SetData(list.Items)

	return res, nil
}

func NewResponseWithListAndLinks(items interface{}, c *gin.Context, total int64) (*Response, error) {
	list, err := NewResponseList(items)
	if nil != err {
		return nil, err
	}

	res := NewResponse()
	res.SetData(list.Items)
	res.SetLinks(res.buildLinks(c, total))

	return res, nil
}

func NewResponseWithError(
	title string,
	details string,
	status int,
	code *string,
) *Response {
	return NewResponse().AddError(title, details, status, code)
}

func (r *Response) AddError(
	title string,
	details string,
	status int,
	code *string,
) *Response {
	e := &Error{
		Title:   title,
		Details: details,
		Status:  status,
		Code:    code,
	}
	r.Errors = append(r.Errors, e)
	return r
}

func (r *Response) AddMessage(message string) *Response {
	r.Messages = append(r.Messages, message)
	return r
}

func (r *Response) SetData(data interface{}) *Response {
	r.Data = data
	return r
}

func (r *Response) SetLinks(links interface{}) *Response {
	r.Links = links
	return r
}

func (r *Response) buildLinks(c *gin.Context, total int64) Links {
	limit, err := strconv.ParseInt(c.Request.URL.Query().Get("limit"), 10, 32)
	if nil != err {
		limit = 15
	}

	links := Links{
		Self:  c.Request.URL.String(),
		Next:  r.getNextUrl(c, total, limit),
		Prev:  r.getPrevUrl(c, total, limit),
		First: r.getFirstUrl(c, total),
		Last:  r.getLastUrl(c, total, limit),
	}

	return links
}

func (r *Response) getNextUrl(c *gin.Context, total int64, limit int64) *string {
	offset, _ := strconv.ParseInt(c.Request.URL.Query().Get("offset"), 10, 32)

	if offset > total || offset+limit > total {
		return nil
	}

	offsetNext := limit + offset

	url := *c.Request.URL

	values := url.Query()
	values.Set("offset", strconv.Itoa(int(offsetNext)))
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}

func (r *Response) getPrevUrl(c *gin.Context, total int64, limit int64) *string {
	offset, _ := strconv.ParseInt(c.Request.URL.Query().Get("offset"), 10, 32)

	if offset == 0 {
		return nil
	}

	url := *c.Request.URL

	offsetNext := offset - limit
	values := url.Query()
	values.Set("offset", strconv.Itoa(int(offsetNext)))
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}

func (r *Response) getFirstUrl(c *gin.Context, total int64) *string {
	if total == 0 {
		return nil
	}

	url := *c.Request.URL

	values := url.Query()
	values.Set("offset", "0")
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}

func (r *Response) getLastUrl(c *gin.Context, total int64, limit int64) *string {
	if total == 0 {
		return nil
	}

	pages := int64(math.Round(float64(total / limit)))
	lastOffset := limit * pages

	url := *c.Request.URL

	values := url.Query()
	values.Set("offset", strconv.Itoa(int(lastOffset)))
	url.RawQuery = values.Encode()
	urlString := url.String()
	return &urlString
}
