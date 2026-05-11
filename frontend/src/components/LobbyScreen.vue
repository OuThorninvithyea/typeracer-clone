<script setup>
import { ref, inject, onMounted } from 'vue'

const player = inject('player')
const room = inject('room')
const actions = inject('actions')

const roomCode = ref('')
const error = ref('')
const playerName = ref(player.name)

function onCreate() {
  player.name = playerName.value.trim() || 'Player'
  actions.createRoom()
}
function onJoin() {
  player.name = playerName.value.trim() || 'Player'
  const code = roomCode.value.trim()
  if (!code || code.length < 3) { error.value = 'Enter a valid room code'; return }
  error.value = ''
  actions.joinRoom(code)
}
function joinRoomFromList(id) {
  roomCode.value = id
  onJoin()
}

onMounted(() => actions.listRooms())
</script>

<template>
  <div class="lobby">
    <div class="lobby-content">
      <div class="logo-car">🏎️</div>
      <h1>TypeRacer</h1>
      <p class="subtitle">Race your friends with typing speed!</p>

      <div class="lobby-form">
        <div class="input-group">
          <label>Your Name</label>
          <input v-model="playerName" maxlength="16" placeholder="Enter name..." @keydown.enter="onCreate" />
        </div>
        <button class="btn btn-primary" @click="onCreate">🏁 Create Room</button>

        <div class="or-divider">or</div>

        <div class="input-group">
          <label>Join Room</label>
          <div class="join-row">
            <input v-model="roomCode" maxlength="4" placeholder="CODE" class="room-input" @keydown.enter="onJoin" />
            <button class="btn btn-secondary" @click="onJoin">Join ➜</button>
          </div>
        </div>

        <div v-if="error" class="lobby-error">{{ error }}</div>
      </div>

      <div class="rooms-list">
        <h3>Active Rooms</h3>
        <div v-if="!room._roomsList || room._roomsList.length === 0" class="no-rooms">No active rooms</div>
        <div v-else v-for="r in room._roomsList" :key="r.id" class="room-item" @click="joinRoomFromList(r.id)">
          <span class="room-code">{{ r.id }}</span>
          <span class="room-info">{{ r.players }} player{{ r.players > 1 ? 's' : '' }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.lobby {
  display: flex; align-items: center; justify-content: center;
  height: 100%; background: radial-gradient(ellipse at center, #1a1a3e 0%, #0f0f1a 70%);
}
.lobby-content { text-align: center; max-width: 420px; width: 100%; padding: 2rem; }
.logo-car { font-size: 4rem; margin-bottom: 0.5rem; animation: bounce 2s ease-in-out infinite; }
@keyframes bounce { 0%,100% { transform: translateY(0) } 50% { transform: translateY(-10px) } }
h1 { font-size: 2.8rem; font-weight: 800; color: var(--accent); letter-spacing: -0.03em; }
.subtitle { color: var(--text-dim); margin-top: 0.3rem; font-size: 0.95rem; }
.lobby-form { margin-top: 2rem; display: flex; flex-direction: column; gap: 1rem; align-items: center; }
.input-group { text-align: left; width: 100%; }
.input-group label { font-size: 0.75rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-dim); display: block; margin-bottom: 0.3rem; }
input { width: 100%; padding: 0.7rem 1rem; background: var(--surface); border: 1px solid var(--surface2); border-radius: 8px; color: var(--text); font-family: var(--font); font-size: 0.95rem; outline: none; }
input:focus { border-color: var(--accent); }
.room-input { text-transform: uppercase; letter-spacing: 0.2em; font-weight: 700; text-align: center; max-width: 140px; font-family: var(--mono); }
.join-row { display: flex; gap: 0.5rem; align-items: center; }
.or-divider { color: var(--text-muted); font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.1em; }
.btn { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.7rem 1.5rem; border: none; border-radius: 8px; font-family: var(--font); font-size: 0.95rem; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.btn-primary { background: var(--accent); color: #000; }
.btn-primary:hover { transform: translateY(-2px); box-shadow: 0 4px 20px rgba(226,183,20,0.3); }
.btn-secondary { background: var(--surface2); color: var(--text); }
.btn-secondary:hover { background: #2f2f50; }
.lobby-error { background: rgba(248,81,73,0.12); border: 1px solid rgba(248,81,73,0.3); border-radius: 8px; padding: 0.7rem; color: var(--incorrect); font-size: 0.85rem; width: 100%; }
.rooms-list { margin-top: 2rem; text-align: left; }
.rooms-list h3 { font-size: 0.8rem; text-transform: uppercase; letter-spacing: 0.1em; color: var(--text-dim); margin-bottom: 0.5rem; }
.no-rooms { color: var(--text-muted); font-size: 0.85rem; }
.room-item { display: flex; justify-content: space-between; padding: 0.5rem 0.75rem; background: var(--surface); border-radius: 6px; margin-bottom: 0.3rem; cursor: pointer; }
.room-item:hover { background: var(--surface2); }
.room-code { font-family: var(--mono); font-weight: 700; color: var(--accent); }
.room-info { font-size: 0.8rem; color: var(--text-dim); }
</style>
