package libhbase

import (
	"context"

	"github.com/tsuna/gohbase"
	"github.com/tsuna/gohbase/filter"
	"github.com/tsuna/gohbase/hrpc"
)

type Client struct {
	client gohbase.Client
}

type HBaseMulitSetModel struct {
	Qualifier string
	Value     []byte
}

func New(address string) *Client {
	client := gohbase.NewClient(address)
	c := &Client{}
	c.client = client
	return c
}

func (c *Client) PutSingle(table, row, family, qualifier string, value []byte) error {
	out := make(map[string]map[string][]byte)
	out[family] = map[string][]byte{qualifier: []byte(value)}
	return c.Put(table, row, out)
}

func (c *Client) PutMulti(table, row, family string, values []*HBaseMulitSetModel) error {
	if len(values) == 0 {
		return nil
	}
	out := make(map[string]map[string][]byte)
	vals := make(map[string][]byte)
	for _, v := range values {
		vals[v.Qualifier] = v.Value
	}
	out[family] = vals
	return c.Put(table, row, out)
}

func (c *Client) Put(table, row string, values map[string]map[string][]byte) error {
	putRequest, err := hrpc.NewPutStr(context.Background(), table, row, values)
	if err != nil {
		return err
	}
	if _, err := c.client.Put(putRequest); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetRow(table, row string) ([][]byte, error) {
	return c.Get(table, row)
}

func (c *Client) GetWithQualifier(table, row, family string, qualifier []string) ([][]byte, error) {
	query := map[string][]string{family: qualifier}
	return c.Get(table, row, hrpc.Families(query))
}

func (c *Client) Get(table, row string, options ...func(hrpc.Call) error) ([][]byte, error) {
	getRequest, err := hrpc.NewGetStr(context.Background(), table, row, options...)
	getRsp, err := c.client.Get(getRequest)
	if err != nil {
		return nil, err
	}

	values := make([][]byte, 0)
	for _, v := range getRsp.Cells {
		values = append(values, v.Value)
	}
	return values, nil
}

func (c *Client) GetWithQualifierAndFilter(table, row, family string, qualifier []string, filter filter.Filter) ([][]byte, error) {

	query := map[string][]string{family: qualifier}
	return c.Get(table, row, hrpc.Families(query), hrpc.Filters(filter))
}

func (c *Client) Delete(table, row string,
	values map[string]map[string][]byte, options ...func(hrpc.Call) error) error {
	mutate, err := hrpc.NewDelStr(context.Background(), table, row, values, options...)
	if err != nil {
		return err
	}
	if _, err := c.client.Delete(mutate); err != nil {
		return err
	}
	return nil
}

func (c *Client) DeleteFamily(table, row, family string, values []*HBaseMulitSetModel) error {

	out := make(map[string]map[string][]byte)
	vals := make(map[string][]byte)
	for _, v := range values {
		vals[v.Qualifier] = v.Value
	}
	out[family] = vals

	return c.Delete(table, row, out)
}
