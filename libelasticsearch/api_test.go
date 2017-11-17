package libelasticsearch

import (
	"testing"
)

type Tweet struct {
	User    string `json:"user"`
	Message string `json:"message"`
}

const mapping = `
{
	"settings":{
		"number_of_shards": 1,
		"number_of_replicas": 0
	},
	"mappings":{
		"tweet2":{
			"properties":{
				"user":{
					"type":"keyword"
				},
				"message":{
					"type":"text",
					"store": true,
					"fielddata": true
				}
			}
		}
	}
}`

func TestElastic(t *testing.T) {
	tw := &Tweet{}
	tw.User = "1234"
	tw.Message = "8984"

	option := &Option{}
	option.URL = "http://127.0.0.1:9200"
	option.Username = "elastic"
	option.Password = "dengshudan"
	c := New(option)

	if err := c.CreateIndexWithoutMapping("tweet880"); err != nil {
		panic(err)
	}

	if err := c.PutValue("tweet880", "tweet880", "", tw); err != nil {
		panic(err)
	}
	c.GetByType("tweet880", "tweet880")
}
