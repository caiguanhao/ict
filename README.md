# ict

WebSocket server to connect ICT's bill acceptor, coin acceptor and coin dispenser.

You can see usage example in `html/index.html`.

```
ict -address 127.0.0.1:12345 -ba /dev/ttyS0 -mh /dev/ttyS1 -uca /dev/ttyS2 -serve
```

## Bill Acceptor

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

```js
const uca = new WebSocket('ws://localhost:12345/ict/uca')
uca.send(JSON.stringify({ action: 'enable' }))
// {"action":"status","status":"accepted","type":null}

// once coin is inserted
// receive: {"action":"status","status":"received","type":4}
```

## Coin Dispenser

```js
const mh = new WebSocket('ws://localhost:12345/ict/mini-hopper')
mh.send(JSON.stringify({ action: 'payout+message', value: 2 }))
// 2 coins is dispensed
// receive: {"action":"status","status":"payout-once"}
// receive: {"action":"status","status":"payout-once"}
// receive: {"action":"status","status":"payout-finished"}
```

LICENSE: MIT
