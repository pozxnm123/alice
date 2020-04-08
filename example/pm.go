// Copyright © 2020 AMIS Technologies
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package main

import (
	"sync"
	"time"

	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/host"
)

type peerManager struct {
	id       string
	numPeers uint32
	host     host.Host
}

func newPeerManager(id string, numPeers uint32, host host.Host) *peerManager {
	return &peerManager{
		id:       id,
		numPeers: numPeers,
		host:     host,
	}
}

func (p *peerManager) NumPeers() uint32 {
	return p.numPeers
}

func (p *peerManager) SelfID() string {
	return p.id
}

func (p *peerManager) MustSend(peerID string, message proto.Message) {
	peerAddr, _ := readFullAddr(peerID)
	send(p.host, peerAddr, message)
}

// EnsureAllConnected connects the host to specified peer and sends the message to it.
func (p *peerManager) EnsureAllConnected(peerID string, msg proto.Message, wg *sync.WaitGroup) {
	defer wg.Done()

	logger := log.New("to", peerID)

	for {
		// Read the full address of the peer from config file.
		peerAddr, err := readFullAddr(peerID)
		if err != nil {
			logger.Warn("Cannot read peer address from config file", "err", err)
			return
		}
		// Peer host hasn't been setup. Try to get the peer address later.
		if len(peerAddr) == 0 {
			logger.Debug("Peer address is empty. Sleep 3 seconds and retry")
			time.Sleep(3 * time.Second)
			continue
		}
		// Connect the host to the peer.
		err = connect(p.host, peerAddr)
		if err != nil {
			logger.Warn("Failed to connect to peer", "err", err)
			return
		}
		// Send the message to the peer.
		err = send(p.host, peerAddr, msg)
		if err != nil {
			logger.Warn("Failed to send message to peer", "err", err)
			return
		}
		logger.Debug("Successfully connect to peer")
		return
	}
}
