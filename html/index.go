package html

func Index() string {
	return `<html>
<head>
<title>ICT</title>
<link rel="icon" href="data:;base64,iVBORw0KGgo=">
<style>
table {
  margin-bottom: 10px;
}

input[type="number"] {
  width: 100px;
}

input {
  font-size: 20px;
}

button {
  font-size: 26px;
  margin-bottom: 10px;
}
</style>
</head>
<body>
  <table>
    <tr>
      <td>BA Type 0</td>
      <td><input id="ba_type_0" type="number" min="1" step="1" value="1" /></td>
    </tr>
    <tr>
      <td>BA Type 1</td>
      <td><input id="ba_type_1" type="number" min="1" step="1" value="5" /></td>
    </tr>
    <tr>
      <td>BA Type 2</td>
      <td><input id="ba_type_2" type="number" min="1" step="1" value="10" /></td>
    </tr>
    <tr>
      <td>BA Type 3</td>
      <td><input id="ba_type_3" type="number" min="1" step="1" value="20" /></td>
    </tr>
    <tr>
      <td>BA Type 4</td>
      <td><input id="ba_type_4" type="number" min="1" step="1" value="50" /></td>
    </tr>
    <tr>
      <td>BA Type 5</td>
      <td><input id="ba_type_5" type="number" min="1" step="1" value="100" /></td>
    </tr>
    <tr>
      <td></td>
      <td><label><input id="ba_auto" type="checkbox" checked /> Auto Accept</label></td>
      <td></td>
    </tr>
    <tr>
      <td>UCA Type 1</td>
      <td><input id="uca_type_1" type="number" min="0.01" step="0.01" value="0.05" /></td>
    </tr>
    <tr>
      <td>UCA Type 2</td>
      <td><input id="uca_type_2" type="number" min="0.01" step="0.01" value="0.1" /></td>
    </tr>
    <tr>
      <td>UCA Type 3</td>
      <td><input id="uca_type_3" type="number" min="0.01" step="0.01" value="0.25" /></td>
    </tr>
    <tr>
      <td>UCA Type 4</td>
      <td><input id="uca_type_4" type="number" min="0.01" step="0.01" value="1" /></td>
    </tr>
    <tr>
      <td>Total</td>
      <td><input id="total" type="number" value="0" disabled /></td>
      <td><input id="zero" type="button" value="ZERO"></td>
    </tr>
    <tr>
      <td>Payout</td>
      <td><input id="mh_payout_value" type="number" value="1" min="1" step="1" /></td>
      <td><input id="mh_payout" type="button" value="MH PAYOUT"></td>
    </tr>
  </table>
  <button id="ba_enable">BA ENABLE</button>
  <button id="ba_disable">BA DISABLE</button>
  <button id="ba_reset">BA RESET</button>
  <button id="ba_status">BA STATUS</button>
  <br>
  <button id="ba_accept">BA ACCEPT</button>
  <button id="ba_reject">BA REJECT</button>
  <button id="ba_hold">BA HOLD</button>
  <button id="ba_info">BA INFO</button>
  <br>
  <button id="uca_enable">UCA ENABLE</button>
  <button id="uca_disable">UCA DISABLE</button>
  <button id="uca_status">UCA STATUS</button>
  <button id="uca_info">UCA INFO</button>
  <br>
  <button id="mh_status">MH STATUS</button>
  <button id="mh_reset">MH RESET</button>
  <button id="mh_empty">MH EMPTY</button>
<script>
const ba = new WebSocket('ws://localhost:12345/ict/bill-acceptor')

ba.addEventListener('close', () => {
  document.querySelectorAll('[id^=ba_]').forEach((i) => i.disabled = true)
})

ba.addEventListener('open', () => {
  document.querySelectorAll('[id^=ba_]').forEach((i) => i.disabled = false)
})

ba.addEventListener('message', (e) => {
  console.log(e.data)
  const json = JSON.parse(e.data)
  if (json.action === 'status') {
    if (json.status === 'validated') {
      if (document.querySelector('#ba_auto').checked) {
        ba.send(JSON.stringify({ action: 'accept' }))
      }
    } else if (json.status === 'done') {
      document.querySelector('#total').value = (+document.querySelector('#total').value) + (+document.querySelector('#ba_type_' + json.type).value)
    }
  } else if (json.action === 'init') {
    [ 'enable', 'disable', 'reset', 'status', 'accept', 'reject', 'hold', 'info' ].forEach((a) => {
      if (!json.supported_actions.includes(a)) {
        document.querySelector('#ba_' + a).disabled = true
      }
    })
  }
})

document.querySelector('#zero').addEventListener('click', () => {
  document.querySelector('#total').value = 0
})

document.querySelector('#ba_enable').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'enable' }))
})

document.querySelector('#ba_disable').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'disable' }))
})

document.querySelector('#ba_reset').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'reset' }))
})

document.querySelector('#ba_status').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'status' }))
})

document.querySelector('#ba_accept').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'accept' }))
})

document.querySelector('#ba_reject').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'reject' }))
})

document.querySelector('#ba_hold').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'hold' }))
})

document.querySelector('#ba_info').addEventListener('click', () => {
  ba.send(JSON.stringify({ action: 'info' }))
})

const uca = new WebSocket('ws://localhost:12345/ict/uca')

uca.addEventListener('close', () => {
  document.querySelectorAll('[id^=uca_]').forEach((i) => i.disabled = true)
})

uca.addEventListener('open', () => {
  document.querySelectorAll('[id^=uca_]').forEach((i) => i.disabled = false)
})

uca.addEventListener('message', (e) => {
  console.log(e.data)
  const json = JSON.parse(e.data)
  if (json.action === 'status') {
    if (json.status === 'received') {
      document.querySelector('#total').value = (+document.querySelector('#total').value) + (+document.querySelector('#uca_type_' + json.type).value)
    }
  }
})

document.querySelector('#uca_enable').addEventListener('click', () => {
  uca.send(JSON.stringify({ action: 'enable' }))
})

document.querySelector('#uca_disable').addEventListener('click', () => {
  uca.send(JSON.stringify({ action: 'disable' }))
})

document.querySelector('#uca_status').addEventListener('click', () => {
  uca.send(JSON.stringify({ action: 'status' }))
})

document.querySelector('#uca_info').addEventListener('click', () => {
  uca.send(JSON.stringify({ action: 'info' }))
})

const mh = new WebSocket('ws://localhost:12345/ict/mini-hopper')

mh.addEventListener('close', () => {
  document.querySelectorAll('[id^=mh_]').forEach((i) => i.disabled = true)
})

mh.addEventListener('open', () => {
  document.querySelectorAll('[id^=mh_]').forEach((i) => i.disabled = false)
})

let payoutValue = 1

mh.addEventListener('message', (e) => {
  console.log(e.data)
  const json = JSON.parse(e.data)
  if (json.action === 'status') {
    if (json.status === 'payout-once') {
      document.querySelector('#mh_payout_value').value = (+document.querySelector('#mh_payout_value').value) - 1
    }
    if (json.status === 'payout-finished') {
      setTimeout(() => {
        document.querySelector('#mh_payout').disabled = false
        document.querySelector('#mh_payout_value').value = payoutValue
      }, 500)
    }
  }
})

document.querySelector('#mh_payout').addEventListener('click', () => {
  document.querySelector('#mh_payout').disabled = true
  payoutValue = +(document.querySelector('#mh_payout_value').value)
  mh.send(JSON.stringify({ action: 'payout+message', value: payoutValue }))
})

document.querySelector('#mh_status').addEventListener('click', () => {
  mh.send(JSON.stringify({ action: 'status' }))
})

document.querySelector('#mh_reset').addEventListener('click', () => {
  mh.send(JSON.stringify({ action: 'reset' }))
})

document.querySelector('#mh_empty').addEventListener('click', () => {
  mh.send(JSON.stringify({ action: 'empty+message' }))
})
</script>
</body>
</html>
`
}
