// Diode Network Client
// Copyright 2019 IoT Blockchain Technology Corporation LLC (IBTC)
// Licensed under the Diode License, Version 1.0
package edge

import (
	"bytes"
	"fmt"

	"github.com/diodechain/diode_go_client/blockquick"
	"github.com/diodechain/diode_go_client/crypto/secp256k1"
	"github.com/diodechain/diode_go_client/rlp"
	"github.com/diodechain/diode_go_client/util"
)

var (
	responsePivot     = []byte("response")
	errorPivot        = []byte("error")
	ticketTooOldPivot = []byte("too_old")
	ticketTooLowPivot = []byte("too_low")
	ticketThanksPivot = []byte("thanks!")
	portOpenPivot     = []byte("portopen")
	portSendPivot     = []byte("portsend")
	portClosePivot    = []byte("portclose")
	goodbyePivot      = []byte("goodbye")
	// Maybe remove parse callback and use parse response?
	blockPivot                 = []byte("getblock")
	block2Pivot                = []byte("getblock2")
	blockHeaderPivot           = []byte("getblockheader")
	blockHeader2Pivot          = []byte("getblockheader2")
	blockquickPivot            = []byte("getblockquick")
	blockquick2Pivot           = []byte("getblockquick2")
	blockPeakPivot             = []byte("getblockpeak")
	accountRootsPivot          = []byte("getaccountroots")
	accountValuePivot          = []byte("getaccountvalue")
	accountPivot               = []byte("getaccount")
	stateRootsPivot            = []byte("getstateroots")
	objectPivot                = []byte("getobject")
	nodePivot                  = []byte("getnode")
	ticketPivot                = []byte("getticket")
	helloPivot                 = []byte("hello")
	okPivot                    = []byte("ok")
	errWrongTypeForItems       = fmt.Errorf("items should be array or slice")
	errKeyNotFoundInItems      = fmt.Errorf("key not found")
	ErrFailedToParseTicket     = fmt.Errorf("failed to parse ticket")
	ErrResponseHandlerNotFound = fmt.Errorf("couldn't find handler for response")
	ErrRPCNotSupport           = fmt.Errorf("rpc method not support")
)

type RLP_V2 struct{}

// parse response
func (rlpV2 RLP_V2) parseResponse(buffer []byte) (interface{}, error) {
	if bytes.Contains(buffer, portOpenPivot) {
		return rlpV2.parsePortOpenResponse(buffer)
	} else if bytes.Contains(buffer, blockPivot) {
		return rlpV2.parseBlockResponse(buffer)
	} else if bytes.Contains(buffer, block2Pivot) {
		return rlpV2.parseBlockResponse(buffer)
	} else if bytes.Contains(buffer, blockHeaderPivot) {
		return rlpV2.parseBlockHeaderResponse(buffer)
	} else if bytes.Contains(buffer, blockHeader2Pivot) {
		return rlpV2.parseBlockHeaderResponse(buffer)
	} else if bytes.Contains(buffer, blockquickPivot) {
		return rlpV2.parseBlockquickResponse(buffer)
	} else if bytes.Contains(buffer, blockquick2Pivot) {
		return rlpV2.parseBlockquickResponse(buffer)
	} else if bytes.Contains(buffer, blockPeakPivot) {
		return rlpV2.parseBlockPeakResponse(buffer)
	} else if bytes.Contains(buffer, accountPivot) {
		return rlpV2.parseAccountResponse(buffer)
	} else if bytes.Contains(buffer, accountValuePivot) {
		return rlpV2.parseAccountValueResponse(buffer)
	} else if bytes.Contains(buffer, accountRootsPivot) {
		return rlpV2.parseAccountRootsResponse(buffer)
	} else if bytes.Contains(buffer, stateRootsPivot) {
		return rlpV2.parseStateRootsResponse(buffer)
	} else if bytes.Contains(buffer, objectPivot) {
		return rlpV2.parseDeviceObjectResponse(buffer)
	} else if bytes.Contains(buffer, nodePivot) {
		return rlpV2.parseServerObjResponse(buffer)
	} else if bytes.Contains(buffer, ticketPivot) {
		return rlpV2.parseDeviceTicketResponse(buffer)
	}
	return nil, ErrResponseHandlerNotFound
}

