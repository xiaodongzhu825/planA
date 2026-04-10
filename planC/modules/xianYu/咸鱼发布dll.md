##### FreeCString(str *C.char)

接收其他函数返回值之后，释放内存，参考示例

##### 内存释放示例

```go
func example () {
	// ...其他逻辑
    var res = StartServer (configFile *C.char)
    FreeCString(res) //释放内存
}
```



##### StartServer (configFile *C.char)

启动http服务器，参数配置文件路径，不提供默认使用工程根目录config.ini

返回C字符串启动消息，接收后使用FreeCString进行内存释放



##### StopServer

停止HTTP服务器

返回C字符串停止消息，接收后使用FreeCString进行内存释放



##### GetServerStatus

获取服务器当前状态

返回C字符串指针消息，running/stopped，接收后使用FreeCString进行内存释放



##### GetServerAddress

获取服务器监听地址

返回C字符串指针服务器地址消息，未运行返回空串，接收后使用FreeCString进行内存释放



##### ReloadConfig(configFile *C.char)

重新加载配置文件，参数配置文件路径，不提供默认使用根目录config.ini

返回C字符串加载结果消息，接收后使用FreeCString进行内存释放





### 以下都需要传递appid和appSecret ###

##### ExecuteGoodsCreat(bodyJson *C.char, configFile *C.char)

*管道通信直接调用此函数*

执行商品创建操作，参数商品信息，参考示例

返回C字符串指针创建商品结果信息，接收后使用FreeCString进行内存释放



##### 商品信息参考示例

```json
{
  "appId": 1228288260261189,
  "appSecret": "aq9gAwrwp6WGZkMRqKIXmnu2c2uCm82k",
  "token": "",
  "apiShopId": 0,
  "typePlatform": 4,
  "shopId": 0,
  "shopToken": "",
  "shopName": "",
  "province": 210000,
  "city": 210100,
  "district": 210101,
  "typeClass": "",
  "typeGoods": "",
  "catIds": "d14d229692616168b108d382c4e6ea42",
  "shop": [
    {
		"userName": "xy938400231518",
		"province": 210000,
		"city": 210100,
		"district": 210101,
		"title": "牧羊少年奇幻之旅",
		"content": "牧羊少年奇幻之旅",
		"mainImgs": ["https://img.cdn1.vip/i/68cf5cb4e5840_1758420148.webp"],
		"contentImgs": []
    }
  ],
  "stuffStatus": 90,
  "bookData": [
    {
		"ISBN": "9787530217054",
		"Title": "牧羊少年奇幻之旅",
		"Author": "保罗·柯艾略",
		"Publisher": "北京十月文艺出版",
		"itemBizType": 2,
		"spBizType": 24,
		"prices": [199999, 299999],
		"stock": 100,
		"catIds": "22e1d81dc4cf3a25a7f7e02f36b0b49a"
    }
  ],
  "itemKey": "itemAAAAA1111"
}
```



##### ExecuteGoodsPublish(bodyJson *C.char, configFile *C.char)

*管道通信直接调用此函数*

执行商品上架操作，参数上架信息，参考示例

返回C字符串指针行商品上架结果信息，接收后使用FreeCString进行内存释放

##### 上架信息参考示例

```json
{
  "product_id": 1250927879325125,
  "user_name": ["xy938400231518"],
  "specify_publish_time": "",
  "notify_url": ""
}
```



#### 追加下架，改价，擦亮 ####

##### ExecuteGoodsDownShelf(bodyJson *C.char, configFile *C.char) ######

*管道通信直接调用此函数*

执行商品下架操作，参数管家商品ID，参考示例

返回C字符串指针行商品下架结果信息，接收后使用FreeCString进行内存释放

##### 下架信息参考示例 #####

```json
{
  "product_id": 1250927879325125
}
```



##### ExecuteGoodsFlash(bodyJson *C.char, configFile *C.char) #####

*管道通信直接调用此函数*

执行商品擦亮操作，参数管家商品ID，参考示例

返回C字符串指针行商品擦亮结果信息，接收后使用FreeCString进行内存释放

##### 擦亮信息参考示例 #####

```json
{
  "product_id": 1250927879325125
}
```



##### ExecuteGoodsEditPrice(bodyJson *C.char, configFile *C.char) #####

*管道通信直接调用此函数*

执行商品改价操作，参数管家商品ID，参考示例

返回C字符串指针行商品改价结果信息，接收后使用FreeCString进行内存释放

##### 改价信息参考示例（单位:分） #####

```json
{
  "product_id": 1250927879325125,
  "price": 550000,
  "originalPrice": 770000
}
```



##### ExecuteGoodsEditStock(bodyJson *C.char, configFile *C.char)  #####

*管道通信直接调用此函数*

执行商品改库存操作，参数管家商品ID，参考示例

返回C字符串指针行商品改价结果信息，接收后使用FreeCString进行内存释放

##### 改库存信息参考示例（单位:分） #####

```json
{
  "product_id": 1250927879325125,
  "stock": 10
}
```



##### ExecuteSelectGoodsListPrice(bodyJson *C.char, configFile *C.char)  #####

*管道通信直接调用此函数*

查询店铺列表操作，参数管家商品ID，参考示例

返回C字符串指针行商品改价结果信息，接收后使用FreeCString进行内存释放

##### 查询参考示例（单位:分） #####

```json
{
    //online_time 字段可传空
   "online_time": [
        1690300800,
        1690366883
    ], 
  "product_status": 22
}
```

