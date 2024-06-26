/*
Copyright IBM Corp. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package fabtoken

import (
	"github.com/hyperledger-labs/fabric-smart-client/platform/view/view"
	"github.com/hyperledger-labs/fabric-token-sdk/token/core/common"
	"github.com/hyperledger-labs/fabric-token-sdk/token/core/common/meta"
	"github.com/hyperledger-labs/fabric-token-sdk/token/core/logging"
	"github.com/hyperledger-labs/fabric-token-sdk/token/driver"
	"github.com/hyperledger-labs/fabric-token-sdk/token/token"
	"github.com/pkg/errors"
)

type TransferService struct {
	Logger                  logging.Logger
	PublicParametersManager common.PublicParametersManager[*PublicParams]
	WalletService           driver.WalletService
	TokenLoader             TokenLoader
	Deserializer            driver.Deserializer
}

func NewTransferService(
	logger logging.Logger,
	publicParametersManager common.PublicParametersManager[*PublicParams],
	walletService driver.WalletService,
	tokenLoader TokenLoader,
	deserializer driver.Deserializer,
) *TransferService {
	return &TransferService{
		Logger:                  logger,
		PublicParametersManager: publicParametersManager,
		WalletService:           walletService,
		TokenLoader:             tokenLoader,
		Deserializer:            deserializer,
	}
}

// Transfer returns a TransferAction as a function of the passed arguments
// It also returns the corresponding TransferMetadata
func (s *TransferService) Transfer(txID string, wallet driver.OwnerWallet, ids []*token.ID, Outputs []*token.Token, opts *driver.TransferOptions) (driver.TransferAction, *driver.TransferMetadata, error) {
	// select inputs
	inputIDs, inputTokens, err := s.TokenLoader.GetTokens(ids)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to load tokens")
	}

	var signerIds []view.Identity
	for _, tok := range inputTokens {
		s.Logger.Debugf("Selected output [%s,%s,%s]", tok.Type, tok.Quantity, view.Identity(tok.Owner.Raw))
		signerIds = append(signerIds, tok.Owner.Raw)
	}

	// prepare outputs
	var outs []*Output
	var metas [][]byte
	for _, output := range Outputs {
		outs = append(outs, &Output{
			Output: output,
		})
		meta := &OutputMetadata{}
		metaRaw, err := meta.Serialize()
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed serializing token information")
		}
		metas = append(metas, metaRaw)
	}

	// assemble transfer action
	transfer := &TransferAction{
		Inputs:   inputIDs,
		Outputs:  outs,
		Metadata: meta.TransferActionMetadata(opts.Attributes),
	}

	ws := s.WalletService

	// assemble transfer metadata
	var receivers []view.Identity
	for i, output := range outs {
		if output.Output == nil || output.Output.Owner == nil {
			return nil, nil, errors.Errorf("failed to transfer: invalid output at index %d", i)
		}
		if len(output.Output.Owner.Raw) == 0 { // redeem
			receivers = append(receivers, output.Output.Owner.Raw)
			continue
		}
		recipients, err := s.Deserializer.Recipients(output.Output.Owner.Raw)
		if err != nil {
			return nil, nil, errors.Wrap(err, "failed getting recipients")
		}
		receivers = append(receivers, recipients...)
	}

	var senderAuditInfos [][]byte
	for _, t := range inputTokens {
		auditInfo, err := s.Deserializer.GetOwnerAuditInfo(t.Owner.Raw, ws)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed getting audit info for sender identity [%s]", view.Identity(t.Owner.Raw).String())
		}
		senderAuditInfos = append(senderAuditInfos, auditInfo...)
	}

	var receiverAuditInfos [][]byte
	for _, output := range outs {
		auditInfo, err := s.Deserializer.GetOwnerAuditInfo(output.Output.Owner.Raw, ws)
		if err != nil {
			return nil, nil, errors.Wrapf(err, "failed getting audit info for recipient identity [%s]", view.Identity(output.Output.Owner.Raw).String())
		}
		receiverAuditInfos = append(receiverAuditInfos, auditInfo...)
	}
	outputs, err := transfer.GetSerializedOutputs()
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed getting serialized outputs")
	}

	receiverIsSender := make([]bool, len(receivers))
	for i, receiver := range receivers {
		_, err = ws.OwnerWallet(receiver)
		receiverIsSender[i] = err == nil
	}

	metadata := &driver.TransferMetadata{
		Outputs:            outputs,
		Senders:            signerIds,
		SenderAuditInfos:   senderAuditInfos,
		TokenIDs:           ids,
		OutputsMetadata:    metas,
		Receivers:          receivers,
		ReceiverIsSender:   receiverIsSender,
		ReceiverAuditInfos: receiverAuditInfos,
	}

	s.Logger.Debugf("Transfer metadata: [out:%d, rec:%d]", len(metadata.Outputs), len(metadata.Receivers))

	// done
	return transfer, metadata, nil
}

// VerifyTransfer checks the outputs in the TransferAction against the passed tokenInfos
func (s *TransferService) VerifyTransfer(tr driver.TransferAction, outputsMetadata [][]byte) error {
	// TODO:
	return nil
}

// DeserializeTransferAction un-marshals a TransferAction from the passed array of bytes.
// DeserializeTransferAction returns an error, if the un-marshalling fails.
func (s *TransferService) DeserializeTransferAction(raw []byte) (driver.TransferAction, error) {
	t := &TransferAction{}
	if err := t.Deserialize(raw); err != nil {
		return nil, errors.Wrap(err, "failed deserializing transfer action")
	}
	return t, nil
}
