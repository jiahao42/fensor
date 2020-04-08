graph TD
    Browser[Browser] -->|Proxy: SOCKS| B[V2Ray client]
    B --> |Proxy: VMess| C[V2Ray Server]
    C --> |Proxy: Freedom| D[Destination]


