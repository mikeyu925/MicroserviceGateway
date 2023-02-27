### Microservices Gateway Project 

> 「Implemented by the Goang」



#### 网络代理

网络代理负责 控制 和 管理 对某个网络主机的访问

- 控制：客户端/服务器流量控制、黑白名单、权限校验和URL重写
- 管理：流量统计、编解码、Header头转换、负载均衡、服务发现、连接池



正向代理「客户端代理」：帮助客户端访问无法访问的服务器资源，隐藏用户真实IP 「VPN、浏览器web代理等」



 反向代理「服务端代理」：为服务器做负载均衡、缓存、提供安全校验等，隐藏服务器真实IP。「LVS技术、nginx proxy_pass等」