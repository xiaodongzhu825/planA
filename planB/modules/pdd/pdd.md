# pdd.dll 使用教程
## 1.创建DLL工具实例
### 加载DLL文件
```gotemplate
// PddDLL 拼多多工具DLL结构
type pddDLL struct {
	dll                            *syscall.DLL
	pddGoodsOuterCatMappingGet    *syscall.Proc // 类目预测
	freeCString                   *syscall.Proc // 释放C字符串
}

// 初始化pddDLL
func InitPddDLL() (*pddDLL, error) {
	dllPath := filepath.Join("dll", "pdd.dll")
	if _, err := os.Stat(dllPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("pdd DLL 不存在: %s", dllPath)
	}
	if dll, err := syscall.LoadDLL(dllPath); err != nil {
		return nil, fmt.Errorf("加载pdd DLL 失败: %s", err)
	} else {
		return &pddDLL{
			dll:                            dll,
			pddGoodsOuterCatMappingGet:    dll.MustFindProc("PddGoodsOuterCatMappingGet"),
			freeCString:                   dll.MustFindProc("FreeCString"),
		}, nil
	}
}

dll, err := InitPddDLL()
```

### 获取C字符串
```gotemplate
// cStr 获取C字符串
func (m *pddDLL) cStr(p uintptr) string {
	if p == 0 {
		return ""
	}
	b := []byte{}
	for i := uintptr(0); ; i++ {
		c := *(*byte)(unsafe.Pointer(p + i))
		if c == 0 {
			break
		}
		b = append(b, c)
	}
	s := string(b)
	if m.freeCString != nil {
		m.freeCString.Call(p)
	}
	return s
}
```

## 2. 使用dll函数示例
```gotemplate
// 类目预测
func (m *pddDLL) PddGoodsOuterCatMappingGet(clientId, clientSecret, accessToken,
	outerCatId, outerCatName, outerGoodsName string) (string, error) {
	proc, err := m.dll.FindProc("PddGoodsOuterCatMappingGet")
	if err != nil {
		return "", fmt.Errorf("找不到函数 PddGoodsOuterCatMappingGet: %v", err)
	}
	
	clientIdPtr, _ := syscall.BytePtrFromString(clientId)
	clientSecretPtr, _ := syscall.BytePtrFromString(clientSecret)
	accessTokenPtr, _ := syscall.BytePtrFromString(accessToken)
	outerCatIdPtr, _ := syscall.BytePtrFromString(outerCatId)
	outerCatNamePtr, _ := syscall.BytePtrFromString(outerCatName)
	outerGoodsNamePtr, _ := syscall.BytePtrFromString(outerGoodsName)
	
	resultPtr, _, _ := proc.Call(
		uintptr(unsafe.Pointer(clientIdPtr)),
		uintptr(unsafe.Pointer(clientSecretPtr)),
		uintptr(unsafe.Pointer(accessTokenPtr)),
		uintptr(unsafe.Pointer(outerCatIdPtr)),
		uintptr(unsafe.Pointer(outerCatNamePtr)),
		uintptr(unsafe.Pointer(outerGoodsNamePtr)),
	)
	
	result := m.cStr(resultPtr)
	return result, nil
}
```

