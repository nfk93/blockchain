package objects

import (
	. "github.com/nfk93/blockchain/crypto"
	"testing"
)

func TestState_AddContractTransaction(t *testing.T) {

	var s State
	s.Ledger = make(map[string]int)
	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)
	s.Ledger[pk1.String()] = 100
	s.Ledger[pk2.String()] = 100

	trans := CreateTransaction(pk1, pk2, 50, "transOne", sk1)
	s.AddTransaction(trans, 2)

	if s.Ledger[pk1.String()] != 150 && s.Ledger[pk1.String()] != 48 {
		t.Error("not correct amount!")
	}

}

func TestState_AddBlockReward(t *testing.T) {

	var s State
	s.Ledger = make(map[string]int)
	sk1, pk1 := KeyGen(2048)
	_, pk2 := KeyGen(2048)
	s.Ledger[pk1.String()] = 100
	s.Ledger[pk2.String()] = 100

	trans := CreateTransaction(pk1, pk2, 50, "transOne", sk1)
	s.AddTransaction(trans, 2)
	s.AddBlockReward(pk1, 2)

	if s.Ledger[pk1.String()] != 150 && s.Ledger[pk1.String()] != 50 {
		t.Error("not correct amount!")
	}

}

func TestState_PrepayContracts(t *testing.T) {
	var s State
	_, pk := KeyGen(2048)
	s.ConAccounts = make(map[string]ContractAccount)
	s.ConAccounts["address22"] = ContractAccount{pk, 200}

	s.PrepayContracts("address22", 155)

	if s.ConAccounts["address22"].Prepaid != 355 {
		t.Error("not correct amount!")
	}

}

func TestState_FundContractCall(t *testing.T) {
	var s State
	s.Ledger = make(map[string]int)
	_, pk1 := KeyGen(2048)
	s.Ledger[pk1.String()] = 100

	s.ConAccounts = make(map[string]ContractAccount)
	s.ConAccounts["address22"] = ContractAccount{pk1, 200}

	success := s.FundContractCall(pk1, 50)

	if !success {
		t.Error("Fund account didn't succeed!")
	}

	if s.Ledger[pk1.String()] != 50 && s.ConAccounts["address22"].Prepaid != 250 {
		t.Error("not correct amount!")
	}
}

func TestState_CleanContractLedger(t *testing.T) {
	var s State
	_, pk := KeyGen(2048)
	s.Ledger = make(map[string]int)
	s.Ledger[pk.String()] = 100
	s.ConStake = make(map[string]int)
	s.ConStake["address1"] = 100
	s.ConStake["address2"] = 100
	s.ConStake["address3"] = 100
	s.ConStake["address4"] = 100

	s.ConAccounts = make(map[string]ContractAccount)
	s.ConAccounts["address1"] = ContractAccount{pk, 200}
	s.ConAccounts["address2"] = ContractAccount{pk, 1}
	s.ConAccounts["address3"] = ContractAccount{pk, 0}
	s.ConAccounts["address4"] = ContractAccount{pk, -1}

	expiredContracts := s.CleanContractLedger()

	if _, success := s.ConAccounts["address1"]; !success {
		t.Error("address1 has been deleted...")
	}
	if _, success := s.ConAccounts["address2"]; !success {
		t.Error("address2 has been deleted...")
	}
	if len(expiredContracts) != 2 {
		t.Error("Not all expired contracts has been deleted...")
	}
	if s.Ledger[pk.String()] != 300 {
		t.Error("Not correct amount in ledger for pk1")
	}

}
