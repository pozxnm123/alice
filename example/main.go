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

	"github.com/getamis/sirius/log"
	"github.com/libp2p/go-libp2p-core/network"
)

const (
	dkgProtocol = "/dkg/1.0.0"
)

func main() {
	id := flag.Uint64("id", 0, "id for this node")
	configPath := flag.String("config", "", "config path")
	flag.Parse()
	if *id <= 0 {
		log.Crit("Please provide id")
	}
	if *configPath == "" {
		log.Crit("empty config path")
	}

	config, err := readYamlFile(*configPath)
	if err != nil {
		log.Crit("Failed to read config file", "configPath", *configPath, err)
	}

	// For convenience, set port to id + 10000.
	port := *id + 10000
	// Make a host that listens on the given multiaddress.
	host, err := makeBasicHost(port)
	if err != nil {
		log.Crit("Failed to create a basic host", "err", err)
	}

	// Write the full address to config file.
	writeFullAddr(getID(*id), fmt.Sprintf("/ip4/127.0.0.1/tcp/%d/p2p/%s", port, host.ID().Pretty()))

	// Create a new peer manager with peer number to be 2.
	pm := newPeerManager(getID(*id), uint32(2), host)
	service, err := NewService(config, pm)
	if err != nil {
		log.Crit("Failed to new service", "err", err)
	}
	// Set a stream handler on the host.
	host.SetStreamHandler(dkgProtocol, func(s network.Stream) {
		service.Handle(s)
	})

	// Start DKG process
	service.Process()
}
