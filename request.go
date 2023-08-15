/**
 * @Author: ZhaoYadong
 * @Date: 2023-08-15 13:36:06
 * @LastEditors: ZhaoYadong
 * @LastEditTime: 2023-08-15 13:36:07
 * @FilePath: /src/shield/request.go
 */

package util

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
)

func chechCode(code string) (string, error) {
	code = strings.TrimSpace(code)
	code = strings.Trim(code, "\"")
	if code == "" || len(code) != 36 {
		return "", errors.New("激活码必须是36位长度")
	}
	return code, nil
}

func RemoteActive(machineID, code string) (*CDKey, error) {
	code, err := chechCode(code)
	if err != nil {
		return nil, err
	}
	data := ActiveData{
		MachineID: machineID,
		Code:      code,
	}
	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, errors.New("请求原始数据编码失败")
	}
	sk := []byte(GenerateKey(getSk()))
	encrypted, err := AesEncrypt(dataBytes, sk)
	if err != nil {

		return nil, errors.New("加密失败")
	}
	// 拼接request
	activeRequest := ActiveRequest{
		Data: encrypted,
	}
	activeRequestByte, err := json.Marshal(activeRequest)
	if err != nil {
		return nil, errors.New("请求封装数据编码失败")
	}

	request, err := http.NewRequest("POST", getURL(), bytes.NewReader(activeRequestByte))
	if err != nil {
		return nil, errors.New("远程请求失败,检查网络")
	}
	request.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, errors.New("远程请求失败,无法获取响应")
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.New("读取远程返回内容失败")
	}

	response := &Response{}
	if err := json.Unmarshal(body, response); err != nil {
		return nil, errors.New("解码失败")
	}
	if response.Code != 0 {
		return nil, errors.New("远程返回信息: " + response.Msg)
	}

	responseDataBytes, err := json.Marshal(response.Data)
	if err != nil {
		return nil, errors.New("远程返回内容编码失败")
	}

	var activeResponse ActiveResponse
	if err := json.Unmarshal(responseDataBytes, &activeResponse); err != nil {
		return nil, errors.New("远程返回消息体解码失败")
	}
	decrypt, err := AesDecrypt(activeResponse.Data, sk)
	if err != nil {
		return nil, errors.New("远程返回详情解密失败")
	}
	cdk := new(CDKey)
	if err := json.Unmarshal(decrypt, cdk); err != nil {
		return nil, errors.New("远程返回详情解码失败")
	}

	return cdk, nil
}
