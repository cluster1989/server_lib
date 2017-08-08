package zone_client

import (
	"github.com/wqf/common_lib/codec"
	"github.com/wqf/common_lib/libnet"
	"github.com/wqf/zone_server/zone_route"
)

type Client struct {
	Session *libnet.Session
}

func New(session *libnet.Session) *Client {
	c := new(Client)
	c.Session = session
	c.Session.AddCloseCallback(c.Session.ID(), c.closeCallback)
	return c
}

func (c *Client) ClientLoop() {
	for {
		//接收数据
		reqData, err := c.Session.Recv()
		if err != nil {
			//记录错误
			continue
		}
		if reqData != nil {
			if err = c.parse(reqData); err != nil {
				//记录错误
				continue
			}
		}
	}
}

func (c *Client) ID() uint64 {
	return c.Session.ID()
}

func (c *Client) closeCallback() {
	//关闭回调

}

func (c *Client) Exec(data interface{}) error {
	err := c.Session.Send(data)
	return err
}

func (c *Client) Close() {
	Mananger.DelClient(c)
	c.Session.Close()
}

func (c *Client) parse(reqData []byte) (err error) {
	//先取协议号
	cmdByte := reqData[0:2]
	if cmdByte == nil {
		//记录错误,直接关闭，没有标识
		c.Close()
		return
	}
	cmdInt16 := codec.GetUint16BE(cmdByte)
	cmdObj := reqData[2:]

	if !c.checkRoute(cmdInt16) {
		c.Close()
		//记录消息,直接关闭
		return
	}
	c.route(cmdInt16, cmdObj)
	return
}

func (c *Client) checkRoute(cmdId uint16) bool {

	if _, ok := zone_route.Routes[cmdId]; !ok { //如果时注册排除
		return false
	}

	// if !s.cliManager.Existed(cliID) {
	// 	return false
	// }
	return true
}

func (c *Client) checkClient() bool {
	return Mananger.Existed(c.ID())
}

func (c *Client) exceptionCMDIDS(cmdId uint16) bool {
	for _, oCmdId := range zone_route.ExceptionRouteIds {
		if oCmdId == cmdId {
			return true
		}
	}
	return false
}

func (c *Client) route(cmdId uint16, cmdObj []byte) {
	defer func() {
		//不能在这里出错
		if r := recover(); r != nil {
			//如果出错，向客户端发送错误码----
		}
	}()

	if !c.exceptionCMDIDS(cmdId) && !c.checkClient() {
		c.Close()
		//记录消息,直接关闭
		return
	}

	f, _ := zone_route.Routes[cmdId]
	ret := f(cmdObj)
	c.Exec(ret)
}
