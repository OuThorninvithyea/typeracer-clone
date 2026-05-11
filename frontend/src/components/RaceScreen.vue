<script setup>
import { ref, inject, computed, onMounted, onUnmounted } from 'vue'
import RaceTrack from './RaceTrack.vue'

const player = inject('player')
const room = inject('room')
const actions = inject('actions')

// ─── TYPING STATE ──────────────────────────────────────────────────────────
const charIndex = ref(0)
const typedChars = ref([])
const correctChars = ref(0)
const incorrectChars = ref(0)
const totalKeystrokes = ref(0)
const correctKeystrokes = ref(0)
const isFinished = ref(false)
const startTime = ref(null)
const finalWpm = ref(0)
const countdownValue = ref(null)
const passageEl = ref(null)

// ─── COUNTDOWN ────────────────────────────────────────────────────────────
let countdownTimer = null
onMounted(() => {
  const now = Date.now()
  const diff = room.startTime - now
  if (diff > 0) {
    countdownValue.value = 3
    countdownTimer = setInterval(() => {
      const remaining = Math.ceil((room.startTime - Date.now()) / 1000)
      if (remaining <= 0) {
        clearInterval(countdownTimer)
        countdownTimer = null
        countdownValue.value = null
        startTime.value = Date.now()
        return
      }
      const nums = { 3: '3', 2: '2', 1: '1', 0: '🏁' }
      countdownValue.value = nums[remaining] || remaining
    }, 200)
  } else {
    startTime.value = Date.now()
  }
})

// ─── COMPUTED ──────────────────────────────────────────────────────────────
const passageChars = computed(() => room.passage.split(''))
const progress = computed(() =>
  room.passage.length > 0 ? (charIndex.value / room.passage.length) * 100 : 0
)
const wpm = computed(() => {
  if (!startTime.value) return 0
  const elapsed = (Date.now() - startTime.value) / 1000
  if (elapsed < 1) return 0
  return Math.round((correctChars.value / 5) / (elapsed / 60))
})
const acc = computed(() =>
  totalKeystrokes.value > 0
    ? Math.round((correctKeystrokes.value / totalKeystrokes.value) * 100)
    : 100
)

function charClass(i) {
  if (i < charIndex.value) {
    const expected = passageChars.value[i]
    const typed = typedChars.value[i]
    return typed === expected ? 'p-correct' : 'p-incorrect'
  }
  return i === charIndex.value ? 'p-current' : ''
}

const sortedPlayers = computed(() =>
  [...room.players].sort((a, b) => {
    if (a.finished && !b.finished) return -1
    if (!a.finished && b.finished) return 1
    if (a.finished && b.finished) return (a.finishTime || 0) - (b.finishTime || 0)
    return (b.progress || 0) - (a.progress || 0)
  })
)
const place = computed(() => {
  const idx = sortedPlayers.value.findIndex(p => p.name === player.name)
  return idx >= 0 ? idx + 1 : 1
})
const finishedCount = computed(() => room.players.filter(p => p.finished).length)

// ─── TYPING ────────────────────────────────────────────────────────────────
function handleKeydown(e) {
  if (isFinished.value || countdownValue.value) return
  if (e.ctrlKey || e.metaKey || e.altKey) return
  if (e.key.length > 1 && e.key !== 'Backspace') return

  if (e.key === 'Backspace') {
    e.preventDefault()
    if (charIndex.value > 0) {
      charIndex.value--
      typedChars.value.pop()
      totalKeystrokes.value = Math.max(0, totalKeystrokes.value - 1)
      correctKeystrokes.value = Math.max(0, correctKeystrokes.value - 1)
    }
    return
  }

  if (e.key.length === 1) {
    e.preventDefault()
    if (charIndex.value >= passageChars.value.length) return

    const expected = passageChars.value[charIndex.value]
    const correct = e.key === expected

    typedChars.value.push(e.key)
    if (correct) { correctChars.value++; correctKeystrokes.value++ }
    else { incorrectChars.value++ }
    totalKeystrokes.value++
    charIndex.value++

    const pct = (charIndex.value / passageChars.value.length) * 100
    const done = charIndex.value >= passageChars.value.length

    // Sync own car locally
    const meLocal = room.players.find(p => p.name === player.name)
    if (meLocal) { meLocal.progress = pct; meLocal.wpm = wpm.value; meLocal.accuracy = acc.value }

    actions.sendProgress({
      progress: pct, wpm: wpm.value, accuracy: acc.value,
      charsTyped: totalKeystrokes.value, finished: done,
    })

    if (done) { isFinished.value = true; finalWpm.value = wpm.value }
  }
}

