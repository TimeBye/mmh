# 快速开始

- [如何安装](#如何安装)
- [如何使用](#如何使用)
- [配置调整](#配置调整)

## 如何安装

目前官方仓库提供了 macOS 的预编译二进制文件，Intel 与 Apple Silicon 均可使用；推荐下载 universal 版本

``` sh
# 请自行查看仓库最新 Release 版本号并替换此版本号
export MMH_VERSION=v1.6.0
# 下载
wget https://github.com/mritd/mmh/releases/download/${MMH_VERSION}/mmh_darwin_universal
# 增加可执行权限
chmod +x mmh_darwin_universal
# 安装 mmh 并自动创建 mcp/mcs/mcx/mec/mgo/mping/mtun 快捷命令
./mmh_darwin_universal install
```

`install` 默认安装到当前二进制文件所在目录，不写入系统目录，无需 sudo。如需安装到 `/usr/local/bin` 或其他目录，可通过 `--dir` 指定；目录不在 `PATH` 中时会给出提示：

``` sh
# 安装到 /usr/local/bin（需要 sudo）
sudo ./mmh_darwin_universal install --dir /usr/local/bin

# 安装到用户目录
./mmh_darwin_universal install --dir "$HOME/.local/bin"
```

Intel 芯片也可下载 `mmh_darwin_amd64`，Apple Silicon 芯片也可下载 `mmh_darwin_arm64`，两者与 universal 版本用法一致。

## 如何使用

mmh 安装完成后可直接在命令行运行 `mmh` 命令进行预览；**命令执行后会自动生成样例配置文件，样例配置文件默认存储在 `~/.mmh` 目录。**

``` sh
➜  ~ mmh
Select Login Server (1/1):
┌───┬───┬─────────┬──────┬────────────┬──────┐
│   │ # │  NAME   │ USER │  ADDRESS   │ PORT │
├───┼───┼─────────┼──────┼────────────┼──────┤
│ » │ 1 │ prod11  │ root │ 10.10.4.11 │ 22   │
│   │ 2 │ prod12  │ root │ 10.10.4.12 │ 22   │
└───┴───┴─────────┴──────┴────────────┴──────┘
↑/↓ 选择  ←/→ 翻页  1-9 快速跳转  Enter 确认  q 退出
```

在配置正确的情况下选择任意一个服务器回车即可完成登录，**不过样例配置中的服务器是无法链接的，你需要自行调整增加真实服务器配置。**

## 配置调整

**mmh 默认配置信息存储在 `~/.mmh` 目录，其中默认创建的 `default.yaml` 配置文件为服务器列表配置，调整此配置添加真实服务器即可。**

``` yaml
basic:
  user: root
  password: password
max_proxy: 5
servers:
- name: prod11
  address: 10.10.4.11
  proxy: prod12
- name: prod12
  address: 10.10.4.12
```

### basic

basic 段中包含了一些默认配置，在 servers 段中填写的真实服务器如果缺少相应设置，那么默认会通过 basic 中的字段进行填充

``` yaml
basic:
  user: root
  password: "123456789"
servers:
- name: prod11
  address: 10.10.4.11
```

**以上配置中如果连接 `prod11` 服务器，那么默认用户名为 `root`，密码为 `123456789`；端口不写的情况下全局默认为 `22`。**

### servers

servers 段为一个数组结构，每一个数组元素都视为一个服务器配置，每个服务器可以进行各种自定义配置。

### servers.proxy

每个服务器配置中的 `proxy` 字段用于实现无限跳板功能，**proxy 字段指定了连接本服务器前需要先跳转的服务器:**

``` yaml
basic:
  user: root
  password: "123456789"
servers:
- name: prod11
  address: 10.10.4.11
  proxy: prod12
- name: prod12
  address: 10.10.4.12
  proxy: prod13
- name: prod13
  address: 10.10.4.13
```

以上配置中，连接 `prod11` 服务器时由于其指定了 `proxy` 为 `prod12`，则 mmh 将首先尝试 `prod12`；
其后由于 `prod12` 指定了 `proxy` 为 `prod13`，则 mmh 接着会尝试连接 `prod13`；其实际连接顺序为:

``` sh
本地 -> prod13 -> prod12 -> prod11
```

按照此种模式配置下，mmh 会不断递归寻找需要连接的服务，从而理论上可以实现无限跳板；**不过为了安全
保证(防止用户误配置)，mmh 默认只允许最多 5 个跳板机穿透(可通过配置调整)，超过 5 个将会导致连接失败。**

[首页](.) | [上一页](01-overview) | [下一页](03-config)
