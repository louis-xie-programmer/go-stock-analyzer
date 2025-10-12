<template>
  <div style="padding:20px">
    <h2>符合策略的股票列表</h2>
    <el-table :data="stocks" style="width: 100%; margin-top: 20px;">
      <el-table-column prop="code" label="代码" width="100"/>
      <el-table-column prop="strategy" label="策略"/>
      <el-table-column prop="date" label="日期"/>
    </el-table>
  </div>
</template>

<script>
import axios from 'axios'
export default {
  data() { return { stocks: [] } },
  mounted() { this.fetchResults() },
  methods: {
    async fetchResults() {
      const res = await axios.get('/api/results')
      this.stocks = res.data
    }
  }
}
</script>
