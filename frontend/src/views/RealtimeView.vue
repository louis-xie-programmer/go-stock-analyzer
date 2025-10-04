<template>
  <div class="p-4">
    <h2 class="text-xl font-bold mb-2">📡 实时行情订阅</h2>
    <div class="mb-4 flex space-x-2">
      <input v-model="inputCode" placeholder="输入股票代码 (如 sz000001)" class="border px-2 py-1" />
      <button @click="subscribe" class="bg-green-500 text-white px-3 py-1 rounded">订阅</button>
    </div>
    <table class="table-auto border w-full text-sm">
      <thead>
        <tr class="bg-gray-200">
          <th class="px-2 py-1">代码</th>
          <th>名称</th>
          <th>现价</th>
          <th>涨跌</th>
          <th>最高</th>
          <th>最低</th>
          <th>成交量</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="q in Object.values(quotes)" :key="q.code">
          <td class="border px-2 py-1">{{ q.code }}</td>
          <td class="border px-2 py-1">{{ q.name }}</td>
          <td class="border px-2 py-1">{{ q.price }}</td>
          <td class="border px-2 py-1"
              :class="q.change > 0 ? 'text-red-600' : 'text-green-600'">
              {{ q.change.toFixed(2) }}
          </td>
          <td class="border px-2 py-1">{{ q.high }}</td>
          <td class="border px-2 py-1">{{ q.low }}</td>
          <td class="border px-2 py-1">{{ q.volume }}</td>
        </tr>
      </tbody>
    </table>
  </div>
</template>

<script setup>
import { ref, onMounted, onBeforeUnmount } from 'vue'

const inputCode = ref('')
const quotes = ref({})
let ws = null

onMounted(() => {
  ws = new WebSocket('ws://localhost:8080/ws/realtime')
  ws.onmessage = (event) => {
    const q = JSON.parse(event.data)
    quotes.value[q.code] = q
  }
})

onBeforeUnmount(() => {
  ws && ws.close()
})

const subscribe = () => {
  if (inputCode.value && ws && ws.readyState === WebSocket.OPEN) {
    ws.send(JSON.stringify({ action: "subscribe", symbols: [inputCode.value] }))
    inputCode.value = ''
  }
}
</script>
