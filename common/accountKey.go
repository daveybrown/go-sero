package common

import (
	"github.com/btcsuite/btcutil/base58"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common/address"
)

type AccountKey [64]byte


func BytesToKey(b []byte) AccountKey {
	var key AccountKey;
    copy(key[:],b)
    return key;
}

func Base58ToKey(key string) AccountKey {
	out:= base58.Decode(key)
	return BytesToKey(out[:])
}

func AddressToKey(addr address.AccountAddress) AccountKey {
	return BytesToKey(addr[:])
}

func (key AccountKey) String() string{
	return base58.Encode(key[:])
}

func (key AccountKey) ToUint512() *c_type.Uint512 {
	pubKey := c_type.Uint512{}
	copy(pubKey[:], key[:])
	return &pubKey
}