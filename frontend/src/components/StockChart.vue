<template>
  <div ref="chart" style="width:100%;height:400px"></div>
</template>

<script>
import * as echarts from 'echarts'
export default {
  props: ['klines'],
  mounted() {
    const chart = echarts.init(this.$refs.chart)
    chart.setOption({
      xAxis: { type:'category', data:this.klines.map(k=>k.date)},
      yAxis: { type:'value' },
      series: [{ type:'line', data:this.klines.map(k=>k.close)}]
    })
  }
}
</script>
