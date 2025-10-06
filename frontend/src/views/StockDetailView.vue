<template>
  <div style="padding:20px">
    <h2>个股详情 - {{ stock.code || symbol }}</h2>
    <div v-if="loading">加载中...</div>
    <div v-else>
      <p><strong>代码:</strong> {{ stock.code }}</p>
      <p><strong>名称:</strong> {{ stock.name }}</p>
      <p><strong>现价:</strong> {{ stock.trade }}</p>
      <p><strong>板块:</strong> {{ stock.board }}</p>

      <div style="margin-top:16px">
        <div class="chart-header">
          <h3>{{ chartTitle }}</h3>
          <div class="chart-types">
            <button 
              :class="{ active: chartType === 'timeline' }"
              @click="switchChartType('timeline')">
              分时
            </button>
            <button 
              :class="{ active: chartType === 'kline' }"
              @click="switchChartType('kline')">
              日K
            </button>
          </div>
        </div>
        
        <TimelineChart
          v-if="chartType === 'timeline'"
          :symbol="symbol"
          :data="timelineData"
          :height="420"
        />
        <div 
          v-else
          ref="klineChart" 
          style="height:420px">
        </div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, computed, watch, onMounted, onUnmounted, nextTick } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import * as echarts from 'echarts'
import TimelineChart from '../components/TimelineChart.vue'

const route = useRoute()
const symbol = route.params.symbol
const stock = ref({})
const loading = ref(true)
const chartType = ref('timeline') // 默认显示分时图
const timelineData = ref([])
const klineChart = ref(null)
const wsStatus = ref('disconnected') // WebSocket 状态
const retryCount = ref(0) // WebSocket 重试计数
const MAX_RETRIES = 3
let klineInst = null
let wsConn = null
let wsReconnectTimer = null
const isTradingTime = ref(false)

const chartTitle = computed(() => {
  return chartType.value === 'timeline' ? '分时图' : 'K线（最近120日）'
})

// 新增：监听图表类型变化
watch(chartType, async (newType) => {
  if (newType === 'timeline') {
    if (isTradingTime.value) {
      ensureWebSocketConnection()
    } else {
      loadLatestTimelineData()
    }
  } else {
    await renderKline()
    cleanupWebSocket()
  }
})

async function fetchStock() {
  loading.value = true
  try {
    const res = await axios.get('/api/stocks', { params: { q: symbol, size: 1 } })
    const list = res.data.list || []
    if (list.length > 0) {
      stock.value = list[0]
    } else {
      stock.value = { code: symbol, name: '-', trade: '-' }
    }
  } catch (e) {
    console.error(e)
    stock.value = { code: symbol, name: '-', trade: '-' }
  } finally {
    loading.value = false
  }
}

async function renderKline() {
  try {
    const res = await axios.get('/api/kline', { params: { symbol, datalen: 120 } })
    const data = res.data
    const dates = data.map(item => item.date)
    const values = data.map(item => [item.open, item.close, item.low, item.high])
    
    //if (!klineInst && klineChart.value) {
      klineInst = echarts.init(klineChart.value)
    //}
    
    const option = {
      tooltip: {
        trigger: 'axis',
        axisPointer: { type: 'cross' }
      },
      grid: { left: '10%', right: '10%' },
      xAxis: { 
        type: 'category',
        data: dates,
        scale: true,
        boundaryGap: false,
        axisLine: { onZero: false },
        splitLine: { show: false }
      },
      yAxis: { 
        scale: true,
        splitArea: { show: true }
      },
      dataZoom: [
        { type: 'inside', start: 50, end: 100 },
        { show: true, type: 'slider', bottom: '0%' }
      ],
      series: [{ 
        type: 'candlestick', 
        data: values,
        itemStyle: {
          color: '#f5222d',
          color0: '#52c41a',
          borderColor: '#f5222d',
          borderColor0: '#52c41a'
        }
      }]
    }
    klineInst.setOption(option)
  } catch (e) {
    console.error('kline load failed', e)
  }
}

function cleanupWebSocket() {
  if (wsReconnectTimer) {
    clearTimeout(wsReconnectTimer)
    wsReconnectTimer = null
  }
  
  if (wsConn) {
    try {
      wsConn.onclose = null // 移除重连逻辑
      wsConn.close()
      wsConn = null
    } catch (e) {
      console.error('Error closing websocket:', e)
    }
  }
  wsStatus.value = 'disconnected'
  retryCount.value = 0
}

function ensureWebSocketConnection() {
  if (wsConn?.readyState === WebSocket.OPEN) {
    return // 已连接
  }
  cleanupWebSocket() // 清理旧连接
  connectWebSocket() // 建立新连接
}

function handleReconnect() {
  if (chartType.value !== 'timeline') return
  
  retryCount.value++
  if (retryCount.value <= MAX_RETRIES) {
    const delay = Math.min(1000 * Math.pow(2, retryCount.value), 10000)
    console.log(`Retry ${retryCount.value}/${MAX_RETRIES} in ${delay}ms`)
    
    wsReconnectTimer = setTimeout(() => {
      if (chartType.value === 'timeline') {
        connectWebSocket()
      }
    }, delay)
  } else {
    console.log('Max retry attempts reached')
    retryCount.value = 0
  }
}

