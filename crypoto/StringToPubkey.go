package crypoto

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// ExtractPublicKeyFromHex 从 Hex 字符串中提取 ecdsa.PublicKey
// 对应 Node.js 端的 0x04 + X(32字节) + Y(32字节) 格式
func ExtractPublicKeyFromHex(hexStr string) (*ecdsa.PublicKey, error) {
	// 1. 将 Hex 字符串解码为字节
	pubBytes, err := hex.DecodeString(hexStr)
	if err != nil {
		return nil, fmt.Errorf("hex 解码失败: %v", err)
	}

	// 2. 校验长度
	// P-256 曲线的未压缩公钥长度应为 65 字节 (1字节前缀 + 32字节X + 32字节Y)
	if len(pubBytes) != 65 {
		return nil, fmt.Errorf("公钥长度错误: 期望 65 字节, 实际 %d 字节", len(pubBytes))
	}

	// 3. 校验前缀
	if pubBytes[0] != 0x04 {
		return nil, fmt.Errorf("公钥前缀错误: 期望 0x04 (未压缩格式)")
	}

	// 4. 解析公钥
	// elliptic.Unmarshal 会自动处理 0x04 前缀，并提取 X 和 Y 坐标
	// 注意：必须指定与 Node.js 端相同的曲线，这里是 P-256
	x, y := elliptic.Unmarshal(elliptic.P256(), pubBytes)

	if x == nil {
		return nil, fmt.Errorf("解析公钥失败：坐标无效")
	}

	// 5. 构建 ecdsa.PublicKey 对象
	publicKey := &ecdsa.PublicKey{
		Curve: elliptic.P256(),
		X:     x,
		Y:     y,
	}

	return publicKey, nil
}
func MakeKeyPair() (*ecdsa.PrivateKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	return privateKey, nil
}
func GetPubKeyStr(key *ecdsa.PublicKey) string {
	publicKeyBytes := make([]byte, 65)
	publicKeyBytes[0] = 0x04
	xBytes := key.X.Bytes()
	yBytes := key.Y.Bytes()
	copy(publicKeyBytes[1+32-len(xBytes):1+32], xBytes)
	copy(publicKeyBytes[1+32+32-len(yBytes):1+32+32], yBytes)
	return hex.EncodeToString(publicKeyBytes)
}
