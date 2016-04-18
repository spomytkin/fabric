/*
Licensed to the Apache Software Foundation (ASF) under one
or more contributor license agreements.  See the NOTICE file
distributed with this work for additional information
regarding copyright ownership.  The ASF licenses this file
to you under the Apache License, Version 2.0 (the
"License"); you may not use this file except in compliance
with the License.  You may obtain a copy of the License at
  http://www.apache.org/licenses/LICENSE-2.0
Unless required by applicable law or agreed to in writing,
software distributed under the License is distributed on an
"AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
KIND, either express or implied.  See the License for the
specific language governing permissions and limitations
under the License.
*/

package helper

import (
	"github.com/hyperledger/fabric/consensus"
	"github.com/hyperledger/fabric/core/peer"

	"fmt"
	"github.com/hyperledger/fabric/core/chaincode"
	pb "github.com/hyperledger/fabric/protos"
	"golang.org/x/net/context"
	"sync"
)

type EngineImpl struct {
	consenter    consensus.Consenter
	peerEndpoint *pb.PeerEndpoint
}

func (eng *EngineImpl) GetHandlerFactory() peer.HandlerFactory {
	return NewConsensusHandler
}

func (eng *EngineImpl) ProcessTransactionMsg(msg *pb.Message, tx *pb.Transaction) (response *pb.Response) {
	//TODO: Do we always verify security, or can we supply a flag on the invoke ot this functions so to bypass check for locally generated transactions?
	if tx.Type == pb.Transaction_CHAINCODE_QUERY {
		// The secHelper is set during creat ChaincodeSupport, so we don't need this step
		// cxt := context.WithValue(context.Background(), "security", secHelper)
		cxt := context.Background()
		result, err := chaincode.Execute(cxt, chaincode.GetChain(chaincode.DefaultChain), tx)
		if err != nil {
			response = &pb.Response{Status: pb.Response_FAILURE,
				Msg: []byte(fmt.Sprintf("Error:%s", err))}
		} else {
			response = &pb.Response{Status: pb.Response_SUCCESS, Msg: result}
		}
	} else {
		// Chaincode Transaction
		response = &pb.Response{Status: pb.Response_SUCCESS, Msg: []byte(tx.Uuid)}

		//TODO: Do we need to verify security, or can we supply a flag on the invoke ot this functions
		// If we fail to marshal or verify the tx, don't send it to consensus plugin
		if response.Status == pb.Response_FAILURE {
			return response
		}

		// Pass the message to the consenter (eg. PBFT)
		err := eng.consenter.RecvMsg(msg, eng.peerEndpoint.ID)
		if err != nil {
			response = &pb.Response{Status: pb.Response_FAILURE, Msg: []byte(err.Error())}
		}
	}
	return response
}

func (eng *EngineImpl) setConsenter(consenter consensus.Consenter) *EngineImpl {
	eng.consenter = consenter
	return eng
}

func (eng *EngineImpl) setPeerEndpoint(peerEndpoint *pb.PeerEndpoint) *EngineImpl {
	eng.peerEndpoint = peerEndpoint
	return eng
}

var engineOnce sync.Once

var engine *EngineImpl

func getEngineImpl() *EngineImpl {
	return engine
}

func GetEngine(coord peer.MessageHandlerCoordinator) (peer.Engine, error) {
	var err error
	engine = new(EngineImpl)
	//engineOnce.Do(func() {
	//	engine = new(EngineImpl)
	//	engine.consenter = controller.NewConsenter(NewHelper(coord))
	//	engine.peerEndpoint, err = coord.GetPeerEndpoint()
	//
	//})
	return engine, err
}
