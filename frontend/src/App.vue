<script setup>
import { ref, reactive, computed, provide } from 'vue'
import LobbyScreen from './components/LobbyScreen.vue'
import WaitingRoom from './components/WaitingRoom.vue'
import RaceScreen from './components/RaceScreen.vue'
import ResultsScreen from './components/ResultsScreen.vue'
import ConnectionStatus from './components/ConnectionStatus.vue'

// ─── STATE ────────────────────────────────────────────────────────────────
const screen = ref('lobby') // lobby | waiting | race | results
const ws = ref(null)
const msgQueue = ref([])
const connected = ref(false)

const player = reactive({
  name: 'Player',
  color: '#e2b714',
  roomId: null,
  place: 1,
})

const room = reactive({
  passage: '',
  players: [],
  state: 'waiting',
  startTime: 0,
  finishedOrder: [],
})

// ─── WEBSOCKET ────────────────────────────────────────────────────────────
function connect() {
  if (ws.value && ws.value.readyState <= 1) return
  const proto = location.protocol === 'https:' ? 'wss:' : 'ws:'
  const host = location.hostname
  const port = location.port || (location.protocol === 'https:' ? '' : '3000')
  const url = port ? `${proto}//${host}:${port}/ws` : `${proto}//${host}/ws`

  console.log('🔌 Connecting to', url)
  const socket = new WebSocket(url)
  socket.onopen = () => {
    connected.value = true
    msgQueue.value.forEach(m => socket.send(JSON.stringify(m)))
    msgQueue.value = []
  }
  socket.onclose = () => { connected.value = false; setTimeout(connect, 3000) }
  socket.onmessage = handleMessage
  socket.onerror = () => { connected.value = false }
  ws.value = socket
}

function wsSend(data) {
  if (ws.value && ws.value.readyState === WebSocket.OPEN) {
    ws.value.send(JSON.stringify(data))
  } else {
    msgQueue.value.push(data)
    if (!ws.value || ws.value.readyState === WebSocket.CLOSED) connect()
  }
}

function handleMessage(e) {
  let msg
  try { msg = JSON.parse(e.data) } catch { return }
  switch (msg.type) {
    case 'room_joined':
      player.roomId = msg.roomId
      player.name = msg.playerName
      player.color = msg.playerColor
      room.passage = msg.passage
      room.players = msg.players || []
      screen.value = 'waiting'
      break
    case 'player_joined':
    case 'player_left':
      room.players = msg.players
      break
    case 'error':
      alert(msg.message)
      break
    case 'race_starting':
      room.startTime = msg.startTime
      room.state = 'racing'
      screen.value = 'race'
      break
    case 'players_update':
      room.players = msg.players
      break
    case 'player_finished':
      if (!room.finishedOrder.find(p => p === msg.playerName)) {
        room.finishedOrder.push(msg.playerName)
      }
      const fp = room.players.find(p => p.name === msg.playerName)
      if (fp) { fp.finished = true; fp.place = msg.place }
      break
    case 'race_finished':
      setTimeout(() => { screen.value = 'results' }, 1000)
      break
    case 'rooms_list':
      room._roomsList = msg.rooms
      break
  }
}

// Computed for my player
const me = computed(() => room.players.find(p => p.name === player.name))

// ─── ACTIONS ──────────────────────────────────────────────────────────────
function createRoom() {
  wsSend({ type: 'create_room', name: player.name })
}
function joinRoom(code) {
  wsSend({ type: 'join_room', roomId: code.toUpperCase(), name: player.name })
}
function startRace() {
  wsSend({ type: 'start_race' })
  room.state = 'racing'
}
function sendProgress(data) {
  wsSend({ type: 'progress', ...data })
}
function goLobby() {
  ws.value?.close()
  ws.value = null
  screen.value = 'lobby'
  connected.value = false
  // Reset room
  room.passage = ''
  room.players = []
  room.state = 'waiting'
  room.finishedOrder = []
  setTimeout(connect, 200)
}
function listRooms() {
  wsSend({ type: 'list_rooms' })
}

provide('wsSend', wsSend)
provide('connected', connected)
provide('player', player)
provide('room', room)
provide('me', me)
provide('actions', { createRoom, joinRoom, startRace, sendProgress, goLobby, listRooms })
</script>

<template>
  <div id="app-root" :class="screen">
    <ConnectionStatus />
    <LobbyScreen v-if="screen === 'lobby'" @create="createRoom" @join="joinRoom" />
    <WaitingRoom v-else-if="screen === 'waiting'" />
    <RaceScreen v-else-if="screen === 'race'" />
    <ResultsScreen v-else-if="screen === 'results'" />
  </div>
</template>

<style>
:root {
  --bg: #0f0f1a;
  --surface: #1a1a2e;
  --surface2: #232340;
  --text: #e8e8f0;
  --text-dim: #8888a0;
  --text-muted: #555570;
  --accent: #e2b714;
  --correct: #3fb950;
  --incorrect: #f85149;
  --mono: 'JetBrains Mono', 'Fira Code', monospace;
  --font: 'Inter', 'Segoe UI', system-ui, sans-serif;
}
* { margin: 0; padding: 0; box-sizing: border-box; }
html, body { height: 100%; background: var(--bg); color: var(--text); font-family: var(--font); font-size: 15px; overflow: hidden; }
#app { width: 100%; height: 100vh; }
#app-root { width: 100%; height: 100%; }
.hidden { display: none !important; }
</style>
