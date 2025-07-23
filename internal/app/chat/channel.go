package chat

type IChannelManager interface {
	JoinChannel(client *Client, channelID uint)
	LeaveChannel(client *Client, channelID uint)
	Broadcast(message []byte)
	GetOfflineMessages(uid uint) []string
}
