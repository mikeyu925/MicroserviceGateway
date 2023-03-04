### Microservices Gateway Project 

> 「Implemented by the Golang」



#### 网络代理

网络代理负责 控制 和 管理 对某个网络主机的访问

- 控制：客户端/服务器流量控制、黑白名单、权限校验和URL重写
- 管理：流量统计、编解码、Header头转换、负载均衡、服务发现、连接池



正向代理「客户端代理」：帮助**客户端**访问无法访问的服务器资源，隐藏用户真实IP 「VPN、浏览器web代理等」

>  步骤：
>
> - 代理服务器接收客户端请求，复制，封装成新请求
> - 发送新请求到真实服务器，接收响应
> - 处理响应并返回客户端
>
> 我这里用的是Mac系统，因此讲一下一些相关的配置：
>
> Mac代理设置：关于本机 -> 偏好设置 -> 网络 -> 高级 -> 代理设置.  或者配置 ~/.bash_profile
>
> Mac查看本机ip：`ifconfig`
>
> 注：配置完正向代理，所有的请求都会先经过正向代理

<img src="./README.assets/image-20230301211017351.png" alt="image-20230301211017351" style="zoom:50%;" />



反向代理「服务端代理」：为**服务器**做负载均衡、缓存、提供安全校验等，隐藏服务器真实IP。「LVS技术、nginx proxy_pass等」

> 

<img src="./README.assets/image-20230301211036283.png" alt="image-20230301211036283" style="zoom:50%;" />





#### HTTP代理







#### WebSocket代理

解决服务端无法主动向客户端推送的问题，应用层协议，兼容HTTP，使用相同的端口

- 真正意义的全双工：C/S地位对等
- 长连接
- 服务端可以主动向客户端发消息

使用HTTP Upgrade机制进行握手

使用TCP作为传输层协议，支持TLS

- ws://host:port/path/query   (80端口)	
- wss://host:port/path/query (443端口)

<img src="./README.assets/image-20230302133807026.png" alt="image-20230302133807026" style="zoom:50%;" />

> 101 代表着切换协议
>
> Websocket所有请求都是GET请求





#### TCP代理

基于流式数据及无状态数据的管理，关注流量控制及请求来源控制

> 参考HTTP代理实现了TCP的代理服务器







#### gRPC代理

> 谷歌出品的高性能、开源、通用的RPC框架
>
> - 基于HTTP/2设计,双向流、流控、头部压缩、单TCP多路复用
> - 面向服务端和移动端：节省空间、省电
> - 支持众多语言C++、Go、Java、Python
> - 使用 protobuf 作为IDL
>
> 基本理念：
>
> 1、定义一个服务，指定其能够被远程调用的方法（即函数，包含参数和返回类型） 
>
> 2、在服务端实现这个接口，并运行一个 gRPC服务器来处理客户端调用 
>
> 3、客户端有一个存根，即跟服务端一样的方法回调被调用的 A 方法，唤醒正在等待响应（阻塞）的客户端调用并返回响应结果
>
> 总结：rpc就是像调用本地函数一样嗲用远程函数

<img src="./README.assets/image-20230302154224558.png" alt="image-20230302154224558" style="zoom:50%;" />









#### Bug记录

- 请求完之后没有自动关闭连接

  <img src="./README.assets/image-20230304185915578.png" alt="image-20230304185915578" style="zoom:50%;" />

  > ```go
  > func (c *conn) serve(ctx context.Context) {
  > 	defer func() {
  > 		if err := recover(); err != nil && err != ErrAbortHandler {
  > 				// ... 省略操作
  > 		}
  > 		// 服务完之后关闭连接
  > 		c.rwc.Close()
  > 	}()
  > 	// ... 其他操作
  > }
  > ```

- 







### 相关面试题整理

---

grpc与http对比，grpc为啥好，基本原理是什么？

> 区别：
>
> - rpc是远程过程调用，就是本地去调用一个远程的函数，而http是通过 url和符合restful风格的数据包去发送和获取数据。
> - rpc的一般使用的编解码协议更加高效，比如grpc使用protobuf编解码。而http的一般使用json进行编解码，数据相比rpc更加直观，但是数据包也更大，效率低下
> - rpc一般用在**服务内部的相互调用**，而http则用于和用户交互。
>
> 相似点：
>
> - 都有类似的机制，例如grpc的metadata机制和http的头机制作用相似，而且web框架，和rpc框架中都有拦截器的概念
> - grpc是基于http2协议，可以实现多路复用的长连接，效率更高

---

