### 测试步骤

三台机器：

- A, 没有公网 ip
- B, 没有公网 ip；且 A、B 无法连通
- C, 有公网 ip；ip 是：103.61.39.95

C 上运行：

```
./ping -listen :2222
```

A 上运行：

```
./ping -listen :2222 -connect 103.61.39.95:2222

# 它的 Router identity 是：`42aae0bc14b06db0895f96d0d2c29f622e3ce91f5c6068fd387db642f085dfbf`
```

B 上运行：

```
go run ping.go -listen :2222 -connect 103.61.39.95:2222 -pubkey 42aae0bc14b06db0895f96d0d2c29f622e3ce91f5c6068fd387db642f085dfbf
```

上面 A 和 C 连接了；B 也和 C 连接了；期望 B 可以 ping A；但实际无法 ping 通。
