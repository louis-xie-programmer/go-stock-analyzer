<template>
  <div style="padding:20px">
    <h2>自选股</h2>
    <div style="margin-bottom:10px">
      <button @click="fetchWatchlist">刷新</button>
    </div>
    <table border="1" cellpadding="6">
      <tr><th>代码</th><th>名称</th><th>添加时间</th><th>操作</th></tr>
      <tr v-for="s in watchlist" :key="s.symbol">
        <td><router-link :to="`/stocks/${s.symbol}`">{{ s.symbol }}</router-link></td>
        <td><router-link :to="`/stocks/${s.symbol}`">{{ s.name }}</router-link></td>
        <td>{{ s.added_at }}</td>
        <td>
          <button @click="removeFromWatchlist(s)">移除</button>
        </td>
      </tr>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const watchlist = ref([])

async function fetchWatchlist() {
  try {
    const res = await axios.get('/api/watchlist')
    watchlist.value = res.data
  } catch (e) {
    console.error(e)
    alert('无法获取自选股')
  }
}
async function removeFromWatchlist(s) {
  if (!confirm(`确认从自选中移除 ${s.name} (${s.symbol}) ?`)) return
  try {
    await axios.delete('/api/watchlist/remove', { params: { symbol: s.symbol } })
    await fetchWatchlist()
  } catch (e) {
    console.error(e)
    alert('移除失败')
  }
}

onMounted(fetchWatchlist)
</script>
