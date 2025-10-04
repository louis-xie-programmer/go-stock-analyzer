import { createRouter, createWebHistory } from 'vue-router'
import ResultView from '../views/ResultView.vue'
import DSLTester from '../views/DSLTester.vue'
import RealtimeView from '../views/RealtimeView.vue'
import StockPoolView from '../views/StockPoolView.vue'

const routes = [
  { path: '/', component: ResultView },
  { path: '/dsl', component: DSLTester },
  { path: '/realtime', component: RealtimeView },
    { path: '/stocks', component: StockPoolView },
]

export default createRouter({ history: createWebHistory(), routes })
