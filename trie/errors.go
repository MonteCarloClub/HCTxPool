/*
Copyright (c) 2022 Zhu Zunxiong <liuzunxiong@qq.com>

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
*/

package trie

import (
	"fmt"

	"github.com/MonteCarloClub/KBD/common"
)

// MissingNodeError is returned by the trie functions (TryGet, TryUpdate, TryDelete)
// in the case where a trie node is not present in the local database. It contains
// information necessary for retrieving the missing node.
type MissingNodeError struct {
	NodeHash common.Hash // hash of the missing node
	Path     []byte      // hex-encoded path to the missing node
}

func (err *MissingNodeError) Error() string {
	return fmt.Sprintf("missing trie node %x (path %x)", err.NodeHash, err.Path)
}
