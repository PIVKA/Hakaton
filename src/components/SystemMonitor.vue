<template>
    <div class="system-monitor">
      <div class="threshold-controls">
        <div class="threshold-input">
          <label>RAM Threshold (%):</label>
          <input type="number" v-model.number="ramThreshold" min="0" max="100">
        </div>
        <div class="threshold-input">
          <label>CPU Threshold (%):</label>
          <input type="number" v-model.number="cpuThreshold" min="0" max="100">
        </div>
        <button @click="toggleSound">{{ soundEnabled ? 'Disable' : 'Enable' }} Sound Alerts</button>
      </div>
  
      <div class="metrics-container">
        <div class="metric-card" :class="{ 'threshold-exceeded': currentMetrics.ram_usage > ramThreshold }">
          <h3>RAM Usage</h3>
          <div class="metric-value">{{ currentMetrics.ram_usage.toFixed(2) }}%</div>
          <div class="progress-bar">
            <div class="progress-fill" 
                 :style="{ width: currentMetrics.ram_usage + '%' }"
                 :class="{ 'threshold-exceeded': currentMetrics.ram_usage > ramThreshold }"></div>
          </div>
          <div class="threshold-indicator" :style="{ left: ramThreshold + '%' }"></div>
        </div>
  
        <div class="metric-card" :class="{ 'threshold-exceeded': currentMetrics.cpu_usage > cpuThreshold }">
          <h3>CPU Usage</h3>
          <div class="metric-value">{{ currentMetrics.cpu_usage.toFixed(2) }}%</div>
          <div class="progress-bar">
            <div class="progress-fill" 
                 :style="{ width: currentMetrics.cpu_usage + '%' }"
                 :class="{ 'threshold-exceeded': currentMetrics.cpu_usage > cpuThreshold }"></div>
          </div>
          <div class="threshold-indicator" :style="{ left: cpuThreshold + '%' }"></div>
        </div>
      </div>
  
      <div class="chart-container">
        <h3>Usage History</h3>
        <div class="chart-wrapper">
          <line-chart :chart-data="chartData" :options="chartOptions"></line-chart>
        </div>
      </div>
  
      <div class="connection-status">
        <span :class="['status-indicator', connectionStatus]"></span>
        {{ connectionText }}
        <span v-if="alertActive" class="alert-active">ALERT!</span>
      </div>
  
      <audio ref="alertSound" src="/alert.mp3" preload="auto"></audio>
    </div>
  </template>
  
  <script>
  import { Chart as ChartJS, Title, Tooltip, Legend, LineElement, LinearScale, PointElement, CategoryScale, TimeScale } from 'chart.js'
  
  ChartJS.register(Title, Tooltip, Legend, LineElement, LinearScale, PointElement, CategoryScale, TimeScale)
  
  export default {
    name: 'SystemMonitor',
    data() {
      return {
        ws: null,
        currentMetrics: {
          ram_usage: 0,
          cpu_usage: 0,
          timestamp: 0
        },
        metricsHistory: [],
        ramThreshold: 80,
        cpuThreshold: 70,
        soundEnabled: true,
        alertActive: false,
        connectionStatus: 'disconnected',
        chartOptions: {
          responsive: true,
          maintainAspectRatio: false,
          scales: {
            y: {
              beginAtZero: true,
              max: 100,
              ticks: {
                callback: function(value) {
                  return value + '%'
                }
              }
            },
            x: {
              type: 'time',
              time: {
                unit: 'second',
                displayFormats: {
                  second: 'HH:mm:ss'
                }
              }
            }
          },
          animation: {
            duration: 0
          }
        }
      }
    },
    computed: {
      chartData() {
        return {
          datasets: [
            {
              label: 'RAM Usage',
              data: this.metricsHistory.map(m => ({
                x: new Date(m.timestamp * 1000),
                y: m.ram_usage
              })),
              borderColor: '#42b983',
              backgroundColor: 'rgba(66, 185, 131, 0.1)',
              tension: 0.1
            },
            {
              label: 'CPU Usage',
              data: this.metricsHistory.map(m => ({
                x: new Date(m.timestamp * 1000),
                y: m.cpu_usage
              })),
              borderColor: '#647eff',
              backgroundColor: 'rgba(100, 126, 255, 0.1)',
              tension: 0.1
            }
          ]
        }
      },
      connectionText() {
        const statusMap = {
          disconnected: 'Disconnected',
          connecting: 'Connecting...',
          connected: 'Connected to server',
          error: 'Connection error'
        }
        return statusMap[this.connectionStatus]
      }
    },
    methods: {
      connectWebSocket() {
        this.connectionStatus = 'connecting'
        
        this.ws = new WebSocket('ws://localhost:8080/ws')
        
        this.ws.onopen = () => {
          this.connectionStatus = 'connected'
          console.log('WebSocket connected')
        }
        
        this.ws.onmessage = (event) => {
          const data = JSON.parse(event.data)
          this.updateMetrics(data)
        }
        
        this.ws.onerror = (error) => {
          this.connectionStatus = 'error'
          console.error('WebSocket error:', error)
        }
        
        this.ws.onclose = () => {
          this.connectionStatus = 'disconnected'
          console.log('WebSocket disconnected')
          setTimeout(() => this.connectWebSocket(), 3000)
        }
      },
      updateMetrics(metrics) {
        this.currentMetrics = metrics
        this.metricsHistory.push(metrics)
        
        // Ограничиваем историю последними 60 значениями
        if (this.metricsHistory.length > 60) {
          this.metricsHistory.shift()
        }
        
        // Проверка порогов
        this.checkThresholds()
      },
      checkThresholds() {
        const ramExceeded = this.currentMetrics.ram_usage > this.ramThreshold
        const cpuExceeded = this.currentMetrics.cpu_usage > this.cpuThreshold
        
        this.alertActive = ramExceeded || cpuExceeded
        
        if (this.alertActive && this.soundEnabled) {
          this.playAlertSound()
        }
      },
      playAlertSound() {
        const audio = this.$refs.alertSound
        audio.currentTime = 0
        audio.play().catch(e => console.log('Audio play failed:', e))
      },
      toggleSound() {
        this.soundEnabled = !this.soundEnabled
      }
    },
    mounted() {
      this.connectWebSocket()
    },
    beforeUnmount() {
      if (this.ws) {
        this.ws.close()
      }
    }
  }
  </script>
  
  <style scoped>
  .system-monitor {
    display: flex;
    flex-direction: column;
    gap: 20px;
  }
  
  .threshold-controls {
    display: flex;
    gap: 20px;
    justify-content: center;
    align-items: center;
    flex-wrap: wrap;
    background: white;
    padding: 15px;
    border-radius: 8px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  
  .threshold-input {
    display: flex;
    align-items: center;
    gap: 10px;
  }
  
  .threshold-input input {
    width: 60px;
    padding: 5px;
    border: 1px solid #ddd;
    border-radius: 4px;
  }
  
  .metrics-container {
    display: flex;
    gap: 20px;
    justify-content: center;
    flex-wrap: wrap;
  }
  
  .metric-card {
    position: relative;
    background: white;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
    min-width: 250px;
    flex: 1;
  }
  
  .metric-card.threshold-exceeded {
    animation: pulse 1.5s infinite;
  }
  
  @keyframes pulse {
    0% { box-shadow: 0 0 0 0 rgba(255, 71, 87, 0.4); }
    70% { box-shadow: 0 0 0 10px rgba(255, 71, 87, 0); }
    100% { box-shadow: 0 0 0 0 rgba(255, 71, 87, 0); }
  }
  
  .metric-value {
    font-size: 2rem;
    font-weight: bold;
    margin: 10px 0;
  }
  
  .progress-bar {
    position: relative;
    height: 20px;
    background: #eee;
    border-radius: 10px;
    overflow: hidden;
  }
  
  .progress-fill {
    height: 100%;
    transition: width 0.5s ease;
  }
  
  .metric-card:nth-child(1) .progress-fill {
    background: #42b983;
  }
  
  .metric-card:nth-child(1) .progress-fill.threshold-exceeded {
    background: #ff4757;
  }
  
  .metric-card:nth-child(2) .progress-fill {
    background: #647eff;
  }
  
  .metric-card:nth-child(2) .progress-fill.threshold-exceeded {
    background: #ff4757;
  }
  
  .threshold-indicator {
    position: absolute;
    top: 0;
    bottom: 0;
    width: 2px;
    background: #ff4757;
    transform: translateX(-50%);
  }
  
  .chart-container {
    background: white;
    border-radius: 8px;
    padding: 20px;
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
  
  .chart-wrapper {
    height: 300px;
    position: relative;
  }
  
  .connection-status {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    font-size: 0.9rem;
    color: #666;
  }
  
  .status-indicator {
    width: 10px;
    height: 10px;
    border-radius: 50%;
  }
  
  .status-indicator.disconnected {
    background: #ff4757;
  }
  
  .status-indicator.connecting {
    background: #ffa502;
  }
  
  .status-indicator.connected {
    background: #2ed573;
  }
  
  .status-indicator.error {
    background: #ff4757;
    animation: pulse 1.5s infinite;
  }
  
  .alert-active {
    color: #ff4757;
    font-weight: bold;
    margin-left: 10px;
    animation: blink 1s infinite;
  }
  
  @keyframes blink {
    0% { opacity: 1; }
    50% { opacity: 0.5; }
    100% { opacity: 1; }
  }
  
  button {
    padding: 8px 16px;
    background: #42b983;
    color: white;
    border: none;
    border-radius: 4px;
    cursor: pointer;
    transition: background 0.3s;
  }
  
  button:hover {
    background: #3aa876;
  }
  </style>