<template>
  <div style="padding:20px">
    <h2>策略编辑器（内测）</h2>
    <div style="margin-bottom:8px">
      <input v-model="name" placeholder="策略名称" style="width:400px;padding:6px"/>
      <select v-model="target">
        <option value="watchlist">自选股</option>
        <option :value="'board:'+b" v-for="b in boards" :key="b">{{ b }}</option>
        <option value="all">全部板块</option>
      </select>
      <button @click="save" style="margin-left:8px">保存</button>
      <button @click="run" style="margin-left:8px">运行</button>
    </div>
    <textarea v-model="code" style="width:100%;height:320px;font-family:monospace;"></textarea>
    <div style="margin-top:12px">
      <b>运行结果（命中列表）</b>
      <div v-if="running">执行中...</div>
      <ul v-if="matches.length">
        <li v-for="s in matches" :key="s">{{ s }}</li>
      </ul>
      <div v-if="err" style="color:red">{{ err }}</div>
      <div v-if="duration">耗时: {{ duration }} ms</div>
    </div>
  </div>
</template>

<script setup>
import { ref } from 'vue'
import axios from 'axios'

const name = ref('示例策略')
const code = ref(`// 必须实现：Match(symbol string, klines []map[string]interface{}) bool
func Match(symbol string, klines []map[string]interface{}) bool {
  if len(klines) < 20 { return false }
  sum := 0.0
  for i:=len(klines)-20; i<len(klines); i++ {
    sum += klines[i]["Close"].(float64)
  }
  ma20 := sum / 20.0
  last := klines[len(klines)-1]["Close"].(float64)
  return last > ma20
}`)
const target = ref('watchlist')
const boards = ['上证主板','深证主板','创业板','科创板']

const matches = ref([])
const running = ref(false)
const err = ref('')
const duration = ref(0)

async function save() {
  try {
    const res = await axios.post('/api/strategy', { name: name.value, code: code.value, description: '' })
    alert('保存成功 id=' + res.data.id)
  } catch (e) {
    alert('保存失败: ' + (e.response?.data?.error || e.message))
  }
}

async function run() {
  running.value = true
  err.value = ''
  matches.value = []
  duration.value = 0
  try {
    const res = await axios.post('/api/strategy/run', { code: code.value, target: target.value })
    matches.value = res.data.matches || []
    duration.value = res.data.duration_ms || 0
    if (res.data.error) err.value = res.data.error
  } catch (e) {
    err.value = e.response?.data?.error || e.message
  } finally {
    running.value = false
  }
}
</script>
