package crypto

import (
	"bytes"
	"crypto"
	"crypto/aes"
	"crypto/cipher"
	"crypto/des"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"errors"
	"github.com/tjfoc/gmsm/sm2"
	"github.com/tjfoc/gmsm/sm3"
	"github.com/tjfoc/gmsm/sm4"
	"io"
	"os"
)

// AesEcbEncrypt aes ecb 加密
func AesEcbEncrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	blockSize := aes.BlockSize
	dataLen := len(data)
	padding := blockSize - (dataLen % blockSize)
	paddedLen := dataLen + padding

	paddedData := make([]byte, paddedLen)
	copy(paddedData, data)

	for i := dataLen; i < paddedLen; i++ {
		paddedData[i] = byte(padding)
	}

	cipher, err := aes.NewCipher(generateAesKey(key, len(key)))
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	encrypted := make([]byte, paddedLen)
	for bs := 0; bs < paddedLen; bs += blockSize {
		cipher.Encrypt(encrypted[bs:], paddedData[bs:])
	}

	return encrypted
}

// AesEcbDecrypt aes ecb 解密
func AesEcbDecrypt(encrypted, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	blockSize := aes.BlockSize
	if len(encrypted)%blockSize != 0 {
		panic("aes: encrypted data length is not a multiple of block size")
	}

	cipher, err := aes.NewCipher(generateAesKey(key, len(key)))
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i += blockSize {
		cipher.Decrypt(decrypted[i:], encrypted[i:])
	}

	if len(decrypted) == 0 {
		return nil
	}
	padding := int(decrypted[len(decrypted)-1])
	if padding == 0 || padding > blockSize {
		panic("aes: invalid PKCS#7 padding")
	}
	for i := len(decrypted) - padding; i < len(decrypted); i++ {
		if decrypted[i] != byte(padding) {
			panic("aes: invalid PKCS#7 padding content")
		}
	}

	return decrypted[:len(decrypted)-padding]
}

// AesCbcEncrypt aes cbc 加密
func AesCbcEncrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	padding := aes.BlockSize - len(data)%aes.BlockSize
	padded := append(data, bytes.Repeat([]byte{byte(padding)}, padding)...)

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("aes: failed to generate IV: " + err.Error())
	}

	encrypted := make([]byte, len(padded))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted, padded)

	return append(iv, encrypted...)
}

// AesCbcDecrypt aes cbc 解密
func AesCbcDecrypt(encrypted, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	if len(encrypted) < aes.BlockSize {
		panic("aes: ciphertext too short")
	}

	if len(encrypted)%aes.BlockSize != 0 {
		panic("aes: ciphertext is not a multiple of the block size")
	}

	iv := encrypted[:aes.BlockSize]
	ciphertext := encrypted[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	decrypted := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(decrypted, ciphertext)

	return pkcs7UnPadding(decrypted)
}

// AesCtrCrypt AES CTR算法模式加密
func AesCtrCrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	block, _ := aes.NewCipher(key)

	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)

	dst := make([]byte, len(data))
	stream.XORKeyStream(dst, data)

	return dst
}

// AesCtrEncrypt AES CTR算法模式加密
func AesCtrEncrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("aes: failed to generate IV: " + err.Error())
	}

	stream := cipher.NewCTR(block, iv)
	ciphertext := make([]byte, len(data))
	stream.XORKeyStream(ciphertext, data)

	return append(iv, ciphertext...)
}

// AesCtrDecrypt AES CTR算法模式解密
func AesCtrDecrypt(encrypted, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}
	if len(encrypted) < aes.BlockSize {
		panic("aes: invalid ciphertext length")
	}

	iv := encrypted[:aes.BlockSize]
	ciphertext := encrypted[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	stream := cipher.NewCTR(block, iv)
	plaintext := make([]byte, len(ciphertext))
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext
}

// AesCfbEncrypt AES CFB模式加密
func AesCfbEncrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("aes: failed to generate IV: " + err.Error())
	}

	ciphertext := make([]byte, len(data))
	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(ciphertext, data)

	return append(iv, ciphertext...)
}

