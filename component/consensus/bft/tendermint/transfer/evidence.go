/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 11:11 上午
# @File : evidence.go
# @Description :
# @Attention :
*/
package transfer

import "time"

// Evidence represents any provable malicious activity by a validator.
// Verification logic for each evidence is part of the evidence module.
type Evidence interface {
	// ABCI() []abci.Evidence // forms individual evidence to be sent to the application
	Bytes() []byte         // bytes which comprise the evidence
	Hash() []byte          // hash of the evidence
	Height() int64         // height of the infraction
	String() string        // string format of the evidence
	Time() time.Time       // time of the infraction
	ValidateBasic() error  // basic consistency check
}