# 接口详情
## 1. 类目预测--PddGoodsOuterCatMappingGet
### 请求信息
```gotemplate
dll.PddGoodsOuterCatMappingGet(clientId, clientSecret, accessToken,
outerCatId, outerCatName, outerGoodsName)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| outerCatId | string | 是 | 外部平台类目ID |
| outerCatName | string | 是 | 外部平台类目名称 |
| outerGoodsName | string | 是 | 外部商品名称 |
### 响应示例
```json
{
  "outer_cat_mapping_get_response": {
    "cat_id2": 16028,
    "cat_id3": 16031,
    "cat_id1": 15543,
    "request_id": "17666480184871649",
    "cat_id4": 0
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 2. 快递公司查看--PddLogisticsCompaniesGet
### 请求信息
```gotemplate
dll.PddLogisticsCompaniesGet(clientId, clientSecret)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
### 响应示例
```json
{
  "logistics_companies_get_response": {
    "logistics_companies": [
      {
        "available": 1,
        "code": "SF",
        "id": 1,
        "logistics_company": "顺丰速运"
      },
      {
        "available": 1,
        "code": "STO",
        "id": 2,
        "logistics_company": "申通快递"
      }
    ]
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 3. erp打单信息同步--PddErpOrderSync
### 请求信息
```gotemplate
dll.PddErpOrderSync(clientId, clientSecret, accessToken, logisticsId,
orderSn, orderState, waybillNo)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| logisticsId | string | 是 | 	物流公司ID |
| orderSn | string | 是 | 拼多多订单号 |
| orderState | string | 是 | 订单状态 |
| waybillNo | string | 是 | 	运单号 |
### 响应示例
```json
{
  "erp_order_sync_response": {
    "is_success": true,
    "request_id": "17666480184871650"
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 4. 拼多多订单同步--PddOrderSynchronization
### 请求信息
```gotemplate
dll.PddOrderSynchronization(clientId, clientSecret, accessToken, logisticsCompany, logisticsOnlineSendJson)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| logisticsCompany | string | 是 | 物流公司名称 |
| logisticsOnlineSendJson | string | 是 | 	拼多多订单同步json字符串 |
### 响应示例
```json
{
  "erp_order_sync_response": {
    "is_success": true,
    "request_id": "17666480184871651"
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 5. 商品图片上传接口--PddGoodsImgUpload
### 请求信息
```gotemplate
dll.PddGoodsImgUpload(clientId, clientSecret, accessToken, filePath)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| filePath | string | 是 | 图片文件路径 |
### 响应示例
```json
{
  "goods_img_upload_response": {
    "image_url": "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg",
    "request_id": "17666480184871652"
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 6. 商品新增接口--PddGoodsAdd
### 请求信息
```gotemplate
dll.PddGoodsAdd(clientId, clientSecret, accessToken, goodsAddJson)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| goodsAddJson | string | 是 | 商品信息JSON字符串 |
#### 商品信息JSON结构示例
```json
{
  "goods_name": "测试商品",
  "goods_desc": "商品描述",
  "cat_id": 20111,
  "goods_type": 1,
  "market_price": 9900,
  "is_folt": false,
  "is_pre_sale": false,
  "is_refundable": true,
  "shipment_limit_second": 86400,
  "cost_template_id": 10001,
  "image_url": "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg",
  "carousel_gallery": [
    "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg"
  ],
  "detail_gallery": [
    "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg"
  ],
  "sku_list": [
    {
      "out_sku_sn": "SKU001",
      "price": 8900,
      "quantity": 100,
      "spec_id_list": "1001:10001",
      "sku_properties": [
        {
          "ref_pid": 1001,
          "value": "红色",
          "vid": 10001,
          "punit": "个"
        }
      ],
      "is_onsale": 1,
      "limit_quantity": 10,
      "multi_price": 8500,
      "thumb_url": "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg",
      "weight": 500
    }
  ]
}
```
### 响应示例
```json
{
  "goods_add_response": {
    "goods_id": 123456789,
    "goods_name": "测试商品",
    "goods_sn": "G202501200001",
    "request_id": "17666480184871653"
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 7. 联合拼多多图片上传的商品新增--SelfPddGoodsAdd
### 请求信息
```gotemplate
dll.SelfPddGoodsAdd(clientId, clientSecret, accessToken, filePath, goodsAddJson)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| filePath | string | 是 | 图片文件路径 |
| goodsAddJson | string | 是 | 商品信息JSON字符串（不需包含image_url）|
#### 接口说明
此接口为组合接口，内部执行以下步骤：
1.上传商品主图文件到拼多多服务器
2.获取图片URL并自动填充到商品信息中
3.调用商品新增接口创建商品
#### 商品信息JSON结构示例
```json
{
  "goods_name": "测试商品",
  "goods_desc": "商品描述",
  "cat_id": 20111,
  "goods_type": 1,
  "market_price": 9900,
  "is_folt": false,
  "is_pre_sale": false,
  "is_refundable": true,
  "shipment_limit_second": 86400,
  "cost_template_id": 10001,
  "image_url": "",
  "carousel_gallery": [
    "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg"
  ],
  "detail_gallery": [
    "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg"
  ],
  "sku_list": [
    {
      "out_sku_sn": "SKU001",
      "price": 8900,
      "quantity": 100,
      "spec_id_list": "1001:10001",
      "sku_properties": [
        {
          "ref_pid": 1001,
          "value": "红色",
          "vid": 10001,
          "punit": "个"
        }
      ],
      "is_onsale": 1,
      "limit_quantity": 10,
      "multi_price": 8500,
      "thumb_url": "http://oms-imageimg.pinduoduo.com/upload/2025/01/20/e9a8c1b6e1a84f1d8d7c3a8b9e2f5c7d.jpg",
      "weight": 500
    }
  ]
}
```
### 响应示例
```json
{
  "goods_add_response": {
    "goods_id": 123456790,
    "goods_name": "测试商品",
    "goods_sn": "G202501200002",
    "request_id": "17666480184871654"
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 8. 批量数据解密脱敏接口--PddOpenDecryptMaskBatch
### 请求信息
```gotemplate
dll.PddOpenDecryptMaskBatch(clientId, clientSecret, accessToken, reqJson)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| reqJson | string | 是 | 信息JSON字符串 |
#### 信息JSON结构示例
```json
[
  {
    "data_tag": "251229-272441044622514",
    "encrypted_data": "~AgAAAAPlscEH0psOJAEXpTdsLOWvDJ9bB7IEjIoqNfiDhhJR9NHOxsdZ+PEFluSSCngCikoDU+CP/sSXZJ92ic7+PdNlJNLA7g/6VUMDWF6RvjW9IeRN+lKNarsjWDQR~0~"
  }
]
```
### 响应示例
```json
{
  "open_decrypt_mask_batch_response": {
    "data_decrypt_list": [
      {
        "data_tag": "str",
        "data_type": 0,
        "decrypted_data": "str",
        "encrypted_data": "str",
        "error_code": 0,
        "error_msg": "str"
      }
    ]
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 生成商家自定义的规格--PddGoodsSpecIdGet
### 请求信息
```gotemplate
dll.PddGoodsSpecIdGet(clientId, clientSecret, accessToken, parentSpecId, specName)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| clientId | string | 是 | 拼多多开放平台ClientID |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌 |
| parentSpecId | string | 是 | 拼多多标准规格ID |
| specName | string | 是 | 商家编辑的规格值，如颜色规格下设置白色属性 |
### 响应参数
```json
{
  "goods_spec_id_get_response": {
    "parent_spec_id": 0,
    "spec_id": 0,
    "spec_name": "str"
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 修改商品SKU价格--PddGoodsSkuPriceUpdate
### 请求信息
```gotemplate
dll.PddGoodsSkuPriceUpdate(clientId, clientSecret, accessToken, request)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明                    |
|--|--|--|-----------------------|
| clientId | string | 是 | 拼多多开放平台ClientID       |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret   |
| accessToken | string | 是 | 授权令牌                  |
| request | string | 是 | 价格更新请求JSON字符串         |
#### 请求JSON结构
```json
{
  "goods_id": "必填，商品id，类型为LONG",
  "ignore_edit_warn": "非必填，是否获取商品发布警告信息，默认为忽略，类型为BOOLEAN",
  "market_price": "非必填，参考价（单位分），类型为LONG",
  "market_price_in_yuan": "非必填，参考价（单位元），类型为STRING",
  "sku_price_list": [
    {
      "group_price": "非必填，拼团购买价格（单位分），类型为LONG",
      "is_onsale": "非必填，sku上架状态，0-已下架，1-上架中，类型为INTEGER",
      "single_price": "非必填，单独购买价格（单位分），类型为LONG",
      "sku_id": "必填，sku标识，类型为LONG"
    }
  ],
  "sync_goods_operate": "非必填，提交后上架状态，0:上架,1:保持原样，类型为INTEGER",
  "two_pieces_discount": "非必填，满2件折扣，可选范围0-100，0表示取消，95表示95折，设置需先查询规则接口获取实际可填范围，类型为INTEGER"
}
```
### 响应参数
```json
{
  "goods_update_sku_price_response": {
    "goods_commit_id": 0,
    "is_success": true
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 商品库存更新接口--PddGoodsQuantityUpdate
### 请求信息
```gotemplate
dll.PddGoodsQuantityUpdate(clientId, clientSecret, accessToken, request)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明                  |
|--|--|--|---------------------|
| clientId | string | 是 | 拼多多开放平台ClientID     |
| clientSecret | string | 是 | 拼多多开放平台ClientSecret |
| accessToken | string | 是 | 授权令牌                |
| request | string | 是 | 库存更新请求JSON字符串       |
#### 请求JSON结构  request 字符串
```json
{
  "force_update": "非必填，是否强制更新，仅update_type=1(全量更新)时有效，默认值false；force_update=false时，quantity不能小于预扣库存；force_update=true时，代表强制更新，当quantity<预扣库存时，不报错，直接将quantity清0，类型为BOOLEAN",
  "goods_id": "必填，商品id，类型为LONG",
  "outer_id": "非必填，sku商家编码，类型为STRING",
  "quantity": "必填，库存修改值。当全量更新库存时，quantity必须为大于等于0的正整数；当增量更新库存时，quantity为整数，可小于等于0。若增量更新时传入的库存为负数，则负数与实际库存之和不能小于0。比如当前实际库存为1，传入增量更新quantity=-1，库存改为0，类型为LONG",
  "sku_id": "非必填，sku_id和outer_id必填一个，类型为LONG",
  "update_type": "非必填，库存更新方式，可选。1为全量更新，2为增量更新。如果不填，默认为全量更新，类型为INTEGER"
}
```
### 响应参数
```json
{
  "goods_quantity_update_response": {
    "is_success": false
  }
}
```
### 错误响应示例
```json
{
  "error_response": {
    "error_msg": "公共参数错误:type",
    "sub_msg": "",
    "sub_code": null,
    "error_code": 10001,
    "request_id": "15440104776643887"
  }
}
```

## 获取商品信息接口 -- OutPddAuthGetCommitDetailt
### 请求信息
```gotemplate
dll.OutPddAuthGetCommitDetailt(goodsCommitId, goodsId, accessToken)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| goodsCommitId | string | 是 | 商品提交ID |
| goodsId | string | 是 | 商品ID |
| accessToken | string | 是 | 授权令牌 |
### 响应参数
```json

```


## 获取商品详情信息接口 -- OutPddAuthGetGoodsDetail
### 请求信息
```gotemplate
dll.OutPddAuthGetGoodsDetail(goodsId, accessToken)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| goodsId | string | 是 | 商品ID |
| accessToken | string | 是 | 授权令牌 |
### 响应参数
```json
{
    "bad_fruit_claim": 0,
    "buy_limit": 999999,
    "carousel_gallery_list": [
        "https://img.pddpic.com/open-gw/2025-06-30/59c30d4c-193f-40a3-a639-7af59a381ec5.jpeg",
        "https://img.pddpic.com/open-gw/2023-09-07/02a5c39a-7a90-4530-a338-3e87095a21a9.png",
        "https://img.pddpic.com/open-gw/2023-09-07/02a5c39a-7a90-4530-a338-3e87095a21a9.png",
        "https://img.pddpic.com/open-gw/2023-09-07/02a5c39a-7a90-4530-a338-3e87095a21a9.png",
        "https://img.pddpic.com/open-gw/2023-09-07/02a5c39a-7a90-4530-a338-3e87095a21a9.png",
        "https://img.pddpic.com/open-gw/2023-09-07/02a5c39a-7a90-4530-a338-3e87095a21a9.png",
        "https://img.pddpic.com/open-gw/2023-09-07/02a5c39a-7a90-4530-a338-3e87095a21a9.png",
        "https://img.pddpic.com/open-gw/2023-09-07/02a5c39a-7a90-4530-a338-3e87095a21a9.png",
        "https://img.pddpic.com/open-gw/2025-06-30/4539f740-331b-4687-aa00-5c96855de6cd.jpeg",
        "https://img.pddpic.com/open-gw/2025-06-30/b0e89e39-c97b-475d-9be2-f1909e30acb5.jpeg"
    ],
    "cat_id": 15678,
    "cost_template_id": 655688447565777,
    "country_id": 0,
    "customer_num": 2,
    "customs": "",
    "detail_gallery_list": [
        "https://img.pddpic.com/open-gw/2025-06-30/b691c104-baf8-42b2-97e2-b7258113114b.jpeg",
        "https://img.pddpic.com/open-gw/2023-09-07/53e6f7ff-d15e-4e8f-8625-e293717ca1e4.jpeg",
        "https://img.pddpic.com/open-gw/2023-09-07/ecff591d-32a6-42c9-ba5a-6a42829092a8.jpeg",
        "https://img.pddpic.com/open-gw/2023-10-16/7034f8a0-5d88-49f8-a96f-608abb8cac80.jpeg",
        "https://img.pddpic.com/open-gw/2023-10-16/e10c2b6c-d4de-4fdd-8d48-f0a334735e9a.jpeg",
        "https://img.pddpic.com/open-gw/2023-10-16/c19358fb-0a4d-49ad-bcc8-b2980e938064.jpeg",
        "https://img.pddpic.com/open-gw/2025-06-30/1deeb9c0-7212-432b-a309-f774db6e1adb.jpeg"
    ],
    "goods_desc": "书名：金属工艺学  下  第6版，作者：'邓文英，宋力宏主编'，ISBN：9787040456295，出版社：高等教育出版社",
    "goods_id": 770621582375,
    "goods_name": "金属工艺学  下  第6版  邓文英，宋力宏主编 高等教育出版社 978",
    "goods_property_list": [
        {
            "punit": "",
            "ref_pid": 425,
            "template_pid": 401030,
            "vid": 0,
            "vvalue": "9787040456295"
        },
        {
            "punit": "",
            "ref_pid": 876,
            "template_pid": 401029,
            "vid": 0,
            "vvalue": "金属工艺学  下  第6版"
        },
        {
            "punit": "页",
            "ref_pid": 692,
            "template_pid": 401032,
            "vid": 0,
            "vvalue": "157"
        },
        {
            "punit": "元",
            "ref_pid": 879,
            "template_pid": 401034,
            "vid": 0,
            "vvalue": "24.70"
        },
        {
            "punit": "",
            "ref_pid": 882,
            "template_pid": 401037,
            "vid": 0,
            "vvalue": "邓文英，宋力宏主编"
        },
        {
            "punit": "",
            "ref_pid": 880,
            "template_pid": 401035,
            "vid": 483761,
            "vvalue": "高等教育出版社"
        },
        {
            "punit": "",
            "ref_pid": 888,
            "template_pid": 401043,
            "vid": 0,
            "vvalue": "平装"
        }
    ],
    "goods_type": 1,
    "image_url": "",
    "invoice_status": 0,
    "is_customs": 0,
    "is_folt": 0,
    "is_group_pre_sale": 0,
    "is_pre_sale": 0,
    "is_refundable": 1,
    "is_sku_pre_sale": 0,
    "market_price": 5948,
    "order_limit": 999999,
    "outer_goods_id": "9787040456295",
    "oversea_type": 0,
    "pre_sale_time": 0,
    "privacy_delivery": 0,
    "quan_guo_lian_bao": 0,
    "second_hand": 1,
    "shipment_limit_second": 172800,
    "sku_list": [
        {
            "is_onsale": 1,
            "limit_quantity": 999999,
            "multi_price": 1487,
            "out_sku_sn": "9787040456295",
            "price": 1587,
            "quantity": 0,
            "reserve_quantity": 0,
            "sku_id": 1753931570290,
            "sku_pre_sale_time": 0,
            "spec": [
                {
                    "parent_id": 1216,
                    "parent_name": "尺寸",
                    "spec_id": 27632894279,
                    "spec_name": "单本 无附赠 超七天不退换"
                }
            ],
            "thumb_url": "https://img.pddpic.com/open-gw/2025-06-30/59c30d4c-193f-40a3-a639-7af59a381ec5.jpeg",
            "weight": 500
        }
    ],
    "status": 4,
    "tiny_name": "金属工艺学  下  第6",
    "two_pieces_discount": 96,
    "video_gallery": [],
    "warehouse": "",
    "warm_tips": "",
    "zhi_huan_bu_xiu": 0
}
```

## 生成自定义规格接口 -- OutPddAuthSetSpec
### 请求信息
```gotemplate
dll.OutPddAuthSetSpec(specTypeId, specName, accessToken)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| specTypeId | int | 是 | 规格类型ID |
| specName | string | 是 | 规格名称 |
| accessToken | string | 是 | 授权令牌 |
### 响应参数
```json
{
    "parentSpecId": 3820,
    "specName": "全新",
    "specId": 1080396526
}
```

## 修改价格接口 -- OutPddAuthUpdatePrice
### 请求信息
```gotemplate
dll.OutPddAuthUpdatePrice(jsonData)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明              |
|--|--|--|-----------------|
| jsonData | int | 是 | 	价格修改信息JSON字符串  |
### 响应参数
```json
[
  {
    "success": true,
    "msg": "操作成功"
  },
  {
    "success": false,
    "msg": "操作失败"
  }
]
```

## 修改库存接口 -- OutPddAuthUpdateStock
### 请求信息
```gotemplate
dll.OutPddAuthUpdateStock(jsonData)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明              |
|--|--|--|-----------------|
| jsonData | int | 是 | 	价格修改信息JSON字符串  |
### 响应参数
```json
[
  {
    "success": true,
    "msg": "操作成功"
  },
  {
    "success": false,
    "msg": "操作失败"
  }
]
```

## 12.释放C字符串内存--FreeCString
### 请求信息
```gotemplate
dll.FreeCString(str)
```
### 请求参数
| 参数名 | 类型 | 必填 | 说明       |
|--|--|--|----------|
| str | string | 是 | 需要释放的字符串 |
