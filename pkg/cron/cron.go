package cron

import (
	"github.com/robfig/cron/v3"
	"github.com/ergoapi/zlog"
)

var Cron *Client

type Client struct {
	client *cron.Cron
}

func New() *Client {
	return &Client{client: cron.New()}
}

func (c *Client) Start() {
	zlog.Info("start cron")
	c.client.Start()
}

func (c *Client) Add(spec string, cmd func()) error {
	id, err := c.client.AddFunc(spec, cmd)
	zlog.Info("add cron: %v", id)
	return err
}

func (c *Client) Stop() {
	zlog.Info("stop cron")
	c.client.Stop()
}

func init() {
	Cron = New()
}
