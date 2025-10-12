<template>
  <div style="padding:20px">
    <h2>自选股</h2>
    <div style="margin-bottom:10px">
      <button @click="fetchWatchlist">刷新</button>
    </div>
    <el-table :data="watchlist" style="width: 100%; margin-top: 20px;">
      <el-table-column prop="symbol" label="代码" width="100">
        <template #default="scope">
          <router-link :to="`/stocks/${scope.row.symbol}`">{{ scope.row.symbol }}</router-link>
        </template>
      </el-table-column>
      <el-table-column prop="name" label="名称"/>
      <el-table-column prop="added_at" label="添加时间"/>
      <el-table-column label="操作" width="120">
        <template #default="scope">
          <el-button size="small" type="danger" @click="removeFromWatchlist(scope.row)">移除</el-button>
        </template>
      </el-table-column>
    </el-table>
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