// AesCfbDecrypt AES CFB模式解密
func AesCfbDecrypt(encrypted, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	if len(encrypted) < aes.BlockSize {
		panic("aes: encrypted data too short")
	}

	iv := encrypted[:aes.BlockSize]
	ciphertext := encrypted[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext
}

// AesOfbEncrypt AES OFB模式加密
func AesOfbEncrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("aes: failed to generate IV: " + err.Error())
	}

	ciphertext := make([]byte, len(data))
	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(ciphertext, data)

	return append(iv, ciphertext...)
}

// AesOfbDecrypt AES OFB模式解密
func AesOfbDecrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	if len(data) < aes.BlockSize {
		panic("aes: encrypted data too short")
	}

	iv := data[:aes.BlockSize]
	ciphertext := data[aes.BlockSize:]

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	plaintext := make([]byte, len(ciphertext))
	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(plaintext, ciphertext)

	return plaintext
}

// AesGcmEncrypt AES GCM模式加密
func AesGcmEncrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("aes: failed to create GCM: " + err.Error())
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		panic("aes: failed to generate nonce: " + err.Error())
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	return append(nonce, ciphertext...)
}

// AesGcmDecrypt AES GCM模式解密
func AesGcmDecrypt(data, key []byte) []byte {
	if !isAesKeyLengthValid(len(key)) {
		panic("aes: invalid key length (must be 16, 24, or 32 bytes)")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		panic("aes: failed to create cipher: " + err.Error())
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		panic("aes: failed to create GCM: " + err.Error())
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		panic("aes: ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic("aes: decryption failed: " + err.Error())
	}

	return plaintext
}

// DesEcbEncrypt DES ECB模式加密
func DesEcbEncrypt(data, key []byte) []byte {
	cipher, err := des.NewCipher(generateDesKey(key))
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	blockSize := cipher.BlockSize()
	padded := pkcs5Padding(data, blockSize)
	encrypted := make([]byte, len(padded))

	for i := 0; i < len(padded); i += blockSize {
		cipher.Encrypt(encrypted[i:], padded[i:])
	}

	return encrypted
}

// DesEcbDecrypt DES ECB模式解密
func DesEcbDecrypt(encrypted, key []byte) []byte {
	cipher, err := des.NewCipher(generateDesKey(key))
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	blockSize := cipher.BlockSize()
	if len(encrypted)%blockSize != 0 {
		panic("des: invalid encrypted data length")
	}

	decrypted := make([]byte, len(encrypted))
	for i := 0; i < len(encrypted); i += blockSize {
		cipher.Decrypt(decrypted[i:], encrypted[i:])
	}

	// Remove padding
	return pkcs5UnPadding(decrypted)
}

// DesCbcEncrypt DES CBC模式加密
func DesCbcEncrypt(data, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	blockSize := block.BlockSize()
	data = pkcs7Padding(data, blockSize)

	encrypted := make([]byte, blockSize+len(data))
	iv := encrypted[:blockSize]

	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("des: failed to generate IV: " + err.Error())
	}

	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(encrypted[blockSize:], data)

	return encrypted
}

// DesCbcDecrypt DES CBC模式解密
func DesCbcDecrypt(encrypted, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	blockSize := block.BlockSize()
	if len(encrypted) < blockSize || len(encrypted)%blockSize != 0 {
		panic("des: invalid encrypted data length")
	}

	iv := encrypted[:blockSize]
	ciphertext := encrypted[blockSize:]

	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(ciphertext, ciphertext)

	return pkcs7UnPadding(ciphertext)
}

// DesCtrCrypt DES CTR模式加密
func DesCtrCrypt(data, key []byte) []byte {
	size := len(key)
	if size != 8 {
		panic("key length shoud be 8")
	}

	block, _ := des.NewCipher(key)

	iv := bytes.Repeat([]byte("1"), block.BlockSize())
	stream := cipher.NewCTR(block, iv)

	dst := make([]byte, len(data))
	stream.XORKeyStream(dst, data)

	return dst
}

