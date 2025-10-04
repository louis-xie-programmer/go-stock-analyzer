<template>
  <div class="p-4">
    <h2 class="text-xl font-bold mb-4">📊 股票池（可加入自选）</h2>

    <div class="flex items-center mb-4 space-x-2">
      <input
        v-model="keyword"
        placeholder="输入代码或名称搜索..."
        class="border rounded-lg px-3 py-2 w-64"
        @keyup.enter="fetchData"
      />
      <button @click="fetchData" class="bg-blue-500 text-white px-4 py-2 rounded-lg">🔍 搜索</button>
      <button @click="refresh" class="bg-green-500 text-white px-4 py-2 rounded-lg">🔄 刷新</button>
    </div>

    <table class="w-full border text-sm text-left">
      <thead class="bg-gray-100">
        <tr>
          <th class="p-2 border">代码</th>
          <th class="p-2 border">名称</th>
          <th class="p-2 border">现价</th>
          <th class="p-2 border text-center">操作</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="s in stocks" :key="s.code" class="hover:bg-gray-50">
          <td class="p-2 border">{{ s.code }}</td>
          <td class="p-2 border">{{ s.name }}</td>
          <td class="p-2 border">{{ s.trade.toFixed(2) }}</td>
          <td class="p-2 border text-center">
            <button
              v-if="!isInWatchlist(s)"
              @click="addToWatchlist(s)"
              class="bg-yellow-400 px-3 py-1 rounded text-white"
            >★ 加入</button>
            <button
              v-else
              @click="removeFromWatchlist(s)"
              class="bg-gray-400 px-3 py-1 rounded text-white"
            >✖ 移除</button>
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const keyword = ref('')
const stocks = ref([])
const watchlist = ref([])

async function fetchData() {
  const res = await axios.get('/api/stocks', { params: { q: keyword.value } })
  stocks.value = res.data.list
}

async function fetchWatchlist() {
  const res = await axios.get('/api/watchlist')
  watchlist.value = res.data
}

function isInWatchlist(s) {
  return watchlist.value.some(w => w.symbol === s.symbol)
}

async function addToWatchlist(s) {
  await axios.post('/api/watchlist/add', { symbol: s.symbol, name: s.name })
  await fetchWatchlist()
}

async function removeFromWatchlist(s) {
  await axios.delete('/api/watchlist/remove', { params: { symbol: s.symbol } })
  await fetchWatchlist()
}

function refresh() {
  keyword.value = ''
  fetchData()
  fetchWatchlist()
}

onMounted(() => {
  fetchData()
  fetchWatchlist()
})
</script>
