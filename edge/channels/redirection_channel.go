package channels

type RedirectionChannel struct {
	chunkChannel  chan []byte
	errorChannel  chan error
	returnChannel chan error
}

func (channel *RedirectionChannel) closeAll() {
	close(channel.chunkChannel)
	close(channel.errorChannel)
	close(channel.returnChannel)
}
