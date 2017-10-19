# natscli

Allows you to connect to nats and monitor topics. If no subs are given
it subscribes to '>' which is the nats wildcard for everything.

Usage:
```bash
natscli [-nats http://nats.com:4444] [subs...]
```
