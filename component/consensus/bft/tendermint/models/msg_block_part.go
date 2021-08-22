/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/4/2 10:15 上午
# @File : msg_block_part.go
# @Description :
# @Attention :
*/
package models


type BlockPartMessageWrapper struct {
	Height int64
	Round  int32
	Part   *Part
}

func (b BlockPartMessageWrapper) ValidateBasic() error {
	return nil
}

