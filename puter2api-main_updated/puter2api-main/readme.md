# 教程
https://docs.puter.com/playground/ai-chatgpt/

![](2025-11-27-01-44-09.png)

# 构建教程
```bash
git add .
git commit -m "Add multi-platform binary release"
git tag v1.0.0
git push origin main --tags
```

## 模型列表
```bash
curl --location 'https://puter.com/puterai/chat/models' \
--header 'Accept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7' \
--header 'Accept-Language: zh-CN,zh;q=0.9' \
--header 'Cache-Control: no-cache' \
--header 'Connection: keep-alive' \
--header 'Pragma: no-cache' \
--header 'Sec-Fetch-Dest: document' \
--header 'Sec-Fetch-Mode: navigate' \
--header 'Sec-Fetch-Site: none' \
--header 'Sec-Fetch-User: ?1' \
--header 'Upgrade-Insecure-Requests: 1' \
--header 'User-Agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36' \
--header 'sec-ch-ua: "Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"' \
--header 'sec-ch-ua-mobile: ?0' \
--header 'sec-ch-ua-platform: "macOS"' \
--header 'Cookie: _clck=trp08g%5E2%5Eg1c%5E0%5E2156; __stripe_mid=1ae8b1e7-34aa-4d9c-a0bc-e00348bd4e36e7b24e; __stripe_sid=937ef572-58cc-46d3-bc2e-dbac15385aa77e545a; puter_auth_token=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0IjoicyIsInYiOiIwLjAuMCIsInUiOiJaMmZNZ2FzOFJSeXhSS3M3S1FuWnpnPT0iLCJ1dSI6Im02UmZFbkEzU0VpUjk0TVZadnJYZkE9PSIsImlhdCI6MTc2NDE3MTc0OH0.2pC0C8jAvpFiUpOcVza1V7uCnnfVza9kBLX2p5PcFyw; _clsk=1ydouus%5E1764171905602%5E6%5E1%5Ei.clarity.ms%2Fcollect'
```
## 对话测试
```bash
curl --location 'https://api.puter.com/drivers/call' \
--header 'accept: */*' \
--header 'accept-language: zh-CN,zh;q=0.9' \
--header 'cache-control: no-cache' \
--header 'content-type: text/plain;actually=json' \
--header 'origin: https://docs.puter.com' \
--header 'pragma: no-cache' \
--header 'priority: u=1, i' \
--header 'referer: https://docs.puter.com/' \
--header 'sec-ch-ua: "Chromium";v="142", "Google Chrome";v="142", "Not_A Brand";v="99"' \
--header 'sec-ch-ua-mobile: ?0' \
--header 'sec-ch-ua-platform: "macOS"' \
--header 'sec-fetch-dest: empty' \
--header 'sec-fetch-mode: cors' \
--header 'sec-fetch-site: same-site' \
--header 'user-agent: Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36' \
--data '{
    "interface": "puter-chat-completion",
    "driver": "claude",
    "test_mode": false,
    "method": "complete",
    "args": {
        "messages": [
            {
                "content": "你的模型版本是什么"
            }
        ],
        "model": "claude-opus-4-5",
        "stream": true
    },
    "auth_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0IjoiYXUiLCJ2IjoiMC4wLjAiLCJ1dSI6Im02UmZFbkEzU0VpUjk0TVZadnJYZkE9PSIsImF1IjoiaWRnL2ZEMDdVTkdhSk5sNXpXUGZhUT09IiwicyI6Ii9tdy9XRldQSElqN09QZEVDeGpsU3c9PSIsImlhdCI6MTc2NDE3MTk3Nn0.-zE27rKsvIGXiWAnAHYfJU5jDppoWak4KjE9HWZdfLs"
}'
```