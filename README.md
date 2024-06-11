# GeeRPC
RPC(Remote Procedure Call，远程过程调用)是一种计算机通信协议，允许调用不同进程空间的程序。
RPC 的客户端和服务器可以在一台机器上，也可以在不同的机器上。
程序员使用时，就像调用本地程序一样，无需关注内部的实现细节。

不同的应用程序之间的通信方式有很多，比如浏览器和服务器之间广泛使用的基于 HTTP 协议的 Restful API。
与 RPC 相比，Restful API 有相对统一的标准，因而更通用，兼容性更好，支持不同的语言。
HTTP 协议是基于文本的，一般具备更好的可读性。但是缺点也很明显：

Restful 接口需要额外的定义，无论是客户端还是服务端，都需要额外的代码来处理，而 RPC 调用则更接近于直接调用。
基于 HTTP 协议的 Restful 报文冗余，承载了过多的无效信息，而 RPC 通常使用自定义的协议格式，减少冗余报文。
RPC 可以采用更高效的序列化协议，将文本转为二进制传输，获得更高的性能。
因为 RPC 的灵活性，所以更容易扩展和集成诸如注册中心、负载均衡等功能。

2 RPC 框架需要解决什么问题
RPC 框架需要解决什么问题？或者我们换一个问题，为什么需要 RPC 框架？

我们可以想象下两台机器上，两个应用程序之间需要通信，那么首先，需要确定采用的传输协议是什么？
如果这个两个应用程序位于不同的机器，那么一般会选择 TCP 协议或者 HTTP 协议；那如果两个应用程序位于相同的机器，
也可以选择 Unix Socket 协议。传输协议确定之后，还需要确定报文的编码格式，比如采用最常用的 JSON 或者 XML，
那如果报文比较大，还可能会选择 protobuf 等其他的编码方式，甚至编码之后，再进行压缩。
接收端获取报文则需要相反的过程，先解压再解码。

解决了传输协议和报文编码的问题，接下来还需要解决一系列的可用性问题，例如，连接超时了怎么办？
是否支持异步请求和并发？

如果服务端的实例很多，客户端并不关心这些实例的地址和部署位置，只关心自己能否获取到期待的结果，
那就引出了注册中心(registry)和负载均衡(load balance)的问题。简单地说，即客户端和服务端互相不感知对方的存在，
服务端启动时将自己注册到注册中心，客户端调用时，从注册中心获取到所有可用的实例，选择一个来调用。
这样服务端和客户端只需要感知注册中心的存在就够了。
注册中心通常还需要实现服务动态添加、删除，使用心跳确保服务处于可用状态等功能。
再进一步，假设服务端是不同的团队提供的，如果没有统一的 RPC 框架，各个团队的服务提供方就需要各自
实现一套消息编解码、连接池、收发线程、超时处理等“业务之外”的重复技术劳动，造成整体的低效。
因此，“业务之外”的这部分公共的能力，即是 RPC 框架所需要具备的能力。

3 关于 GeeRPC
Go 语言广泛地应用于云计算和微服务，成熟的 RPC 框架和微服务框架汗牛充栋。
grpc、rpcx、go-micro 等都是非常成熟的框架。一般而言，RPC 是微服务框架的一个子集，
微服务框架可以自己实现 RPC 部分，当然，也可以选择不同的 RPC 框架作为通信基座。
考虑性能和功能，上述成熟的框架代码量都比较庞大，而且通常和第三方库，
例如 protobuf、etcd、zookeeper 等有比较深的耦合，难以直观地窥视框架的本质。
GeeRPC 的目的是以最少的代码，实现 RPC 框架中最为重要的部分，帮助大家理解 RPC 框架在设计时需要考虑什么。
代码简洁是第一位的，功能是第二位的。
因此，GeeRPC 选择从零实现 Go 语言官方的标准库 net/rpc，并在此基础上，新增了协议交换(protocol exchange)、
注册中心(registry)、服务发现(service discovery)、负载均衡(load balance)、超时处理(timeout processing)等特性。
分七天完成，最终代码约 1000 行。

超时处理是 RPC 框架一个比较基本的能力，如果缺少超时处理机制，无论是服务端还是客户端都容易因为网络或其他错误导致挂死，
资源耗尽，这些问题的出现大大地降低了服务的可用性。因此，我们需要在 RPC 框架中加入超时处理的能力。

纵观整个远程调用的过程，需要客户端处理超时的地方有：
与服务端建立连接，导致的超时
发送请求到服务端，写报文导致的超时
等待服务端处理时，等待处理导致的超时（比如服务端已挂死，迟迟不响应）
从服务端接收响应时，读报文导致的超时

需要服务端处理超时的地方有：
读取客户端请求报文时，读报文导致的超时
发送响应报文时，写报文导致的超时
调用映射服务的方法时，处理报文导致的超时

GeeRPC 在 3 个地方添加了超时处理机制。分别是：
1）客户端创建连接时
2）客户端 Client.Call() 整个过程导致的超时（包含发送报文，等待处理，接收报文所有阶段）
3）服务端处理报文，即 Server.handleRequest 超时。

支持 HTTP 协议需要做什么？
使用HTTP协议的CONNECT方法代理客户端和服务端的连接服务。以明文的方式向代理服务器发送连接请求，然后代理服务器返回连接成功。
之后客户端和服务端建立TCP握手并交换加密数据。代理服务器只负责传输彼此的数据包，并不能读取具体数据内容。
事实上，这个过程其实是通过代理服务器将 HTTP 协议转换为 HTTPS 协议的过程。

负载均衡策略
假设有多个服务实例，每个实例提供相同的功能，为了提高整个系统的吞吐量，每个实例部署在不同的机器上。
客户端可以选择任意一个实例进行调用，获取想要的结果。那如何选择呢？取决了负载均衡的策略。
对于 RPC 框架来说，我们可以很容易地想到这么几种策略：
随机选择策略 - 从服务列表中随机选择一个。
轮询算法(Round Robin) - 依次调度不同的服务器，每次调度执行 i = (i + 1) mode n。
加权轮询(Weight Round Robin) - 在轮询算法的基础上，为每个服务实例设置一个权重，高性能的机器赋予更高的权重，也可以根据服务实例的当前的负载情况做动态的调整，例如考虑最近5分钟部署服务器的 CPU、内存消耗情况。
哈希/一致性哈希策略 - 依据请求的某些特征，计算一个 hash 值，根据 hash 值将请求发送到对应的机器。一致性 hash 还可以解决服务实例动态添加情况下，调度抖动的问题。一致性哈希的一个典型应用场景是分布式缓存服务。

注册中心：客户端无需知道服务端的存在
注册中心的好处在于，客户端和服务端都只需要感知注册中心的存在，而无需感知对方的存在。更具体一些：
服务端启动后，向注册中心发送注册消息，注册中心得知该服务已经启动，处于可用状态。一般来说，服务端还需要定期向注册中心发送心跳，证明自己还活着。
客户端向注册中心询问，当前哪天服务是可用的，注册中心将可用的服务列表返回客户端。
客户端根据注册中心得到的服务列表，选择其中一个发起调用。
如果没有注册中心，就像 GeeRPC 第六天实现的一样，客户端需要硬编码服务端的地址，而且没有机制保证服务端是否处于可用状态。当然注册中心的功能还有很多，比如配置的动态同步、通知机制等。比较常用的注册中心有 etcd、zookeeper、consul，一般比较出名的微服务或者 RPC 框架，这些主流的注册中心都是支持的。