# Plan A
## 目录结构
```gotemplate 
controller                          逻辑控制
controlState                        全局状态控制
        |-lock                      状态锁
        |-serviceAlive              服务存活状态
export                              导出的csv文件
initialization                      初始化
        |-config                    初始化配置文件
        |-cron                      初始化定时任务
        |-golabl                    初始化全局变量
        |-middle                    初始化中间件
        |-mysql                     初始化mysql数据库
        |-redis                     初始化redis数据库
        |-router                    初始化路由
        |-sqlite                    初始化sqlite数据库
        |-validator                 初始化验证器
        |-init.go                   初始化文件
logs                                日志
modules                             DLL模块
planB                               模块B
rep                                 工厂模式接口
router                              路由
service                             服务（针对数据库相关操作）
tool                                工具
type                                结构体
        |-mysql                     mysql结构体
        |-redis                     redis结构体
        |-sqlite                    sqlite结构体
        |-validator                 验证器结构体
validator                           验证器
config.yaml                         配置文件
taskDb.db                           sqlite数据库（自动创建）
```