package ethapi

import (
	"errors"

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
const CONTRACT_PREFIX = "SS"

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
