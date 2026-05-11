<script setup>
import { inject, computed } from 'vue'
const player = inject('player')
const room = inject('room')
const actions = inject('actions')

const sorted = computed(() =>
  [...room.players].sort((a, b) => {
    if (a.finished && !b.finished) return -1
    if (!a.finished && b.finished) return 1
    if (a.finished && b.finished) return (a.finishTime || 0) - (b.finishTime || 0)
    return (b.progress || 0) - (a.progress || 0)
  })
)

const place = computed(() => {
  const idx = sorted.value.findIndex(p => p.name === player.name)
  return idx >= 0 ? idx + 1 : 1
})
const trophies = ['🥇', '🥈', '🥉', '4️⃣', '5️⃣', '6️⃣', '7️⃣', '8️⃣']
</script>

<template>
  <div class="results">
    <div class="results-content">
      <h1 class="results-title">🏁 Race Complete!</h1>
      <div class="trophy">{{ trophies[place - 1] || '🏁' }}</div>

      <div class="results-wpm">{{ sorted.find(p => p.name === player.name)?.wpm || 0 }} <small>wpm</small></div>

      <div class="results-players">
        <div v-for="(p, i) in sorted" :key="p.name"
             class="final-entry" :class="{ 'me': p.name === player.name }">
          <span class="fe-pos">{{ trophies[i] || (i + 1) }}</span>
          <span class="fe-name">{{ p.name }}</span>
          <span class="fe-wpm">{{ p.wpm || 0 }} wpm</span>
          <span class="fe-stat">{{ p.accuracy || 0 }}% · {{ Math.round(p.progress || 0) }}%</span>
        </div>
      </div>

      <button class="btn btn-primary" @click="actions.goLobby()">🔄 Play Again</button>
    </div>
  </div>
</template>

<style scoped>
.results {
  display: flex; align-items: center; justify-content: center;
  height: 100%; background: radial-gradient(ellipse at center, #1a1a3e 0%, #0f0f1a 70%);
}
.results-content { text-align: center; max-width: 500px; width: 100%; padding: 2rem; display: flex; flex-direction: column; align-items: center; gap: 1rem; }
.results-title { font-size: 2rem; font-weight: 800; }
.trophy { font-size: 5rem; animation: bounce 1s ease; }
@keyframes bounce { 0% { transform: scale(3) rotate(-20deg); opacity: 0; } 60% { transform: scale(1) rotate(5deg); } }
.results-wpm { font-size: 3.5rem; font-weight: 800; color: var(--accent); font-family: var(--mono); }
.results-wpm small { font-size: 1.2rem; color: var(--text-dim); }
.results-players { width: 100%; margin-top: 0.5rem; }
.final-entry { display: flex; align-items: center; gap: 0.8rem; padding: 0.6rem 1rem; background: var(--surface); border-radius: 8px; margin-bottom: 0.3rem; }
.fe-pos { font-size: 1.1rem; min-width: 30px; }
.fe-name { flex: 1; font-weight: 600; }
.fe-wpm { font-family: var(--mono); color: var(--accent); font-weight: 700; }
.fe-stat { font-size: 0.8rem; color: var(--text-dim); }
.me { border: 1px solid var(--accent); }

.btn { display: inline-flex; align-items: center; gap: 0.5rem; padding: 0.7rem 1.5rem; border: none; border-radius: 8px; font-family: var(--font); font-size: 0.95rem; font-weight: 600; cursor: pointer; transition: all 0.2s; }
.btn-primary { background: var(--accent); color: #000; }
.btn-primary:hover { transform: translateY(-2px); }
</style>
