go install github.com/google/wire/cmd/wire@latest

go install github.com/swaggo/swag/cmd/swag@latest


admin  实例类型
t2.small

mysql  db.t3.small 读写
测试 db.t3.micro


cp dump.rdb /data/cheetah/


下载redis 数据

//事务死锁
SHOW ENGINE INNODB STATUS;
SHOW PROCESSLIST;   
KILL [查询ID或事务ID];

//查询存存储过程
SELECT ROUTINE_NAME
FROM INFORMATION_SCHEMA.ROUTINES
WHERE ROUTINE_TYPE = 'PROCEDURE'
AND ROUTINE_SCHEMA = DATABASE();



# rk-api
//查看内存用量
curl -o heap.pprof http://xxx/debug/pprof/heap
go tool pprof heap.pprof
top

//查看CPU
go tool pprof http://xxx/debug/pprof/profile
top

go tool pprof -http=:7070 http://xxxxxx/debug/pprof/profile 


后台统计页也要截图。对比。   流程 先清理，wingo，九星，流水，游戏返利，充值返利 数据。   这些都是从后台看的。 
用户端的  是第二个流程。会直接清空，然后计算这个月的塞入新的。


insert IGNORE refund_link_game_flow (id,fid, created_at, updated_at, uid, type, currency, number, pc)
SELECT id,id, created_at, updated_at, uid, type, currency, number, pc
FROM flow
WHERE type > 300 limit 2;

replace into

每个月处理，这些表需要删出 上个月的数据，并且备份(也可以采用新建表+日期)
wingo_order,nine_order, refund_game_flow, flow(某些流水)，game_return,recharge_return

另外
hall_invite_relation  中return_cash = 0

<!-- update `user` set cs_game = 0,cs_recharge = 0,withdraw_all =0,recharge_all =0,red_cash = 0 limit 1 -->

{"level":"info","ts":"2024-03-11 11:34:16","msg":"SettlePlayerOrder","order":{"id":523,"UID":8000438,"betType":1,"periodID":"20240311231","ticketNumber":5,"number":3,"betTime":1710136933,"betAmount":40,"fee":2,"delivery":38,"rewardAmount":0,"price":208183,"balance":1290,"status":1,"finishTime":1710137061,"username":"7787808297"}}



# 如果要守护在后台运行
$ nohup ./rk-api &> run.log &
$ tail -f run.log
```

# Linux系统systemd服务配置

可以使用`systemd`配置`rk-api`开机自启，假设可执行文件和相关资源文件放置在`/var/www/rk-api/`目录下，`rk-api`二进制文件需要其他用户可读可执行权限，其余资源文件需要其他用户可读权限，并且已经配置好`config.json`。

在目录`/etc/systemd/system/`下新建文件`rk-api.service`，以下是文件样例。

```ini
[Unit]
Description=rk-api
Documentation=
# 在网络启动完成后运行
After=network.target nss-lookup.target

[Service]
# 使用随机用户执行该服务
DynamicUser=yes
# 指定工作目录
WorkingDirectory=/var/www/rk-api/
# 执行程序
ExecStart=/var/www/rk-api/rk-api

[Install]
WantedBy=multi-user.target
```
保存后使用`systemctl daemon-reload`更新systemd配置文件，使用`systemctl start/stop rk-api`启动/停止服务，使用`systemctl enable/disable rk-api`启用/禁用服务开机自启。

可以使用`journalctl --unit rk-api.service`查看程序日志。

go run cmd/rk-api/main.go -c configs/config-linux.yaml



#REWRITE-START
        if ($host ~ '^dswfrc.xyz'){
            return 302 https://chat.whatsapp.com/Lje7TvIQjwi8ojdbEbgFnG;
        }
#REWRITE-END



#ninestar
#day interest
# game cash return

# 游戏返利，充值返利
# 游戏结算

# 投注金额限制
# good list

#provider 日志


比如。 用户进入游戏商  ，游戏商锁定1000 块，实时给消息让这边扣除用户1000块。 但用户可能没下注就离开了，之后游戏上给消息退回 1000块。   这边就得到2笔 流水   -1000，+1000 。 所以通过流水能计算盈亏 但没法计算下注，因此无法返利。   返利就得定时去查询他们的下单情况，进行返利。


普通分销 --- 


Well kaka, [2025/2/24 15:58]
盈亏：当局牌输赢的钱（赢的玩家会抽水，目前抽水比例为5%）
有效投注：当局牌输赢的钱的绝对值（抽水前）

举例子1，我和A,B  打了一局，我输12，A输12，B赢24（扣除抽水后为22.8）
因此，这一局的结果：
我盈亏为-12
有效投注为12

A盈亏为-12
有效投注为12

B盈亏为22.8
有效投注为24

solana, [2025/2/24 16:00]
如果 score > 0    有效投注= score  ,盈亏 = score -tax 。    如果score <0  ,有效投注 = abs(score),盈亏 = score .