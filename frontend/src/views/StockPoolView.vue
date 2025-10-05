<template>
  <div style="padding:20px">
    <h2>股票池</h2>
    <div style="margin-bottom:10px">
      <input v-model="keyword" placeholder="搜索代码或名称" />
      <button @click="fetchData">搜索</button>
      <button @click="refresh">刷新</button>
    </div>
    <table border="1" cellpadding="6">
      <tr><th>代码</th><th>名称</th><th>现价</th><th>操作</th></tr>
      <tr v-for="s in stocks" :key="s.symbol">
        <td><router-link :to="`/stocks/${s.symbol}`">{{ s.code }}</router-link></td>
        <td><router-link :to="`/stocks/${s.symbol}`">{{ s.name }}</router-link></td>
        <td>{{ s.trade }}</td>
        <td>
          <button @click="addToWatchlist(s)">加入自选</button>
        </td>
      </tr>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const keyword = ref('')
const stocks = ref([])

async function fetchData() {
  const res = await axios.get('/api/stocks', { params: { q: keyword.value, size: 200 } })
  stocks.value = res.data.list
}
function refresh() {
  keyword.value = ''
  fetchData()
}

async function addToWatchlist(s) {
  await axios.post('/api/watchlist/add', { symbol: s.symbol, name: s.name })
  alert('已加入自选')
}

onMounted(fetchData)
</script>
