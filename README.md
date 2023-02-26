# neko-client
Connect to [demodesk/neko](https://github.com/demodesk/neko) backend with self contained vue component. 

For **community edition** neko with GUI and _plug & play_ deployment visit [m1k1o/neko](https://github.com/m1k1o/neko).

## Installation
Code is published to public NPM registry and GitHub npm repository.

```bash
# npm command
npm i @demodesk/neko
# yarn command
yarn add @demodesk/neko
```

### Build

You can set keyboard provider at build time, either `novnc` or the default `guacamole`.

```bash
# by default uses guacamole keyboard
npm run build
# uses novnc keyboard
KEYBOARD=novnc npm run build
```

### Example
API consists of accessing Vue reactive state, calling various methods and subscribing to events. Simple usage:

```html
<!-- import vue -->
<script src="https://unpkg.com/vue"></script>

<!-- import neko -->
<script src="./neko.umd.js"></script>
<link rel="stylesheet" href="./neko.css">

<div id="app">
  <neko ref="neko" server="http://127.0.0.1:3000/api" autologin autoplay />
</div>

<script>
new Vue({
  components: { neko },
  mounted() {
    // access state
    // this.$refs.neko.state.session_id
  
    // call methods
    // this.$refs.neko.setUrl('http://127.0.0.1:3000/api')
    // this.$refs.neko.login('username', 'password')
    // this.$refs.neko.logout()
  
    // subscribe to events
    // this.$refs.neko.events.on('room.control.host', (id) => { })
  },
}).$mount('#app')
</script>
```
