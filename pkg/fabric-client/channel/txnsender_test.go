/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package channel

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	fab "github.com/hyperledger/fabric-sdk-go/api/apifabclient"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn"
	"github.com/hyperledger/fabric-sdk-go/api/apitxn/mocks"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabric-client/mocks"
	"github.com/hyperledger/fabric/protos/common"
	pb "github.com/hyperledger/fabric/protos/peer"
)

func TestCreateTransaction(t *testing.T) {
	channel, _ := setupTestChannel()

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	//Test Empty proposal response scenario
	_, err := channel.CreateTransaction([]*apitxn.TransactionProposalResponse{})

	if err == nil || err.Error() != "At least one proposal response is necessary" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, for empty proposal response scenario")
	}

	//Test invalid proposal header scenario

	txid := apitxn.TransactionID{
		ID: "1234",
	}

	test := &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TxnID:          txid,
				Proposal:       &pb.Proposal{Header: []byte("TEST"), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input := []*apitxn.TransactionProposalResponse{test}

	_, err = channel.CreateTransaction(input)

	if err == nil || err.Error() != "Could not unmarshal the proposal header" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal header scenario")
	}

	//Test invalid proposal payload scenario
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				TxnID:          txid,
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("TEST")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")},
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}

	_, err = channel.CreateTransaction(input)
	if err == nil || err.Error() != "Could not unmarshal the proposal payload" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction, invalid proposal payload scenario")
	}

	//Test proposal response
	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TxnID: txid,
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 99, Payload: []byte("")}},
		},
	}

	input = []*apitxn.TransactionProposalResponse{test}
	_, err = channel.CreateTransaction(input)

	if err == nil || err.Error() != "Proposal response was not successful, error code 99, msg success" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//Test repeated field header nil scenario

	test = &apitxn.TransactionProposalResponse{
		TransactionProposalResult: apitxn.TransactionProposalResult{
			Endorser: "http://peer1.com",
			Proposal: apitxn.TransactionProposal{
				Proposal:       &pb.Proposal{Header: []byte(""), Extension: []byte(""), Payload: []byte("")},
				SignedProposal: &pb.SignedProposal{Signature: []byte(""), ProposalBytes: []byte("")}, TxnID: txid,
			},
			ProposalResponse: &pb.ProposalResponse{Response: &pb.Response{Message: "success", Status: 200, Payload: []byte("")}},
		},
	}

	_, err = channel.CreateTransaction([]*apitxn.TransactionProposalResponse{test})

	if err == nil || err.Error() != "repeated field endorsements has nil element" {
		t.Fatal("Proposal response was supposed to fail in Create Transaction")
	}

	//TODO: Need actual sample payload for success case

}
func TestSendInstantiateProposal(t *testing.T) {
	//Setup channel
	client := mocks.NewMockClient()
	user := mocks.NewMockUserWithMSPID("test", "1234")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)
	client.SetUserContext(user)
	channel, _ := NewChannel("testChannel", client)

	mockCtrl := gomock.NewController(t)
	defer mockCtrl.Finish()
	proc := mock_apitxn.NewMockProposalProcessor(mockCtrl)

	tp := apitxn.TransactionProposal{SignedProposal: &pb.SignedProposal{}}
	tpr := apitxn.TransactionProposalResult{Endorser: "example.com", Status: 99, Proposal: tp, ProposalResponse: nil}

	proc.EXPECT().ProcessTransactionProposal(gomock.Any()).Return(tpr, nil)
	targets := []apitxn.ProposalProcessor{proc}

	//Add a Peer
	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	tresponse, txnid, err := channel.SendInstantiateProposal("", nil, "",
		"", targets)

	if err == nil || err.Error() != "Missing 'chaincodeName' parameter" {
		t.Fatal("Validation for chain code name parameter for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", targets)

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "",
		"", targets)

	if err == nil || err.Error() != "Missing 'chaincodePath' parameter" {
		t.Fatal("Validation for chain code path for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"", targets)

	if err == nil || err.Error() != "Missing 'chaincodeVersion' parameter" {
		t.Fatal("Validation for chain code version for send Instantiate Proposal failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", targets)

	if err != nil || len(tresponse) == 0 || txnid.ID == "" {
		t.Fatal("Send Instantiate Proposal Test failed")
	}

	tresponse, txnid, err = channel.SendInstantiateProposal("qscc", nil, "test",
		"1", nil)
	if err == nil || err.Error() != "Missing peer objects for instantiate CC proposal" {
		t.Fatal("Missing peer objects validation is not working as expected")
	}

}

func TestBroadcastEnvelope(t *testing.T) {

	//Setup channel
	channel, _ := setupTestChannel()

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)

	//Add an orderer
	channel.AddOrderer(orderer)

	peer := mocks.MockPeer{MockName: "Peer1", MockURL: "http://peer1.com", MockRoles: []string{}, MockCert: nil}
	channel.AddPeer(&peer)

	sigEnvelope := &fab.SignedEnvelope{
		Signature: []byte(""),
		Payload:   []byte(""),
	}
	res, err := channel.BroadcastEnvelope(sigEnvelope)

	if err != nil || res == nil {
		t.Fatalf("Test Broadcast Envelope Failed, cause %s", err.Error())
	}

	channel.RemoveOrderer(orderer)
	_, err = channel.BroadcastEnvelope(sigEnvelope)

	if err == nil || err.Error() != "orderers not set" {
		t.Fatal("orderers not set validation on broadcast envelope is not working as expected")
	}

}

func TestSendTransaction(t *testing.T) {

	channel, _ := setupTestChannel()

	response, err := channel.SendTransaction(nil)

	//Expect orderer is nil error
	if response != nil || err == nil || err.Error() != "orderers is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'orderers is nil' error")
	}

	//Create mock orderer
	orderer := mocks.NewMockOrderer("", nil)

	//Add an orderer
	channel.AddOrderer(orderer)

	//Call Send Transaction with nil tx
	response, err = channel.SendTransaction(nil)

	//Expect tx is nil error
	if response != nil || err == nil || err.Error() != "Transaction is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'Transaction is nil' error")
	}

	//Create tx with nil proposal
	txn := apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: nil,
		},
		Transaction: &pb.Transaction{},
	}

	//Call Send Transaction with nil proposal
	response, err = channel.SendTransaction(&txn)

	//Expect proposal is nil error
	if response != nil || err == nil || err.Error() != "proposal is nil" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'proposal is nil' error")
	}

	//Create tx with improper proposal header
	txn = apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte("TEST")},
		},
		Transaction: &pb.Transaction{},
	}
	//Call Send Transaction
	response, err = channel.SendTransaction(&txn)

	//Expect header unmarshal error
	if response != nil || err == nil || err.Error() != "Could not unmarshal the proposal header" {
		t.Fatal("Test SendTransaction failed, it was supposed to fail with 'Could not unmarshal the proposal header' error")
	}

	//Create tx with proper proposal header
	txn = apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{Header: []byte(""), Payload: []byte(""), Extension: []byte("")},
		},
		Transaction: &pb.Transaction{},
	}

	//Call Send Transaction
	response, err = channel.SendTransaction(&txn)

	if response == nil || err != nil {
		t.Fatalf("Test SendTransaction failed, reason : '%s'", err.Error())
	}
}

