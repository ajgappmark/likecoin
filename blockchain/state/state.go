package state

import (
	"encoding/json"
	"errors"
	"math/big"
	"sync"

	"github.com/denisskin/bin"
)

type State struct {
	vals   map[string]Number        //
	sets   map[string]struct{}      //
	keys   []Key                    //
	getter func(key Key) Number     //
	setter func(Key, Number, int64) //
	mx     sync.Mutex               // ??
}

var (
	//errIncrement = errors.New("blockchain/state-error: increment error")
	ErrNegativeValue = errors.New("blockchain/state-error: not enough funds")
	ErrInvalidKey    = errors.New("blockchain/state-error: invalid key")
)

func NewState() *State {
	return NewStateEx(nil, nil)
}

func NewStateEx(
	getter func(Key) Number,
	setter func(Key, Number, int64),
) *State {
	return &State{
		getter: getter,
		setter: setter,
		vals:   map[string]Number{},
		sets:   map[string]struct{}{},
	}
}

func (s *State) NewSubState() *State {
	return NewStateEx(s.Get, s.Set)
}

func (s *State) Copy() *State {
	a := NewState()
	for _, key := range s.Keys() {
		a.Set(key, s.Get(key), 0)
	}
	return a
}

func (s *State) Keys() []Key {
	return s.keys
}

func (s *State) Get(key Key) Number {
	sKey := key.strKey()
	val, ok := s.vals[sKey]
	if ok {
		return new(big.Int).Set(val)
	}
	if s.getter != nil {
		val = s.getter(key)
	}
	if val == nil {
		val = Int(0)
	}
	s.vals[sKey] = val
	return new(big.Int).Set(val)
}

func (s *State) Set(key Key, v Number, tag int64) {
	if v.Sign() < 0 {
		s.Fail(ErrNegativeValue)
		return
	}
	sKey := key.strKey()
	s.vals[sKey] = v
	if _, ok := s.sets[sKey]; !ok {
		s.keys = append(s.keys, key)
		s.sets[sKey] = struct{}{}
	}
	if s.setter != nil {
		s.setter(key, v, tag)
	}
}

func (s *State) Increment(key Key, v Number, tag int64) {
	s.Set(key, new(big.Int).Add(s.Get(key), v), tag)
}

func (s *State) Decrement(key Key, v Number, tag int64) {
	s.Increment(key, new(big.Int).Neg(v), tag)
}

func (s *State) Equal(s1 *State) bool {
	if len(s.keys) != len(s1.keys) {
		return false
	}
	for _, key := range s.keys {
		if s.Get(key).Cmp(s1.Get(key)) != 0 {
			return false
		}
	}
	return true
}

func (s *State) Hash() []byte {
	return bin.Hash256(s)
}

func (s *State) Encode() []byte {
	w := bin.NewBuffer(nil)
	w.WriteVarInt(len(s.keys))
	for _, key := range s.keys {
		w.WriteVar(key.Asset)
		w.WriteVar(key.Address)
		w.WriteBigInt(s.Get(key))
	}
	return w.Bytes()
}

func (s *State) Decode(data []byte) error {
	s.vals = map[string]Number{}
	s.sets = map[string]struct{}{}

	r := bin.NewBuffer(data)
	var key Key
	for n, _ := r.ReadVarInt(); n > 0 && r.Error() == nil; n-- {
		r.ReadVar(&key.Asset)
		r.ReadVar(&key.Address)
		v, _ := r.ReadBigInt()
		s.Set(key, v, 0)
	}
	return r.Error()
}

type stateValue struct {
	Asset string   `json:"asset"`
	Addr  string   `json:"address"`
	Value *big.Int `json:"value"`
}

func (s *State) MarshalJSON() ([]byte, error) {
	var vv []stateValue
	for _, key := range s.keys {
		vv = append(vv, stateValue{
			Asset: key.Asset.String(),
			Addr:  key.Address.String(),
			Value: s.Get(key),
		})
	}
	return json.Marshal(vv)
}

func (s *State) Fail(err error) {
	panic(err)
}

func (s *State) Execute(fn func()) (err error) {
	s.mx.Lock()
	defer s.mx.Unlock()
	defer func() {
		err, _ = recover().(error)
	}()

	fn()

	return
}
