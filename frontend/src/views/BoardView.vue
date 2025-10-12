<template>
  <div style="padding:20px">
    <h2>按板块浏览</h2>
    <div style="margin:10px 0">
      <button v-for="b in boards" :key="b" @click="selectBoard(b)" :style="{marginRight:'8px'}">{{ b }}</button>
    </div>
    <div v-if="selected">
      <h3>{{ selected }}</h3>
      <el-table :data="list" style="width: 100%; margin-top: 20px;">
        <el-table-column prop="code" label="代码" width="100"/>
        <el-table-column prop="name" label="名称"/>
        <el-table-column prop="trade" label="现价" width="100"/>
        <el-table-column label="操作" width="120">
          <template #default="scope">
            <el-button size="small" type="primary" @click="addToWatch(scope.row)">加入自选</el-button>
          </template>
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import axios from 'axios'
import { ElTable } from 'element-plus'

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
  console.log('BoardView mounted')
  // default select first
  selectBoard(boards[0])
})
</script>
