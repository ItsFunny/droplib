/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/3/13 8:40 下午
# @File : helper.go
# @Description :
# @Attention :
*/
package types

func (this *ConsensusParamsProto) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ConsensusParamsProto)
	if !ok {
		that2, ok := that.(ConsensusParamsProto)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if !this.Block.Equal(that1.Block) {
		return false
	}
	if !this.Evidence.Equal(that1.Evidence) {
		return false
	}
	if !this.Validator.Equal(that1.Validator) {
		return false
	}
	// if !this.Version.Equal(that1.Version) {
	// 	return false
	// }
	return true
}
func (this *BlockParamsProto) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*BlockParamsProto)
	if !ok {
		that2, ok := that.(BlockParamsProto)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.MaxBytes != that1.MaxBytes {
		return false
	}
	if this.MaxGas != that1.MaxGas {
		return false
	}
	return true
}
func (this *EvidenceParamsProto) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*EvidenceParamsProto)
	if !ok {
		that2, ok := that.(EvidenceParamsProto)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.MaxAgeNumBlocks != that1.MaxAgeNumBlocks {
		return false
	}
	if this.MaxAgeDuration != that1.MaxAgeDuration {
		return false
	}
	if this.MaxBytes != that1.MaxBytes {
		return false
	}
	return true
}
func (this *ValidatorParamsProto) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*ValidatorParamsProto)
	if !ok {
		that2, ok := that.(ValidatorParamsProto)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if len(this.PubKeyTypes) != len(that1.PubKeyTypes) {
		return false
	}
	for i := range this.PubKeyTypes {
		if this.PubKeyTypes[i] != that1.PubKeyTypes[i] {
			return false
		}
	}
	return true
}
func (this *HashedParamsProto) Equal(that interface{}) bool {
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*HashedParamsProto)
	if !ok {
		that2, ok := that.(HashedParamsProto)
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}
	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	if this.BlockMaxBytes != that1.BlockMaxBytes {
		return false
	}
	if this.BlockMaxGas != that1.BlockMaxGas {
		return false
	}
	return true
}
