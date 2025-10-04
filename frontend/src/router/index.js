import { createRouter, createWebHistory } from 'vue-router'
import ResultView from '../views/ResultView.vue'
import DSLTester from '../views/DSLTester.vue'
import RealtimeView from '../views/RealtimeView.vue'

const routes = [
  { path: '/', component: ResultView },
  { path: '/dsl', component: DSLTester },
  { path: '/realtime', component: RealtimeView },
]

export default createRouter({ history: createWebHistory(), routes })
