# ict

WebSocket server to connect ICT's bill acceptor, coin acceptor and coin dispenser.

You can see usage example in `html/index.html`.

```
ict -address 127.0.0.1:12345 -ba /dev/ttyS0 -mh /dev/ttyS1 -uca /dev/ttyS2 -serve
```

## Bill Acceptor

![Bill Acceptor](https://user-images.githubusercontent.com/1284703/87120763-d499b880-c2b3-11ea-8744-54605eecb461.gif)

```js
const ba = new WebSocket('ws://localhost:12345/ict/bill-acceptor')
ba.send(JSON.stringify({ action: 'enable' }))
// {"action":"status","status":"enable","type":null}

// once bill is inserted
// receive: {"action":"status","status":"validated","type":0}
// send:    {"action":"accept"}
// receive: {"action":"status","status":"done","type":0}
```

## Coin Acceptor

![Coin Acceptor](https://user-images.githubusercontent.com/1284703/87120765-d5324f00-c2b3-11ea-832c-876f215e5c69.gif)

```js
const uca = new WebSocket('ws://localhost:12345/ict/uca')
uca.send(JSON.stringify({ action: 'enable' }))
// {"action":"status","status":"accepted","type":null}

// once coin is inserted
// receive: {"action":"status","status":"received","type":4}
```

## Coin Dispenser

![Coin Dispenser](https://user-images.githubusercontent.com/1284703/87120761-d2cff500-c2b3-11ea-8f70-161141ebe09c.gif)

```js
const mh = new WebSocket('ws://localhost:12345/ict/mini-hopper')
mh.send(JSON.stringify({ action: 'payout+message', value: 2 }))
// 2 coins is dispensed
// receive: {"action":"status","status":"payout-once"}
// receive: {"action":"status","status":"payout-once"}
// receive: {"action":"status","status":"payout-finished"}
```

## Windows

You can use `ict` in Windows by specifying COM ports:

![Windows](https://user-images.githubusercontent.com/1284703/87120788-e713f200-c2b3-11ea-81b5-07c7c0d87d59.png)

LICENSE: MIT
