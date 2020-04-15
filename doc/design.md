# 单向匿名聊天

## 思路

通过 Telegram Bot 在 Web 端和 Telegram 账户之间转发消息。

### 完整流程

1. 服务端初始化 Bot，并启动服务。
2. Web 端打开首页，连接上服务端的 Websocket。
3. 服务端为该 Web 端，生成唯一的随机用户名。
4. Web 端发送消息。
5. 服务端接受到消息，带着用户名发送给指定的 Telegram 账户。消息发送成功后，记录下用户名和该消息 ID 的关系。
6. Telegram 账户回复该消息，服务端根据回复消息的 ID 找到 Websocket 客户端，并推送消息。

## 问题

### 如何标示不同用户

通过 [faker](https://github.com/bxcodec/faker) 生成用户名。

### 在转发消息回 Web 端时，如何区分不同用户

在内存里保存一个 Map，以随机的用户名为 Key，以 Telegram 的消息 ID 组成的数组为值。

### 清理资源

#### Web 端关闭

1. 清理 Websocket 连接。
2. 清理 Web 端相对应的用户名与消息 ID 关系。

#### 服务端关闭

1. 关闭 Goroutines。