func (rlpV2 RLP_V2) parseError(buffer []byte) (Error, error) {
	var response errorResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	_ = decodeStream.Decode(&response)
	err := Error{
		// TODO: method required?
		// Method:  response.Payload[0],
		// TODO: response.Payload[1] will be ? string
		Message: response.Payload[len(response.Payload)-1],
	}
	return err, nil
}

// parse response of rpc call
func (rlpV2 RLP_V2) parseBlockPeakResponse(buffer []byte) (interface{}, error) {
	var response blockPeakResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	return response.Payload.BlockNumber, nil
}

// TODO: parse block
func (rlpV2 RLP_V2) parseBlockResponse(buffer []byte) (interface{}, error) {
	var response blockResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	// response.Payload.Block.Coinbase
	// response.Payload.Block.Header
	// response.Payload.Block.Receipts
	// response.Payload.Block.Transactions
	return response, nil
}

// TODO: check error from findItemInItems
// TODO: use big.Int instead of uint64?
func (rlpV2 RLP_V2) parseBlockHeaderResponse(buffer []byte) (interface{}, error) {
	var response blockHeaderResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	// get value
	txHash, _ := findItemInItems(response.Payload.Items, "transaction_hash")
	stateHash, _ := findItemInItems(response.Payload.Items, "state_hash")
	blockHash, _ := findItemInItems(response.Payload.Items, "block_hash")
	prevBlock, _ := findItemInItems(response.Payload.Items, "previous_block")
	nonce, _ := findItemInItems(response.Payload.Items, "nonce")
	minerSig, _ := findItemInItems(response.Payload.Items, "miner_signature")
	timestamp, _ := findItemInItems(response.Payload.Items, "timestamp")
	number, _ := findItemInItems(response.Payload.Items, "number")
	// also can decompress pubkey and marshal to pubkey bytes
	dminerPubkey := secp256k1.DecompressPubkeyBytes(response.Payload.MinerPubkey)
	header, err := blockquick.NewHeader(
		txHash.Value,
		stateHash.Value,
		prevBlock.Value,
		minerSig.Value,
		dminerPubkey,
		util.DecodeBytesToUint(timestamp.Value),
		util.DecodeBytesToUint(number.Value),
		util.DecodeBytesToUint(nonce.Value),
	)
	if err != nil {
		return nil, err
	}
	hash := header.Hash()
	if !bytes.Equal(hash[:], blockHash.Value) {
		return nil, fmt.Errorf("Blockhash != real hash %v %v", blockHash.Value, header)
	}
	return header, nil
}

func (rlpV2 RLP_V2) parseBlockquickResponse(buffer []byte) (interface{}, error) {
	var response blockquickResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	return response.Payload.Items, nil
}

// TODO: check error from findItemInItems
// TODO: use big.Int instead of uint64?
func (rlpV2 RLP_V2) parseDeviceTicketResponse(buffer []byte) (interface{}, error) {
	if bytes.Contains(buffer, ticketThanksPivot) {
		var response ticketThanksResponse
		decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
		err := decodeStream.Decode(&response)
		if err != nil {
			return nil, err
		}
		// create empty ticket
		ticket := DeviceTicket{}
		return ticket, nil
	} else if bytes.Contains(buffer, ticketTooLowPivot) {
		var response ticketTooLowResponse
		decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
		err := decodeStream.Decode(&response)
		if err != nil {
			return nil, err
		}
		err = ErrTicketTooLow
		ticket := DeviceTicket{
			BlockHash:        response.Payload.BlockHash,
			TotalConnections: response.Payload.TotalConnections,
			TotalBytes:       response.Payload.TotalBytes,
			LocalAddr:        response.Payload.LocalAddr,
			DeviceSig:        response.Payload.DeviceSig,
			Err:              err,
		}
		return ticket, nil
	} else if bytes.Contains(buffer, ticketTooOldPivot) {
		var response ticketTooOldResponse
		decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
		err := decodeStream.Decode(&response)
		if err != nil {
			return nil, err
		}
		err = ErrTicketTooOld
		ticket := DeviceTicket{
			Err: err,
		}
		return ticket, nil
	}
	return nil, ErrFailedToParseTicket
}

