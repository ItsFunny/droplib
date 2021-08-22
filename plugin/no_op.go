/*
# -*- coding: utf-8 -*-
# @Author : joker
# @Time : 2021/7/23 3:27 下午
# @File : no_op.go
# @Description :
# @Attention :
*/
package plugin



type NoOpPacker struct {

}

func (h *NoOpPacker) Pack(bytes []byte) ([]byte, error) {
	return bytes, nil
}
