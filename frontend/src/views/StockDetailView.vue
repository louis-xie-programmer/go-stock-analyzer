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
        <h3>K线（最近120日）</h3>
        <div id="chart" style="height:420px"></div>
      </div>

    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import axios from 'axios'
import * as echarts from 'echarts'

const route = useRoute()
const symbol = route.params.symbol
const stock = ref({})
const loading = ref(true)
const chartEl = ref(null)
let chartInst = null

async function fetchStock() {
  loading.value = true
  try {
    // use the existing /api/stocks?q= to find the single stock
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
    const el = document.getElementById('chart') || chartEl.value
    chartInst = echarts.init(el)
    const option = {
      xAxis: { data: dates },
      yAxis: { scale: true },
      series: [{ type: 'candlestick', data: values }]
    }
    chartInst.setOption(option)
  } catch (e) {
    console.error('kline load failed', e)
  }
}

onMounted(async () => {
  await fetchStock()
  await renderKline()
})
</script>
