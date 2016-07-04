UIC
===

本项目fork自 xiaomi/openfalcon/fe项目，在此基础上做了符合本公司的update。大致有：
  - 从ldap中验证用户信息，不允许用户注册；
  - 取消密码更改，将相关字段设为只读；
  - 从ldap中搜索用户，可为普通永固设置admin管理员权限，不可删除用户；
  - 升级用户组管理，主要增加了用户组的组管理员的角色设定；
  - 取消普通用户对其他用户信息的编辑权限，只有管理员才有读写权限；
  - 设置用户组管理员的操作权限，只有组管理员才可对本组进行读写权限；
  
待做：
  - 规范化代码，使全部代码符合golint规范；
  
===

鉴于很多用户反馈UIC太难安装了（虽然我觉得其实很容易……），用Go语言重新实现了一个，也就是这个falcon-fe了。

另外，监控系统组件比较多，有不少web组件，比如uic、portal、alarm、dashboard，没有一个统一的地方汇总查看，falcon-fe也做了一些快捷配置，类似监控系统的hao123导航了

# 安装Go语言环境

```
cd ~
wget http://dinp.qiniudn.com/go1.4.1.linux-amd64.tar.gz
tar zxvf go1.4.1.linux-amd64.tar.gz
mkdir -p workspace/src
echo "" >> .bashrc
echo 'export GOROOT=$HOME/go' >> .bashrc
echo 'export GOPATH=$HOME/workspace' >> .bashrc
echo 'export PATH=$GOROOT/bin:$GOPATH/bin:$PATH' >> .bashrc
echo "" >> .bashrc
source .bashrc
```

# 编译安装fe模块

fe模块使用了beego框架，最近beego升级了，api不向后兼容，好多朋友出现安装fe失败的问题。这里有个临时方案，我保存了一个老版本的beego

```
cd $GOPATH/src/github.com/astaxie
rm -rf beego
git clone https://git.coding.net/ulricqin/beego.bak.git beego
```

之后找时间使用godep处理一下依赖

```
cd $GOPATH/src/github.com/open-falcon
git clone https://github.com/open-falcon/fe.git
cd fe
go get ./...
./control build
./control start
```

# 配置介绍

```
{
    "log": "debug",
    "company": "MI", # 填写自己公司的名称，用于生成联系人二维码
    "http": {
        "enabled": true,
        "listen": "0.0.0.0:1234" # 自己随便搞个端口，别跟现有的重复了，可以使用8080，与老版本保持一致
    },
    "cache": {
        "enabled": true,
        "redis": "127.0.0.1:6379", # 这个redis跟judge、alarm用的redis不同，这个只是作为缓存来用
        "idle": 10,
        "max": 1000,
        "timeout": {
            "conn": 10000,
            "read": 5000,
            "write": 5000
        }
    },
    "salt": "0i923fejfd3", # 搞一个随机字符串
    "canRegister": true,
    "ldap": {
        "enabled": false,
        "addr": "ldap.example.com:389",
        "baseDN": "dc=example,dc=com",
        "bindDN": "cn=mananger,dc=example,dc=com",#允许匿名查询的话，填""值即可
        "bindPasswd": "12345678",
        "userField": "uid", #用于认证的属性，通常为 uid 或 sAMAccountName(AD)。也可以使用诸如mail的属性，这样认证的用户名就是邮箱(前提ldap里有)
        "attributes": ["sn","mail","telephoneNumber"] #数组顺序重要，依次为姓名，邮箱，电话在ldap中的属性名。fe将按这些属性名去ldap中查询新用户的属性，并插入到fe的数据库内。
    },
    "uic": {
        "addr": "root:password@tcp(127.0.0.1:3306)/fe?charset=utf8&loc=Asia%2FChongqing",
        "idle": 10,
        "max": 100
    },
    "shortcut": {
        "falconPortal": "http://11.11.11.11:5050/", 浏览器可访问的portal地址
        "falconDashboard": "http://11.11.11.11:7070/", 浏览器可访问的dashboard地址
        "falconAlarm": "http://11.11.11.11:6060/" 浏览器可访问的alarm的http地址
    }
}
```

# 设置root账号的密码

该项目中的注册用户是有不同角色的，目前分三种角色：普通用户、管理员、root账号。系统启动之后第一件事情应该是设置root的密码，浏览器访问：http://fe.example.com/root?password=abc （此处假设你的项目访问地址是fe.example.com，也可以使用ip）,这样就设置了root账号的密码为abc。普通用户可以支持注册。
