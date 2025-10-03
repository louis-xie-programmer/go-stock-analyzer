<template>
  <div style="padding:20px">
    <h2>符合策略的股票列表</h2>
    <table border="1" cellspacing="0" cellpadding="5">
      <tr>
        <th>股票代码</th>
        <th>策略</th>
        <th>日期</th>
      </tr>
      <tr v-for="s in stocks" :key="s.code + s.strategy">
        <td>{{ s.code }}</td>
        <td>{{ s.strategy }}</td>
        <td>{{ s.date }}</td>
      </tr>
    </table>
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
