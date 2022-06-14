# JmsPull

帮助我们定时拉取 JustMySocks 订阅链接并缓存下来的工具

## 来由
居家常备的 JustMySocks

- 订阅链接可能因为网络环境变化而无法访问
- 尽管 Jms 提供了一些镜像地址比如 https://justmysocks5.net 帮助我们继续获取订阅链接
- 当有些时候，哪怕有镜像地址，也有可能获取失败
- 获取失败的情况下一些客户端可能会将服务器信息全部清掉（哪怕还有一两个链接尚能苟延残喘）

为了解决这个问题

- JmsPull 使用同时请求多镜像地址 + 定时任务(5min/一次) 的形式持续获取最新的订阅链接
- 提供一套 API 替代掉 Jms 的订阅链接

## 配置
配置文件位于 ./conf/config.yaml 里面

首次启动的时候会创建配置模板

```yaml
api-key: JmsPull 接口的 key 参数
corn-interval: 5
domain-sub-link: '最近一次拿到的订阅连接(域名服务器)'
domain-sub-link-update-time: "2022-06-14 17:58:00"
ip-sub-link: '最近一次拿到的订阅连接(ip服务器)'
ip-sub-link-update-time: "2022-06-14 17:58:00"
jms-id: justmysocks 订阅链接里面的 id 参数
jms-server: justmysocks 订阅链接里面的 server 参数
```

需要进行配置的是 `jms-id`, `jms-server`, `api-key` 三个参数
corn 自动拉取的间隔可以通过 corn-interval 设置，单位分钟，默认 5 分钟拉取一次 jms 配置

## API 接口

http 服务器运行于 `38888`

请求示例可查看 [api-test.http](api-test.http)

### 订阅链接获取接口
- `/api/sublink`
- 获取缓存的订阅链接， 替代掉 jms 官方订阅获取接口
- 参数
  - `key`
    - 配置文件里面的 api-key 值
  - `type`
    - 可选值为 `all`; `d`, `domain`, `usedomains`, `ip`
      - `all`
        - 获取域名订阅链接 + ip 订阅链接
      - `d`, `domain`, `usedomains`
        - 获取域名订阅链接
      - `ip`
        - 获取 ip 订阅链接
    - 默认为 `ip`
  - `now`
    - 不从缓存信息拿，而是直接访问 jms 的链接获取接口，来拿到订阅信息
    - 可选值为 `1`
    - 使用了该参数之后，`type` 参数不能为 `all` (TODO 会改掉)


### 状态查看接口
- `/api/health`
- 查看当前各个订阅链接状态以及最后更新时间
- 参数
    - `key`: 配置文件里面的 api-key 值
