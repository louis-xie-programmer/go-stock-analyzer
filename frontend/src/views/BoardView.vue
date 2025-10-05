<template>
  <div style="padding:20px">
    <h2>按板块浏览</h2>
    <div style="margin:10px 0">
      <button v-for="b in boards" :key="b" @click="selectBoard(b)" :style="{marginRight:'8px'}">{{ b }}</button>
    </div>
    <div v-if="selected">
      <h3>{{ selected }}</h3>
      <table border="1" cellpadding="6">
        <tr><th>代码</th><th>名称</th><th>现价</th><th>操作</th></tr>
        <tr v-for="s in list" :key="s.symbol">
          <td>{{ s.code }}</td>
          <td>{{ s.name }}</td>
          <td>{{ s.trade }}</td>
          <td><button @click="addToWatch(s)">加入自选</button></td>
        </tr>
      </table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'

const boards = ["上证主板","深证主板","创业板","科创板"]
const selected = ref("")
const list = ref([])

function selectBoard(b) {
  selected.value = b
  fetchBoard(b)
}

async function fetchBoard(b) {
  const res = await axios.get('/api/stocks', { params: { board: b, size: 500 } })
  list.value = res.data.list
}

async function addToWatch(s) {
  await axios.post('/api/watchlist/add', { symbol: s.symbol, name: s.name })
  alert('已加入自选')
}

onMounted(() => {
  // default select first
  selectBoard(boards[0])
})
</script>