func (rlpV2 RLP_V2) parseDeviceObjectResponse(buffer []byte) (interface{}, error) {
	var response objectResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	serverID := [20]byte{}
	copy(serverID[:], response.Payload.Ticket.ServerID)
	fleetAddr := [20]byte{}
	copy(fleetAddr[:], response.Payload.Ticket.FleetAddr)
	deviceObj := &DeviceTicket{
		ServerID:         serverID,
		BlockNumber:      int(response.Payload.Ticket.PeakBlock),
		BlockHash:        nil,
		FleetAddr:        fleetAddr,
		TotalConnections: response.Payload.Ticket.TotalConnections,
		TotalBytes:       response.Payload.Ticket.TotalBytes,
		DeviceSig:        response.Payload.Ticket.DeviceSig,
		ServerSig:        response.Payload.Ticket.ServerSig,
		LocalAddr:        response.Payload.Ticket.LocalAddr,
	}
	return deviceObj, nil
}

// TODO: decode merkle tree from message
func (rlpV2 RLP_V2) parseAccountResponse(buffer []byte) (interface{}, error) {
	var response accountResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	storageRoot, _ := findItemInItems(response.Payload.Items, "storageRoot")
	nonce, _ := findItemInItems(response.Payload.Items, "nonce")
	code, _ := findItemInItems(response.Payload.Items, "code")
	balance, _ := findItemInItems(response.Payload.Items, "balance")
	dnonce := util.DecodeBytesToInt(nonce.Value)
	dbalance := util.DecodeBytesToInt(balance.Value)
	stateTree, err := rlpV2.NewMerkleTree(response.Payload.MerkleProof)
	if err != nil {
		return nil, err
	}
	account := &Account{
		StorageRoot: storageRoot.Value,
		Nonce:       int64(dnonce),
		Code:        code.Value,
		Balance:     int64(dbalance),
		stateTree:   stateTree,
	}
	return account, nil
}

func (rlpV2 RLP_V2) parseAccountRootsResponse(buffer []byte) (interface{}, error) {
	var response accountRootsResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	accountRoots := &AccountRoots{
		AccountRoots: response.Payload.AccountRoots,
	}
	return accountRoots, nil
}

func (rlpV2 RLP_V2) parseAccountValueResponse(buffer []byte) (interface{}, error) {
	var response accountValueResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	accountTree, err := rlpV2.NewMerkleTree(response.Payload.MerkleProof)
	if err != nil {
		return nil, err
	}
	accountValue := &AccountValue{
		accountTree: accountTree,
	}
	return accountValue, nil
}

func (rlpV2 RLP_V2) parsePortOpenResponse(buffer []byte) (interface{}, error) {
	var response portOpenResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	portOpen := &PortOpen{
		Ok:  (response.Payload.Result == "ok"),
		Ref: response.Payload.Ref,
	}
	return portOpen, nil
}

func (rlpV2 RLP_V2) parseServerObjResponse(buffer []byte) (interface{}, error) {
	var response serverObjectResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	serverObj := &ServerObj{
		Host:       response.Payload.ServerObject.Host,
		EdgePort:   response.Payload.ServerObject.EdgePort,
		ServerPort: response.Payload.ServerObject.ServerPort,
		Sig:        response.Payload.ServerObject.ServerSig,
	}
	return serverObj, nil
}

