package manager

import (
	"github.com/MonteCarloClub/KBD/chain_manager"
	"github.com/MonteCarloClub/KBD/common"
	"github.com/MonteCarloClub/KBD/model/accounts"
	"github.com/MonteCarloClub/KBD/model/event"
	"github.com/MonteCarloClub/KBD/model/kbpool"
	"github.com/MonteCarloClub/KBD/types"
)

type Backend interface {
	AccountManager() *accounts.Manager
	BlockProcessor() *types.BlockProcessor
	ChainManager() *chain_manager.ChainManager
	TxPool() *kbpool.TxPool
	BlockDb() common.Database
	StateDb() common.Database
	EventMux() *event.TypeMux
}
