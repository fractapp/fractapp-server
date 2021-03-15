package adaptors

import "fractapp-server/types"

type Adaptor interface {
	Connect() error
	Subscribe() error
	Unsubscribe()
	WaitNewBlock() (uint64, error)
	Err() <-chan error
	Transfers(blockNumber uint64) ([]types.Tx, error)
}
