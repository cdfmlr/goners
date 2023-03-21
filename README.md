# goners

Goner's Oafish Network Explorer & Reliable Sniffer.

## Usage

### devices

### pcap

### http

服务端：

```sh
sudo go run ./cmd http
```

客户端：

```sh
curl localhost:9800/devices | json_pp
[
   {
      "addrs" : [
         {
            "ip" : "fe80::xxx:xxx:xxx",
            "ip_type" : 64,
            "ip_type_str" : "LinkLocalUnicast",
            "network_name" : "ip+net",
            "prefix" : 64
         },
         {
            "ip" : "xxx.xxx.xxx.xxx",
            "ip_type" : 132,
            "ip_type_str" : "Private, GlobalUnicast",
            "network_name" : "ip+net",
            "prefix" : 19
         },
      ],
      "hardware_addr" : "xx:xx:xx:xx:xx:xx",
      "index" : 11,
      "name" : "en0"
   },
   ...
]

$ curl -X POST -d '{"device": "lo0" }' -i localhost:9800/pcap
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Tue, 21 Mar 2023 01:32:28 GMT
Content-Length: 53

{"session_id":"7261481c-c9ec-44a8-9748-b80d4b750b8c"}⏎                                                                       

$ curl -X DELETE -d '{"session_id": "7261481c-c9ec-44a8-9748-b80d4b750b8c" }' -i localhost:9800/pcap
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Tue, 21 Mar 2023 01:32:49 GMT
Content-Length: 61

{"deleted_session_id":"7261481c-c9ec-44a8-9748-b80d4b750b8c"}
```

WebSocket:

```sh
```

