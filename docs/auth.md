```bash
openssl genrsa 2048 2>/dev/null | openssl rsa -traditional 2>/dev/null | awk '{printf "%s\\n", $0}'
```
