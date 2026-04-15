package crypoto

import (
	"crypto/ecdh"
	"crypto/rand"
	"encoding/hex"
)

// ExtractPublicKeyFromHex 从 Hex 字符串中提取 ecdsa.PublicKey
// 对应 Node.js 端的 0x04 + X(32字节) + Y(32字节) 格式
func ExtractPublicKeyFromHex(hexStr string) (*ecdh.PublicKey, error) {
	bytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, err
	}
	// 直接从字节还原公钥，ecdh 库会自动校验格式
	return ecdh.P256().NewPublicKey(bytes)
}
func MakeKeyPair() (*ecdh.PrivateKey, error) {
	privateKey, err := ecdh.P256().GenerateKey(rand.Reader)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
func GetPubKeyStr(key *ecdh.PublicKey) string {
	return hex.EncodeToString(key.Bytes())
}
