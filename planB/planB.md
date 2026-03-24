# Plan B
## 目录结构
```gotemplate 
dispatcher                             具体执行的平台操作（工厂模式，发布商品、上下架等）
        |-kongfuzi                     孔夫子
        |-pinduoduo                    拼多多
        |-xianyu                       闲鱼
initialization                         初始化
        |-config                       初始化配置文件
        |-golabl                       初始化全局变量
        |-platform                     初始化任务平台（拼多多、闲鱼等）
        |-pool                         初始化协程池
        |-redis                        初始化redis
        |-speed                        初始化限速器
        |-task                         初始化任务（获取header与footer）
        |-taskType                     初始化任务类型（发布商品、上下架等）
        |-init.go                      初始化文件
interfaces                             工厂模式接口
logic                                  逻辑执行
modules                                DLL模块
service                                服务（针对数据库相关操作）
tool                                   工具
type                                   结构体
        |-pinduoduo                    拼多多结构体
        |-xianyu                       闲鱼结构体
validation                             验证器


```