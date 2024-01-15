/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package driver

import (
	"math/big"
	"strconv"
	"strings"
	"time"

	view2 "github.com/hyperledger-labs/fabric-smart-client/platform/view"
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger-labs/fabric-token-sdk/token/driver"
	"github.com/hyperledger-labs/fabric-token-sdk/token/token"
)

// ActionType is the type of transaction
type ActionType int

const (
	// Issue is the action type for issuing tokens.
	Issue ActionType = iota
	// Transfer is the action type for transferring tokens.
	Transfer
	// Redeem is the action type for redeeming tokens.
	Redeem
)

// SearchDirection defines the direction of a search.
type SearchDirection int

const (
	// FromLast defines the direction of a search from the last key.
	FromLast SearchDirection = iota
	// FromBeginning defines the direction of a search from the first key.
	FromBeginning
)

// MovementDirection defines the direction of a movement.
type MovementDirection int

const (
	// Sent amount transferred from.
	Sent MovementDirection = iota
	// Received amount transferred to.
	Received
	// All amount transferred to and from.
	All
)

// TxStatus is the status of a transaction
type TxStatus string

const (
	// Unknown is the status of a transaction that is unknown
	Unknown TxStatus = "Unknown"
	// Pending is the status of a transaction that has been submitted to the ledger
	Pending TxStatus = "Pending"
	// Confirmed is the status of a transaction that has been confirmed by the ledger
	Confirmed TxStatus = "Confirmed"
	// Deleted is the status of a transaction that has been deleted due to a failure to commit
	Deleted TxStatus = "Deleted"
)

// MovementRecord is a record of a movement of assets.
// Given a Token Transaction, a movement record is created for each enrollment ID that participated in the transaction
// and each token type that was transferred.
// The movement record contains the total amount of the token type that was transferred to/from the enrollment ID
// in a given token transaction.
type MovementRecord struct {
	// TxID is the transaction ID
	TxID string
	// EnrollmentID is the enrollment ID of the account that is receiving or sending
	EnrollmentID string
	// TokenType is the type of token
	TokenType string
	// Amount is positive if tokens are received. Negative otherwise
	Amount *big.Int
	// Status is the status of the transaction
	Status TxStatus
}

// TransactionRecord is a more finer-grained version of a movement record.
// Given a Token Transaction, for each token action in the Token Request,
// a transaction record is created for each unique enrollment ID found in the outputs.
// The transaction record contains the total amount of the token type that was transferred to/from that enrollment ID
// in that action.
type TransactionRecord struct {
	// TxID is the transaction ID
	TxID string
	// ActionType is the type of action performed by this transaction record
	ActionType ActionType
	// SenderEID is the enrollment ID of the account that is sending tokens
	SenderEID string
	// RecipientEID is the enrollment ID of the account that is receiving tokens
	RecipientEID string
	// TokenType is the type of token
	TokenType string
	// Amount is positive if tokens are received. Negative otherwise
	Amount *big.Int
	// Timestamp is the time the transaction was submitted to the db
	Timestamp time.Time
	// Status is the status of the transaction
	Status TxStatus
}

func (t *TransactionRecord) String() string {
	var s strings.Builder
	s.WriteString("{")
	s.WriteString(t.TxID)
	s.WriteString(" ")
	s.WriteString(strconv.Itoa(int(t.ActionType)))
	s.WriteString(" ")
	s.WriteString(t.SenderEID)
	s.WriteString(" ")
	s.WriteString(t.RecipientEID)
	s.WriteString(" ")
	s.WriteString(t.TokenType)
	s.WriteString(" ")
	s.WriteString(t.Amount.String())
	s.WriteString(" ")
	s.WriteString(t.Timestamp.String())
	s.WriteString(" ")
	s.WriteString(string(t.Status))
	s.WriteString("}")
	return s.String()
}

// ValidationRecord is a record that contains information about the validation of a given token request
type ValidationRecord struct {
	// TxID is the transaction ID
	TxID string
	// TokenRequest is the token request marshalled
	TokenRequest []byte
	// Metadata is the metadata produced by the validator when evaluating the token request
	Metadata map[string][]byte
	// Timestamp is the time the transaction was submitted to the db
	Timestamp time.Time
	// Status is the status of the transaction
	Status TxStatus
}

// TransactionIterator is an iterator for transactions
type TransactionIterator interface {
	Close()
	Next() (*TransactionRecord, error)
}

// ValidationRecordsIterator is an iterator for transactions
type ValidationRecordsIterator interface {
	Close()
	Next() (*ValidationRecord, error)
}

// QueryMovementsParams defines the parameters for querying movements.
// Movement records will be filtered by EnrollmentID, TokenType, and Status.
// SearchDirection tells if the search should start from the oldest to the newest records or vice versa.
// MovementDirection which amounts to consider. Sent correspond to a negative amount,
// Received to a positive amount, and All to both.
type QueryMovementsParams struct {
	// EnrollmentIDs is the enrollment IDs of the accounts to query
	EnrollmentIDs []string
	// TokenTypes is the token types to query
	TokenTypes []string
	// TxStatuses is the statuses of the transactions to query
	TxStatuses []TxStatus
	// SearchDirection is the direction of the search
	SearchDirection SearchDirection
	// MovementDirection is the direction of the movement
	MovementDirection MovementDirection
	// NumRecords is the number of records to return
	// If 0, all records are returned
	NumRecords int
}

