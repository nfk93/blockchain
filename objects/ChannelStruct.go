package objects

type ChannelStruct struct {
	TransClientInput chan Transaction
	TransFromP2P     chan TransData
	BlockFromP2P     chan Block
	BlockToP2P       chan Block
	BlockToTrans     chan Block
	StateFromTrans   chan State
	FinalizeToTrans  chan string
	BlockFromTrans   chan Block
	TransToTrans     chan CreateBlockData
}

func CreateChannelStruct() ChannelStruct {
	tci := make(chan Transaction)
	blockChannel1 := make(chan Block)
	blockChannel2 := make(chan Block)
	blockChannel3 := make(chan Block)
	blockChannel4 := make(chan Block)
	transChannel := make(chan TransData)
	blockDataChannel := make(chan CreateBlockData)
	stringChannel := make(chan string)
	stateChannel := make(chan State)
	return ChannelStruct{tci, transChannel, blockChannel1,
		blockChannel2, blockChannel3, stateChannel,
		stringChannel, blockChannel4, blockDataChannel}
}