// DesCtrEncrypt DES CTR模式加密
func DesCtrEncrypt(data, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	iv := make([]byte, block.BlockSize())
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("des: failed to generate IV: " + err.Error())
	}

	stream := cipher.NewCTR(block, iv)

	encrypted := make([]byte, len(data))
	stream.XORKeyStream(encrypted, data)

	// 返回前缀包含 IV，便于解密
	return append(iv, encrypted...)
}

// DesCtrDecrypt DES CTR模式解密
func DesCtrDecrypt(encrypted, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	blockSize := block.BlockSize()
	if len(encrypted) < blockSize {
		panic("des: ciphertext too short")
	}

	iv := encrypted[:blockSize]
	ciphertext := encrypted[blockSize:]

	stream := cipher.NewCTR(block, iv)

	decrypted := make([]byte, len(ciphertext))
	stream.XORKeyStream(decrypted, ciphertext)

	return decrypted
}

// DesCfbEncrypt DES CFB模式加密
func DesCfbEncrypt(data, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	iv := make([]byte, des.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("des: failed to generate IV: " + err.Error())
	}

	encrypted := make([]byte, des.BlockSize+len(data))

	copy(encrypted[:des.BlockSize], iv)

	stream := cipher.NewCFBEncrypter(block, iv)
	stream.XORKeyStream(encrypted[des.BlockSize:], data)

	return encrypted
}

// DesCfbDecrypt DES CFB模式解密
func DesCfbDecrypt(encrypted, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	if len(encrypted) < des.BlockSize {
		panic("des: encrypted data too short")
	}

	iv := encrypted[:des.BlockSize]
	ciphertext := encrypted[des.BlockSize:]

	stream := cipher.NewCFBDecrypter(block, iv)
	stream.XORKeyStream(ciphertext, ciphertext)

	return ciphertext
}

// DesOfbEncrypt DES OFB模式加密
func DesOfbEncrypt(data, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	data = pkcs7Padding(data, des.BlockSize)

	iv := make([]byte, des.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic("des: failed to generate IV: " + err.Error())
	}

	encrypted := make([]byte, des.BlockSize+len(data))
	copy(encrypted[:des.BlockSize], iv)

	stream := cipher.NewOFB(block, iv)
	stream.XORKeyStream(encrypted[des.BlockSize:], data)

	return encrypted
}

// DesOfbDecrypt DES OFB模式解密
func DesOfbDecrypt(data, key []byte) []byte {
	if len(key) != 8 {
		panic("des: key length must be 8 bytes")
	}

	block, err := des.NewCipher(key)
	if err != nil {
		panic("des: failed to create cipher: " + err.Error())
	}

	if len(data) < des.BlockSize {
		panic("des: encrypted data too short")
	}

	iv := data[:des.BlockSize]
	ciphertext := data[des.BlockSize:]

	stream := cipher.NewOFB(block, iv)
	decrypted := make([]byte, len(ciphertext))
	stream.XORKeyStream(decrypted, ciphertext)

	decrypted = pkcs7UnPadding(decrypted)

	return decrypted
}

// GenerateRsaKeyFile 在当前目录下创建rsa私钥文件和公钥文件
func GenerateRsaKeyFile(keySize int, priKeyFile, pubKeyFile string) error {
	// private key
	privateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return err
	}

	derText := x509.MarshalPKCS1PrivateKey(privateKey)

	block := pem.Block{
		Type:  "rsa private key",
		Bytes: derText,
	}

	file, err := os.Create(priKeyFile)
	if err != nil {
		panic(err)
	}
	err = pem.Encode(file, &block)
	if err != nil {
		return err
	}

	file.Close()

	// public key
	publicKey := privateKey.PublicKey

	derpText, err := x509.MarshalPKIXPublicKey(&publicKey)
	if err != nil {
		return err
	}

	block = pem.Block{
		Type:  "rsa public key",
		Bytes: derpText,
	}

	file, err = os.Create(pubKeyFile)
	if err != nil {
		return err
	}

	err = pem.Encode(file, &block)
	if err != nil {
		return err
	}

	file.Close()

	return nil
}

