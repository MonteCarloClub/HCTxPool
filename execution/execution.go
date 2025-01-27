package execution

import (
	"math/big"

	"github.com/MonteCarloClub/KBD/model/state"
	vm2 "github.com/MonteCarloClub/KBD/model/vm"

	. "github.com/MonteCarloClub/KBD/block_error"
	"github.com/MonteCarloClub/KBD/common"
	"github.com/MonteCarloClub/KBD/crypto"
	"github.com/MonteCarloClub/KBD/params"
)

type Execution struct {
	env     vm2.Environment
	address *common.Address
	input   []byte
	evm     vm2.VirtualMachine

	Gas, price, value *big.Int
}

func NewExecution(env vm2.Environment, address *common.Address, input []byte, gas, gasPrice, value *big.Int) *Execution {
	exe := &Execution{env: env, address: address, input: input, Gas: gas, price: gasPrice, value: value}
	exe.evm = vm2.NewVm(env)
	return exe
}

func (self *Execution) Call(codeAddr common.Address, caller vm2.ContextRef) ([]byte, error) {
	// Retrieve the executing code
	code := self.env.State().GetCode(codeAddr)

	return self.exec(&codeAddr, code, caller)
}

func (self *Execution) Create(caller vm2.ContextRef) (ret []byte, err error, account *state.StateObject) {
	// Input must be nil for create
	code := self.input
	self.input = nil
	ret, err = self.exec(nil, code, caller)
	// Here we get an error if we run into maximum stack depth,
	// See: https://github.com/ethereum/yellowpaper/pull/131
	// and YP definitions for CREATE instruction
	if err != nil {
		return nil, err, nil
	}
	account = self.env.State().GetStateObject(*self.address)
	return
}

func (self *Execution) exec(contextAddr *common.Address, code []byte, caller vm2.ContextRef) (ret []byte, err error) {
	env := self.env
	evm := self.evm
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(self.Gas, self.price)

		return nil, vm2.DepthError
	}

	vsnapshot := env.State().Copy()
	var createAccount bool
	if self.address == nil {
		// Generate a new address
		nonce := env.State().GetNonce(caller.Address())
		env.State().SetNonce(caller.Address(), nonce+1)

		addr := crypto.CreateAddress(caller.Address(), nonce)

		self.address = &addr
		createAccount = true
	}
	snapshot := env.State().Copy()

	var (
		from = env.State().GetStateObject(caller.Address())
		to   *state.StateObject
	)
	if createAccount {
		to = env.State().CreateAccount(*self.address)
	} else {
		to = env.State().GetOrNewStateObject(*self.address)
	}

	err = env.Transfer(from, to, self.value)
	if err != nil {
		env.State().Set(vsnapshot)

		caller.ReturnGas(self.Gas, self.price)

		return nil, ValueTransferErr("insufficient funds to transfer value. Req %v, has %v", self.value, from.Balance())
	}

	context := vm2.NewContext(caller, to, self.value, self.Gas, self.price)
	context.SetCallCode(contextAddr, code)

	ret, err = evm.Run(context, self.input)
	if err != nil {
		env.State().Set(snapshot)
	}

	return
}
