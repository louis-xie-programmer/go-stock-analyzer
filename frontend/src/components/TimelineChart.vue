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
function formatTime(timeStr) {
  if (!timeStr) return ''
  if (typeof timeStr !== 'string') timeStr = String(timeStr)
  // 如果包含空格，取后半部分
  if (timeStr.indexOf(' ') >= 0) {
    const parts = timeStr.split(' ')
    timeStr = parts[1] || parts[0]
  }
  // 如果形如 HH:MM:SS，取 HH:MM
  if (/^\d{1,2}:\d{1,2}:\d{1,2}$/.test(timeStr)) {
    return timeStr.substring(0,5)
  }
  // 如果形如 HH:MM
  if (/^\d{1,2}:\d{1,2}$/.test(timeStr)) {
    const p = timeStr.split(':')
    return p[0].padStart(2,'0') + ':' + p[1].padStart(2,'0')
  }
  // fallback
  return timeStr
}

function renderChart() {
  if (!chartEl.value || !props.data.length) return
  
  const times = props.data.map(item => formatTime(item.time))
  const prices = props.data.map(item => item.price)
  const volumes = props.data.map(item => item.volume)
  
  const basePriceRaw = props.data[0] && props.data[0].price
  const basePrice = Number(basePriceRaw)
  
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
        const time = times[params[0].dataIndex]
        const price = prices[params[0].dataIndex]
        const volume = volumes[params[0].dataIndex]
        const change = ((price - basePrice) / basePrice * 100).toFixed(2)
        return `时间：${time}<br/>
                价格：${price}<br/>
                成交量：${volume}<br/>
                涨跌幅：${change}%`
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
        max: 'dataMax'
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
        max: 'dataMax'
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
        markLine: (isFinite(basePrice) ? {
          data: [{ yAxis: basePrice }],
          label: { formatter: '{c}' }
        } : { data: [] }),
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