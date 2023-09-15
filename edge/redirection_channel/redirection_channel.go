package redirection_channel

type Message struct {
	Body []byte
	Err  error
}

type RedirectionChannel struct {
	MessageChannel chan Message
	ReturnChannel  chan error
}

func NewRedirectionChannel(messageChannelSize int) RedirectionChannel {
	return RedirectionChannel{
		make(chan Message, messageChannelSize),
		make(chan error, 1),
	}
}
