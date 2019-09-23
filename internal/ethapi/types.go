package ethapi

import (
	"errors"
	"fmt"

	"github.com/sero-cash/go-sero/accounts"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/seroparam"
	"github.com/sero-cash/go-czero-import/superzk"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/utils"
)

const PK_PREFIX = "SP"
const TK_PREFIX = "ST"
const PKR_PREFIX = "SC"
const CONTECT_PREFIX = "SS"

type PKAddress [64]byte

func (b PKAddress) ToUint512() c_type.Uint512 {
	result := c_type.Uint512{}
	copy(result[:], b[:])

	return result
}

func (b PKAddress) ToPkr() c_type.PKr {
	pk := c_type.Uint512{}
	copy(pk[:], b[:])
	return superzk.Pk2PKr(&pk, nil)
}

func (b PKAddress) String() string {
	return string(toAddressString(PK_PREFIX, b[:]))
}

func (b PKAddress) MarshalText() ([]byte, error) {
	return toAddressString(PK_PREFIX, b[:]), nil
}

func toAddressString(protocol string, b []byte) []byte {
	currentBlockNumber := txtool.Ref_inst.Bc.GetCurrenHeader().Number.Uint64()
	if currentBlockNumber > seroparam.SIP5() {
		addr := utils.NewAddressByBytes(b[:])
		addr.SetProtocol(protocol)
		return []byte(addr.ToCode())
	} else {
		return []byte(base58.Encode(b[:]))
	}
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		err := validPk(addr)
		if err != nil {
			return err
		}
		copy(b[:], addr.Bytes)
		return nil
	}
}

type TKAddress [64]byte

func (b TKAddress) ToTk() c_type.Tk {
	result := c_type.Tk{}
	copy(result[:], b[:])

	return result
}

func (b TKAddress) ToAccounAddress() address.AccountAddress {
	var addr address.AccountAddress
	copy(addr[:], b[:])
	return addr
}

func (b TKAddress) MarshalText() ([]byte, error) {
	return toAddressString(TK_PREFIX, b[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *TKAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if !addr.MatchProtocol("ST") {
			return errors.New("address protocol is not tk")
		}
		if len(addr.Bytes) == 64 {
			copy(b[:], addr.Bytes)
		} else {
			return errors.New("ivalid TK")
		}
		return nil
	}
}

type PKrAddress [96]byte

func (b PKrAddress) ToPKr() *c_type.PKr {
	result := &c_type.PKr{}
	copy(result[:], b[:])
	return result
}

func (b PKrAddress) MarshalText() ([]byte, error) {
	return toAddressString(PKR_PREFIX, b[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKrAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		err := validPkr(addr)
		if err != nil {
			return err
		}
		copy(b[:], addr.Bytes)
		return nil
	}
}

type FromAddress []byte

func (b FromAddress) ToPKr() (ret c_type.PKr) {
	if b.IsPkr() {
		copy(ret[:], b[:])
	} else {
		pk := c_type.Uint512{}
		copy(pk[:], b[:])
		return superzk.Pk2PKr(&pk, nil)
	}
	return
}

func (b FromAddress) IsPkr() bool {
	return len(b) == 96
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *FromAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return ErrEmptyString
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		isConctract, err := isContractAddress(addr.Bytes)
		if err != nil {
			return err
		}
		if isConctract {
			return errors.New("from address can not be conract Address")
		} else {

			if len(addr.Bytes) == 96 {
				err := validPkr(addr)
				if err != nil {
					return err
				}
				*b = addr.Bytes
				return nil

			} else if len(addr.Bytes) == 64 {
				err := validPk(addr)
				if err != nil {
					return err
				}
				*b = addr.Bytes
				return nil
			} else {
				return errors.New("from Address must be length 64 or 96")
			}
		}
	}
}

type ToAddress []byte

func (b ToAddress) ToPKr() (ret c_type.PKr) {
	copy(ret[:], b[:])
	return
}

func (b ToAddress) IsPkr() bool {
	return len(b) == 96
}

func (b ToAddress) ToAddress() (ret common.Address) {
	copy(ret[:], b[:])
	return
}

func (b ToAddress) toPk() (ret c_type.Uint512) {
	copy(ret[:], b[:])
	return
}

func (b ToAddress) String() string {
	return string(toAddressString("SC", b[:]))
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *ToAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return ErrEmptyString
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		isConctract, err := isContractAddress(addr.Bytes)
		if err != nil {
			return err
		}
		if isConctract {
			*b = addr.Bytes
			return nil
		} else {

			if len(addr.Bytes) == 96 {
				err := validPkr(addr)
				if err != nil {
					return err
				}
				*b = addr.Bytes
				return nil

			} else if len(addr.Bytes) == 64 {
				err := validPk(addr)
				if err != nil {
					return err
				}
				pk := c_type.Uint512{}
				copy(pk[:], addr.Bytes)
				pkr := superzk.Pk2PKr(&pk, nil)
				*b = pkr[:]
				return nil
			} else {
				return errors.New("AllMixedAddress must be length 64 or 96")
			}
		}
	}
}
func validPkr(addr utils.Address) (e error) {
	currentBlockNumber := txtool.Ref_inst.Bc.GetCurrenHeader().Number.Uint64()
	if currentBlockNumber > seroparam.SIP6() {
		if addr.Version == "0" {
			return errors.New(fmt.Sprintf("after %d block must be new pkr address", seroparam.SIP6()))
		}
	}
	if !addr.MatchProtocol("SP") {
		return errors.New("address protocol is not pk")
	}
	if len(addr.Bytes) == 96 {
		pkr := c_type.PKr{}
		copy(pkr[:], addr.Bytes)
		if !superzk.IsPKrValid(&pkr) {
			e = errors.New("ivalid Pkr")
		}
	} else {
		e = errors.New("pkr address must be 96 bytes")
	}
	return
}

