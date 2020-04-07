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
	"fmt"

	"github.com/getamis/alice/crypto/tss/dkg"
	"github.com/getamis/sirius/log"
)

func readFullAddr(id string) (string, error) {
	c, err := readYamlFile(getFilePath(id))
	if err != nil {
		log.Error("Cannot read YAML file", "err", err)
		return "", err
	}

	return c.FullAddr, nil
}

func writeFullAddr(id string, fulladdr string) error {
	c, err := readYamlFile(getFilePath(id))
	if err != nil {
		log.Error("Cannot read YAML file", "err", err)
		return err
	}
	c.FullAddr = fulladdr
	err = writeYamlFile(c, getFilePath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func readDKGSetup(id string) (uint32, uint32, error) {
	c, err := readYamlFile(getFilePath(id))
	if err != nil {
		log.Error("Cannot read YAML file", "err", err)
		return uint32(0), uint32(0), err
	}
	return c.Threshold.DKG, c.Rank, nil
}

func writeDKGResult(id string, result *dkg.Result) error {
	c, err := readYamlFile(getFilePath(id))
	if err != nil {
		log.Error("Cannot read YAML file", "err", err)
		return err
	}
	c.DKGResult.Share = result.Share.String()
	c.DKGResult.Pubkey.X = result.PublicKey.GetX().String()
	c.DKGResult.Pubkey.Y = result.PublicKey.GetY().String()
	for peerID, bk := range result.Bks {
		c.DKGResult.BKs[peerID] = bk.GetX().String()
	}
	// Clear the full address of this host.
	c.FullAddr = ""
	err = writeYamlFile(c, getFilePath(id))
	if err != nil {
		log.Error("Cannot write YAML file", "err", err)
		return err
	}
	return nil
}

func getID(id int) string {
	return fmt.Sprintf("id-%d", id)
}

func getPeerIDs(selfID string) []string {
	var peerIDs []string
	for i := 1; i <= 3; i++ {
		peerID := getID(i)
		if peerID != selfID {
			peerIDs = append(peerIDs, peerID)
		}
	}
	return peerIDs
}
