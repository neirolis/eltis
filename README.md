# ELTIS DP5000.B2-XX device service

Service for interaction with the call block(door), connected to the server by USB-COM adapter.

## API

`GET /` or `POST /` - open main door (id:0)

`GET /open/:id` or `POST /open/:id` - to open specified door.