// Periodic sync (100ms)
let syncTimer = null
onMounted(() => {
  document.addEventListener('keydown', handleKeydown)
  syncTimer = setInterval(() => {
    if (!isFinished.value && totalKeystrokes.value > 0) {
      const pct = (charIndex.value / passageChars.value.length) * 100
      const meLocal = room.players.find(p => p.name === player.name)
      if (meLocal) { meLocal.progress = pct; meLocal.wpm = wpm.value }
      actions.sendProgress({
        progress: pct, wpm: wpm.value, accuracy: acc.value,
        charsTyped: totalKeystrokes.value, finished: false,
      })
    }
  }, 100)
})
onUnmounted(() => {
  document.removeEventListener('keydown', handleKeydown)
  if (syncTimer) clearInterval(syncTimer)
})
</script>

<template>
  <div class="race">
    <!-- Countdown -->
    <div v-if="countdownValue" class="countdown-overlay">
      <span class="countdown-num">{{ countdownValue }}</span>
    </div>

    <!-- Finished Overlay -->
    <div v-if="isFinished && !countdownValue" class="finished-overlay">
      <div class="fo-content">
        <div class="fo-icon">✅</div>
        <div class="fo-title">You finished!</div>
        <div class="fo-wpm">{{ finalWpm }}<small> wpm</small></div>
        <div class="fo-waiting">⏳ Waiting... ({{ finishedCount }}/{{ room.players.length }})</div>
      </div>
    </div>

    <!-- Leaderboard -->
    <div class="leaderboard">
      <div class="lb-header">🏁 RANKINGS</div>
      <div v-for="(p, i) in sortedPlayers.slice(0, 6)" :key="p.name"
           class="lb-entry" :class="{ 'lb-me': p.name === player.name }">
        <span class="lb-pos">{{ i + 1 }}</span>
        <span class="lb-name">{{ p.name }}</span>
        <span class="lb-wpm">{{ p.wpm || 0 }}</span>
      </div>
    </div>

    <!-- 3D Track -->
    <RaceTrack :players="room.players" :myName="player.name" />

    <!-- Progress Bars -->
    <div class="progress-container">
      <div v-for="p in sortedPlayers" :key="p.name" class="progress-bar-wrap">
        <span class="prog-name">{{ p.name }}</span>
        <div class="prog-track">
          <div class="prog-fill" :style="{ width: Math.round(p.progress || 0) + '%', background: p.color }"></div>
        </div>
        <span class="prog-pct">{{ Math.round(p.progress || 0) }}%</span>
      </div>
    </div>

    <!-- Typing Area -->
    <div class="typing-area">
      <div class="passage-display" ref="passageEl">
        <span v-for="(ch, i) in passageChars" :key="i"
              :class="['p-char', charClass(i)]">{{ ch }}</span>
      </div>
    </div>

    <!-- Stats -->
    <div class="race-stats">
      <div class="rstat"><span class="rstat-val">{{ wpm }}</span> <small>wpm</small></div>
      <div class="rstat"><span class="rstat-val">{{ acc }}</span> <small>%</small></div>
      <div class="rstat"><span class="rstat-val">{{ Math.round(progress) }}</span> <small>%</small></div>
      <div class="rstat"><span class="rstat-val">#{{ place }}</span> <small>place</small></div>
    </div>
  </div>
</template>

<style scoped>
.race { display: flex; flex-direction: column; height: 100vh; position: relative; }

