package routes

import (
	"flare-indexer/database"
	"flare-indexer/services/context"
	"flare-indexer/services/utils"
	globalUtils "flare-indexer/utils"
	"flare-indexer/utils/contracts/mirroring"
	"flare-indexer/utils/staking"
	"fmt"
	"net/http"

	"gorm.io/gorm"
)

type GetMirroringRequest struct {
	TxID string `json:"txId"`
}

type MirroringStakeData struct {
	TxID         string `json:"txId"`
	StakingType  uint8  `json:"stakingType"`
	InputAddress string `json:"inputAddress"`
	NodeId       string `json:"nodeId"`
	StartTime    uint64 `json:"startTime"`
	EndTime      uint64 `json:"endTime"`
	Weight       uint64 `json:"weight"`
}

type MirroringResponse struct {
	StakeData   MirroringStakeData `json:"stakeData"`
	MerkleProof []string           `json:"merkleProof"`
	TxInput     string             `json:"txInput"`
}

type GetMirroringResponse []MirroringResponse

type mirroringRouteHandlers struct {
	db     *gorm.DB
	epochs staking.EpochInfo
}

func newMirroringRouteHandlers(ctx context.ServicesContext) *mirroringRouteHandlers {
	return &mirroringRouteHandlers{
		db:     ctx.DB(),
		epochs: staking.NewEpochInfo(&ctx.Config().Epochs),
	}
}

func (rh *mirroringRouteHandlers) listMirroringTransactions() utils.RouteHandler {
	handler := func(request GetMirroringRequest) (GetMirroringResponse, *utils.ErrorHandler) {
		tx, err := database.FetchPChainTx(rh.db, request.TxID)
		if err != nil {
			return GetMirroringResponse{}, utils.InternalServerErrorHandler(err)
		}
		if tx == nil {
			return GetMirroringResponse{}, utils.HttpErrorHandler(http.StatusNotFound, "tx not found")
		}
		response, err := rh.createMirroringData(tx)
		if err != nil {
			return GetMirroringResponse{}, utils.InternalServerErrorHandler(err)
		}
		return response, nil
	}
	return utils.NewRouteHandler(handler, http.MethodPost, GetMirroringRequest{}, GetMirroringResponse{})
}

func AddMirroringRoutes(router utils.Router, ctx context.ServicesContext) {
	rh := newMirroringRouteHandlers(ctx)

	mirroringSubrouter := router.WithPrefix("/mirroring", "Mirroring")
	mirroringSubrouter.AddRoute("/tx_data", rh.listMirroringTransactions())
}

func (rh *mirroringRouteHandlers) createMirroringData(tx *database.PChainTx) ([]MirroringResponse, error) {
	epoch := rh.epochs.GetEpochIndex(*tx.StartTime)
	startTimestamp, endTimestamp := rh.epochs.GetTimeRange(epoch)
	txs, err := database.GetPChainTxsForEpoch(&database.GetPChainTxsForEpochInput{
		DB:             rh.db,
		StartTimestamp: startTimestamp,
		EndTimestamp:   endTimestamp,
	})
	if err != nil {
		return nil, err
	}
	txs = staking.DedupeTxs(txs)
	merkleTree, err := staking.BuildTree(txs)
	if err != nil {
		return nil, err
	}

	var mirroringData []MirroringResponse
	for _, txData := range txs {
		if txData.ID != tx.ID {
			continue
		}
		stakeData, err := staking.ToStakeData(&txData)
		if err != nil {
			return nil, err
		}
		merkleProof, err := staking.GetMerkleProof(merkleTree, &txData)
		if err != nil {
			return nil, err
		}
		txDataBytes, err := createMirrorTransactionBytes(stakeData, merkleProof)
		if err != nil {
			return nil, err
		}
		merkleProofStrings := make([]string, len(merkleProof))
		for i, proof := range merkleProof {
			merkleProofStrings[i] = globalUtils.BytesToHexString(proof[:])
		}
		mirroringData = append(mirroringData, MirroringResponse{
			StakeData: MirroringStakeData{
				TxID:         globalUtils.BytesToHexString(stakeData.TxId[:]),
				StakingType:  stakeData.StakingType,
				InputAddress: globalUtils.BytesToHexString(stakeData.InputAddress[:]),
				NodeId:       globalUtils.BytesToHexString(stakeData.NodeId[:]),
				StartTime:    stakeData.StartTime,
				EndTime:      stakeData.EndTime,
				Weight:       stakeData.Weight,
			},
			MerkleProof: merkleProofStrings,
			TxInput:     globalUtils.BytesToHexString(txDataBytes),
		})
	}
	if len(mirroringData) == 0 {
		return nil, fmt.Errorf("no mirroring data found for tx %s", tx.ID)
	}
	return mirroringData, nil
}

func createMirrorTransactionBytes(stakeData *mirroring.IPChainStakeMirrorVerifierPChainStake, merkleProof [][32]byte) ([]byte, error) {
	abi, err := mirroring.MirroringMetaData.GetAbi()
	if err != nil {
		return nil, err
	}
	packed, err := abi.Pack("mirrorStake", *stakeData, merkleProof)
	if err != nil {
		return nil, err
	}
	return packed, nil
}
