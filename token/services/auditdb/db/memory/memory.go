/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package memory

import (
	"crypto/sha256"
	"fmt"

	view2 "github.com/hyperledger-labs/fabric-smart-client/platform/view"
	"github.com/hyperledger-labs/fabric-token-sdk/token"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/auditdb"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/auditdb/db/sql"
	"github.com/hyperledger-labs/fabric-token-sdk/token/services/db/driver"
	sqldb "github.com/hyperledger-labs/fabric-token-sdk/token/services/db/sql"
	"github.com/pkg/errors"
	_ "modernc.org/sqlite"
)

type Driver struct {
	*sql.Driver
}

// Open returns a pure go sqlite implementation in memory for testing purposes.
func (d *Driver) Open(sp view2.ServiceProvider, tmsID token.TMSID) (driver.AuditTransactionDB, error) {
	h := sha256.New()
	if _, err := h.Write([]byte(tmsID.String())); err != nil {
		return nil, err
	}

	sqlDB, err := d.Driver.OpenSQLDB(
		"sqlite",
		fmt.Sprintf("file:%x?mode=memory&cache=shared", h.Sum(nil)),
		10,
		false,
	)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to open memory db for [%s]", tmsID)
	}

	return sqldb.NewTransactionDB(sqlDB, "memory", true)
}

func init() {
	auditdb.Register("memory", &Driver{Driver: sql.NewDriver()})
}
