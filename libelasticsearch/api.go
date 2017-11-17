package libelasticsearch

import (
	"context"
	"encoding/json"

	"github.com/olivere/elastic"
	"github.com/olivere/elastic/config"
)

type Client struct {
	ctx context.Context
	*elastic.Client
}

type Option struct {
	URL      string
	Index    string
	Username string
	Password string
}

// 新建es的client
func New(option *Option) *Client {

	config := &config.Config{}
	config.URL = option.URL
	config.Username = option.Username
	config.Password = option.Password
	config.Index = option.Index

	client, err := elastic.NewClientFromConfig(config)

	if err != nil {
		panic(err)
	}

	c := &Client{}
	c.Client = client
	c.ctx = context.Background()

	info, code, err := client.Ping(config.URL).Do(c.ctx)
	if err != nil {
		panic(err)
	}

	return c
}

func (c *Client) CreateIndexWithoutMapping(index string) error {

	exist, err := c.IndexExists(index).Do(c.ctx)
	if err != nil {
		return err
	}

	if exist {
		return nil
	}

	createIndex, err := c.Client.CreateIndex(index).Do(c.ctx)
	if createIndex.Acknowledged {

	}
	return err
}

func (c *Client) PutValue(index, logtype, id string, val interface{}) error {
	b, err := json.Marshal(val)
	if err != nil {
		return err
	}

	if len(id) == 0 {
		_, err = c.Index().Index(index).Type(logtype).BodyString(string(b)).Do(c.ctx)
	} else {
		_, err = c.Index().Index(index).Type(logtype).Id("1").BodyString(string(b)).Do(c.ctx)
	}

	return err
}

func (c *Client) GetByType(index, logtype string) ([]*SearchHit, error) {

	result, err := c.Client.Search().Index(index).Type(logtype).Do(c.ctx)
	if err != nil {

		return nil, err
	}

	b, _ := json.Marshal(result)

	return result.Hits.Hits, nil
}
