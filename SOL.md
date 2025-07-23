API 游戏列表

1. 热度
2. 游戏名
3. 游戏图标 地址
3. 游戏公司logo 地址
4. 当前在玩人数
5. 游戏类别(公司棋牌,区块链，外接游戏)
6. 游戏评分Rating
7. 游戏描述 Description 


type Game struct {
    ID              uint      `gorm:"primaryKey"`
    Name            string    `gorm:"size:255"`
    IconURL         string    `gorm:"size:512"`
    CompanyLogoURL  string    `gorm:"size:512"`
    CurrentPlayers  int
    Category        string    `gorm:"size:255"`
    Heat            int
    CreatedAt       time.Time
    UpdatedAt       time.Time
}



API 排行榜(个人,所有人)

1.游戏类型
2.输赢金额


上面是2个前端需求，我 需要在前端页面展示 3类游戏，还有可以选择全部。  列表项显示游戏  。

另外显示排行榜。  我用的golang  gorm , gin 。  该怎么设计api , gorm。  另外帮我考虑下 数据结构 未来可能扩充的字段。




1.获取授权链接


考虑平台停机  


ssh -i "jhkj_admin" root@172.20.20.1 -o ServerAliveInterval=60 -o ServerAliveCountMax=10





VerifyGithubUser       

 {"oauthUser": "github.User{Login:\"stripluagio\", ID:12146485, NodeID:\"MDQ6VXNlcjEyMTQ2NDg1\", AvatarURL:\"https://avatars.githubusercontent.com/u/12146485?v=4\", HTMLURL:\"https://github.com/stripluagio\", GravatarID:\"\", Blog:\"\", PublicRepos:11, PublicGists:1, Followers:0, Following:0, CreatedAt:github.Timestamp{2015-04-28 02:54:12 +0000 UTC}, UpdatedAt:github.Timestamp{2024-12-20 05:14:33 +0000 UTC}, Type:\"User\", SiteAdmin:false, TotalPrivateRepos:1, OwnedPrivateRepos:1, PrivateGists:0, DiskUsage:169, Collaborators:0, Plan:github.Plan{Name:\"free\", Space:976562499, Collaborators:0, PrivateRepos:10000}, URL:\"https://api.github.com/users/stripluagio\", EventsURL:\"https://api.github.com/users/stripluagio/events{/privacy}\", FollowingURL:\"https://api.github.com/users/stripluagio/following{/other_user}\", FollowersURL:\"https://api.github.com/users/stripluagio/followers\", GistsURL:\"https://api.github.com/users/stripluagio/gists{/gist_id}\", OrganizationsURL:\"https://api.github.com/users/stripluagio/orgs\", ReceivedEventsURL:\"https://api.github.com/users/stripluagio/received_events\", ReposURL:\"https://api.github.com/users/stripluagio/repos\", StarredURL:\"https://api.github.com/users/stripluagio/starred{/owner}{/repo}\", SubscriptionsURL:\"https://api.github.com/users/stripluagio/subscriptions\"}"}


{"oauthUser": "github.User{Login:\"stripluagio\", ID:12146485, NodeID:\"MDQ6VXNlcjEyMTQ2NDg1\", AvatarURL:\"https://avatars.githubusercontent.com/u/12146485?v=4\", HTMLURL:\"https://github.com/stripluagio\", GravatarID:\"\", Blog:\"\", PublicRepos:11, PublicGists:1, Followers:0, Following:0, CreatedAt:github.Timestamp{2015-04-28 02:54:12 +0000 UTC}, UpdatedAt:github.Timestamp{2024-12-20 05:14:33 +0000 UTC}, Type:\"User\", SiteAdmin:false, TotalPrivateRepos:1, OwnedPrivateRepos:1, PrivateGists:0, DiskUsage:169, Collaborators:0, Plan:github.Plan{Name:\"free\", Space:976562499, Collaborators:0, PrivateRepos:10000}, URL:\"https://api.github.com/users/stripluagio\", EventsURL:\"https://api.github.com/users/stripluagio/events{/privacy}\", FollowingURL:\"https://api.github.com/users/stripluagio/following{/other_user}\", FollowersURL:\"https://api.github.com/users/stripluagio/followers\", GistsURL:\"https://api.github.com/users/stripluagio/gists{/gist_id}\", OrganizationsURL:\"https://api.github.com/users/stripluagio/orgs\", ReceivedEventsURL:\"https://api.github.com/users/stripluagio/received_events\", ReposURL:\"https://api.github.com/users/stripluagio/repos\", StarredURL:\"https://api.github.com/users/stripluagio/starred{/owner}{/repo}\", SubscriptionsURL:\"https://api.github.com/users/stripluagio/subscriptions\"}"}



设计通讯

基础的通知给游戏 webhook 

游戏每隔几秒拉取一次数据。

给平台的webhook  下注 ,派奖 等。

平台停止服务，游戏停止服务。


告诉 平台 流水变化。扣钱消息




有多个H5平台独立运营的，有平台数据库，平台用户，平台服务，平台运营后台。另外有一个游戏团队，做了一个h5游戏，有游戏数据库，也有h5游戏前端页，后端服务，游戏后台。  

现在老板提出需求。 希望H5游戏，能够接入到多个H5平台。  
要求使用平台的用户登陆 平台页面有游戏图标点击进入游戏(iframe)    
原先独立的游戏用户依然能够正常登陆，同时原先独立的游戏页面还增加从平台登陆按钮，输入平台名，平台账号密码也可以登陆。
游戏内金币和平台同步，个人信息也要跟平台一致。当平台金币发生改变，游戏那边也得同步。 当游戏内金币改变，平台也得同步。平台的后台查询用户金币和游戏后台查询的要一致。

多个平台，只愿意提供金额查询。另外希望游戏有金额输赢(需要提供详细流水方便平台记录)等变动时通知平台
余额(总余额，冻结余额，可用余额)

建议提供适配层(帮我取个更好的名字)， 适配层无法直接访问game,platform的DB。都通过sdk HTTP 调用
调用transfer (参数变动金额，详细流水来源) ，唯一的事务ID，重传(平台可能关闭)。 
大概流程分布式锁，新的事务ID，查询游戏平台金币，得到最余额，通知平台变更(流水详细)，平台挂机则指数重传，通知游戏同步余额， 返回最新余额给游戏。
协调 game 和platform SDK(调用时接口会加密)。 

当平台 需要轮询余额，当发生改变时通知 平台更新最新余额。  

提供适配层的golang完整代码。



还有冻结解冻通知。所以当平台本身金额变化的时候，只能通过金额查询，去让游戏知晓。



简体中文‌：zh-CN
‌繁体中文‌：zh 或 zh-TW
‌英语‌：en 或 en-US