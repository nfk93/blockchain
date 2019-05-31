package objects

import (
	"bytes"
	. "github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/smart"
	"sort"
	"strconv"
)

type State struct {
	Ledger     map[string]uint64
	ConStake   map[string]uint64
	ConOwners  map[string]PublicKey
	ParentHash string
	TotalStake uint64
}

func NewInitialState(key PublicKey) State {
	initialStake := uint64(100000000000000) // 1 mil
	ledger := make(map[string]uint64)
	conStake := make(map[string]uint64)
	conledger := make(map[string]PublicKey)
	ledger[key.Hash()] = initialStake
	return State{ledger, conStake, conledger, "", initialStake}
}

//Returns gasCost
func (s *State) AddTransaction(t Transaction, gasCost uint64) uint64 {
	//TODO: Handle checks of legal transactions
	amountWithFees := t.Amount + gasCost

	if !t.VerifyTransaction() {
		// fmt.Println("The transactions didn't verify", t)
		return 0
	}

	// Sender has to be able to pay both the amount and the fee
	if s.Ledger[t.From.Hash()] < amountWithFees {
		// fmt.Println("Not enough money on senders account")
		return 0
	}

	s.Ledger[t.From.Hash()] -= amountWithFees
	s.Ledger[t.To.Hash()] += t.Amount
	s.TotalStake -= gasCost // Take the fee out of the system
	return gasCost
}

func (s *State) PayBlockRewardOrRemainGas(pk PublicKey, reward uint64) {
	s.Ledger[pk.Hash()] += reward
	s.TotalStake += reward // putting back the fees and an block reward if anyone claim it
}

func (s State) toString() string {
	var buf bytes.Buffer

	sortedLedger := make(map[string]uint64)

	keys := make([]string, 0, len(s.Ledger))
	for k := range s.Ledger {
		sortedLedger[k] = s.Ledger[k]
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, k := range keys {
		buf.WriteString(k)
		buf.WriteString(strconv.Itoa(int(sortedLedger[k])))
	}
	buf.WriteString(s.ParentHash)
	buf.WriteString(strconv.Itoa(int(s.TotalStake)))

	return buf.String()
}

func (s *State) SignHashedState(sk SecretKey) string {
	return Sign(HashSHA(s.toString()), sk)
}

func (s State) VerifyHashedState(sig string, pk PublicKey) bool {
	return Verify(HashSHA(s.toString()), sig, pk)
}

// Opens account for contract and moves prepaid to its account
func (s *State) InitializeContractAccount(addr string, owner PublicKey) {
	s.ConOwners[addr] = owner
}

// Used for handling contract layer transaction to users
func (s *State) AddContractTransaction(t smart.ContractTransaction) {
	s.Ledger[t.To] += t.Amount
}

// Returns true if caller has enough funds on account to pay for call
func (s *State) FundContractCall(callerAccount PublicKey, amount uint64, gas uint64) bool {
	if s.Ledger[callerAccount.Hash()] >= amount+gas {
		s.TotalStake -= gas
		s.Ledger[callerAccount.Hash()] -= amount + gas
		return true
	}
	return false
}

//Used to refund money from contracts back into the original user ledger
func (s *State) returnAmountFromContracts(callerAccount PublicKey, amount uint64) {
	s.Ledger[callerAccount.Hash()] += amount
}

func (s *State) HandleContractInit(contractInit ContractInitialize, blockhash string, parenthash string, slot uint64) uint64 {
	if s.Ledger[contractInit.Owner.Hash()] > contractInit.Prepaid+contractInit.Gas {
		var addr string
		var remainGas uint64
		var err error

		if blockhash == "" {
			addr, remainGas, err = smart.InitiateContractOnNewBlock(contractInit.Code, contractInit.Gas,
				contractInit.Prepaid, contractInit.StorageLimit)
		} else {
			addr, remainGas, err = smart.InitiateContract(contractInit.Code, contractInit.Gas, contractInit.Prepaid,
				contractInit.StorageLimit, blockhash)
		}

		s.PayBlockRewardOrRemainGas(contractInit.Owner, remainGas)
		if err != nil {
			return contractInit.Gas - remainGas
		} else {
			s.TotalStake -= contractInit.Prepaid
			s.InitializeContractAccount(addr, contractInit.Owner)
			return contractInit.Gas - remainGas
		}
	} else {
		return 0
	}
}

func (s *State) HandleContractCall(contract ContractCall, blockhash string, parenthash string, slot uint64) uint64 {
	// Transfer funds from caller to contract
	if !s.FundContractCall(contract.Caller, contract.Amount, contract.Gas) {
		return 0
	}
	var newContractLedger map[string]uint64
	var transferList []smart.ContractTransaction
	var remainingGas uint64
	var callerr error

	// Run contracts in smart contract layer
	if blockhash == "" {
		newContractLedger, transferList, remainingGas, callerr = smart.CallContractOnNewBlock(contract.Address,
			contract.Entry, contract.Params, contract.Amount, contract.Gas)
	} else {
		newContractLedger, transferList, remainingGas, callerr = smart.CallContract(contract.Address,
			contract.Entry, contract.Params, contract.Amount, contract.Gas, blockhash)
	}

	// Calc how much gas used and refund not used gas to caller
	gasUsed := contract.Gas - remainingGas
	s.PayBlockRewardOrRemainGas(contract.Caller, remainingGas)

	// If contract not successful, then return amount to caller
	if callerr != nil {
		s.returnAmountFromContracts(contract.Caller, contract.Amount)
		return gasUsed
	}
	// If contract succeeded, execute the transactions from the contract layer
	s.ConStake = newContractLedger
	for _, t := range transferList {
		s.AddContractTransaction(t)
	}
	return gasUsed
}

// Get list of contract addresses that expire from the smart contract layer
// pay contract stake back to owner and delete account
func (s *State) CleanExpiredContract(expiring []string) {
	for _, conAddr := range expiring {
		owner := s.ConOwners[conAddr]
		s.Ledger[owner.Hash()] += s.ConStake[conAddr]
		delete(s.ConStake, conAddr)
		// TODO Is this needed? Does anyway get a new state from scl...
		// not sure what you mean, but I think it is very important, because of the resulting change in state -Niko
	}

}
