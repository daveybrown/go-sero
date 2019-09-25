package prepare

import (
	"bytes"

	"github.com/sero-cash/go-czero-import/c_superzk"

	"github.com/sero-cash/go-czero-import/superzk"

	"github.com/sero-cash/go-sero/zero/utils"

	"github.com/sero-cash/go-sero/zero/txtool"

	"github.com/pkg/errors"
	"github.com/sero-cash/go-czero-import/c_type"
	"github.com/sero-cash/go-sero/common"
)

func GenTxParam(param *PreTxParam, gen TxParamGenerator, state TxParamState) (txParam *txtool.GTxParam, e error) {
	if len(param.Receptions) > 500 {
		return nil, errors.New("receptions count must <= 500")
	}
	utxos, err := SelectUtxos(param, gen)
	if err != nil {
		return nil, err
	}

	if param.RefundTo == nil {
		if av, err := param.IsSzk(); err != nil {
			e = err
			return
		} else {
			if param.RefundTo = gen.DefaultRefundTo(&param.From, av); param.RefundTo == nil {
				return nil, errors.New("can not find default refund to")
			}
		}
	} else {
		if av, err := param.IsSzk(); err != nil {
			if c_superzk.IsSzkPKr(param.RefundTo) {
				if av != AV_SUPERZK {
					return nil, errors.New("refundto must be the same version with recv address")
				}
			} else {
				if av != AV_CZERO {
					return nil, errors.New("refundto must be the same version with recv address")
				}
			}
		}
	}

	bparam := BeforeTxParam{
		Fee:        param.Fee,
		GasPrice:   *param.GasPrice,
		Utxos:      utxos,
		RefundTo:   *param.RefundTo,
		Receptions: param.Receptions,
		Cmds:       param.Cmds,
	}
	txParam, e = BuildTxParam(state, &bparam)
	return
}

func IsPk(addr c_type.PKr) bool {
	byte32 := common.Hash{}
	return bytes.Equal(byte32[:], addr[64:96])
}

func CreatePkr(pk *c_type.Uint512, index uint64) c_type.PKr {
	r := c_type.Uint256{}
	copy(r[:], common.LeftPadBytes(utils.EncodeNumber(index), 32))
	if index == 0 {
		return superzk.Pk2PKr(pk, nil)
	} else {
		return superzk.Pk2PKr(pk, &r)
	}
}