// TODO: check error from jsonparser
func (rlpV2 RLP_V2) parseStateRootsResponse(buffer []byte) (interface{}, error) {
	var response stateRootsResponse
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&response)
	if err != nil {
		return nil, err
	}
	stateRoots := &StateRoots{
		StateRoots: response.Payload.StateRoots,
	}
	return stateRoots, nil
}

// parse inbound request
func (rlpV2 RLP_V2) parseInboundPortOpenRequest(buffer []byte) (interface{}, error) {
	var inboundRequest portOpenInboundRequest
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&inboundRequest)
	if err != nil {
		return nil, err
	}
	deviceID := util.EncodeToString(inboundRequest.Payload.DeviceID)
	portOpen := &PortOpen{
		Port:        inboundRequest.Payload.Port,
		Ref:         inboundRequest.Payload.Ref,
		RawDeviceID: inboundRequest.Payload.DeviceID,
		DeviceID:    deviceID,
		Ok:          true,
	}
	return portOpen, nil
}

func (rlpV2 RLP_V2) parseInboundPortSendRequest(buffer []byte) (interface{}, error) {
	var inboundRequest portSendInboundRequest
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&inboundRequest)
	if err != nil {
		return nil, err
	}
	portSend := &PortSend{
		Ref:  inboundRequest.Payload.Ref,
		Data: inboundRequest.Payload.Data,
		Ok:   true,
	}
	return portSend, nil
}

func (rlpV2 RLP_V2) parseInboundPortCloseRequest(buffer []byte) (interface{}, error) {
	var inboundRequest portCloseInboundRequest
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&inboundRequest)
	if err != nil {
		return nil, err
	}
	portClose := &PortClose{
		Ref: inboundRequest.Payload.Ref,
		Ok:  true,
	}
	return portClose, nil
}

// TODO: should test it
func (rlpV2 RLP_V2) parseInboundGoodbyeRequest(buffer []byte) (interface{}, error) {
	var inboundRequest goodbyeInboundRequest
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	err := decodeStream.Decode(&inboundRequest)
	if err != nil {
		return nil, err
	}
	goodbye := Goodbye{
		Reason: inboundRequest.Payload.Reason,
	}
	return goodbye, nil
}

func (rlpV2 RLP_V2) parseInboundRequest(buffer []byte) (req interface{}, err error) {
	if bytes.Contains(buffer, portOpenPivot) {
		return rlpV2.parseInboundPortOpenRequest(buffer)
	} else if bytes.Contains(buffer, portSendPivot) {
		return rlpV2.parseInboundPortSendRequest(buffer)
	} else if bytes.Contains(buffer, portClosePivot) {
		return rlpV2.parseInboundPortCloseRequest(buffer)
	} else if bytes.Contains(buffer, goodbyePivot) {
		return rlpV2.parseInboundGoodbyeRequest(buffer)
	}
	return
}

func (rlpV2 RLP_V2) IsResponseType(rawData []byte) bool {
	return bytes.Contains(rawData, responsePivot)
}

func (rlpV2 RLP_V2) IsErrorType(rawData []byte) bool {
	return bytes.Contains(rawData, errorPivot)
}

func (rlpV2 RLP_V2) ResponseID(buffer []byte) uint64 {
	var response responseID
	decodeStream := rlp.NewStream(bytes.NewReader(buffer), 0)
	decodeStream.Decode(&response)
	return response.RequestID
}

// NewMerkleTree returns merkle tree of given byte of json
// eg: ["0x", "0x1", ["0x2bbfda354b607b8cdd7d52c29344c76c17d76bb7d9187874a994144b55eaf931","0x0000000000000000000000000000000000000000000000000000000000000001"]]
func (rlpV2 RLP_V2) NewMerkleTree(rawTree []interface{}) (mt MerkleTree, err error) {
	mt = MerkleTree{
		mtp:     RLPMerkleTreeParser{},
		RawTree: rawTree,
	}
	rootHash, module, leaves, err := mt.parse()
	if err != nil {
		return
	}
	mt.RootHash = rootHash
	mt.Module = module
	mt.Leaves = leaves
	return
}

