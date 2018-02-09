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
      path: '/pipelines/detail',
      component: lazyLoading('pipelines/detail')
    }
  ]
})

// Menu should have 2 levels.
function generateRoutesFromMenu (menu = [], routes = []) {
  for (let i = 0, l = menu.length; i < l; i++) {
    let item = menu[i]
    if (item.path && item.subroute) {
      for (let x = 0, y = item.subroute.length; x < y; x++) {
        routes.push(item.subroute[x])
      }
    } else if (item.path) {
      routes.push(item)
    }
  }
  return routes
}