func TestBuildChannelHeader(t *testing.T) {

	header, err := BuildChannelHeader(common.HeaderType_CHAINCODE_PACKAGE, "test", "", 1, "1234", time.Time{})

	if err != nil || header == nil {
		t.Fatalf("Test Build Channel Header failed, cause : '%s'", err.Error())
	}

}

func TestSignPayload(t *testing.T) {

	client := mocks.NewMockInvalidClient()
	user := mocks.NewMockUser("test")
	cryptoSuite := &mocks.MockCryptoSuite{}
	client.SaveUserToStateStore(user, true)
	client.SetCryptoSuite(cryptoSuite)
	channel, _ := NewChannel("testChannel", client)

	signedEnv, err := channel.SignPayload([]byte(""))

	if err == nil {
		t.Fatal("Test Sign Payload was supposed to fail")
	}

	channel, _ = setupTestChannel()
	signedEnv, err = channel.SignPayload([]byte(""))

	if err != nil || signedEnv == nil {
		t.Fatal("Test Sign Payload Failed")
	}

}

func TestConcurrentOrderers(t *testing.T) {
	// Determine number of orderers to use - environment can override
	const numOrderersDefault = 10000
	numOrderersEnv := os.Getenv("TEST_MASSIVE_ORDERER_COUNT")
	numOrderers, err := strconv.Atoi(numOrderersEnv)
	if err != nil {
		numOrderers = numOrderersDefault
	}

	channel, err := setupMassiveTestChannel(0, numOrderers)
	if err != nil {
		t.Fatalf("Failed to create massive channel: %s", err)
	}

	txn := apitxn.Transaction{
		Proposal: &apitxn.TransactionProposal{
			Proposal: &pb.Proposal{},
		},
		Transaction: &pb.Transaction{},
	}
	_, err = channel.SendTransaction(&txn)
	if err != nil {
		t.Fatalf("SendTransaction returned error: %s", err)
	}
}
