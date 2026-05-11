<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'

const props = defineProps({
  players: { type: Array, default: () => [] },
  myName: { type: String, default: '' },
})

const canvas = ref(null)
let ctx = null
let animFrame = null
let w = 0, h = 0
let roadOffset = 0

function resize() {
  if (!canvas.value) return
  const rect = canvas.value.parentElement.getBoundingClientRect()
  w = canvas.value.width = rect.width
  h = canvas.value.height = rect.height
  ctx = canvas.value.getContext('2d')
}

function draw() {
  if (!ctx) return
  roadOffset = (roadOffset + 1) % 100
  const W = w, H = h
  ctx.clearRect(0, 0, W, H)

  // Sky
  const sky = ctx.createLinearGradient(0, 0, 0, H * 0.55)
  sky.addColorStop(0, '#0a0a1a')
  sky.addColorStop(1, '#1a1a3a')
  ctx.fillStyle = sky
  ctx.fillRect(0, 0, W, H * 0.55)

  // Stars
  ctx.fillStyle = 'rgba(255,255,255,0.4)'
  for (let i = 0; i < 40; i++) {
    const sx = ((i * 137 + 12345) % W)
    const sy = ((i * 89 + 37035) % (H * 0.45))
    ctx.beginPath(); ctx.arc(sx, sy, ((i * 7 + 12345) % 2) + 0.5, 0, Math.PI * 2); ctx.fill()
  }

  // Grass
  ctx.fillStyle = '#0f2a0f'
  ctx.fillRect(0, H * 0.55, W, H * 0.45)

  // Road
  const numSegs = 40
  const vpY = H * 0.45
  const roadTop = W * 0.015
  const roadBot = W * 0.55

  for (let i = 0; i < numSegs; i++) {
    const t1 = i / numSegs
    const t2 = (i + 1) / numSegs
    const x1 = W/2 - (roadBot/2)*(1-t1) - (roadTop/2)*t1
    const x2 = W/2 + (roadBot/2)*(1-t1) + (roadTop/2)*t1
    const x3 = W/2 + (roadBot/2)*(1-t2) + (roadTop/2)*t2
    const x4 = W/2 - (roadBot/2)*(1-t2) - (roadTop/2)*t2
    const y1 = vpY + (H - vpY) * (1 - t1)
    const y2 = vpY + (H - vpY) * (1 - t2)
    const b = 30 + 40 * (1 - t1)
    ctx.fillStyle = `rgb(${b},${b+5},${b+10})`
    ctx.beginPath(); ctx.moveTo(x1,y1); ctx.lineTo(x2,y1); ctx.lineTo(x3,y2); ctx.lineTo(x4,y2); ctx.closePath(); ctx.fill()

    // Rumble strips
    const sw = (x2-x1)*0.02
    const lit = Math.floor((i + roadOffset/5)/2) % 2 === 0
    ctx.fillStyle = lit ? '#e2b714' : '#8a7a10'
    ctx.beginPath(); ctx.moveTo(x1,y1); ctx.lineTo(x1+sw*2,y1); ctx.lineTo(x4+sw*2,y2); ctx.lineTo(x4,y2); ctx.closePath(); ctx.fill()
    ctx.beginPath(); ctx.moveTo(x2-sw*2,y1); ctx.lineTo(x2,y1); ctx.lineTo(x3,y2); ctx.lineTo(x3-sw*2,y2); ctx.closePath(); ctx.fill()

    // Lane dashes
    if (i % 4 === Math.floor(roadOffset/10) % 4) {
      ctx.fillStyle = '#e2b714'
      const cx1 = x1 + (x2-x1)*0.48, cx2 = x1 + (x2-x1)*0.52
      ctx.beginPath(); ctx.moveTo(cx1,y1); ctx.lineTo(cx2,y1); ctx.lineTo(x3-(x3-x4)*0.48,y2); ctx.lineTo(x3-(x3-x4)*0.52,y2); ctx.closePath(); ctx.fill()
    }
  }

  // Finish line
  ctx.fillStyle = '#ff4757'
  ctx.fillRect(W/2 - W*0.12, vpY - 2, W*0.24, 3)

  // ─── Draw cars ────────────────────────────────────────────────────────
  const players = props.players
  if (players && players.length > 0) {
    players.forEach(p => {
      const prog = p.progress || 0
      const t = (prog / 100) * 0.85
      const y = vpY + (H - vpY) * (1 - t)
      const scale = 0.6 + 0.8 * (1 - t)
      const cw = 70 * scale, ch = 32 * scale
      const x = W / 2
      const isMe = p.name === props.myName

      // Shadow
      ctx.fillStyle = 'rgba(0,0,0,0.3)'
      ctx.beginPath(); ctx.ellipse(x, y + ch*0.5 + 3, cw*0.55, ch*0.15, 0, 0, Math.PI*2); ctx.fill()

      // Name
      ctx.fillStyle = isMe ? '#e2b714' : (p.color || '#aaa')
      ctx.font = `${Math.max(9, 11*scale)}px Inter, sans-serif`
      ctx.textAlign = 'center'
      ctx.fillText(p.name, x, y - ch/2 - 6)

      // Body
      ctx.fillStyle = p.color || '#666'
      rRect(ctx, x - cw/2, y - ch/2, cw, ch, 4); ctx.fill()

      // Windshield
      ctx.fillStyle = 'rgba(255,255,255,0.25)'
      rRect(ctx, x - cw*0.25, y - ch*0.35, cw*0.5, ch*0.4, 2); ctx.fill()

      // Wheels
      ctx.fillStyle = '#222'
      ctx.fillRect(x - cw*0.35, y - ch*0.3, cw*0.12, ch*0.2)
      ctx.fillRect(x + cw*0.35 - cw*0.12, y - ch*0.3, cw*0.12, ch*0.2)

      // Gold ring for me
      if (isMe) {
        ctx.strokeStyle = '#e2b714'
        ctx.lineWidth = 2
        rRect(ctx, x - cw/2 - 3, y - ch/2 - 3, cw + 6, ch + 6, 6); ctx.stroke()
      }
    })
  }

  animFrame = requestAnimationFrame(draw)
}

function rRect(ctx, x, y, w, h, r) {
  ctx.beginPath()
  ctx.moveTo(x+r, y); ctx.lineTo(x+w-r, y)
  ctx.quadraticCurveTo(x+w, y, x+w, y+r)
  ctx.lineTo(x+w, y+h-r)
  ctx.quadraticCurveTo(x+w, y+h, x+w-r, y+h)
  ctx.lineTo(x+r, y+h)
  ctx.quadraticCurveTo(x, y+h, x, y+h-r)
  ctx.lineTo(x, y+r)
  ctx.quadraticCurveTo(x, y, x+r, y)
  ctx.closePath()
}

onMounted(() => {
  resize()
  draw()
  window.addEventListener('resize', resize)
})
onUnmounted(() => {
  if (animFrame) cancelAnimationFrame(animFrame)
  window.removeEventListener('resize', resize)
})
</script>

<template>
  <div class="track-wrap">
    <canvas ref="canvas"></canvas>
  </div>
</template>

<style scoped>
.track-wrap { width: 100%; height: 280px; overflow: hidden; }
.track-wrap canvas { width: 100%; height: 100%; display: block; }
@media (max-width: 640px) {
  .track-wrap { height: 180px; }
}
</style>