func (rlpV2 RLP_V2) NewErrorResponse(method string, err error) Message {
	ret := []byte(fmt.Sprintf("[\"error\", \"%s\", \"%+v\"]", method, err.Error()))
	return Message{
		Len:    len(ret),
		Buffer: ret,
	}
}

func (rlpV2 RLP_V2) NewMessage(requestID uint64, method string, args ...interface{}) ([]byte, func(buffer []byte) (interface{}, error), error) {
	switch method {
	case "hello":
		request := helloRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.Flag = args[0].(uint64)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, nil, nil
	case "portsend":
		request := portSendRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.Ref = args[0].(uint64)
		request.Payload.Data = args[1].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, nil, nil
	case "portclose":
		request := portCloseRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.Ref = args[0].(uint64)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, nil, nil
	case "getblock":
		request := blockRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.BlockNumber = args[0].(uint64)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseBlockResponse, nil
	case "getblockpeak":
		request := blockPeakRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseBlockPeakResponse, nil
	case "getblockheader2":
		request := blockHeaderRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.BlockNumber = args[0].(uint64)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseBlockHeaderResponse, nil
	case "getblockquick2":
		request := blockquickRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.LastValid = args[0].(uint64)
		request.Payload.WindowSize = args[1].(uint64)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseBlockquickResponse, nil
	case "getaccount":
		request := accountRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.BlockNumber = args[0].(uint64)
		request.Payload.Address = args[1].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseAccountResponse, nil
	case "getaccountroots":
		request := accountRootsRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.BlockNumber = args[0].(uint64)
		request.Payload.Address = args[1].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseAccountRootsResponse, nil
	case "getaccountvalue":
		request := accountValueRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.BlockNumber = args[0].(uint64)
		request.Payload.Address = args[1].([]byte)
		request.Payload.Key = args[2].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseAccountValueResponse, nil
	case "ticket":
		request := ticketRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.BlockNumber = args[0].(uint64)
		request.Payload.FleetAddr = args[1].([]byte)
		request.Payload.TotalConnections = args[2].(uint64)
		request.Payload.TotalBytes = args[3].(uint64)
		request.Payload.LocalAddr = args[4].([]byte)
		request.Payload.DeviceSig = args[5].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseDeviceTicketResponse, nil
	case "portopen":
		request := portOpenRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.DeviceID = args[0].([]byte)
		request.Payload.Port = args[1].(uint64)
		request.Payload.Mode = args[2].(string)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parsePortOpenResponse, nil
	case "getobject":
		request := objectRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.DeviceID = args[0].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseDeviceObjectResponse, nil
	case "getnode":
		request := serverObjectRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.ServerID = args[0].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseServerObjResponse, nil
	case "getstateroots":
		request := stateRootsRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.BlockNumber = args[0].(uint64)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, rlpV2.parseStateRootsResponse, nil
	default:
		return nil, nil, ErrRPCNotSupport
	}
}

func (rlpV2 RLP_V2) NewResponseMessage(requestID uint64, responseType string, method string, args ...interface{}) ([]byte, func(buffer []byte) (interface{}, error), error) {
	switch method {
	case "portopen":
		request := portOpenOutboundResponse{}
		request.RequestID = requestID
		request.Payload.ResponseType = responseType
		request.Payload.Ref = args[0].(uint64)
		request.Payload.Result = args[1].(string)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, nil, nil
	case "portsend":
		request := portSendRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.Ref = args[0].(uint64)
		request.Payload.Data = args[1].([]byte)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, nil, nil
	case "portclose":
		request := portCloseRequest{}
		request.RequestID = requestID
		request.Payload.Method = method
		request.Payload.Ref = args[0].(uint64)
		decodedRlp, err := rlp.EncodeToBytes(request)
		if err != nil {
			return nil, nil, err
		}
		return decodedRlp, nil, nil
	default:
		return nil, nil, ErrRPCNotSupport
	}
}