// RsaEncrypt RSA加密
func RsaEncrypt(data []byte, pubKeyFileName string) []byte {
	file, err := os.Open(pubKeyFileName)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	defer file.Close()
	buf := make([]byte, fileInfo.Size())

	_, err = file.Read(buf)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(buf)

	pubInterface, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		panic(err)
	}
	pubKey := pubInterface.(*rsa.PublicKey)

	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, data)
	if err != nil {
		panic(err)
	}

	return cipherText
}

// RsaDecrypt RSA解密
func RsaDecrypt(data []byte, privateKeyFileName string) []byte {
	file, err := os.Open(privateKeyFileName)
	if err != nil {
		panic(err)
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
	}
	buf := make([]byte, fileInfo.Size())
	defer file.Close()

	_, err = file.Read(buf)
	if err != nil {
		panic(err)
	}

	block, _ := pem.Decode(buf)

	priKey, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		panic(err)
	}

	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, priKey, data)
	if err != nil {
		panic(err)
	}

	return plainText
}

// GenerateRsaKeyPair 生成RSA密钥对
func GenerateRsaKeyPair(keySize int) (*rsa.PrivateKey, *rsa.PublicKey) {
	privateKey, _ := rsa.GenerateKey(rand.Reader, keySize)
	return privateKey, &privateKey.PublicKey
}

// RsaEncryptOAEP RSA加密OAEP
func RsaEncryptOAEP(data []byte, label []byte, key rsa.PublicKey) ([]byte, error) {
	encryptedBytes, err := rsa.EncryptOAEP(sha256.New(), rand.Reader, &key, data, label)
	if err != nil {
		return nil, err
	}

	return encryptedBytes, nil
}

// RsaDecryptOAEP RSA解密OAEP
func RsaDecryptOAEP(ciphertext []byte, label []byte, key rsa.PrivateKey) ([]byte, error) {
	decryptedBytes, err := rsa.DecryptOAEP(sha256.New(), rand.Reader, &key, ciphertext, label)
	if err != nil {
		return nil, err
	}

	return decryptedBytes, nil
}

// RsaSign RSA签名
func RsaSign(hash crypto.Hash, data []byte, privateKeyFileName string) ([]byte, error) {
	privateKey, err := loadRasPrivateKey(privateKeyFileName)
	if err != nil {
		return nil, err
	}

	hashed, err := hashData(hash, data)
	if err != nil {
		return nil, err
	}

	return rsa.SignPKCS1v15(rand.Reader, privateKey, hash, hashed)
}

// RsaVerifySign RSA验证签名
func RsaVerifySign(hash crypto.Hash, data, signature []byte, pubKeyFileName string) error {
	publicKey, err := loadRsaPublicKey(pubKeyFileName)
	if err != nil {
		return err
	}

	hashed, err := hashData(hash, data)
	if err != nil {
		return err
	}

	return rsa.VerifyPKCS1v15(publicKey, hash, hashed, signature)
}

// GenerateSm2KeyPair 生成SM2 密钥对
func GenerateSm2KeyPair() (*sm2.PrivateKey, *sm2.PublicKey) {
	privateKey, _ := sm2.GenerateKey(rand.Reader)
	return privateKey, &privateKey.PublicKey
}

// GenerateSm2KeyPairWithHex 生成SM2密钥并返回密钥对的Hex
func GenerateSm2KeyPairWithHex() (string, string) {
	privateKey, _ := sm2.GenerateKey(rand.Reader)
	return hex.EncodeToString(privateKey.D.Bytes()), hex.EncodeToString(append(privateKey.PublicKey.X.Bytes(), privateKey.PublicKey.Y.Bytes()...))
}

// Sm2Encrypt SM2 加密
func Sm2Encrypt(data []byte, pubKey *sm2.PublicKey) ([]byte, error) {
	return sm2.Encrypt(pubKey, data, rand.Reader, sm2.C1C3C2)
}

// Sm2EncryptWithHex SM2 加密
func Sm2EncryptWithHex(data []byte, pubKeyHex string) ([]byte, error) {
	pubKey, err := hexToSm2PublicKey(pubKeyHex)
	if err != nil {
		panic(err)
	}
	return Sm2Encrypt(data, pubKey)
}

// Sm2Decrypt  SM2 解密
func Sm2Decrypt(ciphertext []byte, privKey *sm2.PrivateKey) ([]byte, error) {
	return sm2.Decrypt(privKey, ciphertext, sm2.C1C3C2)
}

