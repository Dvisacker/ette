const { client } = require('websocket')

const _client = new client()

let state = true

_client.on('connectFailed', e => { console.error(`[!] Failed to connect : ${e}`); process.exit(1); })

// connect for listening to any block being mined
// event
_client.on('connect', c => {
    c.on('close', d => {
        console.log(`[!] Closed connection : ${d}`)
        process.exit(0)
    })

    // receiving json encoded message
    c.on('message', d => { console.log(JSON.parse(d.utf8Data)) })

    // periodic subscription & unsubscription request performed
    handler = _ => {

        c.send(JSON.stringify({ name: 'block', type: state ? 'subscribe' : 'unsubscribe' }))
        state = !state

    }

    setInterval(handler, 10000)
    handler()

})

_client.connect('ws://localhost:7000/v1/ws', null)
