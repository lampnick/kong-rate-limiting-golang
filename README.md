# kong-rate-limiting-golang

#### 特性
- 使用golang编写的一个kong限流插件
- 限流支持并发
- 精准限流

#### 插件开发流程
1. 定义一个结构体类型保存配置文件
```
用lua写的插件通过schema来指定怎样读取和验证来自数据库和Admin API中的配置数据。由于GO是静态类型语言，都需要用配置结构体定义
type MyConfig struct {
    Path   string //这里配置的会在konga添加插件时显示出来
    Reopen bool
}
公有属性将会被配置数据填充，如果希望在数据库中使用不同的名称，可以使用encoding/json加tag的方式
type MyConfig struct {
    Path   string `json:my_file_path`
    Reopen bool   `json:reopen`
}
```
2. 使用New()创建一个实例
```
你的go插件必须定义一个名叫New的函数来创建这个类型的实例并返回一个interface{}类型
func New() interface{} {
    return &MyConfig{}
}
```
3. 添加处理阶段方法
```
你可以在请求的生命周期的各个阶段实现自定义的逻辑。如在"access"阶段,定义一个名为Access的方法
func (conf *MyConfig) Access (kong *pdk.PDK) {
  ...
}
你可以实现自定义逻辑的阶段方法有如下几种
Certificate
Rewrite
Access
Preread
Log
```
4. 编译go插件
```
go build -buildmode plugin  nick-rate-limiting.go
```
5. 将生成的.so文件放到go_plugins_dir定义的目录中
```.env
cp nick-rate-limiting.so ../plugins/
```
6. 重启kong（平滑重启）
```
kong prepare && kong reload
```
#### 部署
- kong配置文件修改
```
plugins = bundled,nick-rate-limiting
go_plugins_dir = /etc/kong/plugins
go_pluginserver_exe = /usr/local/bin/go-pluginserver
```
- 构建go-pluginserver
```
在go-pluginserver中执行go build github.com/Kong/go-pluginserver
会生成 go-pluginserver文件，复制到/usr/local/bin目录
```
-  编译go插件
```
go build -buildmode plugin  custom-rate-limiting.go && cp custom-rate-limiting.so
```
- 将生成的.so文件放到go_plugins_dir定义的目录中
```.env
cp custom-rate-limiting.so ../plugins/
```
- 重启kong
```
kong prepare && kong reload
- 在konga中配置插件
- 测试请求是否正常，规则是否生效
```

### 相关截图
- konga json配置
    ```
    [{
        "type": "header,query,body",
        "key": "orderId",
        "value": "orderId1,orderId2,orderId3"
    }, {
        "type": "query",
        "key": "username",
        "value": "nick,jack,star"
    }]
    ```
- konga 配置
![image](http://www.lampnick.com/wp-content/uploads/2020/09/kong-config.png)

- postman 显示header
![image](http://www.lampnick.com/wp-content/uploads/2020/09/kong-post-header.png)

- siege压测效果图
![image](http://www.lampnick.com/wp-content/uploads/2020/09/rate-limiting.png)