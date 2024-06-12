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
	return "/cr"
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
	return "/current-room"
}

type CommandStartGame struct{}

func (cmd CommandStartGame) String() string {
	return "/start"
}

type CommandCurrentGame struct{}

func (cmd CommandCurrentGame) String() string {
	return "/current-game"
}

type CommandStay struct{}

func (cmd CommandStay) String() string {
	return "/stay"
}

type CommandExtraCard struct{}

func (cmd CommandExtraCard) String() string {
	return "/more"
}

type CommandRemoveRoom struct{}

func (cmd CommandRemoveRoom) String() string {
	return "/close-room"
}

type CommandQuitRoom struct{}

func (cmd CommandQuitRoom) String() string {
	return "/quit"
}
