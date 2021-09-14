// SPDX-FileCopyrightText: 2020-present Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: LicenseRef-ONF-Member-1.0

package utils

import (
	"crypto/md5"
	"fmt"

	uuid2 "github.com/google/uuid"
	topoapi "github.com/onosproject/onos-api/go/onos/topo"
	e2server "github.com/onosproject/onos-e2t/pkg/southbound/e2ap/server"
	"github.com/onosproject/onos-lib-go/pkg/env"
	"github.com/onosproject/onos-lib-go/pkg/uri"
)

// GetE2TID gets E2T URI
func GetE2TID() topoapi.ID {
	return topoapi.ID(uri.NewURI(
		uri.WithScheme("e2"),
		uri.WithOpaque(env.GetPodID())).String())
}

// GetE2ControlRelationID gets E2 relation ID
func GetE2ControlRelationID(connID e2server.ConnID) topoapi.ID {
	return topoapi.ID(connID)
}

func GetCellID(conn *e2server.E2APConn, cell *topoapi.E2Cell) topoapi.ID {
	return topoapi.ID(uri.NewURI(uri.WithOpaque(fmt.Sprintf("%s/%s", conn.E2NodeID, cell.CellGlobalID.Value))).String())
}

func GetCellRelationID(conn *e2server.E2APConn, cell *topoapi.E2Cell) topoapi.ID {
	bytes := md5.Sum([]byte(fmt.Sprintf("%s/%s", conn.E2NodeID, cell.CellGlobalID.Value)))
	uuid, err := uuid2.FromBytes(bytes[:])
	if err != nil {
		panic(err)
	}
	return topoapi.ID(uri.NewURI(
		uri.WithScheme("uuid"),
		uri.WithOpaque(uuid.String())).String())
}