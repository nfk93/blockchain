package consensus

import (
	"github.com/nfk93/blockchain/objects/genesisdata"
	"testing"
)
import . "github.com/nfk93/blockchain/objects"

var p2pOutTrans chan Transaction
var p2pOutBlock chan Block
var p2pInBlock chan Block
var genesis genesisdata.GenesisData

func resetMocksAndStart() {
	p2pOutTrans = make(chan Transaction)
	p2pOutBlock = make(chan Block)
	p2pInBlock = make(chan Block)
	genesis = genesisdata.GenesisData{}
	StartConsensus(genesis, p2pOutTrans, p2pOutBlock, p2pInBlock)
}

func TestSmokeTest(t *testing.T) {
	resetMocksAndStart()
}
