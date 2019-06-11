package consensus

import (
	"bytes"
	"fmt"
	. "github.com/nfk93/blockchain/crypto"
	o "github.com/nfk93/blockchain/objects"
	"github.com/nfk93/blockchain/transaction"
	"log"
	"sort"
	"strconv"
	"sync"
	"time"
)

var slotLength time.Duration
var currentSlot uint64
var slotLock sync.RWMutex
var hardness float64
var genesisTime time.Time
var finalizeInterval uint64
var sk SecretKey
var pk PublicKey

type FinalData struct {
	stake           map[string]uint64
	totalstake      uint64
	leadershipNonce string
	blockHash       string
}

var finalData = make(map[uint64]FinalData)
var finalLock sync.RWMutex

func runSlot() { //Calls drawLottery every slot and increments the currentSlot after slotLength time.
	currentSlot = 1
	offset := time.Since(genesisTime)
	for {
		if (currentSlot)%finalizeInterval == 0 {
			finalize(currentSlot - (finalizeInterval))
		}
		drawLottery(currentSlot)
		timeSinceGenesis := time.Since(genesisTime) - offset
		if saveGraphFiles {
			go func() {
				blocks.rlock()
				defer blocks.runlock()
				copy_ := make(map[string]o.Block)
				for k, v := range blocks.m {
					copy_[k] = v
				}
				err := printBlockTreeGraphToFile(fmt.Sprintf("slot%d", currentSlot), copy_)
				if err != nil {
					log.Println(fmt.Sprintf("error saving tree: %s", err.Error()))
				}
			}()
		}
		func() {
			checkPendingBlocks()
		}()
		sleepyTime := time.Duration(currentSlot)*slotLength - timeSinceGenesis
		if sleepyTime > 0 {
			time.Sleep(sleepyTime)
		}
		currentSlot += 1
	}
}

func getCurrentSlot() uint64 {
	slotLock.RLock()
	defer slotLock.RUnlock()
	return currentSlot
}

func getEpoch(slot uint64) uint64 {
	return slot / finalizeInterval
}

func processGenesisData(genesisData o.GenesisData, blockHash string) {
	// TODO  -  Use GenesisTime when going away from two-phase implementation
	hardness = genesisData.Hardness
	slotLength = genesisData.SlotDuration
	genesisFinalData := FinalData{stake: genesisData.InitialState.Ledger, totalstake: genesisData.InitialState.TotalStake,
		leadershipNonce: genesisData.Nonce, blockHash: blockHash}
	finalData[0] = genesisFinalData
	finalizeInterval = genesisData.FinalizeInterval
	genesisTime = genesisData.GenesisTime
	go runSlot()
	go transaction.StartTransactionLayer(channels, saveGraphFiles)
}

func finalize(slot uint64) {
	if isVerbose {
		log.Println("Finalizing at slot", slot)
	}
	if slot == 0 {
		// do nothing
	} else {
		func() {
			epoch := getEpoch(slot)
			blocks.rlock()
			defer blocks.runlock()
			finalLock.Lock()
			defer finalLock.Unlock()
			head := blocks.get(getCurrentHead())
			for {
				parent := blocks.get(head.ParentPointer)
				if parent.Slot < slot {
					finalHash := head.LastFinalized
					if isVerbose {
						log.Println("Finalizing block", finalHash[:6]+"...")
					}
					newNonce := newLeadershipNonce(head)
					channels.FinalizeToTrans <- finalHash
					state := <-channels.StateFromTrans
					finData := getFinalData(state, newNonce, head.CalculateBlockHash())
					finalData[epoch] = finData
					break
				}
				head = parent
			}
		}()
	}
	if isVerbose {
		log.Println(fmt.Sprintf("Finalized slot %d successfully", slot))
		PrintCurrentStake()
	}
}

func getFinalData(state o.State, leadershipNonce, blockHash string) FinalData {
	m := state.Ledger
	// add contract accounts to owners mining pool
	for k, v := range state.ConStake {
		conowner := state.ConOwners[k]
		conownerhash := conowner.Hash()
		m[conownerhash] += v
	}
	return FinalData{stake: m, totalstake: state.TotalStake, leadershipNonce: leadershipNonce, blockHash: blockHash}
}

func newLeadershipNonce(finalBlock o.Block) string {
	var buf bytes.Buffer
	previousFinal := finalBlock.LastFinalized
	head := finalBlock
	for {
		buf.WriteString(head.BlockNonce.Nonce)
		head = blocks.get(head.ParentPointer)
		if head.ParentPointer == previousFinal {
			break
		}
	}
	return HashSHA(buf.String())
}

func drawLottery(slot uint64) {
	finalLock.RLock()
	defer finalLock.RUnlock()
	blocks.rlock()
	defer blocks.runlock()
	fd := finalData[getFinalDataIndex(slot)]
	leadershipNonce := fd.leadershipNonce
	lastfinalized := fd.blockHash
	winner, draw := CalculateDraw(hardness, sk, pk, slot, fd)
	if winner {
		if isVerbose {
			log.Println("We won slot " + strconv.Itoa(int(slot)))
		}
		generateBlock(draw, slot, leadershipNonce, lastfinalized, getCurrentHead())
	}
}

//Sends all unused transactions to the transaction layer for the transaction layer to process for the new block
func generateBlock(draw string, slot uint64, leadershipNonce, lastfinalized, parentHash string) {
	blockData := o.CreateBlockData{
		getUnusedTransactions(),
		sk,
		pk,
		slot,
		draw,
		o.CreateNewBlockNonce(leadershipNonce, sk, slot),
		lastfinalized,
		parentHash}
	channels.TransToTrans <- blockData
	block := <-channels.BlockFromTrans
	go func() {
		channels.BlockToP2P <- block
	}()
}

func getLotteryPower(pk PublicKey, slot uint64) float64 {
	finalLock.RLock()
	defer finalLock.RUnlock()
	fd, exists := finalData[getEpoch(slot)]
	if !exists {
		log.Println("ERROR: attempted getting lottery power from unfinalized epoch")
		return 0
	} else {
		return float64(fd.stake[pk.Hash()]) / float64(fd.totalstake)
	}
}

func PrintCurrentStake() {
	slot := getCurrentSlot()
	lastFinalEpoch := getEpoch(slot) - 1
	var keyList []string
	finalLock.RLock()
	defer finalLock.RUnlock()
	stake := finalData[lastFinalEpoch].stake
	for k := range stake {
		keyList = append(keyList, k)
	}
	sort.Strings(keyList)

	for _, k := range keyList {
		log.Printf("Keyhash: %v, Stake: %v\n", k[:10], stake[k])
	}
}
