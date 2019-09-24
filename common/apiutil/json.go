package apiutil

import (
	"fmt"
	"math/big"

	"github.com/sero-cash/go-sero/common"

	"github.com/sero-cash/go-sero/common/address"

	"github.com/btcsuite/btcutil/base58"

	"github.com/sero-cash/go-sero/accounts"
	"github.com/sero-cash/go-sero/zero/txtool"
	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-czero-import/c_czero"
	"github.com/sero-cash/go-czero-import/c_superzk"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/pkg/errors"

	"github.com/sero-cash/go-czero-import/seroparam"
)

type decError struct{ msg string }

func (err decError) Error() string { return err.msg }

var (
	ErrEmptyString = &decError{"empty input string"}
	ErrSyntax      = &decError{"invalid hex string"}
)

const PK_PREFIX = "SP"
const TK_PREFIX = "ST"
const PKR_PREFIX = "SC"
const CONTRACT_PREFIX = "SS"

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
	return string(ToAddressString(PK_PREFIX, b[:]))
}

func (b PKAddress) MarshalText() ([]byte, error) {
	return ToAddressString(PK_PREFIX, b[:]), nil
}

func ValidPk(addr utils.Address) (e error) {
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

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		err := ValidPkr(addr)
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

func (b TKAddress) MarshalText() ([]byte, error) {
	return ToAddressString(TK_PREFIX, b[:]), nil
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

func ValidPkr(addr utils.Address) (e error) {
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

func (b PKrAddress) ToBase58() string {
	return base58.Encode(b[:])
}

func (b PKrAddress) ToPKr() *c_type.PKr {
	result := &c_type.PKr{}
	copy(result[:], b[:])

	return result
}

func (b PKrAddress) MarshalText() ([]byte, error) {
	return ToAddressString(PKR_PREFIX, b[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *PKrAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return nil
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		err := ValidPkr(addr)
		if err != nil {
			return err
		}
		copy(b[:], addr.Bytes)
		return nil
	}
}

func ToAddressString(protocol string, b []byte) []byte {
	currentBlockNumber := txtool.Ref_inst.Bc.GetCurrenHeader().Number.Uint64()
	if currentBlockNumber > seroparam.SIP5() {
		addr := utils.NewAddressByBytes(b[:])
		addr.SetProtocol(protocol)
		return []byte(addr.ToCode())
	} else {
		return []byte(base58.Encode(b[:]))
	}
}

func isString(input []byte) bool {
	return len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"'
}

type Big big.Int

func (b Big) MarshalJSON() ([]byte, error) {
	i := big.Int(b)
	by, err := i.MarshalJSON()
	if err != nil {
		return nil, err
	}
	if seroparam.IsExchangeValueStr() {
		bytes := make([]byte, len(by)+2)
		copy(bytes[1:len(bytes)-1], by[:])
		bytes[0] = '"'
		bytes[len(bytes)-1] = '"'
		return bytes, nil
	} else {
		return by, err
	}
}

// UnmarshalJSON implements json.Unmarshaler.
func (b *Big) UnmarshalJSON(input []byte) error {
	if isString(input) {
		input = input[1 : len(input)-1]
	}
	i := big.Int{}
	if e := i.UnmarshalText(input); e != nil {
		return e
	} else {
		*b = Big(i)
		return nil
	}
}

func (b *Big) ToInt() *big.Int {
	return (*big.Int)(b)
}

type ToAddress [96]byte

func (b ToAddress) ToPKr() (ret c_type.PKr) {
	copy(ret[:], b[:])
	return
}

func (b ToAddress) IsConract() bool {
	flag, _ := isContractAddress(b[:])
	return flag
}

func (b ToAddress) String() string {
	if b.IsConract() {
		return string(ToAddressString(CONTRACT_PREFIX, b[:]))
	} else {
		return string(ToAddressString(PKR_PREFIX, b[:]))
	}
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *ToAddress) UnmarshalText(input []byte) error {
	if len(input) == 0 {
		return errors.New("empty string")
	}
	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		isConctract, err := isContractAddress(addr.Bytes)
		if err != nil {
			return err
		}
		if isConctract {
			copy(b[:], addr.Bytes)
			return nil
		} else {

			if len(addr.Bytes) == 96 {
				err := ValidPkr(addr)
				if err != nil {
					return err
				}
				copy(b[:], addr.Bytes)
				return nil

			} else if len(addr.Bytes) == 64 {
				err := ValidPk(addr)
				if err != nil {
					return err
				}
				pk := c_type.Uint512{}
				copy(pk[:], addr.Bytes)
				pkr := superzk.Pk2PKr(&pk, nil)
				copy(b[:], pkr[:])
				return nil
			} else {
				return errors.New("ToAddress must be length 64 or 96")
			}
		}
	}
}

func isContractAddress(b []byte) (bool, error) {
	var addr common.Address
	copy(addr[:], b)
	return txtool.Ref_inst.Bc.IsContract(addr)
}

type ContractAddress c_type.PKr

func (b ContractAddress) MarshalText() ([]byte, error) {
	return ToAddressString(CONTRACT_PREFIX, b[:]), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *ContractAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return errors.New("empty string")
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if !addr.MatchProtocol("SS") {
			return errors.New("address protocol is not contract")
		}
		out := addr.Bytes
		if len(out) == 96 {
			addr := common.Address{}
			copy(addr[:], out)
			if isContract, err := txtool.Ref_inst.Bc.IsContract(addr); err != nil {
				return err
			} else {
				if isContract {
					copy(b[:], out)
					return nil
				} else {
					return errors.New("this 96 bytes not contract address")
				}
			}
		} else if len(out) == 64 {
			contract_addr := common.Address{}
			copy(contract_addr[:], out)
			if isContract, err := txtool.Ref_inst.Bc.IsContract(contract_addr); err != nil {
				return err
			} else {
				if isContract {
					copy(b[:], contract_addr[:])
					return nil
				} else {
					return errors.New("this 64 bytes not contract address")
				}
			}
		} else {
			return errors.New("ContractAddress must be length 64 or 96")
		}
	}

}

type MixAddress []byte

func (b MixAddress) ToPkr() c_type.PKr {
	pkr := c_type.PKr{}
	if len(b) == 64 {
		pk := c_type.Uint512{}
		copy(pk[:], b[:])
		pkr = superzk.Pk2PKr(&pk, nil)
	} else {
		copy(pkr[:], b[:])
	}
	return pkr
}

func (b MixAddress) IsPkr() bool {
	return len(b) == 96
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *MixAddress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return nil
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if len(addr.Bytes) == 96 {
			err := ValidPkr(addr)
			if err != nil {
				return err
			}
			*b = addr.Bytes
			return nil
		} else if len(addr.Bytes) == 64 {
			err := ValidPkr(addr)
			if err != nil {
				return err
			}
			*b = addr.Bytes
			return nil
		} else {
			return errors.New("invalid mix address")
		}
	}
}

type MixBase58Adrress []byte

func (b MixBase58Adrress) ToPkr() c_type.PKr {
	pkr := c_type.PKr{}
	if len(b) == 64 {
		pk := c_type.Uint512{}
		copy(pk[:], b[:])
		pkr = superzk.Pk2PKr(&pk, nil)
	} else {
		copy(pkr[:], b[:])
	}
	return pkr
}

func (b MixBase58Adrress) IsPkr() bool {
	return len(b) == 96
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (b *MixBase58Adrress) UnmarshalText(input []byte) error {

	if len(input) == 0 {
		return ErrEmptyString
	}

	if addr, e := utils.NewAddressByString(string(input)); e != nil {
		return e
	} else {
		if addr.IsHex {
			return errors.New("is not base58 address")
		}
		if len(addr.Bytes) == 96 {
			err := ValidPkr(addr)
			if err != nil {
				return err
			}
			*b = addr.Bytes
			return nil
		} else if len(addr.Bytes) == 64 {
			err := ValidPkr(addr)
			if err != nil {
				return err
			}
			*b = addr.Bytes
			return nil
		} else {
			return errors.New("invalid mix address")
		}
	}
}

func TkToPkAddress(c_tk *c_type.Tk) PKAddress {
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

func AccountAddressToTkAddress(addr address.AccountAddress) TKAddress {
	var tk TKAddress
	copy(tk[:], addr[:])
	return tk
}

func GetFromPK(account accounts.Account) c_type.Uint512 {
	addr := account.GetPKByHeight()
	return addr
}