// QueryTransactionsParams defines the parameters for querying transactions.
// One can filter by sender, by recipient, and by time range.
type QueryTransactionsParams struct {
	// SenderWallet is the wallet of the sender
	// If empty, any sender is accepted
	// If the sender does not match but the recipient matches, the transaction is returned
	SenderWallet string
	// RecipientWallet is the wallet of the recipient
	// If empty, any recipient is accepted
	// If the recipient does not match but the sender matches, the transaction is returned
	RecipientWallet string
	// From is the start time of the query
	// If nil, the query starts from the first transaction
	From *time.Time
	// To is the end time of the query
	// If nil, the query ends at the last transaction
	To *time.Time
	// ActionTypes is the list of action types to accept
	// If empty, any action type is accepted
	ActionTypes []ActionType
	// Statuses is the list of transaction status to accept
	// If empty, any status is accepted
	Statuses []TxStatus
}

// QueryValidationRecordsParams defines the parameters for querying validation records.
type QueryValidationRecordsParams struct {
	// From is the start time of the query
	// If nil, the query starts from the first transaction
	From *time.Time
	// To is the end time of the query
	// If nil, the query ends at the last transaction
	To *time.Time
	// Statuses is the list of transaction status to accept
	// If empty, any status is accepted
	Statuses []TxStatus
	// Filter defines a custom filter function.
	// If specified, this filter will be applied.
	// the filter returns true if the record must be selected, false otherwise.
	Filter func(record *ValidationRecord) bool
}

// TokenTransactionDB defines the interface for a token transactions database
type TokenTransactionDB interface {
	// Close closes the database
	Close() error

	// BeginUpdate begins a new update to the database
	BeginUpdate() error

	// Commit commits the current update to the database
	Commit() error

	// Discard discards the current update to the database
	Discard() error

	// SetStatus sets the status of a transaction
	SetStatus(txID string, status TxStatus) error

	// GetStatus returns the status of a given transaction.
	// It returns an error if the transaction is not found
	GetStatus(txID string) (TxStatus, error)

	// AddMovement adds a movement record to the database.
	// Each token transaction can be seen as a list of movements.
	AddMovement(record *MovementRecord) error

	// AddTransaction adds a transaction record to the database.
	AddTransaction(record *TransactionRecord) error

	// QueryTransactions returns a list of transactions that match the given criteria
	QueryTransactions(params QueryTransactionsParams) (TransactionIterator, error)

	// QueryMovements returns a list of movement records
	QueryMovements(params QueryMovementsParams) ([]*MovementRecord, error)

	// QueryValidations returns a list of validation  records
	QueryValidations(params QueryValidationRecordsParams) (ValidationRecordsIterator, error)

	// AddValidationRecord adds a new validation records for the given params
	AddValidationRecord(txID string, tr []byte, meta map[string][]byte) error

	// AddTokenRequest binds the passed transaction id to the passed token request
	AddTokenRequest(txID string, tr []byte) error

	// GetTokenRequest returns the token request bound to the passed transaction id, if available.
	// It returns nil without error if the key is not found.
	GetTokenRequest(txID string) ([]byte, error)

	// AddTransactionEndorsementAck records the signature of a given endorser for a given transaction
	AddTransactionEndorsementAck(txID string, endorser view.Identity, sigma []byte) error

	// GetTransactionEndorsementAcks returns the endorsement signatures for the given transaction id
	GetTransactionEndorsementAcks(txID string) (map[string][]byte, error)

	// TokenDB provides all the features to store and retrieve tokens
	TokenDB
}

type TokenDB interface {
	StoreOwnerToken(tr TokenRecord, owners []string) error
	StoreIssuedToken(tr TokenRecord) error
	StoreAuditToken(tr TokenRecord) error
	Delete(namespace, txID string, index uint64, deletedBy string) error
	IsMine(namespace, txID string, index uint64) (bool, error)
	UnspentTokensIterator(namespace string) (driver.UnspentTokensIterator, error)
	UnspentTokensIteratorBy(namespace string, id, typ string) (driver.UnspentTokensIterator, error)
	ListUnspentTokens(namespace string) (*token.UnspentTokens, error)
	ListUnspentTokensBy(ns, ownerEID, typ string) (*token.UnspentTokens, error)
	ListAuditTokens(namespace string, ids ...*token.ID) ([]*token.Token, error)
	ListHistoryIssuedTokens(namespace string) (*token.IssuedTokens, error)
	GetTokenInfos(namespace string, ids []*token.ID, callback driver.QueryCallbackFunc) error
	GetAllTokenInfos(namespace string, ids []*token.ID) ([][]byte, error)
	GetTokens(namespace string, inputs ...*token.ID) ([]string, []*token.Token, error)
	WhoDeletedTokens(namespace string, inputs ...*token.ID) ([]string, []bool, error)
}

type TokenRecord struct {
	// TxID is the ID of the transaction that created the token
	TxID string
	// Index is the index in the transaction
	Index uint64
	// Namespace is the namespace where the token was created
	Namespace string
	// IssuerRaw represents the serialization of the issuer identity
	// if this is an IssuedToken.
	IssuerRaw []byte
	// OwnerRaw is the serialization of the owner identity
	OwnerRaw []byte
	// InfoRaw is the output of TMS.GetTokenInfo (json encoded)
	InfoRaw string
	// Quantity is the number of units of Type carried in the token.
	// It is encoded as a string containing a number in base 16. The string has prefix ``0x''.
	Quantity string
	// Type is the type of token
	Type string
	// Amount is the Quantity converted to decimal
	Amount uint64
	// TxStatus is the status of the transaction
	TxStatus TxStatus
	// SettledAt is the settlement time if known
	SettledAt time.Time
}

// Driver is the interface for a database driver
type Driver interface {
	// Open opens a database connection
	Open(sp view2.ServiceProvider, name string) (TokenTransactionDB, error)
}
