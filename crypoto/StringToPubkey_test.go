package crypoto

import (
	"testing"
)

func TestExtractPublicKeyFromHex(t *testing.T) {
	pair, err := MakeKeyPair()
	if err != nil {
		t.Fatalf("生成密钥对失败: %v", err)
		return
	}
	pubKeyStr := GetPubKeyStr(&pair.PublicKey)
	t.Logf("公钥:\n %s", pubKeyStr)
	fromHex, err := ExtractPublicKeyFromHex(pubKeyStr)
	if err != nil {
		t.Fatalf("提取公钥失败: %v", err)
		return
	}
	if pair.PublicKey.X.Cmp(fromHex.X) != 0 || pair.PublicKey.Y.Cmp(fromHex.Y) != 0 {
		t.Fatalf("公钥不一致")
		return
	}
	keyStr2 := GetPubKeyStr(fromHex)
	if keyStr2 != pubKeyStr {
		t.Fatalf("公钥不一致")
		return
	}
	t.Logf("公钥2:\n %s", keyStr2)
}

func TestExtractPublicKeyFromHex2(t *testing.T) {
	pubKey := "045938a88208ef46f2cd01682279aab8660184a768e6d97ed1628eda3d442c35d15bcb660775fdeac0c23b1e67ae787cb7d87242029fc4d3675252a80a0f3a453a"
	hex, err := ExtractPublicKeyFromHex(pubKey)
	if err != nil {
		t.Fatalf("提取公钥失败: %v", err)
		return
	}
	t.Logf("公钥:\n %s", GetPubKeyStr(hex))
}
