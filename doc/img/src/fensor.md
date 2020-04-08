graph TD
    browser[Browser] -->|Proxy: SOCKS| client[fensor client]
    client --> db[Redis]
    db --> if_url_in_db{found URL status?}
    if_url_in_db --> |Y| if_url_blocked{DNS blocked?}
    if_url_blocked --> |Y, Proxy: VMess| server[fensor server]
    if_url_blocked --> |N, Proxy: Freedom| dst
    server --> |Proxy: Freedom| dst[Destination]
    if_url_in_db --> |N| local_dns[Local DNS server]
    local_dns --> local_dns_resolved{Domain resolved?}
    local_dns_resolved --> |N| global_dns[Global DNS server]
    local_dns_resolved --> |Y, Proxy: Freedom| dst
    global_dns --> global_dns_resolved{Domain resolved?}
    global_dns_resolved --> |Y, Proxy: Freedom| dst
    global_dns_resolved --> |N, Proxy: Vmess| server


