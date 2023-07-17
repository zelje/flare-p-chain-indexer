package cronjob

import (
	"context"

	"github.com/ava-labs/avalanchego/vms/platformvm/api"
	"github.com/ybbus/jsonrpc/v3"
)

type PermissionedValidators struct {
	Validators []*api.PermissionedValidator
}

func CallPChainGetConnectedValidators(client jsonrpc.RPCClient) ([]*api.PermissionedValidator, error) {
	ctx := context.Background()
	response, err := client.Call(ctx, "platform.getCurrentValidators")
	if err != nil {
		return nil, err
	}

	reply := PermissionedValidators{}
	err = response.GetObject(&reply)

	return reply.Validators, err
}