func validPk(addr utils.Address) (e error) {
	currentBlockNumber := txtool.Ref_inst.Bc.GetCurrenHeader().Number.Uint64()
	if currentBlockNumber > seroparam.SIP6() {
		if addr.Version == "0" {
			return errors.New(fmt.Sprintf("after %d block must be new pk address", seroparam.SIP6()))
		}
	}
	if !addr.MatchProtocol("SP") {
		return errors.New("address protocol is not pk")
	}

	if len(addr.Bytes) == 64 {
		pk := c_type.Uint512{}
		copy(pk[:], addr.Bytes)
		if !superzk.IsPKValid(&pk) {
			e = errors.New("ivalid Pk")
		}
	} else {
		e = errors.New("pk address must be 64 bytes")
	}
	return
}

func isContractAddress(b []byte) (bool, error) {
	var addr common.Address
	copy(addr[:], b)
	return txtool.Ref_inst.Bc.IsContract(addr)
}

func AccountAddressToPkAddress(addr address.AccountAddress) PKAddress {
	var pk PKAddress
	copy(pk[:], addr[:])
	return pk
}

func AddressToPkrAddress(addr common.Address) PKrAddress {
	var pkr PKrAddress
	copy(pkr[:], addr[:])
	return pkr
}

func AccountAddressToTkAddress(addr address.AccountAddress) TKAddress {
	var tk TKAddress
	copy(tk[:], addr[:])
	return tk
}

func TkToPkAddress(tk address.AccountAddress) PKAddress {
	c_tk := tk.ToTK()
	var c_pk c_type.Uint512
	height := txtool.Ref_inst.Bc.GetCurrenHeader().Number.Uint64()
	if height >= seroparam.SIP5() {
		c_pk = c_superzk.Tk2Pk(c_tk)
	} else {
		c_pk = c_czero.Tk2Pk(c_tk)
	}
	var pk PKAddress
	copy(pk[:], c_pk[:])
	return pk
}

func GetFromPK(account accounts.Account) c_type.Uint512 {
	addr := account.GetPKByHeight()
	return *addr.ToUint512()
}
