import Vue from 'vue'
import Router from 'vue-router'
import menuModule from '../store/modules/menu'
import lazyLoading from '../store/modules/menu/lazyLoading'
Vue.use(Router)

export default new Router({
  mode: 'hash', // Demo is living in GitHub.io, so required!
  linkActiveClass: 'is-active',
  scrollBehavior: () => ({ y: 0 }),
  routes: [
    ...generateRoutesFromMenu(menuModule.state.items),
    {
      path: '*',
      redirect: '/overview'
    },
    {
      name: 'Pipeline Detail',
      path: '/pipeline/detail',
      component: lazyLoading('pipeline/detail')
    },
    {
      name: 'Pipeline Logs',
      path: '/pipeline/log',
      component: lazyLoading('pipeline/log')
    }
  ]
})

// Menu should have 1 level.
function generateRoutesFromMenu (menu = [], routes = []) {
  for (let i = 0, l = menu.length; i < l; i++) {
    routes.push(menu[i])
  }
  return routes
}
