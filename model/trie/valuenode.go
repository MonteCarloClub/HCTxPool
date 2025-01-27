package trie

import "github.com/MonteCarloClub/KBD/common"

type ValueNode struct {
	trie  *Trie
	data  []byte
	dirty bool
}

func NewValueNode(trie *Trie, data []byte) *ValueNode {
	return &ValueNode{trie, data, false}
}

func (self *ValueNode) Value() Node { return self } // Best not to call :-)
func (self *ValueNode) Val() []byte { return self.data }
func (self *ValueNode) Dirty() bool { return self.dirty }
func (self *ValueNode) Copy(t *Trie) Node {
	return &ValueNode{t, common.CopyBytes(self.data), self.dirty}
}
func (self *ValueNode) RlpData() interface{} { return self.data }
func (self *ValueNode) Hash() interface{}    { return self.data }

func (self *ValueNode) setDirty(dirty bool) {
	self.dirty = dirty
}
