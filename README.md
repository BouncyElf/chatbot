# chatbot
chatbot is a chat-bot app based on websocket.

# usage
start server
```bash
docker compose up --build
```

create your own identity
```bash
curl 'http://localhost:8080/register?name=<your name here>'
```

start cli and chat~
```bash
./client -s 8080 -i 1
```
> NOTE that the `-i` means your id. you can get it after regiter

# configuration
update `bot-config.yaml` and config it yourself.