function connectWebSocket() {
  if (wsConn || retryCount.value >= MAX_RETRIES) return

  wsStatus.value = 'connecting'
  const wsHost = import.meta.env.DEV ? 'localhost:8080' : window.location.host
  wsConn = new WebSocket(`ws://${wsHost}/ws/realtime`)
  
  // 连接超时处理
  const connectTimeout = setTimeout(() => {
    if (wsConn?.readyState === WebSocket.CONNECTING) {
      wsConn.close()
      wsStatus.value = 'error'
      handleReconnect()
    }
  }, 5000)
  
  wsConn.onopen = () => {
    clearTimeout(connectTimeout)
    console.log('WebSocket connected')
    wsStatus.value = 'connected'
    retryCount.value = 0 // 重置重试计数
    
    wsConn.send(JSON.stringify({
      action: 'subscribe',
      symbols: [symbol]
    }))
  }
  
  wsConn.onmessage = async (event) => {
    if (chartType.value !== 'timeline') return
    try {
      const data = JSON.parse(event.data)
      // 格式化time为HH:mm:ss
      let tstr = data.time
      if (typeof tstr === 'number') {
        const date = new Date(tstr)
        tstr = date.toLocaleTimeString('zh-CN', { hour12: false })
      } else if (/^\d{1,2}:\d{1,2}:\d{1,2}$/.test(tstr)) {
        // 已是字符串
      } else {
        try {
          tstr = new Date(tstr).toLocaleTimeString('zh-CN', { hour12: false })
        } catch {}
      }
      timelineData.value.push({
        time: tstr,
        price: data.price,
        volume: data.volume
      })
    } catch (e) {
      console.error('Error parsing message:', e)
    }
  }
  
  wsConn.onerror = (event) => {
    console.error('WebSocket error:', {
      type: event.type,
      timestamp: event.timeStamp,
      target: event.target?.url || 'unknown'
    })
    wsStatus.value = 'error'
  }
  
  wsConn.onclose = (event) => {
    console.log('WebSocket closed:', {
      code: event.code,
      reason: event.reason || 'No reason provided',
      wasClean: event.wasClean
    })
    wsStatus.value = 'disconnected'
    wsConn = null
    handleReconnect()
  }
}

onMounted(async () => {
  await fetchStock()
  // 验证是否在交易时间
  try {
    const res = await axios.get('/api/is_market_open', { params: { symbol } })
    isTradingTime.value = res.data.is_open
  } catch (e) {
    console.error('is_market_open check failed', e)
    isTradingTime.value = false
  }
  if (chartType.value === 'kline') {
    await renderKline()
  } else {
    if (isTradingTime.value) {
      ensureWebSocketConnection()
    } else {
      loadLatestTimelineData()
    }
  }
  window.addEventListener('resize', handleResize)
})

onUnmounted(() => {
  cleanupWebSocket() // 清理 WebSocket
  
  if (klineInst) {
    klineInst.dispose()
    klineInst = null
  }
  
  window.removeEventListener('resize', handleResize)
})

function switchChartType(type) {
  if (type === chartType.value) return // 避免重复切换
  if (type === 'kline') {
    chartType.value = type
    nextTick(() => renderKline())
  } else {
    chartType.value = type
    if (isTradingTime.value) {
      ensureWebSocketConnection()
    } else {
      loadLatestTimelineData()
    }
  }
}

// 加载最近一个交易日的分时数据
async function loadLatestTimelineData() {
  try {
    const res = await axios.get('/api/timeline', { params: { symbol } })
    // 假设返回格式为 [{time, price, volume}]
    const raw = res.data || []
    raw.forEach(item => {
      let tstr = item.time
      if (typeof tstr === 'number') {
        const date = new Date(tstr)
        tstr = date.toLocaleTimeString('zh-CN', { hour12: false })
      } else if (/^\d{1,2}:\d{1,2}:\d{1,2}$/.test(tstr)) {
        // already hh:mm:ss
      } else {
        try {
          tstr = new Date(tstr).toLocaleTimeString('zh-CN', { hour12: false })
        } catch {}
      }
      // normalize to HH:MM:SS
      if (/^\d{1}:/.test(tstr)) {
        // pad hour if needed
        const parts = tstr.split(':')
        parts[0] = parts[0].padStart(2, '0')
        tstr = parts.join(':')
      }
      const price = Number(item.price ?? item.close ?? 0)
      const volume = Number(item.volume ?? 0)
      timelineData.value.push({
        time: tstr,
        price: price,
        volume: volume
      })
    })
    let lastPrice = raw[0]?.price ?? null
    timelineData.value.forEach(item => {
      if (item.price != null) {
        lastPrice = item.price
      } else if (lastPrice != null) {
        item.price = lastPrice
      }
    })
  } catch (e) {
    console.error('loadLatestTimelineData failed', e)
  }
}

const handleResize = () => {
  if (klineInst) klineInst.resize()
}
</script>

<style scoped>
.chart-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
}

.chart-types {
  display: flex;
  gap: 8px;
}

.chart-types button {
  padding: 4px 12px;
  border: 1px solid #d9d9d9;
  background: white;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.3s;
}

.chart-types button:hover {
  border-color: #1890ff;
  color: #1890ff;
}

.chart-types button.active {
  background: #1890ff;
  border-color: #1890ff;
  color: white;
}
</style>