func Sm2DecryptWithHex(ciphertext []byte, privKeyHex string) ([]byte, error) {
	privKey, err := hexToSm2PrivateKey(privKeyHex)
	if err != nil {
		panic(err)
	}
	return Sm2Decrypt(ciphertext, privKey)
}

// Sm2Sign SM2
func Sm2Sign(privKey *sm2.PrivateKey, msg []byte) ([]byte, error) {
	return privKey.Sign(rand.Reader, msg, nil)
}

// Sm2SignWithHex 使用hex私钥签名
func Sm2SignWithHex(msg []byte, privKeyHex string) ([]byte, error) {
	privKey, err := hexToSm2PrivateKey(privKeyHex)
	if err != nil {
		panic(err)
	}
	return Sm2Sign(privKey, msg)
}

// Sm2Verify SM2 验签
func Sm2Verify(pubKey *sm2.PublicKey, msg, sig []byte) bool {
	return pubKey.Verify(msg, sig)
}

// Sm2VerifyWithHex 使用hex公钥验签
func Sm2VerifyWithHex(msg, sig []byte, pubKeyHex string) (bool, error) {
	pubKey, err := hexToSm2PublicKey(pubKeyHex)
	if err != nil {
		panic(err)
	}
	return Sm2Verify(pubKey, msg, sig), nil
}

// Sm3Hash SM3 哈希计算 (完整实现)
func Sm3Hash(data []byte) []byte {
	h := sm3.New()
	h.Write(data)
	return h.Sum(nil)
}

// Sm3HashWithHex SM3 哈希计算
func Sm3HashWithHex(data []byte) string {
	return hex.EncodeToString(Sm3Hash(data))
}

// Sm4EcbEncrypt SM4 ECB模式加密
func Sm4EcbEncrypt(key, plaintext []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("SM4: invalid key size (must be 16 bytes)")
	}

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	plaintext = pkcs7Padding(plaintext, bs)

	ciphertext := make([]byte, len(plaintext))
	for i := 0; i < len(plaintext); i += bs {
		block.Encrypt(ciphertext[i:], plaintext[i:])
	}

	return ciphertext, nil
}

// Sm4EcbDecrypt SM4 ECB模式解密
func Sm4EcbDecrypt(key, ciphertext []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("SM4: invalid key size (must be 16 bytes)")
	}

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	if len(ciphertext)%bs != 0 {
		return nil, errors.New("SM4: invalid ciphertext length")
	}

	plaintext := make([]byte, len(ciphertext))
	for i := 0; i < len(ciphertext); i += bs {
		block.Decrypt(plaintext[i:], ciphertext[i:])
	}

	return pkcs7UnPadding(plaintext), nil
}

// Sm4CbcEncrypt SM4 CBC模式加密
func Sm4CbcEncrypt(key, iv, plaintext []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("SM4: invalid key size (must be 16 bytes)")
	}
	if len(iv) != 16 {
		return nil, errors.New("SM4: invalid iv size (must be 16 bytes)")
	}

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	plaintext = pkcs7Padding(plaintext, bs)

	ciphertext := make([]byte, len(plaintext))
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, plaintext)

	return ciphertext, nil
}

// Sm4CbcDecrypt SM4 CBC模式解密
func Sm4CbcDecrypt(key, iv, ciphertext []byte) ([]byte, error) {
	if len(key) != 16 {
		return nil, errors.New("SM4: invalid key size (must be 16 bytes)")
	}
	if len(iv) != 16 {
		return nil, errors.New("SM4: invalid iv size (must be 16 bytes)")
	}

	block, err := sm4.NewCipher(key)
	if err != nil {
		return nil, err
	}

	bs := block.BlockSize()
	if len(ciphertext)%bs != 0 {
		return nil, errors.New("SM4: invalid ciphertext length")
	}

	plaintext := make([]byte, len(ciphertext))
	mode := cipher.NewCBCDecrypter(block, iv)
	mode.CryptBlocks(plaintext, ciphertext)

	return pkcs7UnPadding(plaintext), nil
}
