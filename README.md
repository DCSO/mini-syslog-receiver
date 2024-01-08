# mini-syslog-receiver

This is a small syslog server that can be used to receive syslog data for format
discovery and gathering of example logs required to evaluate edge node input.

It is a simple, portable binary that can be handed out to the data provider to
test-drive their data taps (e.g. appliances that support syslog output, etc.)

## Usage
```
$ ./mini-syslog-receiver -h
NAME:
   mini-syslog-receiver - receive and dump syslog data

USAGE:
   mini-syslog-receiver [global options] 

GLOBAL OPTIONS:
   --listen value, -l value   address to listen on (0.0.0.0 means all interfaces) (default: "0.0.0.0")
   --port value, -p value     port to listen on (default: 514)
   --sample value, -m value   sample up to <value> log entries, then exit (default: 1000)
   --tcp, -t                  use TCP instead of UDP (default: false)
   --tls, -s                  use TLS for TCP server (default: false)
   --tls-key value            TLS key file to use for TCP/TLS server
   --tls-chain value          TLS chain file to use for TCP/TLS server
   --outfile value, -o value  file to write output to (print to console if empty)
   --help, -h                 show help
```

The default (i.e. if no parameters are given) the tool will listen on all
interfaces on port UDP/514 (the syslog default) and dump received data as JSON
to the console it was started from. Note that on UNIX systems (e.g. Linux,
macOS) this needs to be done with root privileges because we are opening a
privileged port (< 1024)! On Windows machines the user will have to confirm a
security popup if a privileged port is used.

```
$ sudo ./mini-syslog-receiver
2024/01/08 14:04:53 using UDP 0.0.0.0:514
```

One can specify a high port to avoid this:

```
$ ./mini-syslog-receiver -p 10002
2024/01/08 14:05:18 using UDP 0.0.0.0:10002
```

Use the `-o` parameter to write to a file:
```
$ ./mini-syslog-receiver -o out.json -p 10002 -t yes
2024/01/08 14:07:21 using TCP 0.0.0.0:10002
2024/01/08 14:07:21 writing to file out.json
```

For TLS, one also needs to specify a public/private key pair from a pair of
files (`--tls-chain`/`--tls-key`). The files must contain PEM encoded data. The
certificate file (`--tls-chain`) may contain intermediate certificates following
the leaf certificate to form a certificate chain.

```
$ ./mini-syslog-receiver -p 10002 -t --tls --tls-key server-key.pem --tls-chain server-cert.pem 
2024/01/08 16:32:11 using TCP/TLS 0.0.0.0:10002
```

You can use the `--sample`/`-m` option to limit the dump to a certain number of
log items to avoid logging excessive log amounts:

```
$ ./mini-syslog-receiver -p 10002 -t -sample 2
2024/01/08 16:38:22 using TCP 0.0.0.0:10002
{"app_name":"someapp","client":"[::1]:58786","facility":1,"hostname":"EXAMPLE","message":"foobar","msg_id":"-","priority":13,"proc_id":"-","severity":5,"structured_data":"[timeQuality tzKnown=\"1\" isSynced=\"1\" syncAccuracy=\"961000\"]","timestamp":"2024-01-08T16:38:24.634075+01:00","tls_peer":"","version":1}
{"app_name":"someapp","client":"[::1]:58798","facility":1,"hostname":"EXAMPLE","message":"foobar","msg_id":"-","priority":13,"proc_id":"-","severity":5,"structured_data":"[timeQuality tzKnown=\"1\" isSynced=\"1\" syncAccuracy=\"961000\"]","timestamp":"2024-01-08T16:38:24.928816+01:00","tls_peer":"","version":1}
2024/01/08 16:38:24 sample limit of 2 log entries reached
$
```
The default is to log 1000 log items. Set the value to 0 to enable unlimited
logging.

The server can be stopped at any time using Control-C.

## Testing

You can test whether the server works by logging manually into the server. Start
it, e.g. like this for port 10002 TCP:

```
$ ./mini-syslog-receiver -o out.json -p 10002 -t yes
2024/01/08 14:09:46 using TCP 0.0.0.0:10002
2024/01/08 14:09:46 writing to file out.json
```

then log a message and observe the output:

```
$ logger -T -P 10002 -n localhost "foobar" 
$ jq . < out.json
{
  "app_name": "someapp",
  "client": "[::1]:54434",
  "facility": 1,
  "hostname": "EXAMPLE",
  "message": "foobar",
  "msg_id": "-",
  "priority": 13,
  "proc_id": "-",
  "severity": 5,
  "structured_data": "[timeQuality tzKnown=\"1\" isSynced=\"1\" syncAccuracy=\"614000\"]",
  "timestamp": "2024-01-08T14:10:17.467904+01:00",
  "tls_peer": "",
  "version": 1
}
```

## Distribution

Please find the binaries in the release section:
https://github.com/DCSO/mini-syslog-receiver/releases

There are binaries for various combinations of operating system and
architecture:

* `mini-syslog-receiver-darwin-amd64` -- for macOS on Intel
* `mini-syslog-receiver-darwin-arm64` -- for macOS on ARM (i.e. M1/M2/...)
* `mini-syslog-receiver-windows-amd64` -- for 64-bit Windows (most common)
* `mini-syslog-receiver-windows-i386` -- for 32-bit Windows (older platforms)
* `mini-syslog-receiver-linux-amd64` -- for 64-bit Intel Linux
* `mini-syslog-receiver-linux-i386` -- for 32-bit Intel Linux

## Copyright

Copyright (c) 2024, DCSO Deutsche Cyber-Sicherheitsorganisation GmbH
