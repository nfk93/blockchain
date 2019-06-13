package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"github.com/nfk93/blockchain/transaction"
	"testing"
)

func TestState_AddContractTransaction(t *testing.T) {

	var s State
	s.Ledger = make(map[string]uint64)
	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)
	s.Ledger[pk1.Hash()] = 100
	s.Ledger[pk2.Hash()] = 100
	s.TotalStake = 100 + 100

	trans := CreateTransaction(pk1, pk2, 50, "transOne", sk1)
	s.AddTransaction(trans, 2)

	if s.Ledger[pk1.Hash()] != 150 && s.Ledger[pk1.Hash()] != 48 {
		t.Error("not correct amount!")
	}
	if s.TotalStake != 100+100-2 {
		t.Error("TotalStake is not correct...")
	}

}

func TestState_AddBlockReward(t *testing.T) {

	var s State
	s.Ledger = make(map[string]uint64)
	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)
	s.Ledger[pk1.Hash()] = 100
	s.Ledger[pk2.Hash()] = 100
	s.TotalStake = 100 + 100

	trans := CreateTransaction(pk1, pk2, 50, "transOne", sk1)
	s.AddTransaction(trans, 2)
	s.AddAmountToAccount(pk1, 2)

	if s.Ledger[pk1.Hash()] != 150 && s.Ledger[pk1.Hash()] != 50 {
		t.Error("not correct amount!")
	}
	if s.TotalStake != 100+100-2+2 {
		t.Error("TotalStake is not correct...")
	}

}

func TestState_FundContractCall(t *testing.T) {
	var s State
	s.Ledger = make(map[string]uint64)
	_, pk1 := KeyGen(2048)
	s.Ledger[pk1.Hash()] = 100
	s.TotalStake = 100

	s.ConOwners = make(map[string]PublicKey)
	s.ConOwners["address22"] = pk1

	success := s.FundContractCall(pk1, 50, 20)

	if !success {
		t.Error("Fund account didn't succeed!")
	}
	if s.Ledger[pk1.Hash()] != 100-50-20 {
		t.Error("Owners account is not correct...")
	}
	if s.TotalStake != 100-20 {
		t.Error("TotalStake is not correct...")
	}
}

func initState() {
	channels := CreateChannelStruct()
	go transaction.StartTransactionLayer(channels, false)
}

func BenchmarkState_AddTransactionTransactions(b *testing.B) {
	sk, pk := KeyGen(2048)
	_, pk1 := KeyGen(2048)
	transaction := CreateTransaction(pk, pk1, 100000, "", sk)
	state := NewInitialState(pk)
	b.ResetTimer()
	state.AddTransaction(transaction, uint64(200000))
}

//func TestState_CleanContractLedger(t *testing.T) {
//	var s State
//	_, pk := KeyGen(2048)
//	s.Ledger = make(map[string]uint64)
//	s.Ledger[pk.Hash()] = 100
//	s.ConStake = make(map[string]uint64)
//	s.ConStake["address1"] = 100
//	s.ConStake["address2"] = 100
//	s.ConStake["address3"] = 100
//	s.ConStake["address4"] = 100
//	s.TotalStake = 100 + 4*100
//
//	s.ConOwners = make(map[string]PublicKey)
//	s.ConOwners["address1"] = pk
//	s.ConOwners["address2"] = pk
//	s.ConOwners["address3"] = pk
//	s.ConOwners["address4"] = pk
//
//	expiredContracts := s.CleanExpiredContract()
//
//	if _, success := s.ConOwners["address1"]; !success {
//		t.Error("address1 has been deleted...")
//	}
//	if _, success := s.ConOwners["address2"]; !success {
//		t.Error("address2 has been deleted...")
//	}
//	if len(expiredContracts) != 2 {
//		t.Error("Not all expired contracts has been deleted...")
//	}
//	if s.Ledger[pk.Hash()] != 300 {
//		t.Error("Not correct amount in ledger for pk1")
//	}
//	if s.TotalStake != 100+4*100 {
//		t.Error("TotalStake is not correct...")
//	}
//
//}

//func TestState_CollectStorageCost(t *testing.T) {
//	var s State
//	_, pk := KeyGen(2048)
//
//	s.ConOwners = make(map[string]ContractAccount)
//	s.ConOwners["address1"] = ContractAccount{pk, 200, 15}
//	s.ConOwners["address2"] = ContractAccount{pk, 1, 5}
//	s.ConOwners["address3"] = ContractAccount{pk, 0, 5}
//	s.ConOwners["address4"] = ContractAccount{pk, 20, 5}
//
//	if s.CollectStorageCost(3) != 45+1+0+15 {
//		t.Error("Wrong amount collected")
//	}
//
//	if s.ConOwners["address1"].Prepaid != 200-45 {
//		t.Error("Wrong amount in prepaid for 1")
//	}
//	if s.ConOwners["address2"].Prepaid != 1-1 {
//		t.Error("Wrong amount in prepaid for 2")
//	}
//	if s.ConOwners["address3"].Prepaid != 0-0 {
//		t.Error("Wrong amount in prepaid for 3")
//	}
//	if s.ConOwners["address4"].Prepaid != 20-15 {
//		t.Error("Wrong amount in prepaid for 4")
//	}
//}

//func TestHandleContractCall(t *testing.T) {
//
//	// TODO: make an actual test now
//
//	var s State
//	_, pk := KeyGen(2048)
//	con1 := "contract1"
//
//	s.ConOwners = map[string]PublicKey{}
//	s.ConOwners[con1] = pk
//	s.ConStake = map[string]uint64{}
//	s.ConStake[con1] = 200
//	s.Ledger = map[string]uint64{}
//	s.Ledger[pk.Hash()] = 500
//	s.TotalStake = 200 + 500
//
//	contract := ContractCall{}
//	contract.Caller = pk
//	contract.Gas = 13
//	contract.Amount = 150
//
//	_, err := s.handleContractCall(contract)
//
//	if err != nil {
//		t.Error("ContractCall failed...")
//	}
//	if s.Ledger[pk.Hash()] != 337 {
//		t.Error("Not correct amount in callers account...")
//	}
//	if s.TotalStake != 200+500-13 {
//		t.Error("TotalStake is not correct...")
//	}
//
//}
