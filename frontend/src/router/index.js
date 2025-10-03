import { createRouter, createWebHistory } from 'vue-router'
import ResultView from '../views/ResultView.vue'
import DSLTester from '../views/DSLTester.vue'

const routes = [
    { path: '/', component: ResultView },
    { path: '/dsl', component: DSLTester }
]

export default createRouter({ history: createWebHistory(), routes })
