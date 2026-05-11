<script setup>
import { inject } from 'vue'
const connected = inject('connected')
</script>

<template>
  <div class="conn-status" :class="{ 'conn-on': connected, 'conn-off': !connected }">
    {{ connected ? '● connected' : '● disconnected' }}
  </div>
</template>

<style scoped>
.conn-status {
  position: fixed; bottom: 10px; right: 10px; z-index: 999;
  padding: 4px 10px; border-radius: 6px; font-size: 0.7rem; font-family: var(--mono);
  transition: all 0.3s;
}
.conn-on { background: rgba(63,185,80,0.15); color: #3fb950; }
.conn-off { background: rgba(248,81,73,0.15); color: #f85149; }
</style>
