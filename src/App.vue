<template>
  <div id="app">
    <header class="header">
      <h1>System Monitor</h1>
    </header>
    <main class="main-content">
      <SystemMonitor />
    </main>
  </div>
</template>

<script>
import SystemMonitor from './components/SystemMonitor.vue'

export default {
  name: 'App',
  components: {
    SystemMonitor
  }
}
</script>

<style>
#app {
  font-family: Avenir, Helvetica, Arial, sans-serif;
  text-align: center;
  color: #2c3e50;
  margin: 0;
  padding: 0;
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.header {
  background-color: #42b983;
  color: white;
  padding: 20px;
  box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
}

.main-content {
  flex: 1;
  padding: 20px;
  max-width: 1000px;
  margin: 0 auto;
  width: 100%;
}
</style>