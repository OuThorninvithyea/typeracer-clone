<script setup>
import { inject, computed } from 'vue'
const player = inject('player')
const room = inject('room')
const actions = inject('actions')

const canStart = computed(() => room.players.length >= 2)
</script>

<template>
  <div class="waiting">
    <div class="waiting-content">
      <div class="waiting-header">
        <h2>Room: <span class="room-code-lg">{{ player.roomId }}</span></h2>
        <p class="share-hint">Share this code with friends!</p>
      </div>
      <div class="players-grid">
        <div v-for="p in room.players" :key="p.name" class="player-card" :style="{ borderColor: p.color }">
          <div class="p-avatar">🏎️</div>
          <div class="p-name">{{ p.name }}</div>
        </div>
      </div>
      <div class="passage-preview">{{ room.passage.substring(0, 120) }}...</div>
      <button v-if="canStart" class="btn btn-primary btn-large" @click="actions.startRace()">🏁 Start Race!</button>
      <p v-else class="waiting-hint">Waiting for players to join...</p>
      <button class="btn btn-text" @click="actions.goLobby()">← Leave Room</button>
    </div>
  </div>
</template>

<style scoped>
.waiting { display: flex; align-items: center; justify-content: center; height: 100%; }
.waiting-content { text-align: center; max-width: 600px; width: 100%; padding: 2rem; display: flex; flex-direction: column; align-items: center; gap: 1.5rem; }
h2 { font-size: 1.3rem; }
.room-code-lg { font-family: var(--mono); color: var(--accent); font-size: 2rem; letter-spacing: 0.3em; }
.share-hint { color: var(--text-dim); font-size: 0.85rem; }
.players-grid { display: flex; gap: 1rem; flex-wrap: wrap; justify-content: center; }
.player-card { background: var(--surface); border-radius: 12px; padding: 1rem 1.5rem; text-align: center; min-width: 100px; border: 2px solid transparent; }
.p-avatar { font-size: 2rem; }
.p-name { font-size: 0.85rem; margin-top: 0.3rem; font-weight: 600; }
.passage-preview { background: var(--surface); border-radius: 8px; padding: 1rem; font-size: 0.85rem; color: var(--text-dim); max-width: 500px; line-height: 1.6; font-family: var(--mono); }
.btn { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.7rem 1.5rem; border: none; border-radius: 8px; font-family: var(--font); font-size: 0.95rem; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.btn-primary { background: var(--accent); color: #000; }
.btn-primary:hover { transform: translateY(-2px); }
.btn-large { padding: 1rem 2.5rem; font-size: 1.1rem; border-radius: 12px; }
.btn-text { background: transparent; color: var(--text-dim); }
.btn-text:hover { color: var(--text); }
.waiting-hint { color: var(--text-muted); font-size: 0.85rem; animation: pulse 1.5s infinite; }
@keyframes pulse { 0%,100% { opacity: 1 } 50% { opacity: 0.4 } }
</style>
