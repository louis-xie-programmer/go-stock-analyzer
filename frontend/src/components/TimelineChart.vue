<template>
  <div class="timeline-chart" :style="{ height: height + 'px' }">
    <div ref="chartEl" style="width:100%;height:100%"></div>
  </div>
</template>

<script setup>
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps({
  symbol: { type: String, required: true },
  height: { type: Number, default: 420 },
  data: { type: Array, default: () => [] }
})

const chartEl = ref(null)
let chart = null

// 格式化时间轴标签为 HH:mm（输入可为 'YYYY-MM-DD HH:MM:SS' 或 'HH:MM:SS' 或 'HH:MM'）
// normalize to HH:MM:SS string
function normalizeToHHMMSS(timeStr) {
  if (!timeStr) return ''
  if (typeof timeStr !== 'string') timeStr = String(timeStr)
  if (timeStr.indexOf(' ') >= 0) {
    const parts = timeStr.split(' ')
    timeStr = parts[1] || parts[0]
  }
  if (/^\d{1,2}:\d{1,2}:\d{1,2}$/.test(timeStr)) {
    const parts = timeStr.split(':')
    return parts[0].padStart(2,'0') + ':' + parts[1].padStart(2,'0') + ':' + parts[2].padStart(2,'0')
  }
  if (/^\d{1,2}:\d{1,2}$/.test(timeStr)) {
    const p = timeStr.split(':')
    return p[0].padStart(2,'0') + ':' + p[1].padStart(2,'0') + ':00'
  }
  // try Date parse
  try {
    const d = new Date(timeStr)
    return d.toTimeString().split(' ')[0]
  } catch {
    return ''
  }
}

// format label as HH:mm (used for axis labels)
function formatLabelHHMM(hhmmss) {
  if (!hhmmss) return ''
  return hhmmss.substring(0,5)
}

// generate fixed x axis times from 09:30:00 to 15:00:00 step 1s
function generateFixedTimes() {
  const times = []
  const start = 9 * 3600 + 30 * 60
  const end = 15 * 3600 + 6
  for (let t = start; t <= end; t++) {
    const h = String(Math.floor(t / 3600)).padStart(2, '0')
    const m = String(Math.floor((t % 3600) / 60)).padStart(2, '0')
    const s = String(t % 60).padStart(2, '0')
    times.push(`${h}:${m}:${s}`)
  }
  return times
}

function renderChart() {
  if (!chartEl.value) return

  // fixed axis
  const times = generateFixedTimes()
  const idxMap = new Map()
  times.forEach((t, i) => idxMap.set(t, i))

  // create arrays filled with null/0
  const prices = new Array(times.length).fill(null)
  const volumes = new Array(times.length).fill(0)

  // map incoming data to axis
  props.data.forEach(item => {
    const t = normalizeToHHMMSS(item.time)
    if (!t) return
    const i = idxMap.get(t)
    if (i === undefined) return
    const p = Number(item.price)
    const v = Number(item.volume || 0)
    prices[i] = isFinite(p) ? p : null
    volumes[i] = isFinite(v) ? v : 0
  })

  prices[0] = props.data[0]?.price ?? null
  prices.forEach((p, i) => {
    if (p === null && i > 0) {
      // carry forward last known price
      prices[i] = prices[i - 1]
    }
  })

  // find basePrice = first non-null price
  let basePrice = null
  for (let i = 0; i < prices.length; i++) {
    if (prices[i] != null) { basePrice = prices[i]; break }
  }
  
  // 计算涨跌颜色
  const colors = prices.map(p => p >= basePrice ? '#f5222d' : '#52c41a')
  
  if (!chart) {
    chart = echarts.init(chartEl.value)
  }
  
  const option = {
    title: { text: '分时图', left: 'center' },
    tooltip: {
      trigger: 'axis',
      axisPointer: { type: 'cross' },
      formatter: (params) => {
        const idx = params[0]?.dataIndex
        const time = times[idx]
        const price = prices[idx]
        const volume = volumes[idx]
        const change = (basePrice ? ((price - basePrice) / basePrice * 100).toFixed(2) : '0.00')
        return `时间：${formatLabelHHMM(time)}<br/>价格：${price == null ? '-' : price}<br/>成交量：${volume}<br/>涨跌幅：${price == null ? '-' : change + '%'} `
      }
    },
    grid: [
      { left: '10%', right: '8%', height: '60%' },
      { left: '10%', right: '8%', top: '75%', height: '20%' }
    ],
    xAxis: [
      {
        type: 'category',
        data: times,
        scale: true,
        boundaryGap: false,
        axisLine: { onZero: false },
        splitLine: { show: false },
        min: 'dataMin',
        max: 'dataMax',
        axisLabel: {
          formatter: (val) => {
            // show label only on minute boundary to avoid clutter
            return val && val.endsWith(':00') ? formatLabelHHMM(val) : ''
          }
        }
      },
      {
        type: 'category',
        gridIndex: 1,
        data: times,
        scale: true,
        boundaryGap: false,
        axisLine: { onZero: false },
        splitLine: { show: false },
        min: 'dataMin',
        max: 'dataMax',
        axisLabel: { show: false }
      }
    ],
    yAxis: [
      {
        scale: true,
        splitArea: { show: true }
      },
      {
        scale: true,
        gridIndex: 1,
        splitNumber: 2,
        axisLabel: { show: false },
        axisLine: { show: false },
        axisTick: { show: false },
        splitLine: { show: false }
      }
    ],
    dataZoom: [
      {
        type: 'inside',
        xAxisIndex: [0, 1],
        start: 0,
        end: 100
      },
      {
        show: true,
        xAxisIndex: [0, 1],
        type: 'slider',
        bottom: '0%',
        start: 0,
        end: 100
      }
    ],
    series: [
      {
        name: '价格',
        type: 'line',
        data: prices,
        smooth: true,
        showSymbol: false,
        lineStyle: {
          color: '#1890ff',
          width: 1
        },
        markLine: (basePrice != null ? { data: [{ yAxis: basePrice }], label: { formatter: '{c}' } } : { data: [] }),
        areaStyle: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
            { offset: 0, color: 'rgba(24,144,255,0.3)' },
            { offset: 1, color: 'rgba(24,144,255,0.1)' }
          ])
        }
      },
      {
        name: '成交量',
        type: 'bar',
        xAxisIndex: 1,
        yAxisIndex: 1,
        data: volumes,
        itemStyle: {
          color: (params) => colors[params.dataIndex]
        }
      }
    ]
  }
  
  chart.setOption(option)
}

watch(() => props.data, renderChart, { deep: true })

const _resizeHandler = () => chart?.resize()
onMounted(() => {
  renderChart()
  window.addEventListener('resize', _resizeHandler)
})

onUnmounted(() => {
  chart?.dispose()
  window.removeEventListener('resize', _resizeHandler)
})
</script>

<style scoped>
.timeline-chart {
  width: 100%;
  min-height: 300px;
}
</style>