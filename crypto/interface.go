package crypto

// Cipher 定义了加密算法的接口，支持加密和解密操作
type Cipher interface {
	// Encrypt 使用指定的密钥和初始化向量（IV）加密明文数据，返回加密后的数据
	Encrypt(plain []byte) ([]byte, error)

	// Decrypt 使用指定的密钥和初始化向量（IV）解密密文数据，返回解密后的数据
	Decrypt(cipher []byte) ([]byte, error)

	// Name 返回算法名称
	Name() string
}

// Hasher 定义了哈希算法的接口，支持计算数据的哈希值
type Hasher interface {
	// Sum 计算输入数据的哈希值，返回哈希结果
	Sum(data []byte) ([]byte, error)

	// Name 返回哈希算法名称
	Name() string
}

// KeyExchanger 定义了密钥协商算法的接口
// 适用于ECDH、SM2密钥交换等场景
type KeyExchanger interface {
	// DeriveSharedSecret 计算共享密钥，用于根据对方公钥字节计算共享密钥
	DeriveSharedSecret(peerPubBytes []byte) ([]byte, error)

	// PublicKeyBytes 获取本地公钥字节
	PublicKeyBytes() []byte
}

// Signer 定义了签名算法的接口
// 适用于ECDSA、SM2等签名场景
type Signer interface {
	// Sign 返回签名字符串
	Sign(data []byte) (string, error)

	// Name 返回算法名称
	Name() string
}

// Verifier 定义了验签算法的接口
// 适用于ECDSA、SM2等签名场景
type Verifier interface {
	// Verify 验证签名
	Verify(data []byte, signature string) (bool, error)

	// Name 返回算法名称
	Name() string
}

// 编译期断言：具体实现满足对应接口
var (
	_ Cipher       = (*AESCipher)(nil)
	_ Cipher       = (*AESGCMCipher)(nil)
	_ Cipher       = (*SM2Cipher)(nil)
	_ Cipher       = (*SM4Cipher)(nil)
	_ Cipher       = (*RSACipher)(nil)
	_ Hasher       = (*SHA256Hasher)(nil)
	_ Hasher       = (*SHA512Hasher)(nil)
	_ Hasher       = (*SM3Hasher)(nil)
	_ KeyExchanger = (*ECDHCipher)(nil)
	_ Signer       = (*ECDSACipher)(nil)
	_ Signer       = (*SM2Cipher)(nil)
	_ Verifier     = (*ECDSACipher)(nil)
	_ Verifier     = (*SM2Cipher)(nil)
)
