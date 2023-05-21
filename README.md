# Go_distributed_system

一个基于Go语言的简单分布式系统，实现了基本的分布式服务

![结构图](C:\GoProject\src\Go_distributed_system\images\结构图.png)

![思维导图](C:\GoProject\src\Go_distributed_system\images\思维导图.jpg)

## 特点

- 混合模型：服务注册与健康检查是集中式的，web service采用的是点对点模式，优点是有利于负载均衡，对服务失败的防范更加健壮，缺点是架构更加复杂，且hub的作用范围难以界定
- 组件分为：service注册服务器：管理各类service的注册与service的健康检查、用户门户：实现web逻辑以及API网关、日志服务：实现集中日志的写入、业务服务：实现业务的逻辑以及数据的持久化
- 不使用框架，用go标准库编写，主要用于强化我对go的实战，数据传输使用HTTP，协议为JSON
