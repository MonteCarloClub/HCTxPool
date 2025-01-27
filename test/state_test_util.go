package test

import (
	"bytes"
	"fmt"
	"io"
	"math/big"
	"strconv"

	"github.com/MonteCarloClub/KBD/frame"

	"github.com/MonteCarloClub/KBD/model/kbpool"
	state2 "github.com/MonteCarloClub/KBD/model/state"
	"github.com/MonteCarloClub/KBD/model/vm"

	"github.com/MonteCarloClub/KBD/block_error"
	"github.com/MonteCarloClub/KBD/common"
	"github.com/MonteCarloClub/KBD/crypto"
	"github.com/cloudwego/kitex/pkg/klog"
)

func RunStateTestWithReader(r io.Reader, skipTests []string) error {
	tests := make(map[string]VmTest)
	if err := readJson(r, &tests); err != nil {
		return err
	}

	if err := runStateTests(tests, skipTests); err != nil {
		return err
	}

	return nil
}

func RunStateTest(p string, skipTests []string) error {
	tests := make(map[string]VmTest)
	if err := readJsonFile(p, &tests); err != nil {
		return err
	}

	if err := runStateTests(tests, skipTests); err != nil {
		return err
	}

	return nil

}

func runStateTests(tests map[string]VmTest, skipTests []string) error {
	skipTest := make(map[string]bool, len(skipTests))
	for _, name := range skipTests {
		skipTest[name] = true
	}

	for name, test := range tests {
		if skipTest[name] {
			klog.Infof("Skipping state test %v", name)
			return nil
		}

		if err := runStateTest(test); err != nil {
			return fmt.Errorf("%s: %s\n", name, err.Error())
		}

		klog.Infof("State test passed: %v", name)
		//fmt.Println(string(statedb.Dump()))
	}
	return nil

}

func runStateTest(test VmTest) error {
	frame.Init()
	stateDB := frame.GetState()
	for addr, account := range test.Pre {
		klog.Infof("runStateTest addr = %v , account = %v", addr, account)
		obj := StateObjectFromAccount(frame.GetDB(), addr, account)
		stateDB.SetStateObject(obj)
		for a, v := range account.Storage {
			obj.SetState(common.HexToHash(a), common.HexToHash(v))
		}
	}

	// XXX Yeah, yeah...
	env := make(map[string]string)
	env["currentCoinbase"] = test.Env.CurrentCoinbase
	env["currentDifficulty"] = test.Env.CurrentDifficulty
	env["currentGasLimit"] = test.Env.CurrentGasLimit
	env["currentNumber"] = test.Env.CurrentNumber
	env["previousHash"] = test.Env.PreviousHash
	if n, ok := test.Env.CurrentTimestamp.(float64); ok {
		env["currentTimestamp"] = strconv.Itoa(int(n))
	} else {
		env["currentTimestamp"] = test.Env.CurrentTimestamp.(string)
	}

	var (
		ret []byte
		// gas  *big.Int
		// err  error
	)

	ret, _, _ = RunState(stateDB, env, test.Transaction)

	// // Compare expected  and actual return
	rexp := common.FromHex(test.Out)
	if bytes.Compare(rexp, ret) != 0 {
		return fmt.Errorf("return failed. Expected %x, got %x\n", rexp, ret)
	}

	// check post state
	for addr, account := range test.Post {
		obj := stateDB.GetStateObject(common.HexToAddress(addr))
		if obj == nil {
			continue
		}
		klog.Infof("address:%v obj.Balance:%v \t account.Balance:%v", obj.Address().Hex(), obj.Balance(), common.Big(account.Balance))
		if obj.Balance().Cmp(common.Big(account.Balance)) != 0 {
			return fmt.Errorf("(%x) balance failed. Expected %v, got %v => %v\n", obj.Address().Bytes()[:4], account.Balance, obj.Balance(), new(big.Int).Sub(common.Big(account.Balance), obj.Balance()))
		}

		if obj.Nonce() != common.String2Big(account.Nonce).Uint64() {
			return fmt.Errorf("(%x) nonce failed. Expected %v, got %v\n", obj.Address().Bytes()[:4], account.Nonce, obj.Nonce())
		}

		for addr, value := range account.Storage {
			v := obj.GetState(common.HexToHash(addr))
			vexp := common.HexToHash(value)

			if v != vexp {
				return fmt.Errorf("(%x: %s) storage failed. Expected %x, got %x (%v %v)\n", obj.Address().Bytes()[0:4], addr, vexp, v, vexp.Big(), v.Big())
			}
		}
	}

	stateDB.Sync()
	if common.HexToHash(test.PostStateRoot) != stateDB.Root() {
		return fmt.Errorf("Post state root error. Expected %s, got %x", test.PostStateRoot, stateDB.Root())
	}
	frame.GetDB().Close()
	return nil
}

func RunState(statedb *state2.StateDB, env, tx map[string]string) ([]byte, *big.Int, error) {
	var (
		key   = crypto.ToECDSA(common.Hex2Bytes(tx["secretKey"]))
		data  = common.FromHex(tx["data"])
		gas   = common.Big(tx["gasLimit"])
		price = common.Big(tx["gasPrice"])
		value = common.Big(tx["value"])
		nonce = common.Big(tx["nonce"]).Uint64()
		caddr = common.HexToAddress(env["currentCoinbase"])
	)

	var to *common.Address
	if len(tx["to"]) > 2 {
		t := common.HexToAddress(tx["to"])
		to = &t
	}
	// Set pre compiled contracts
	vm.Precompiled = vm.PrecompiledContracts()

	snapshot := statedb.Copy()
	coinbase := statedb.GetOrNewStateObject(caddr)
	coinbase.SetGasLimit(common.Big(env["currentGasLimit"]))

	message := NewMessage(crypto.PubkeyToAddress(key.PublicKey), to, data, value, gas, price, nonce)
	vmenv := NewEnvFromMap(statedb, env, tx)
	vmenv.origin = crypto.PubkeyToAddress(key.PublicKey)
	ret, _, err := kbpool.ApplyMessage(vmenv, message, coinbase)
	if block_error.IsNonceErr(err) || block_error.IsInvalidTxErr(err) || state2.IsGasLimitErr(err) {
		statedb.Set(snapshot)
	}
	statedb.SyncObjects()

	return ret, vmenv.Gas, err
}
