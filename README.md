# ELTIS DP5000.B2-XX device service

Service for interaction with the call block(door), connected to the server by USB-COM adapter.

## API

`GET /` or `POST /` - open main door (id:0)

`GET /open/:id` or `POST /open/:id` - to open specified door.

## SERIAL DATA

Each packet should be 30bytes, with payload in start of it.

### Initialize driver

When connecting, before opening the doors, it is necessary to initialize the driver at least once:

`[30]bytes{0x7F, 0x7F, 0x0A, 0x01, ...0x00}`

With response: `0x7F, 0x7F, 0x8A, 0x01, ...0x00`

### Open door

After the driver is initialized, you can open the doors:

`[30]bytes{0x7F, 0x40, 0x06, 0x0f, ...0x00}`

To specify the door ID, must specify: `0x40`+`doorID`, for example: `0x40`+`1` = `0x41` to open the door with the ID `1`.

## TODO:

1. Authorization
2. Discovery serial interface by vendorID:productID
