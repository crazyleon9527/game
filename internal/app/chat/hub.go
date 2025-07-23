package chat

import (
	"rk-api/internal/app/utils"
	"rk-api/pkg/logger"
	"sync"
)

type channelSubscribers struct {
	clients map[uint]*Client
	mu      sync.RWMutex
}

type Subscribe struct {
	Channel string
	UID     uint
}

type Broadcast struct {
	Message []byte
	Channel string
}

type Hub struct {
	channelSubs map[string]*channelSubscribers
	clients     map[uint]*Client
	Register    chan *Client
	Unregister  chan *Client
	Join        chan *Subscribe
	Broadcast   chan *Broadcast
}

func NewHub() *Hub {
	return &Hub{
		channelSubs: make(map[string]*channelSubscribers),
		clients:     make(map[uint]*Client),
		Broadcast:   make(chan *Broadcast),
		Register:    make(chan *Client),
		Unregister:  make(chan *Client),
		Join:        make(chan *Subscribe),
	}
}

// 用户是否订阅了该频道
func (h *Hub) IsSubscribed(uid uint, channel string) bool {
	client, ok := h.clients[uid]
	if !ok {
		return false // 用户不存在
	}
	_, subscribed := client.Channels[channel]
	return subscribed // 用户是否存在于 channel 中
}

// 初始化所有频道
func (h *Hub) InitializeChannels(defaultChannels []string) {
	for _, channel := range defaultChannels {
		// 直接检查频道是否存在
		if _, exists := h.channelSubs[channel]; !exists {
			// 如果频道不存在，则创建一个新的频道
			h.channelSubs[channel] = &channelSubscribers{
				clients: make(map[uint]*Client),
			}
		}
	}
}

func (h *Hub) Subscribe(client *Client, channel string) {
	if client == nil {
		return
	}
	logger.Info("Subscribing client to channel:", channel)
	if _, ok := h.channelSubs[channel]; !ok {
		h.channelSubs[channel] = &channelSubscribers{
			clients: make(map[uint]*Client),
		}
	}
	client.Channels[channel] = struct{}{}
	h.channelSubs[channel].clients[client.UID] = client
}

func (h *Hub) Unsubscribe(client *Client, channel string) {
	if _, ok := h.channelSubs[channel]; ok {
		h.channelSubs[channel].mu.Lock()
		delete(h.channelSubs[channel].clients, client.UID)
		h.channelSubs[channel].mu.Unlock()
	}
}

func (h *Hub) Run() {
	defer utils.PrintPanicStack()
	for {
		select {
		case client := <-h.Register:
			logger.Info("Registering client:", client.UID)
			h.clients[client.UID] = client

		case client := <-h.Unregister:
			logger.Info("Unregistering client:", client.UID)
			delete(h.clients, client.UID)
			for channel := range client.Channels {
				h.Unsubscribe(client, channel)
			}
		case join := <-h.Join:
			logger.Info("Joining channel:", join.Channel)
			h.Subscribe(h.clients[join.UID], join.Channel)
		case broadcast := <-h.Broadcast:
			if _, ok := h.channelSubs[broadcast.Channel]; !ok {
				continue
			}
			// logger.Info("Broadcasting message to channel-----------", h.channelSubs["english"], broadcast.Channel)
			for _, client := range h.channelSubs[broadcast.Channel].clients {
				logger.Info("Sending message to client:", client.UID)
				select {
				case client.Send <- broadcast.Message:
				default:
					close(client.Send)
					delete(h.clients, client.UID)
				}
			}
		}
	}
}
