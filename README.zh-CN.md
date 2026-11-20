# caddy-ip2region

[English](README.md) | 简体中文

基于 [ip2region](https://github.com/lionsoul2014/ip2region) xdb 的 Caddy 国家/地区匹配器，模块 ID 为 `http.matchers.ip2region`。配置形态与 `maxmind_geolocation` 类似，但直接读取 ip2region 数据文件，不依赖 MaxMind MMDB。

## 功能

- 按客户端 IP 所属国家或地区匹配请求。
- 支持 `allow_countries` 白名单和 `deny_countries` 黑名单。
- 同时支持 IPv4 与 IPv6 xdb，IPv6 可选。
- 内置查询器池，默认 20 个查询器，并发安全。
- 实现 `Cleanup`，Caddy 热重载时自动释放旧查询服务。
- 单独识别中国香港、中国澳门和中国台湾。
- 兼容两种 xdb 字段格式，优先使用数据中的 ISO 国家码。

## 环境要求

- Go 1.25.1 或更高版本
- Caddy v2.11.4
- ip2region Golang binding

## 构建

使用 `xcaddy` 将模块编译进 Caddy：

```bash
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest
xcaddy build v2.11.4 \
  --with github.com/huimang/caddy-ip2region=./
```

确认模块已注册：

```bash
./caddy list-modules | grep ip2region
```

输出应包含：

```text
http.matchers.ip2region
```

## Caddyfile 配置

命名匹配器：

```caddyfile
@asia ip2region {
    db_path /opt/caddy/data/ip2region/ip2region_v4.xdb
    db_path_v6 /opt/caddy/data/ip2region/ip2region_v6.xdb
    allow_countries CN HK MO TW SG MY ID TH VN PH BN KH LA MM TL
    searchers 20
}
```

在站点中使用：

```caddyfile
gymlin.com {
    @middleEast ip2region {
        db_path /opt/caddy/data/ip2region/ip2region_v4.xdb
        db_path_v6 /opt/caddy/data/ip2region/ip2region_v6.xdb
        allow_countries AE SA QA KW BH OM YE IQ IR JO LB SY IL PS TR EG
    }

    @asia ip2region {
        db_path /opt/caddy/data/ip2region/ip2region_v4.xdb
        db_path_v6 /opt/caddy/data/ip2region/ip2region_v6.xdb
        allow_countries CN HK MO TW SG MY ID TH VN PH BN KH LA MM TL
    }

    handle @middleEast {
        redir https://me.gymlin.com{uri} 302
    }

    handle @asia {
        redir https://cn.gymlin.com{uri} 302
    }

    handle {
        respond "Select your region" 200
    }
}
```

### 配置项

| 配置项 | 必填 | 说明 |
| --- | --- | --- |
| `db_path` | 是 | IPv4 xdb 文件路径 |
| `db_path_v6` | 否 | IPv6 xdb 文件路径；不配置时 IPv6 请求不参与匹配 |
| `allow_countries` | 二选一 | 允许的国家或地区 ISO 3166-1 二字码列表 |
| `deny_countries` | 二选一 | 拒绝的国家或地区 ISO 3166-1 二字码列表 |
| `searchers` | 否 | 查询器池大小，默认 20，必须大于 0 |

`allow_countries` 与 `deny_countries` 至少配置一个，否则配置校验失败。

匹配规则：

- 只配置 `allow_countries`：命中列表才匹配。
- 只配置 `deny_countries`：命中列表不匹配，其余匹配。
- 两者都配置：必须在 `allow_countries` 中，且不能在 `deny_countries` 中。
- IP 无法解析、数据缺失或国家码无法识别时不匹配。

国家码统一转为大写后比较，Caddyfile 中写 `cn` 或 `CN` 均可。

### JSON 配置

```json
{
  "match": [
    {
      "ip2region": {
        "db_path": "/opt/caddy/data/ip2region/ip2region_v4.xdb",
        "db_path_v6": "/opt/caddy/data/ip2region/ip2region_v6.xdb",
        "allow_countries": ["CN", "HK", "MO", "TW"],
        "searchers": 20
      }
    }
  ]
}
```

## 客户端 IP 来源

匹配器优先使用 Caddy 的 `client_ip` 变量，取不到时回退到请求的 `RemoteAddr`。

如果 Caddy 前面还有反向代理或负载均衡，需要在 Caddy 中配置可信代理，否则匹配到的是代理 IP：

```caddyfile
{
    servers {
        trusted_proxies static 10.0.0.0/8 192.168.0.0/16
    }
}
```

## 数据格式

当前使用的 xdb 字段格式：

```text
国家|省份|城市|ISP|国家码
```

示例：

```text
中国|浙江省|杭州市|阿里|CN
中国|香港特别行政区|0|电讯盈科|CN
United Arab Emirates|Dubai|0|0|AE
```

同时兼容标准 ip2region 格式：

```text
国家|区域|省份|城市|ISP
```

解析规则：

1. 最后一个字段为两位字母时，直接作为 ISO 国家码返回。
2. 中国条目优先按省份识别 `HK`、`MO`、`TW`，其余返回 `CN`。
3. `Reserved` 或 `0` 等无法识别的条目返回空，匹配器视为不匹配。
4. 缺少国家码字段时，按内置的国家名映射兜底；新增国家时可在 `country.go` 的 `countryISO` 中补充。

## 数据更新

Caddy 在配置加载时打开 xdb。替换数据文件后需要重载：

```bash
sudo install -m 0644 ip2region_v4.xdb /opt/caddy/data/ip2region/ip2region_v4.xdb
sudo install -m 0644 ip2region_v6.xdb /opt/caddy/data/ip2region/ip2region_v6.xdb
sudo systemctl reload caddy
```

注意事项：

- xdb 文件必须对 Caddy 运行用户可读。
- 建议使用 `install` 或先写临时文件再原子替换，避免 Caddy 在文件写入过程中打开半成品。
- ip2region 数据没有固定的官方更新节奏；如果需要稳定的更新 SLA，可考虑 GeoLite2 或 DB-IP。

## 测试

```bash
go vet ./...
go test ./...
```

测试覆盖：

- Caddy 模块注册
- Caddyfile 解析
- 必填项和国家码校验
- 客户端 IP 回退
- 国家码提取和港澳台识别

## 项目结构

```text
caddy-ip2region/
├── matcher.go        # 模块注册、加载、匹配、Caddyfile 解析
├── country.go        # region 字符串解析与国家码映射
├── matcher_test.go   # 模块与配置测试
├── country_test.go   # 国家码提取测试
├── go.mod
└── go.sum
```

## 部署提示

将自定义 Caddy 安装到独立路径，保留系统包自带的二进制用于回滚：

```bash
sudo install -m 0755 ./caddy /opt/caddy/bin/caddy
```

用 systemd drop-in 修改执行路径：

```ini
[Service]
ExecStart=
ExecStart=/opt/caddy/bin/caddy run --environ --config /opt/caddy/etc/Caddyfile
ExecReload=
ExecReload=/opt/caddy/bin/caddy reload --config /opt/caddy/etc/Caddyfile --force
```

```bash
sudo systemctl daemon-reload
sudo systemctl restart caddy
```

回滚时恢复上一版二进制并重启；如果要退回不带本模块的官方 Caddy，必须先把 Caddyfile 中的 `ip2region` 匹配器移除或替换，否则配置无法加载。

## 已知限制

- 仅提供国家或地区级别匹配，不提供城市级匹配。
- 国家识别准确度依赖 ip2region 数据源。
- 未配置 `db_path_v6` 时，IPv6 客户端不会命中。
- 国家名映射表只覆盖常用国家，未知国家名需要自行补充。

## 许可证

本项目使用 MIT 许可证，详见 [LICENSE](LICENSE)。
