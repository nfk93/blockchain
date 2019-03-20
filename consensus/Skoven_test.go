package consensus

import (
	. "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/objects/genesisdata"
	"testing"
)

var transFromP2P chan Transaction
var blockFromP2P chan Block
var blockToP2P chan Block
var genesis genesisdata.GenesisData

func resetMocksAndStart() {
	transFromP2P = make(chan Transaction)
	blockFromP2P = make(chan Block)
	blockToP2P = make(chan Block)
	genesis = genesisdata.GenesisData{}
	StartConsensus(genesis, transFromP2P, blockFromP2P, blockToP2P)
}

func TestSmokeTest(t *testing.T) {
	resetMocksAndStart()
}
