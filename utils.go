/**
 * @Author: ZhaoYadong
 * @Date: 2023-07-07 22:51:21
 * @LastEditors: ZhaoYadong
 * @LastEditTime: 2023-08-16 13:00:41
 * @FilePath: /src/shield/utils.go
 */
 package util

 import (
	 "bytes"
	 "crypto/aes"
	 "crypto/cipher"
	 "crypto/hmac"
	 "crypto/sha256"
	 "encoding/hex"
	 "encoding/json"
	 "errors"
	 "io/fs"
	 "os"
	 "strings"
	 "time"
 
	 "github.com/denisbrodbeck/machineid"
 )
 
 var (
	 ValidateFileName = "LICENSE"
	 Sk               = ""
	 sk               = "40999aac89e7622f3ca71fba1d972fd94a31c3bfb"
	 URL              = ""
	 url              = "http://127.0.0.1:8888/active"
 )
 
 func getSk() string {
	 if Sk != "" {
		 return Sk
	 }
	 return sk
 }
 
 func getURL() string {
	 if URL != "" {
		 return URL
	 }
	 return url
 }
 
 //高级加密标准（Adevanced Encryption Standard ,AES）
 // 16,24,32位字符串的话，分别对应AES-128，AES-192，AES-256 加密方法
 
 // PKCS7 填充模式
 func PKCS7Padding(ciphertext []byte, blockSize int) []byte {
	 padding := blockSize - len(ciphertext)%blockSize
	 //Repeat()函数的功能是把切片[]byte{byte(padding)}复制padding个，然后合并成新的字节切片返回
	 padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	 return append(ciphertext, padtext...)
 }
 
 // 填充的反向操作，删除填充字符串
 func PKCS7UnPadding(origData []byte) ([]byte, error) {
	 //获取数据长度
	 length := len(origData)
	 if length == 0 {
		 return nil, errors.New("加密字符串错误！")
	 } else {
		 //获取填充字符串长度
		 unpadding := int(origData[length-1])
		 //截取切片，删除填充字节，并且返回明文
		 return origData[:(length - unpadding)], nil
	 }
 }
 
 // 实现加密
 func AesEncrypt(origData []byte, key []byte) ([]byte, error) {
	 //创建加密算法实例
	 block, err := aes.NewCipher(key)
	 if err != nil {
		 return nil, err
	 }
	 //获取块的大小
	 blockSize := block.BlockSize()
	 //对数据进行填充，让数据长度满足需求
	 origData = PKCS7Padding(origData, blockSize)
	 //采用AES加密方法中CBC加密模式
	 blocMode := cipher.NewCBCEncrypter(block, key[:blockSize])
	 crypted := make([]byte, len(origData))
	 //执行加密
	 blocMode.CryptBlocks(crypted, origData)
	 return crypted, nil
 }
 
 // 实现解密
 func AesDecrypt(cypted []byte, key []byte) ([]byte, error) {
	 //创建加密算法实例
	 block, err := aes.NewCipher(key)
	 if err != nil {
		 return nil, err
	 }
	 //获取块大小
	 blockSize := block.BlockSize()
	 //创建加密客户端实例
	 blockMode := cipher.NewCBCDecrypter(block, key[:blockSize])
	 origData := make([]byte, len(cypted))
	 //这个函数也可以用来解密
	 blockMode.CryptBlocks(origData, cypted)
	 //去除填充字符串
	 origData, err = PKCS7UnPadding(origData)
	 if err != nil {
		 return nil, err
	 }
	 return origData, err
 }
 
 var keyLen = 32
 
 func GenerateKey(key string) string {
	 len := len(key)
	 if len < keyLen {
		 mac := hmac.New(sha256.New, []byte(key))
		 key = hex.EncodeToString(mac.Sum(nil))
	 }
	 return key[:keyLen]
 }
 
 // 获取硬件识别码
 func GetMachineID() (string, error) {
	 key, err := machineid.ID()
	 if err != nil {
		 return "", errors.New("读取设备识别码失败")
	 }
	 return key, nil
 }
 
 // 校验文件检测
 // *Encrypt 加密文件内容
 // bool 是否需要重新激活
 // error 常规错误提示
 func ChechEncryptFile(appSymbol string, machineID string) (*Encrypt, bool, error) {
	 // 1.检查校验文件
	 data, err := os.ReadFile(ValidateFileName)
	 if err != nil {
		 return nil, true, errors.New("机要文件不存在或已损坏,请检查后再启动")
	 }
	 // 2.处理硬件识别码
	 machineID = GenerateKey(machineID)
 
	 // 3.解析校验文件内容
	 encryptBytes, err := AesDecrypt(data, []byte(machineID))
	 if err != nil {
		 return nil, false, errors.New("解密机要文件内容失败")
	 }
 
	 // 4.转码
	 encrypt := new(Encrypt)
	 if err := json.Unmarshal(encryptBytes, encrypt); err != nil {
		 return nil, false, errors.New("转码机要文件内容失败")
	 }
 
	 // 5.比较当前系统时间是否调整过
	 if encrypt.SyncedAt.After(time.Now()) {
		 return nil, false, errors.New("系统时间回拨过,请更正系统时间")
	 }
 
	 // 6.检查激活码开始时间
	 if encrypt.StartAt.After(time.Now()) {
		 return nil, false, errors.New("还未到激活码开始时间")
	 }
 
	 // 7.检查激活码是否过期
	 if encrypt.Genre != 3 && encrypt.EndAt.Before(time.Now()) {
		 return nil, true, errors.New("激活码已过期,请重新激活")
	 }
	 // 8.检查激活文件适用范围
	 if !encrypt.Scope[strings.ToLower(appSymbol)] {
		 return nil, true, errors.New("激活码不适用于当前应用,请重新激活")
	 }
 
	 return encrypt, false, nil
 }
 
 // 生成校验文件
 func GenerateValidateFlie(encrypt Encrypt, machineID string) error {
	 // 1.处理硬件识别码
	 machineID = GenerateKey(machineID)
	 // 2.转码
	 encryptJsonBytes, err := json.Marshal(encrypt)
	 if err != nil {
		 return errors.New("转码机要文件内容失败")
	 }
 
	 // 3.加密
	 encryptBytes, err := AesEncrypt(encryptJsonBytes, []byte(machineID))
	 if err != nil {
		 return errors.New("加密机要文件内容失败")
	 }
 
	 // 4.写入机要文件
	 if err := os.WriteFile(ValidateFileName, encryptBytes, fs.ModePerm); err != nil {
		 return errors.New("生成机要文件失败")
	 }
	 return nil
 }
 
 func TimingSync(encrypt Encrypt, machineID string, d time.Duration, ch chan bool) {
	 defer close(ch)
	 ticker := time.NewTicker(d)
	 defer ticker.Stop()
	 for range ticker.C {
		 encrypt.SyncedAt = time.Now()
		 if err := GenerateValidateFlie(encrypt, machineID); err != nil {
			 println(err.Error())
			 return
		 }
		 if encrypt.Genre != 3 && encrypt.EndAt.Before(encrypt.SyncedAt) {
			 println("证书已过期")
			 return
		 }
	 }
 }
 
 func Run(appSymbol string, obj Runable) {
	 defer func() {
		 if err := recover(); err != nil {
			 println("证书不合法")
		 }
	 }()
	 ch := make(chan bool)
	 encrypt, machineID, err := startVerify(appSymbol, ch)
	 if err != nil {
		 println(err.Error(), " 请检查证书是否正确")
		 return
	 }
	 go obj.Exec()
 
	 TimingSync(*encrypt, machineID, time.Second*15, ch)
 }
 
 func startVerify(appSymbol string, ch chan bool) (*Encrypt, string, error) {
	 machineID, err := GetMachineID()
	 if err != nil {
 
		 return nil, "", err
	 }
	 // 检查本地激活文件是否存在,如果存在检查激活是否到期,否则退出
	 encrypt, _, err := ChechEncryptFile(appSymbol, machineID)
	 if err != nil {
		 return nil, "", err
	 }
	 // if encrypt.Genre != 3 {
	 // 	go TimingSync(*encrypt, machineID, time.Second*15, ch)
	 // }
 
	 return encrypt, machineID, nil
 }
 