package channels

type RedirectionChannel struct {
	ChunkChannel  chan []byte
	ErrorChannel  chan error
	ReturnChannel chan error
}

func (channel *RedirectionChannel) closeAll() {
	close(channel.ChunkChannel)
	close(channel.ErrorChannel)
	close(channel.ReturnChannel)
}

func NewRedirectionChannel(chunkChannelSize int) RedirectionChannel {
	return RedirectionChannel{
		make(chan []byte, chunkChannelSize),
		make(chan error, 1),
		make(chan error, 1),
	}
}
