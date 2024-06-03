package p2p

type Message struct {
	Payload any
	From    string
}

type CommandHelp struct{}

func (cmd CommandHelp) String() string {
	return "/help"
}
