package chat

import (
	"log"
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"time"

	"github.com/gorilla/websocket"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

type IMessageProcessor interface {
	ProcessMessage(message []byte)
}

type Client struct {
	UID       uint
	Conn      *websocket.Conn
	Channels  map[string]struct{}
	Send      chan []byte
	Quit      chan struct{}
	Processor IMessageProcessor
	Hub       *Hub // 添加对 Hub 的引用
}

func (c *Client) Start() {
	go c.writePump()
	go c.readPump()
}

// 添加清理方法
func (c *Client) Cleanup() {
	c.Hub.Unregister <- c
	close(c.Send)
	c.Conn.Close()
}

func (c *Client) readPump() {
	defer utils.PrintPanicStack()
	defer func() {
		c.Cleanup() // 在连接断开时调用清理
	}()

	// 设置读取超时
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}
		c.Processor.ProcessMessage(message)
	}
}

func (c *Client) writePump() {
	defer utils.PrintPanicStack()
	ticker := time.NewTicker(pingPeriod) // 心跳间隔（建议55秒）
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// 设置写超时（建议10秒）
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))

			if !ok {
				// 通道关闭时发送关闭消息
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			logger.Infof("client  send message: %s", string(message))

			// 创建文本写入器
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}

			// 批量写入（处理积压消息）
			n := len(c.Send)
			logger.Infof("client  send message count: %d", n)
			for i := 0; i < n; i++ {
				w, err := c.Conn.NextWriter(websocket.TextMessage)
				if err != nil {
					return
				}
				w.Write(<-c.Send)

				if err := w.Close(); err != nil {
					return
				}
			}

		case <-ticker.C:
			// 发送心跳包
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// // 添加消息队列以处理高并发
// type MessageQueue struct {
//     messages chan *entities.ChatMessage
//     workers  int
// }

// func NewMessageQueue(workers int) *MessageQueue {
//     mq := &MessageQueue{
//         messages: make(chan *entities.ChatMessage, 1000),
//         workers:  workers,
//     }
//     mq.Start()
//     return mq
// }

// func (mq *MessageQueue) Start() {
//     for i := 0; i < mq.workers; i++ {
//         go mq.worker()
//     }
// }

// func (mq *MessageQueue) worker() {
//     for msg := range mq.messages {
//         // 处理消息
//         // 存储到数据库
//         // 推送到相关客户端
//     }
// }
