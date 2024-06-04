package p2p

type Message struct {
	Payload any
	From    string
}

type CommandHelp struct{}

func (cmd CommandHelp) String() string {
	return "/help"
}

type CommandCreateRoom struct{}

func (cmd CommandCreateRoom) String() string {
	return "/create-room"
}

type CommandListRoom struct{}

func (cmd CommandListRoom) String() string {
	return "/list-room"
}

type CommandJoinRoom struct{}

func (cmd CommandJoinRoom) String() string {
	return "/join-room"
}

type CommandCurrentRoom struct{}

func (cmd CommandCurrentRoom) String() string {
	return "/cr"
}
