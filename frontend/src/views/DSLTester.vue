<template>
  <div style="padding:20px">
    <h2>DSL策略测试器</h2>
    <input v-model="expr" placeholder="输入DSL表达式" style="width:400px"/>
    <button @click="test">测试</button>
    <pre>{{ result }}</pre>
  </div>
</template>

<script>
import axios from 'axios'
export default {
  data() { return { expr: '', result: '' } },
  methods: {
    async test() {
      const res = await axios.post('/api/dsl/test', { expr: this.expr })
      this.result = JSON.stringify(res.data,null,2)
    }
  }
}
</script>
