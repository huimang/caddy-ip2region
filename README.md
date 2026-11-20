# caddy-ip2region

English | [简体中文](README.zh-CN.md)

A Caddy module that uses ip2region data to redirect clients to different sites based on their IP geolocation.

The module registers the `http.matchers.ip2region` request matcher. It works like `maxmind_geolocation`, but reads ip2region xdb files directly instead of MaxMind MMDB files.

## Features

- Match requests by the country or region of the client IP.
- Support `allow_countries` allowlists and `deny_countries` denylists.
- Support IPv4 and optional IPv6 xdb files.
- Use a concurrency-safe searcher pool with a configurable size.
- Implement `Cleanup` so Caddy releases old searchers on configuration reload.
- Support both common ip2region field layouts and prefer the ISO country code stored in the data.
- Validate allowlists and denylists against the full ISO 3166-1 alpha-2 list.
- Resolve English and Simplified Chinese country names for all ISO codes through CLDR.

## Requirements

- Go 1.25.1 or later
- Caddy v2.11.4
- ip2region Golang binding

## Build

Build Caddy with this module using `xcaddy`:

```bash
go install github.com/caddyserver/xcaddy/cmd/xcaddy@latest

# Build from this checkout
xcaddy build v2.11.4 \
  --with github.com/huimang/caddy-ip2region/v2=./

# Or build from the published module
xcaddy build v2.11.4 \
  --with github.com/huimang/caddy-ip2region/v2@latest
```

Verify that the module is registered:

```bash
./caddy list-modules | grep ip2region
```

The output should contain:

```text
http.matchers.ip2region
```

## Caddyfile

Define a named matcher:

```caddyfile
@asia ip2region {
    db_path /opt/caddy/data/ip2region/ip2region_v4.xdb
    db_path_v6 /opt/caddy/data/ip2region/ip2region_v6.xdb
    allow_countries CN SG MY ID TH VN PH BN KH LA MM TL
    searchers 20
}
```

Use it in a site block:

```caddyfile
example.com {
    @middleEast ip2region {
        db_path /opt/caddy/data/ip2region/ip2region_v4.xdb
        db_path_v6 /opt/caddy/data/ip2region/ip2region_v6.xdb
        allow_countries AE SA QA KW BH OM YE IQ IR JO LB SY IL PS TR EG
    }

    @asia ip2region {
        db_path /opt/caddy/data/ip2region/ip2region_v4.xdb
        db_path_v6 /opt/caddy/data/ip2region/ip2region_v6.xdb
        allow_countries CN SG MY ID TH VN PH BN KH LA MM TL
    }

    handle @middleEast {
        redir https://me.example.com{uri} 302
    }

    handle @asia {
        redir https://asia.example.com{uri} 302
    }

    handle {
        respond "Select your region" 200
    }
}
```

### Options

| Option | Required | Description |
| --- | --- | --- |
| `db_path` | Yes | Path to the IPv4 xdb file |
| `db_path_v6` | No | Path to the IPv6 xdb file; IPv6 requests do not match when omitted |
| `allow_countries` | One of the two | ISO 3166-1 alpha-2 codes to allow |
| `deny_countries` | One of the two | ISO 3166-1 alpha-2 codes to deny |
| `searchers` | No | Searcher pool size, default `20`, must be greater than zero |

At least one of `allow_countries` and `deny_countries` must be configured, otherwise validation fails. Every code is checked against the 249 officially assigned ISO 3166-1 alpha-2 codes; invalid codes also fail validation. `UK` is accepted as an alias for `GB`, `EL` for `GR`, and the user-assigned code `XK` is accepted for Kosovo.

Matching rules:

- Only `allow_countries`: a request matches only when its country code is in the list.
- Only `deny_countries`: a request matches unless its country code is in the list.
- Both configured: the country code must be in `allow_countries` and must not be in `deny_countries`.
- Unparseable IPs, missing data and unknown country codes never match.

Country codes are normalized to upper case, so both `cn` and `CN` work in the Caddyfile. Aliases are normalized before validation.

### JSON configuration

```json
{
  "match": [
    {
      "ip2region": {
        "db_path": "/opt/caddy/data/ip2region/ip2region_v4.xdb",
        "db_path_v6": "/opt/caddy/data/ip2region/ip2region_v6.xdb",
        "allow_countries": ["CN", "SG", "MY", "ID"],
        "searchers": 20
      }
    }
  ]
}
```

## Client IP

The matcher uses Caddy's `client_ip` variable first and falls back to the request `RemoteAddr`.

If Caddy is deployed behind another proxy or load balancer, configure trusted proxies. Otherwise the matcher sees the proxy address instead of the client address:

```caddyfile
{
    servers {
        trusted_proxies static 10.0.0.0/8 192.168.0.0/16
    }
}
```

## Data format

The currently used xdb layout is:

```text
country|province|city|isp|country_code
```

Examples:

```text
中国|浙江省|杭州市|阿里|CN
United Arab Emirates|Dubai|0|0|AE
```

The standard ip2region layout is also supported:

```text
country|region|province|city|isp
```

Parsing rules:

1. When the last field is a valid ISO 3166-1 alpha-2 code, it is used directly.
2. Chinese entries default to `CN`.
3. `Reserved`, `0` and other unrecognized entries return an empty code and never match.
4. When the country-code field is missing, the matcher resolves country names in English and Simplified Chinese through CLDR, plus common short forms and historical names. Add non-standard names to `countryNameAliases` in `iso.go`.

## Updating data

Caddy opens the xdb files while loading the configuration. Replace the files and reload Caddy:

```bash
sudo install -m 0644 ip2region_v4.xdb /opt/caddy/data/ip2region/ip2region_v4.xdb
sudo install -m 0644 ip2region_v6.xdb /opt/caddy/data/ip2region/ip2region_v6.xdb
sudo systemctl reload caddy
```

Notes:

- The xdb files must be readable by the user that runs Caddy.
- Prefer `install` or an atomic rename so Caddy never opens a partially written file.
- ip2region does not publish data on a fixed schedule. Use GeoLite2 or DB-IP if you need a predictable update SLA.

## Testing

```bash
go vet ./...
go test ./...
```

The tests cover:

- Caddy module registration
- Caddyfile parsing
- Required options and country-code validation
- Client IP fallback
- Full ISO 3166-1 code list and extraction for every code
- English and Simplified Chinese country name fallback
- Country-code extraction and special-region detection

## Project layout

```text
caddy-ip2region/
├── matcher.go        # module registration, provisioning, matching and Caddyfile parsing
├── country.go        # region parsing and special-region detection
├── iso.go            # ISO country codes, country names and aliases
├── matcher_test.go   # module and configuration tests
├── country_test.go   # country-code extraction tests
├── iso_test.go       # ISO coverage and name fallback tests
├── go.mod
└── go.sum
```

## Deployment notes

Install the custom Caddy binary in a dedicated path and keep the distribution binary for rollback:

```bash
sudo install -m 0755 ./caddy /opt/caddy/bin/caddy
```

Use a systemd drop-in to change the executable:

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

To roll back, restore the previous binary and restart Caddy. If you switch back to a Caddy build without this module, remove or replace the `ip2region` matchers in the Caddyfile first; otherwise the configuration cannot be loaded.

## Limitations

- Country or region matching only; no city-level matching.
- Accuracy depends on the ip2region data source.
- IPv6 clients do not match when `db_path_v6` is not configured.
- The name fallback uses CLDR English and Simplified Chinese names plus a small alias table; unusual or vendor-specific names may need explicit aliases.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).
