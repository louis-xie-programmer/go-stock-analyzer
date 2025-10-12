<template>
  <div style="padding:20px">
    <h2>股票池</h2>
    <div style="margin-bottom:10px">
      <input v-model="keyword" placeholder="搜索代码或名称" />
      <button @click="fetchData">搜索</button>
      <button @click="refresh">刷新</button>
    </div>
    <el-table :data="stocks" style="width: 100%; margin-top: 20px;">
      <el-table-column prop="code" label="代码" width="100">
        <template #default="scope">
          <router-link :to="`/stocks/${scope.row.symbol}`">{{ scope.row.code }}</router-link>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称">
        <template #default="scope">
          <router-link :to="`/stocks/${scope.row.symbol}`">{{ scope.row.name }}</router-link>
        </template>
      </el-table-column>
      <el-table-column prop="trade" label="现价" width="100"/>
      <el-table-column label="操作" width="120">
        <template #default="scope">
          <el-button size="small" type="primary" @click="addToWatchlist(scope.row)">加入自选</el-button>
        </template>
      </el-table-column>
    </el-table>
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
