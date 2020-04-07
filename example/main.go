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
	"flag"
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/btcsuite/btcd/btcec"
	"github.com/getamis/alice/crypto/tss"
	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/alice/crypto/tss/message/types"
	"github.com/getamis/sirius/log"
	"github.com/golang/protobuf/proto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/network"
)

const (
	dkgProtocol     = "/dkg/1.0.0"
	signerProtocol  = "/signer/1.0.0"
	reshareProtocol = "/reshare/1.0.0"
)

// For simplicity, we use S256 curve in this example.
var curve = btcec.S256()

type listener struct{}

func newListener() *listener {
	return &listener{}
}

func (l *listener) OnStateChanged(oldState types.MainState, newState types.MainState) {
	// Handle state changed.
}

type peerManager struct {
	id       string
	numPeers uint32
	host     host.Host
	dkg      *dkg.DKG
}

func newPeerManager(id string, numPeers uint32, host host.Host) *peerManager {
	return &peerManager{
		id:       id,
		numPeers: numPeers,
		host:     host,
	}
}

func (p *peerManager) setDKG(dkg *dkg.DKG) {
	p.dkg = dkg
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

func (p *peerManager) handleDKGData(s network.Stream) {
	data := &dkg.Message{}
	buf, err := ioutil.ReadAll(s)
	if err != nil {
		s.Reset()
		log.Warn("Cannot read data from stream", "err", err)
		return
	}
	s.Close()

	// unmarshal it
	proto.Unmarshal(buf, data)
	if err != nil {
		log.Error("Cannot unmarshal data", "err", err)
		return
	}

	log.Info("Received request", "from", s.Conn().RemotePeer())
	p.dkg.AddMessage(data)
}

// connectToPeer connects the host to specified peer and sends the message to it.
func (p *peerManager) connectToPeer(peerID string, msg proto.Message, wg *sync.WaitGroup) {
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

func (p *peerManager) startDKG() {
	// 1. Start a DKG process.
	threshold, rank, err := readDKGSetup(p.id)
	if err != nil {
		log.Warn("Cannot read DKG setup from config", "err", err)
		return
	}
	d, err := dkg.NewDKG(curve, p, threshold, rank, newListener())
	if err != nil {
		log.Warn("Cannot create a new DKG", "threshold", threshold, "rank", rank, "err", err)
		return
	}

	d.Start()
	p.setDKG(d)
	log.Info("Started a DKG", "threshold", threshold, "rank", rank)

	var wg sync.WaitGroup

	// 2. Connect the host to peers and send the peer message to them.
	msg := p.dkg.GetPeerMessage()
	peerIDs := getPeerIDs(p.SelfID())
	for _, peerID := range peerIDs {
		wg.Add(1)
		go p.connectToPeer(peerID, msg, &wg)
	}
	wg.Wait()

	// 3. Get the DKG result and write it to the config file.
	for {
		result, err := p.dkg.GetResult()
		if err != nil {
			if err == tss.ErrNotReady {
				log.Debug("DKG process is still processing. Sleep 3 seconds and retry")
				time.Sleep(3 * time.Second)
				continue
			}
			log.Warn("Cannot get result from DKG", "err", err)
			return
		}
		writeDKGResult(p.id, result)
		break
	}
	log.Info("Got dkg result")
}

func main() {
	id := flag.Int("id", -1, "id for this node")
	flag.Parse()

	if *id == -1 {
		log.Error("Please provide ID with -id")
		return
	}

	// For convenience, set port to id + 10000.
	port := *id + 10000

	// Make a host that listens on the given multiaddress.
	host, err := makeBasicHost(port)
	if err != nil {
		log.Warn("Cannot create a basic host", "err", err)
		return
	}

	// Write the full address to config file.
	writeFullAddr(getID(*id), fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, host.ID().Pretty()))

	// Create a new peer manager with peer number to be 2.
	pm := newPeerManager(getID(*id), uint32(2), host)

	// Set a stream handler on the host.
	host.SetStreamHandler(dkgProtocol, func(s network.Stream) {
		pm.handleDKGData(s)
	})

	// Start DKG process
	pm.startDKG()
}
