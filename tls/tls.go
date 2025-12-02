package utils

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
)

// 双向验证：server端提供证书(cert和key)，还必须配置cerfiles即CA根证书，因为需要验证client端提供的证书。另外client也端必须提供一样的内容，即client端的证书(cert/key)以供server端验证，并且提供CA根证书验证server端提供的证书。
// 单向验证：server端提供证书(cert和key)，不需要配置certfiles即CA根证书，而在客户端必须提供CA根证书，用来验证server端的证书是否有效。另外client端也不需要自己的证书，因为它不需要想server端提供验证。

// LoadServerTlsConfigFile 创建服务端TLS证书认证配置
// keyFile 服务端私钥文件路径，必须提供
// certFile 服务端证书文件路径，必须提供
// caFile CA根证书，如果提供则为双向认证，否则为单向认证
// insecureSkipVerify 用来控制客户端是否证书和服务器主机名。如果设置为true,则不会校验证书以及证书中的主机名和服务器主机名是否一致。
func LoadServerTlsConfigFile(keyFile, certFile, caFile string, insecureSkipVerify bool) (*tls.Config, error) {
	if keyFile == "" || certFile == "" {
		return nil, fmt.Errorf("KeyFile and CertFile must both be present[key: %v, cert: %v]", keyFile, certFile)
	}

	var cfg tls.Config
	cfg.InsecureSkipVerify = insecureSkipVerify
	//cfg.ServerName = "host.docker.internal"
	//cfg.MinVersion = tls.VersionTLS13

	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Println("tls.LoadX509KeyPair error:", err)
		return nil, err
	}

	cfg.Certificates = []tls.Certificate{tlsCert}

	if caFile != "" {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		cp, err := newCertPoolWithCaFile(caFile)
		if err != nil {
			log.Fatalln("read cert file error:", err)
			return nil, err
		}

		cfg.RootCAs = cp
		cfg.ClientCAs = cp
	} else {
		cfg.ClientAuth = tls.NoClientCert
	}

	return &cfg, nil
}

func LoadServerTlsConfigString(keyPEMBlock, certPEMBlock, caPEMBlock []byte, insecureSkipVerify bool) (*tls.Config, error) {
	if len(keyPEMBlock) == 0 || len(certPEMBlock) == 0 {
		return nil, fmt.Errorf("KeyPEMBlock and CertPEMBlock must both be present[key: %v, cert: %v]", keyPEMBlock, certPEMBlock)
	}

	var cfg tls.Config
	cfg.InsecureSkipVerify = insecureSkipVerify
	//cfg.ServerName = "host.docker.internal"
	//cfg.MinVersion = tls.VersionTLS13

	tlsCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		log.Println("tls.X509KeyPair error:", err)
		return nil, err
	}

	cfg.Certificates = []tls.Certificate{tlsCert}

	if len(caPEMBlock) != 0 {
		cfg.ClientAuth = tls.RequireAndVerifyClientCert
		cp, err := newCertPool(caPEMBlock)
		if err != nil {
			log.Fatalln("read cert PEM error:", err)
			return nil, err
		}

		cfg.RootCAs = cp
		cfg.ClientCAs = cp
	} else {
		cfg.ClientAuth = tls.NoClientCert
	}

	return &cfg, nil
}

// LoadClientTlsConfigFile 创建客户端端TLS证书认证配置
// keyFile 客户端私钥文件路径
// certFile 客户端证书文件路径
// caFile CA根证书
func LoadClientTlsConfigFile(keyFile, certFile, caFile string) (*tls.Config, error) {
	var cfg tls.Config
	//cfg.InsecureSkipVerify = info.InsecureSkipVerify
	//cfg.ServerName = "host.docker.internal"
	//cfg.MinVersion = tls.VersionTLS13

	if keyFile == "" || certFile == "" {
		return &cfg, nil
	}

	tlsCert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalln("read pair file error:", err)
		return nil, err
	}

	cfg.Certificates = []tls.Certificate{tlsCert}

	if caFile != "" {
		cp, err := newCertPoolWithCaFile(caFile)
		if err != nil {
			log.Fatalln("read cert file error:", err)
			return nil, err
		}

		cfg.RootCAs = cp
	}

	return &cfg, nil
}

func LoadClientTlsConfigString(keyPEMBlock, certPEMBlock, caPEMBlock []byte) (*tls.Config, error) {
	if len(keyPEMBlock) == 0 || len(certPEMBlock) == 0 {
		return nil, fmt.Errorf("KeyPEMBlock and CertPEMBlock must both be present[key: %v, cert: %v]", keyPEMBlock, certPEMBlock)
	}

	var cfg tls.Config
	//cfg.InsecureSkipVerify = info.InsecureSkipVerify
	//cfg.ServerName = "host.docker.internal"
	//cfg.MinVersion = tls.VersionTLS13

	tlsCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		log.Fatalln("read pair PEM error:", err)
		return nil, err
	}

	cfg.Certificates = []tls.Certificate{tlsCert}

	if len(caPEMBlock) != 0 {
		cp, err := newCertPool(caPEMBlock)
		if err != nil {
			log.Fatalln("read cert PEM error:", err)
			return nil, err
		}

		cfg.RootCAs = cp
	}

	return &cfg, nil
}

// newCertPool creates x509 certPool with provided CA file
func newCertPoolWithCaFile(caFile string) (*x509.CertPool, error) {
	pemByte, err := os.ReadFile(caFile)
	if err != nil {
		return nil, err
	}

	return newCertPool(pemByte)
}

func newCertPool(caPEMBlock []byte) (*x509.CertPool, error) {
	certPool := x509.NewCertPool()

	var block *pem.Block
	var cert *x509.Certificate
	var err error
	for {
		block, caPEMBlock = pem.Decode(caPEMBlock)
		if block == nil {
			return certPool, nil
		}

		cert, err = x509.ParseCertificate(block.Bytes)
		if err != nil {
			return nil, err
		}

		certPool.AddCert(cert)
	}

	//if !certPool.AppendCertsFromPEM(caPEMBlock) {
	//	return nil, fmt.Errorf("can't add CA cert")
	//}
	//return certPool, nil
}
