import { createRouter, createWebHistory } from 'vue-router'
import ResultView from '../views/ResultView.vue'
import DSLTester from '../views/DSLTester.vue'
import StockPoolView from '../views/StockPoolView.vue'
import StockDetailView from '../views/StockDetailView.vue'
import BoardView from '../views/BoardView.vue'
import WatchlistView from '../views/WatchlistView.vue'
import StrategyEditor from '../views/StrategyEditor.vue'

const routes = [
  { path: '/', component: BoardView },
  { path: '/dsl', component: DSLTester },
  { path: '/stocks', component: StockPoolView },
  { path: '/stocks/:symbol', component: StockDetailView },
  { path: '/watchlist', component: WatchlistView },
  { path: '/strategy', component: StrategyEditor },
  { path: '/result', component: ResultView },
]

export default createRouter({ history: createWebHistory(), routes })
