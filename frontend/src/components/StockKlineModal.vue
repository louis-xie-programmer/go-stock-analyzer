<template>
  <div class="modal" style="position:fixed;left:0;top:0;right:0;bottom:0;background:rgba(0,0,0,0.5)">
    <div style="width:900px;margin:40px auto;background:#fff;padding:16px;position:relative">
      <button style="position:absolute;right:8px;top:8px" @click="$emit('close')">关闭</button>
      <div ref="chart" style="height:420px"></div>
    </div>
  </div>
</template>

<script>
import * as echarts from 'echarts'
import axios from 'axios'
import { onMounted, ref } from 'vue'
export default {
  props: ['symbol'],
  setup(props) {
    const chart = ref(null)
    let inst = null
    onMounted(async () => {
      chart.value = document.getElementById('chart') || null
      // find by template ref workaround
      const el = document.getElementById('chart') || document.querySelector('.modal div')
      if (!el) return
      inst = echarts.init(el)
      const res = await axios.get('/api/kline', { params: { symbol: props.symbol, datalen: 120 } })
      const data = res.data
      // transform to candlestick arrays
      const dates = data.map(item => item.date)
      const values = data.map(item => [item.open, item.close, item.low, item.high])
      const option = {
        xAxis: { data: dates },
        yAxis: { scale: true },
        series: [{ type: 'candlestick', data: values }]
      }
      inst.setOption(option)
    })
    return {}
  }
}
</script>