.countdown-overlay {
  position: fixed; inset: 0; z-index: 100;
  display: flex; align-items: center; justify-content: center;
  background: rgba(15,15,26,0.85);
}
.countdown-num { font-size: 8rem; font-weight: 800; color: var(--accent); animation: pop 0.8s ease; }
@keyframes pop { 0% { transform: scale(2); opacity: 0; } 50% { transform: scale(1); opacity: 1; } }

.finished-overlay {
  position: fixed; inset: 0; z-index: 50;
  display: flex; align-items: center; justify-content: center;
  background: rgba(15,15,26,0.8); backdrop-filter: blur(6px);
}
.fo-content { text-align: center; }
.fo-icon { font-size: 3.5rem; margin-bottom: 0.5rem; }
.fo-title { font-size: 1.5rem; font-weight: 700; }
.fo-wpm { font-size: 3rem; font-weight: 800; color: var(--accent); font-family: var(--mono); margin: 0.5rem 0; }
.fo-wpm small { font-size: 1.2rem; color: var(--text-dim); }
.fo-waiting { color: var(--text-dim); font-size: 0.9rem; animation: pulse 1.5s infinite; margin-top: 0.5rem; }
@keyframes pulse { 0%,100% { opacity: 1 } 50% { opacity: 0.4 } }

.leaderboard {
  position: absolute; top: 10px; left: 10px; z-index: 10;
  background: rgba(15,15,26,0.85); backdrop-filter: blur(8px);
  border-radius: 10px; padding: 0.5rem; min-width: 160px; border: 1px solid var(--surface2);
}
.lb-header { font-size: 0.6rem; text-transform: uppercase; letter-spacing: 0.12em; color: var(--text-dim); padding: 0.3rem 0.5rem; }
.lb-entry { display: flex; align-items: center; gap: 0.5rem; padding: 0.3rem 0.5rem; font-size: 0.8rem; border-radius: 4px; }
.lb-pos { font-weight: 700; color: var(--text-dim); min-width: 18px; }
.lb-name { flex: 1; }
.lb-wpm { font-family: var(--mono); color: var(--accent); font-weight: 700; }
.lb-me { background: rgba(226,183,20,0.1); }

.progress-container { display: flex; flex-direction: column; gap: 3px; padding: 0 20px; }
.progress-bar-wrap { display: flex; align-items: center; gap: 8px; }
.prog-name { font-size: 0.65rem; font-weight: 600; min-width: 60px; text-align: right; color: var(--text-dim); white-space: nowrap; overflow: hidden; text-overflow: ellipsis; }
.prog-track { flex: 1; height: 6px; background: var(--surface2); border-radius: 3px; overflow: hidden; }
.prog-fill { height: 100%; border-radius: 3px; transition: width 0.15s; width: 0%; }
.prog-pct { font-size: 0.65rem; font-family: var(--mono); color: var(--text-dim); min-width: 35px; }

.typing-area { flex: 1; padding: 0.5rem 2rem; display: flex; flex-direction: column; justify-content: center; max-width: 860px; margin: 0 auto; width: 100%; overflow: auto; }
.passage-display { font-family: var(--mono); font-size: 1.15rem; line-height: 1.8; user-select: none; }
.p-char { color: var(--text-muted); }
.p-correct { color: var(--text); }
.p-incorrect { color: var(--incorrect); background: rgba(248,81,73,0.15); border-radius: 2px; }
.p-current { background: rgba(226,183,20,0.2); border-radius: 2px; }

.race-stats { display: flex; justify-content: center; gap: 2rem; padding: 0.5rem; }
.rstat { text-align: center; }
.rstat-val { font-size: 1.6rem; font-weight: 700; font-family: var(--mono); color: var(--accent); font-variant-numeric: tabular-nums; }
.rstat small { display: block; font-size: 0.65rem; text-transform: uppercase; letter-spacing: 0.08em; color: var(--text-dim); margin-top: 0.1rem; }

@media (max-width: 640px) {
  .leaderboard { display: none; }
  .typing-area { padding: 0.5rem 1rem; }
  .passage-display { font-size: 0.95rem; }
}
</style>
