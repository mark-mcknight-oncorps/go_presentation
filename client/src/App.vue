<script setup lang="ts">
  const socket = new WebSocket('ws://localhost:8080/ws');
  socket.onmessage = (evt) => {
    const square = document.getElementById(evt.data);
    if (!square) return;
    if (square.classList.contains('alive')) {
      square.className = 'square';
    } else {
      square.className = 'square alive';
    }
  }
  const start = () => {
    socket.send('start');
  }
</script>

<template>
  <div class="main">
    <div class="grid">
      <div class="row" v-for="y in 20">
        <div class="square" v-for="x in 20" :id="x + ',' + y"></div>
      </div>
      <div>
        <button @click="start">Start</button>
      </div>
    </div>
  </div>
</template>

<style scoped>
.main {
  width: 100%;
}
.grid {
  margin: 0 auto;
}
.square {
  display: inline-block;
  width: 30px;
  height: 30px;
  margin: 3px;
  border: 1px solid black;
  background: white;
}
.alive {
  background: black;
}
</style>
