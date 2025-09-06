- 查看 show （所有数据为列表）
    - 交易所 exchanges
    - 用户 users
    - 资产 assets
    - 订单 orders（默认未成交列表） orders
    - 仓位 positions（默认未平仓列表）
    - 策略 strategies（所有可用策略，增加空策略，直接下单）
    - 标的 symbols（可选参数 标的名称）
    - 策略订单 ss （默认未提交订单的策略订单列表）

- 激活 use
  - 交易所 exchanges
  - 用户 users

- 设置 update
    - 杠杆 leverage
    - 全逐仓 margin-type

- 创建 create
  - 用户 users
  - 标的 symbols（每个交易所独立增加，可批量增加，可选参数 杠杆/全逐仓）
  - 策略订单 ss

- 关闭 close
    - 仓位 positions
    - 策略订单 ss

- 删除 delete
    - 用户 users
    - 标的 symbols（每个交易所独立增加，可批量增加，可选参数 杠杆/全逐仓）
    - 策略订单 ss 

```code
foxflow[okx:user1]> create users username=user1 ak=xxxx sk=xxxx trade_type=xxxx
foxflow[okx:user1]> create users --username=user1 --ak=xxxx --sk=xxxx --trade_type=xxxx
foxflow[okx:user1]> create ss --symbol=BTC/USDT --side=buy --posSide=long --px=200 --sz=10 --stry=s1:v>100
```


