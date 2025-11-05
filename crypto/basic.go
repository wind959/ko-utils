package crypto

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
)

// Base64StdEncode base64编码编码字符串
func Base64StdEncode(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// Base64StdDecode base64解码字符串
func Base64StdDecode(s string) string {
	b, _ := base64.StdEncoding.DecodeString(s)
	return string(b)
}

// Md5String md5加密字符串
func Md5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

// Md5StringWithBase64 md5加密字符串并返回base64编码
func Md5StringWithBase64(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// Md5Byte md5加密字节数组
func Md5Byte(data []byte) string {
	h := md5.New()
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Md5ByteWithBase64 md5加密字节数组并返回base64编码
func Md5ByteWithBase64(data []byte) string {
	h := md5.New()
	h.Write(data)
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// HmacMd5 hmac md5加密
func HmacMd5(str, key string) string {
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte("")))
}

// HmacMd5WithBase64 hmac md5加密并返回base64编码
func HmacMd5WithBase64(data, key string) string {
	h := hmac.New(md5.New, []byte(key))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum([]byte("")))
}

// HmacSha1 hmac sha1加密
func HmacSha1(str, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte("")))
}

// HmacSha1WithBase64 hmac sha1加密并返回base64编码
func HmacSha1WithBase64(str, key string) string {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(h.Sum([]byte("")))
}

// HmacSha256 hmac sha256加密
func HmacSha256(str, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte("")))
}

// HmacSha256WithBase64 hmac sha256加密并返回base64编码
func HmacSha256WithBase64(str, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(h.Sum([]byte("")))
}

// HmacSha512 hmac sha512加密
func HmacSha512(str, key string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum([]byte("")))
}

// HmacSha512WithBase64 hmac sha512加密并返回base64编码
func HmacSha512WithBase64(str, key string) string {
	h := hmac.New(sha512.New, []byte(key))
	h.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(h.Sum([]byte("")))
}

// Sha1 sha1加密
func Sha1(str string) string {
	sha1 := sha1.New()
	sha1.Write([]byte(str))
	return hex.EncodeToString(sha1.Sum([]byte("")))
}

// Sha1WithBase64 sha1加密并返回base64编码
func Sha1WithBase64(str string) string {
	s1 := sha1.New()
	s1.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(s1.Sum([]byte("")))
}

// Sha256 sha256加密
func Sha256(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	return hex.EncodeToString(hash.Sum([]byte("")))
}

// Sha256WithBase64 sha256加密并返回base64编码
func Sha256WithBase64(str string) string {
	hash := sha256.New()
	hash.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(hash.Sum([]byte("")))
}

// Sha512 sha512加密
func Sha512(str string) string {
	sha512 := sha512.New()
	sha512.Write([]byte(str))
	return hex.EncodeToString(sha512.Sum([]byte("")))
}

// Sha512WithBase64 sha512加密并返回base64编码
func Sha512WithBase64(str string) string {
	sha512 := sha512.New()
	sha512.Write([]byte(str))
	return base64.StdEncoding.EncodeToString(sha512.Sum([]byte("")))
}

// Xor 异或
func Xor(data ...interface{}) []byte {
	if len(data) == 0 {
		panic("data is nil")
	}
	// 转换为字节数组并找到最小长度
	byteSlices := make([][]byte, len(data))
	minLen := -1

	for i, pwd := range data {
		switch v := pwd.(type) {
		case string:
			byteSlices[i] = []byte(v)
		case []byte:
			byteSlices[i] = v
		default:
			return nil // 不支持的类型
		}

		if minLen == -1 || len(byteSlices[i]) < minLen {
			minLen = len(byteSlices[i])
		}
	}
	// 执行异或运算
	result := make([]byte, minLen)
	for i := 0; i < minLen; i++ {
		for _, bs := range byteSlices {
			result[i] ^= bs[i]
		}
	}
	return result
}

// XorWithHex 异或返回 hex
func XorWithHex(data ...interface{}) string {
	return hex.EncodeToString(Xor(data...))
}

// XorWithBase64 异或返回 base64
func XorWithBase64(data ...interface{}) string {
	return base64.StdEncoding.EncodeToString(Xor(data...))
}
