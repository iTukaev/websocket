# Описание

Сервис получает данные по киберспортивным матчам CS:GO от провайдера, структурирует, 
очищает от лишнего и отправляет всем активным вебсокет клиентам, 
если призошли изменения с данными.

При первом подключении клиента отправляет актуальные данные Line и Upcoming.

Параметры хранит в переменных окружения. Описание в разделе "ENV"

# Сборка бинарника

go build -o websocked ./cmd/*.go

# ENV

Переменные окружения берутся из файла .env автоматически при запуске программы, 
если они не были определены ранее в ENV среды выполнения.

В данный момент в .env прописаны следующие значения:

_$ tickers timeout for next request to provider_

__LIVE_TIMEOUT=3__

__UPCOMING_TIMEOUT=20__

_$ provider request URLs_

__LIVE_URL=```https://.....```__

__UPCOMING_URL=```https://.....```__

_$ field to marker json which are messaged to websocket_

__LIVE_JSON=```live```__

__UPCOMING_JSON=```upcoming```__

_$ websocket port_

__WS_ADDR=```:8081```__

# Что получаем а выходе

Сервер отправляет в вебсокет соединение структуры двух типов:
1. Live - матчи, которые идут в данный момент в лайве. 
Тело сообщения содержит поле {"parameter":"live",......}
2. Upcoming - предстоящие матчи. 
Тело сообщения содержит поле {"parameter":"upcoming",......}

Примеры ответов:

_{"parameter":"live","body":[{"game_id":352722396,"game_start":1643463540,
"game_oc_list":[{"oc_group_name":"1x2","oc_name":"Team Fryzex",...}],[{}],...}]}_

_{"parameter":"upcoming","body":[{"game_id":351424884,"game_start":1643472000,
"game_oc_list":[{"oc_group_name":"1x2","oc_name":"Conquer",...}],[{}],...}]}_

# Проверка работы без внешних зависимостей

Можно использовать либо плагин для Chrome - __Simple WebSocket Client__, 
либо Postman.

По умолчанию доступ открыт на ```ws://localhost:8081/ws```
