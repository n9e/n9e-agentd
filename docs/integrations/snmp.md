# SNMP 

## 概述
简单网络管理协议 (SNMP)是用于监视网络连接设备（例如路由器、交换机、服务器和防火墙）的标准。从网络设备收集 SNMP 指标。

SNMP 使用 sysObjectID（系统对象标识符）来唯一标识设备，使用 OID（对象标识符）来唯一标识管理对象。OID 遵循分层树模式：根下是 ISO，编号为 1。下一级是 ORG，编号为3，依此类推，每个级别之间用 . 分隔。

MIB（Management Information Base）充当 OID 和人类可读名称之间的转换器，并组织层次结构的子集。由于树的结构方式，大多数 SNMP 值都以相同的对象集开头：

- 1.3.6.1.1: (MIB-II) 一种保存系统信息（如正常运行时间、接口和网络堆栈）的标准。
- 1.3.6.1.4.1：保存供应商特定信息的标准。

## 安装
 [n9e-agentd](https://github.com/n9e/n9e-agentd) 代理程序中已经包含 SNMP 采集器，无需额外安装。

## 配置
agentd 网络设备监控支持从单个设备收集指标，或自动发现整个子网上的设备（网络内设备的IP必须唯一）。

可以根据网络上存在的设备数量以及网络的动态程度（添加或删除设备的频率）选择适合的采集策略：

- 对于小型且大部分为静态的网络，请参阅 [监控单个设备](#监控单个设备)。
- 对于更大的或动态的网络，请参阅 [自动发现](#自动发现)。

无论采用何种采集策略，都可以利用 agentd 的 sysObjectID 映射设备配置文件自动从您的设备收集相关指标。


##### 监控单个设备

将 IP 地址和任何其他设备元数据（作为标签）snmp.d/conf.yaml包含在代理配置目录conf.d/根文件夹中的文件中。有关所有可用的配置选项，请参阅示例 snmp.d/conf.yaml。

1. SNMPv2
```yaml
# /opt/n9e/agentd/conf.d/snmp.d/conf.yaml
initConfig:
  loader: core
instances:
- ipAddress: "1.2.3.4"
  communityString: “sample-string”
  tags:
    - "key1:val1"
    - "key2:val2"
```

2. SNMPv3
```yaml
# /opt/n9e/agentd/conf.d/snmp.d/conf.yaml
initConfig:
  loader: core
instances:
- ipAddress: "1.2.3.4"
  snmpVersion: 3			# optional, if omitted we will autodetect which version of SNMP you are using
  user: "user"
  authProtocol: "fakeAuth"
  authKey: "fakeKey"
  #privProtocol:
  #privKey:
  tags:
    - "key1:val1"
    - "key2:val2"
```

- 重启 agentd
```bash
sudo systemctl restart n9e-agentd
```

##### 自动发现

指定单个设备的替代方法是使用 Autodiscovery 来自动发现网络上的所有设备。

自动发现配置的网段，并检查来自目标设备的响应。然后，agentd 代理查找 sysObjectID 发现的设备并找到设备对应 的  [配置文件](https://github.com/n9e/n9e-agentd/tree/main/misc/conf.d/snmp.d/profiles)。配置文件中包含各种设备收集的预定义指标列表。

将自动发现与网络设备监控结合使用：
- 编辑 agentd.yaml 配置文件，设置 要扫描的所有子网。以下示例提供了自动发现所需的参数、默认值。

1. SNMPv2
```yaml
# /opt/n9e/agentd/etc/agentd.yaml
agent:
  listeners:
    - name: snmp
  snmpListener:
    workers: 100 # number of workers used to discover devices concurrently
    discoveryInterval: 3600 # interval between each autodiscovery in seconds
    configs:
      - network: 1.2.3.4/24 # CIDR notation, we recommend no larger than /24 blocks
        version: 2
        port: 161
        community: ***
        tags:
          - "key1:val1"
          - "key2:val2"
        loader: core # use SNMP corecheck implementation
      - network: 2.3.4.5/24
        version: 2
        port: 161
        community: ***
        tags:
          - "key1:val1"
          - "key2:val2"
        loader: core
```

2. SNMPv3
```yaml
# /opt/n9e/agentd/etc/agentd.yaml
agent:
  listeners:
    - name: snmp
  snmpListener:
    workers: 100 # number of workers used to discover devices concurrently
    discoveryInterval: 3600 # interval between each autodiscovery in seconds
    configs:
      - network: 1.2.3.4/24 # CIDR notation, we recommend no larger than /24 blocks
        version: 3
        user: "user"
        authenticationProtocol: "fakeAuth"
        authenticationKey: "fakeKey"
        #privacyProtocol:
        #privacyKey:
        tags:
          - "key1:val1"
          - "key2:val2"
        loader: core
      - network: 2.3.4.5/24
        version: 3
        snmpVersion: 3
        user: "user"
        authenticationProtocol: "fakeAuth"
        authenticationKey: "fakeKey"
        #privacyProtocol:
        #privacyKey: 
        tags:
          - "key1:val1"
          - "key2:val2"
        loader: core
```

<b>注意</b>：agentd 会自动发现设备IP，然后依次采集每一个正常应答的设备。

